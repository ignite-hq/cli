package integration_test

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/starport/starport/pkg/cmdrunner/step"
	"github.com/tendermint/starport/starport/pkg/randstr"
	starportconf "github.com/tendermint/starport/starport/services/serve/conf"
	"golang.org/x/sync/errgroup"
)

func TestRelayerWithMultipleChains(t *testing.T) {
	t.Parallel()

	relayerWithMultipleChains(t, 3)
}

func relayerWithMultipleChains(t *testing.T, chainCount int) {
	type Chain struct {
		name, path string
		servers    starportconf.Servers
		logsWriter *io.PipeWriter
		logsReader *io.PipeReader
		connstr    *bytes.Buffer
	}

	var (
		env              = newEnv(t)
		chains           []*Chain
		ctx, serveCancel = context.WithCancel(env.Ctx())
	)
	defer serveCancel()

	// init & scaffold chains.
	newChain := func() *Chain {
		var (
			name    = randstr.Runes(15)
			path    = env.Scaffold(name, Stargate)
			servers = env.RandomizeServerPorts(path)
		)
		r, w := io.Pipe()
		return &Chain{
			name:       name,
			path:       path,
			servers:    servers,
			logsWriter: w,
			logsReader: r,
			connstr:    &bytes.Buffer{},
		}
	}
	for i := 0; i < chainCount; i++ {
		chains = append(chains, newChain())
	}

	// serve chains.
	sg, ctx := errgroup.WithContext(ctx)
	for _, chain := range chains {
		chain := chain
		sg.Go(func() error {
			ok := env.Serve(fmt.Sprintf("should serve app %q", chain.name),
				chain.path,
				ExecCtx(ctx),
				ExecStdout(chain.logsWriter),
			)
			if !ok {
				return errors.New("cannot serve")
			}
			return nil
		})
	}

	// wait for chains to be properly served. we could have skip this but having this step
	// is useful to test if chains will restart and detect chains added by `starport chain add`.
	for _, chain := range chains {
		require.NoError(t, env.IsAppServed(ctx, chain.servers), "some apps cannot get online in time")
	}

	// retrieve each chain's relayer connection string.
	for _, chain := range chains {
		env.Must(env.Exec(fmt.Sprintf("should get base64 relayer connection string from chain %q", chain.name),
			step.New(
				step.Exec(
					"starport",
					"chain",
					"me",
				),
				step.Workdir(chain.path),
				step.Stdout(chain.connstr),
			),
			ExecCtx(ctx),
		))
	}

	// cross connect all chains with each other.
	for _, srcchain := range chains {
		for _, dstchain := range chains {
			if srcchain == dstchain {
				continue
			}
			env.Must(env.Exec(fmt.Sprintf("adding chain %q to chain %q", dstchain.name, srcchain.name),
				step.New(
					step.Exec(
						"starport",
						"chain",
						"add",
						dstchain.connstr.String(),
					),
					step.Workdir(srcchain.path),
				),
				ExecCtx(ctx),
			))
		}
	}

	// check each chain's logs to see if they report that they're connected with
	// other chains successfully. we should expect len(chains) x (len(chains) - 1)
	// connections at total. but things aren't stable yet so, for we expect at least len(chains)
	// connections.
	var (
		capturedLines []string
		mc            sync.Mutex
	)
	cg := &errgroup.Group{}
	for _, chain := range chains {
		chain := chain
		cg.Go(func() error {
			r := bufio.NewReader(chain.logsReader)
			for {
				line, _, err := r.ReadLine()
				if err != nil {
					return err
				}
				if strings.Contains(string(line), "linked") {
					mc.Lock()
					isDone := len(capturedLines) == len(chains)
					if isDone {
						for _, chain := range chains {
							chain.logsReader.Close()
						}
						mc.Unlock()
						return nil
					}
					capturedLines = append(capturedLines, string(line))
					mc.Unlock()
				}
			}
		})
	}
	err := cg.Wait()
	if err == io.ErrClosedPipe {
		err = nil
	}
	require.NoError(t, err, "not enough linked chains")

	serveCancel()
	// wait untill all chains stop serving.
	// a chain will stop serving either by a failure or cancelation.
	// failure is not expected. so, test will exit with error in case of a failure in any of the served chains.
	if err := sg.Wait(); err != nil {
		t.FailNow()
	}
}
