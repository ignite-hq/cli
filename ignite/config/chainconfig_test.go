package config_test

import (
	"bytes"
	"fmt"
	"github.com/ignite/cli/ignite/config/chain/base"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ignite/cli/ignite/config"
)

func TestCheckVersion(t *testing.T) {
	// Arrange
	cfg := bytes.NewBufferString(
		fmt.Sprintf("version: %d", config.LatestVersion),
	)

	// Act
	err := config.CheckVersion(cfg)

	// Assert
	require.NoError(t, err)
}

func TestCheckVersionWithOutdatedVersion(t *testing.T) {
	// Arrange
	cfg := bytes.NewBufferString("version: 0")
	wantError := config.VersionError{}

	// Act
	err := config.CheckVersion(cfg)

	// Assert
	require.ErrorAs(t, err, &wantError)
	require.Equal(t, wantError.Version, base.Version(0))
}
