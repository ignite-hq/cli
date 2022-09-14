package chainconfig

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"

	"github.com/ignite/cli/ignite/chainconfig/config"
)

// Parse reads a config file.
// When the version of the file beign read is not the latest
// it is automatically migrated to the latest version.
func Parse(configFile io.Reader) (*Config, error) {
	var buf bytes.Buffer

	// Read the config file version first to know how to decode it
	version, err := ReadConfigVersion(io.TeeReader(configFile, &buf))
	if err != nil {
		return DefaultConfig(), err
	}

	// Decode the current config file version and assign default
	// values for the fields that are empty
	c, err := decodeConfig(&buf, version)
	if err != nil {
		return DefaultConfig(), err
	}

	// Make sure that the empty fields contain default values
	// after reading the config from the YAML file
	if err = c.SetDefaults(); err != nil {
		return DefaultConfig(), err
	}

	// Finally make sure the config is the latest one before validating it
	cfg, err := ConvertLatest(c)
	if err != nil {
		return DefaultConfig(), err
	}

	return cfg, validateConfig(cfg)
}

// ParseFile parses a config from a file path.
func ParseFile(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return DefaultConfig(), err
	}

	defer file.Close()

	return Parse(file)
}

// ReadConfigVersion reads the config version.
func ReadConfigVersion(configFile io.Reader) (config.Version, error) {
	c := struct {
		Version config.Version `yaml:"version"`
	}{}

	err := yaml.NewDecoder(configFile).Decode(&c)

	return c.Version, err
}

// LocateDefault locates the default path for the config file.
// Returns ErrConfigNotFound when no config file found.
func LocateDefault(root string) (path string, err error) {
	for _, name := range ConfigFileNames {
		path = filepath.Join(root, name)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		} else if !os.IsNotExist(err) {
			return "", err
		}
	}

	return "", ErrConfigNotFound
}

// FaucetHost returns the faucet host to use.
func FaucetHost(cfg *Config) string {
	// We keep supporting Port option for backward compatibility
	// TODO: drop this option in the future
	host := cfg.Faucet.Host
	if cfg.Faucet.Port != 0 {
		host = fmt.Sprintf(":%d", cfg.Faucet.Port)
	}

	return host
}

// CreateConfigDir creates config directory if it is not created yet.
func CreateConfigDir() error {
	path, err := ConfigDirPath()
	if err != nil {
		return err
	}

	return os.MkdirAll(path, 0755)
}

func decodeConfig(r io.Reader, version config.Version) (config.Converter, error) {
	c, ok := Versions[version]
	if !ok {
		return nil, &UnsupportedVersionError{version}
	}

	cfg, err := c.Clone()
	if err != nil {
		return nil, err
	}

	if err = cfg.Decode(r); err != nil {
		return nil, err
	}

	return cfg, nil
}

func validateConfig(c *Config) error {
	if len(c.Accounts) == 0 {
		return &ValidationError{"at least one account is required"}
	}

	if len(c.Validators) == 0 {
		return &ValidationError{"at least one validator is required"}
	}

	for _, validator := range c.Validators {
		if validator.Name == "" {
			return &ValidationError{"validator 'name' is required"}
		}

		if validator.Bonded == "" {
			return &ValidationError{"validator 'bonded' is required"}
		}
	}

	// TODO: We should validate all of the required config fields

	return nil
}
