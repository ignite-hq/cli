package cosmosfaucet

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/ignite/cli/v29/ignite/pkg/errors"
)

// faucetTimeout used to set a timeout while transferring coins from a faucet.
const faucetTimeout = time.Second * 20

// TryRetrieve tries to retrieve tokens from a faucet. faucet address is used when it's provided.
// otherwise, it'll try to guess the faucet address from the rpc address of the chain.
// a non-nil error is returned if faucet's address cannot be determined or when coin retrieval is unsuccessful.
func TryRetrieve(
	ctx context.Context,
	chainID,
	rpcAddress,
	faucetAddress,
	accountAddress string,
) (string, error) {
	var faucetURL *url.URL
	var err error

	if faucetAddress != "" {
		// use if there is a user given faucet address.
		faucetURL, err = url.Parse(faucetAddress)
	} else {
		// find faucet url. can be the user given, otherwise it is the guessed one.
		faucetURL, err = discoverFaucetURL(ctx, chainID, rpcAddress)
	}
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(ctx, faucetTimeout)
	defer cancel()

	fc := NewClient(faucetURL.String())

	resp, err := fc.Transfer(ctx, TransferRequest{
		AccountAddress: accountAddress,
	})
	if err != nil {
		return "", errors.Wrap(err, "faucet is not operational")
	}
	if resp.Error != "" {
		return "", errors.Errorf("faucet is not operational: %s", resp.Error)
	}

	return resp.Hash, nil
}

func discoverFaucetURL(ctx context.Context, chainID, rpcAddress string) (*url.URL, error) {
	// guess faucet address otherwise.
	guessedURLs, err := guessFaucetURLs(rpcAddress)
	if err != nil {
		return nil, err
	}

	for _, u := range guessedURLs {
		// check if the potential faucet server accepts connections.
		address := u.Host
		if u.Scheme == "https" {
			address += ":443"
		}
		if _, err := net.DialTimeout("tcp", address, time.Second); err != nil {
			continue
		}

		// ensure that this is a real faucet server.
		info, err := NewClient(u.String()).FaucetInfo(ctx)
		if err != nil || info.ChainID != chainID || !info.IsAFaucet {
			continue
		}

		return u, nil
	}

	return nil, errors.New("no faucet available, please send coins to the address")
}

// guess tries to guess all possible faucet addresses.
func guessFaucetURLs(rpcAddress string) ([]*url.URL, error) {
	u, err := url.Parse(rpcAddress)
	if err != nil {
		return nil, err
	}

	var guessedURLs []*url.URL

	possibilities := []struct {
		port          string
		subname       string
		nameSeparator string
	}{
		{"4500", "", "."},
		{"", "faucet", "."},
		{"", "4500", "-"},
	}

	// creating guesses addresses by basing RPC address.
	for _, poss := range possibilities {
		guess, _ := url.Parse(u.String())                  // copy the original url.
		for _, scheme := range []string{"http", "https"} { // do for both schemes.
			guess, _ := url.Parse(guess.String()) // copy guess.
			guess.Scheme = scheme

			// try with port numbers.
			if poss.port != "" {
				guess.Host = fmt.Sprintf("%s:%s", u.Hostname(), "4500")
				guessedURLs = append(guessedURLs, guess)
				continue
			}

			// try with subnames.
			if poss.subname != "" {
				bases := []string{
					// try with appending subname to the default name.
					// e.g.: faucet.my.domain.
					u.Hostname(),
				}

				// try with replacing the subname for 1 level.
				// e.g.: faucet.domain.
				sp := strings.SplitN(u.Hostname(), poss.nameSeparator, 2)
				if len(sp) == 2 {
					bases = append(bases, sp[1])
				}
				for _, basename := range bases {
					guess, _ := url.Parse(guess.String()) // copy guess.
					guess.Host = fmt.Sprintf("%s%s%s", poss.subname, poss.nameSeparator, basename)
					guessedURLs = append(guessedURLs, guess)
				}
			}
		}
	}

	return guessedURLs, nil
}
