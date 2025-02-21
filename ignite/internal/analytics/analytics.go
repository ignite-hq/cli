package analytics

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"

	"github.com/ignite/cli/v29/ignite/config"
	"github.com/ignite/cli/v29/ignite/pkg/matomo"
	"github.com/ignite/cli/v29/ignite/pkg/randstr"
	"github.com/ignite/cli/v29/ignite/pkg/sentry"
	"github.com/ignite/cli/v29/ignite/version"
)

const (
	telemetryEndpoint  = "https://matomo-cli.ignite.com"
	envDoNotTrack      = "DO_NOT_TRACK"
	envCI              = "CI"
	envGitHubActions   = "GITHUB_ACTIONS"
	igniteAnonIdentity = "anon_identity.json"
)

var matomoClient matomo.Client

// anonIdentity represents an analytics identity file.
type anonIdentity struct {
	// Name represents the username.
	Name string `json:"name" yaml:"name"`
	// DoNotTrack represents the user track choice.
	DoNotTrack bool `json:"doNotTrack" yaml:"doNotTrack"`
}

func init() {
	matomoClient = matomo.New(
		telemetryEndpoint,
		matomo.WithIDSite(4),
		matomo.WithSource("https://cli.ignite.com"),
	)
}

// SendMetric send command metrics to analytics.
func SendMetric(wg *sync.WaitGroup, cmd *cobra.Command) {
	if cmd.Name() == "version" {
		return
	}

	dntInfo, err := checkDNT()
	if err != nil || dntInfo.DoNotTrack {
		return
	}

	versionInfo, err := version.GetInfo(context.Background())
	if err != nil {
		return
	}

	var (
		path         = cmd.CommandPath()
		scaffoldType = ""
	)
	if strings.Contains(path, "ignite scaffold") {
		splitCMD := strings.Split(path, " ")
		if len(splitCMD) > 2 {
			scaffoldType = splitCMD[2]
		}
	}

	met := matomo.Metric{
		Name:            cmd.Name(),
		Cmd:             path,
		ScaffoldType:    scaffoldType,
		OS:              versionInfo.OS,
		Arch:            versionInfo.Arch,
		Version:         versionInfo.CLIVersion,
		CLIVersion:      versionInfo.CLIVersion,
		GoVersion:       versionInfo.GoVersion,
		SDKVersion:      versionInfo.SDKVersion,
		BuildDate:       versionInfo.BuildDate,
		SourceHash:      versionInfo.SourceHash,
		ConfigVersion:   versionInfo.ConfigVersion,
		Uname:           versionInfo.Uname,
		CWD:             versionInfo.CWD,
		BuildFromSource: versionInfo.BuildFromSource,
		IsCI:            getIsCI(),
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = matomoClient.SendMetric(dntInfo.Name, met)
	}()
}

// EnableSentry enable errors reporting to Sentry.
func EnableSentry(ctx context.Context, wg *sync.WaitGroup) {
	dntInfo, err := checkDNT()
	if err != nil || dntInfo.DoNotTrack {
		return
	}

	closeSentry, err := sentry.InitSentry(ctx)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err == nil {
			defer closeSentry()
		}
	}()
}

// checkDNT check if the user allow to track data or if the DO_NOT_TRACK
// env var is set https://consoledonottrack.com/
func checkDNT() (anonIdentity, error) {
	if dnt := os.Getenv(envDoNotTrack); dnt != "" {
		if dnt, err := strconv.ParseBool(dnt); err != nil || dnt {
			return anonIdentity{DoNotTrack: true}, nil
		}
	}

	globalPath, err := config.DirPath()
	if err != nil {
		return anonIdentity{}, err
	}
	if err := os.Mkdir(globalPath, 0o700); err != nil && !os.IsExist(err) {
		return anonIdentity{}, err
	}

	identityPath := filepath.Join(globalPath, igniteAnonIdentity)
	data, err := os.ReadFile(identityPath)
	if err != nil && !os.IsNotExist(err) {
		return anonIdentity{}, err
	}

	var i anonIdentity
	if err := json.Unmarshal(data, &i); err == nil {
		return i, nil
	}

	i.Name = randstr.Runes(16)
	i.DoNotTrack = false

	prompt := promptui.Select{
		Label: "Ignite uses anonymized metrics to enhance the application, " +
			"focusing on features such as command usage. We do not collect " +
			"identifiable personal information. Your privacy is important to us. " +
			"For more details, please visit our Privacy Policy at https://ignite.com/privacy " +
			"and our Terms of Use at https://ignite.com/terms-of-use. " +
			"Do you consent to the collection of these usage metrics for analytics purposes?",
		Items: []string{"Yes", "No"},
	}
	resultID, _, err := prompt.Run()
	if err != nil {
		return anonIdentity{}, err
	}

	if resultID != 0 {
		i.DoNotTrack = true
	}

	data, err = json.Marshal(&i)
	if err != nil {
		return i, err
	}

	return i, os.WriteFile(identityPath, data, 0o600)
}

func getIsCI() bool {
	ci, err := strconv.ParseBool(os.Getenv(envCI))
	if err != nil {
		return false
	}

	if ci {
		return true
	}

	ci, err = strconv.ParseBool(os.Getenv(envGitHubActions))
	if err != nil {
		return false
	}

	return ci
}
