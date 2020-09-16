package scaffolder

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/gobuffalo/genny"
	"github.com/tendermint/starport/starport/pkg/gomodulepath"
	"github.com/tendermint/starport/starport/templates/app"
)

var (
	commitMessage = "Initialized with Starport"
	devXAuthor    = &object.Signature{
		Name:  "Developer Experience team at Tendermint",
		Email: "hello@tendermint.com",
		When:  time.Now(),
	}
)

// InitOption configures scaffolding.
type InitOption func(*initOptions)

// initOptions keeps set of options to apply scaffolding.
type initOptions struct {
	addressPrefix string
}

// AddressPrefix configures address prefix for the app.
func AddressPrefix(prefix string) InitOption {
	return func(o *initOptions) {
		o.addressPrefix = prefix
	}
}

// Init initializes a new app with name and given options.
// path is the relative path to the scaffoled app.
func (s *Scaffolder) Init(name string, options ...InitOption) (path string, err error) {
	opts := &initOptions{}
	for _, o := range options {
		o(opts)
	}
	pathInfo, err := gomodulepath.Parse(name)
	if err != nil {
		return "", err
	}
	g, err := app.New(&app.Options{
		ModulePath:       pathInfo.RawPath,
		AppName:          pathInfo.Package,
		BinaryNamePrefix: pathInfo.Root,
		AddressPrefix:    opts.addressPrefix,
	})
	if err != nil {
		return "", err
	}
	run := genny.WetRunner(context.Background())
	run.With(g)
	pwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	run.Root = filepath.Join(pwd, pathInfo.Root)
	if err := run.Run(); err != nil {
		return "", err
	}
	if err := initGit(pathInfo.Root); err != nil {
		return "", err
	}
	return pathInfo.Root, nil
}

func initGit(path string) error {
	repo, err := git.PlainInit(path, false)
	if err != nil {
		return err
	}
	wt, err := repo.Worktree()
	if err != nil {
		return err
	}
	if _, err := wt.Add("."); err != nil {
		return err
	}
	_, err = wt.Commit(commitMessage, &git.CommitOptions{
		All:    true,
		Author: devXAuthor,
	})
	return err
}
