package starportconf

import (
	"io"

	"github.com/goccy/go-yaml"
	"github.com/pkg/errors"
)

var (
	// FileNames holds a list of appropriate names for the config file.
	FileNames = []string{"config.yml", "config.yaml"}
)

// Config is the user given configuration to do additional setup
// during serve.
type Config struct {
	Accounts []Account `yaml:"accounts"`
}

// Account holds the options related to setting up Cosmos wallets.
type Account struct {
	Name  string   `yaml:"name"`
	Coins []string `yaml:"coins"`
}

// Parse parses config.yml into UserConfig.
func Parse(r io.Reader) (Config, error) {
	var conf Config
	if err := yaml.NewDecoder(r).Decode(&conf); err != nil {
		return conf, err
	}
	return conf, validate(conf)
}

// validate validates user config.
func validate(conf Config) error {
	if len(conf.Accounts) == 0 {
		errors.New("at least 1 account is needed")
	}
	return nil
}
