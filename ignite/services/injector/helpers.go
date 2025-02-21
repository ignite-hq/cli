package injector

import (
	"path/filepath"

	"github.com/ignite/cli/v29/ignite/templates/module"
)

const (
	PathAppConfigGo = module.PathAppConfigGo
	PathAppGo       = module.PathAppGo
	PathCommands    = "cmd/commands.go"
)

func (i *injector) appPath() string {
	return filepath.Join(i.chain.AppPath(), PathAppGo)
}

func (i *injector) appConfigPath() string {
	return filepath.Join(i.chain.AppPath(), PathAppConfigGo)
}

func (i *injector) commandsPath() (string, error) {
	appPath := i.chain.AppPath()
	binaryName, err := i.chain.Binary()
	if err != nil {
		return "", err
	}

	return filepath.Join(appPath, "cmd", binaryName, PathCommands), nil
}
