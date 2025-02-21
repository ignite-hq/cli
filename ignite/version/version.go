package version

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path"
	"runtime"
	"runtime/debug"
	"strings"
	"text/tabwriter"

	"github.com/blang/semver/v4"
	"github.com/google/go-github/v48/github"

	chainconfig "github.com/ignite/cli/v29/ignite/config/chain"
	"github.com/ignite/cli/v29/ignite/pkg/cmdrunner/exec"
	"github.com/ignite/cli/v29/ignite/pkg/cmdrunner/step"
	"github.com/ignite/cli/v29/ignite/pkg/cosmosbuf"
	"github.com/ignite/cli/v29/ignite/pkg/cosmosver"
	"github.com/ignite/cli/v29/ignite/pkg/errors"
	"github.com/ignite/cli/v29/ignite/pkg/xexec"
)

const (
	errOldCosmosSDKVersionStr = `Your chain has been scaffolded with an older version of Cosmos SDK: %s

Please, follow the migration guide to upgrade your chain to the latest version at https://docs.ignite.com/migration`

	versionDev     = "development"
	versionNightly = "nightly"
)

// Version is the semantic version of Ignite CLI.
var Version = versionDev

type Info struct {
	CLIVersion      string
	GoVersion       string
	SDKVersion      string
	BufVersion      string
	BuildDate       string
	SourceHash      string
	ConfigVersion   string
	OS              string
	Arch            string
	Uname           string
	CWD             string
	BuildFromSource bool
}

// CheckNext checks whether there is a new version of Ignite CLI.
func CheckNext(ctx context.Context) (isAvailable bool, version string, err error) {
	if Version == versionDev || Version == versionNightly {
		return false, "", nil
	}

	tagName, err := getLatestReleaseTag(ctx)
	if err != nil {
		return false, "", err
	}

	currentVersion, err := semver.ParseTolerant(Version)
	if err != nil {
		return false, "", err
	}

	latestVersion, err := semver.ParseTolerant(tagName)
	if err != nil {
		return false, "", err
	}

	isAvailable = latestVersion.GT(currentVersion)

	return isAvailable, tagName, nil
}

func getLatestReleaseTag(ctx context.Context) (string, error) {
	latest, _, err := github.
		NewClient(nil).
		Repositories.
		GetLatestRelease(ctx, "ignite", "cli")
	if err != nil {
		return "", err
	}

	if latest.TagName == nil {
		return "", nil
	}

	return *latest.TagName, nil
}

// fromSource check if the binary was build from source using the CLI version.
func fromSource() bool {
	return Version == versionDev
}

// resolveDevVersion creates a string for version printing if the version being used is "development".
// the version will be of the form "LATEST-dev" where LATEST is the latest tagged release.
func resolveDevVersion(ctx context.Context) string {
	// do nothing if built with specific tag
	if Version != versionDev && Version != versionNightly {
		return Version
	}

	tag, err := getLatestReleaseTag(ctx)
	if err != nil {
		return Version
	}

	// if the module version is higher than the latest tag, use the module version
	if info, ok := debug.ReadBuildInfo(); ok {
		if version := path.Base(info.Main.Path); version > tag {
			tag = fmt.Sprintf("%s.0.0", version)
		}
	}

	if Version == versionDev {
		return tag + "-dev"
	}
	if Version == versionNightly {
		return tag + "-nightly"
	}

	return Version
}

// Long generates a detailed version info.
func Long(ctx context.Context) (string, error) {
	var (
		w = &tabwriter.Writer{}
		b = &bytes.Buffer{}
	)

	info, err := GetInfo(ctx)
	if err != nil {
		return "", err
	}

	write := func(k, v string) {
		fmt.Fprintf(w, "%s:\t%s\n", k, v)
	}
	w.Init(b, 0, 8, 0, '\t', 0)

	write("Ignite CLI version", info.CLIVersion)
	write("Ignite CLI build date", info.BuildDate)
	write("Ignite CLI source hash", info.SourceHash)
	write("Ignite CLI config version", info.ConfigVersion)
	write("Cosmos SDK version", info.SDKVersion)
	write("Buf.Build version", info.BufVersion)

	write("Your OS", info.OS)
	write("Your arch", info.Arch)
	write("Your go version", info.GoVersion)
	write("Your uname -a", info.Uname)

	if info.CWD != "" {
		write("Your cwd", info.CWD)
	}

	if err := w.Flush(); err != nil {
		return "", err
	}

	return b.String(), nil
}

// GetInfo gets the CLI info.
func GetInfo(ctx context.Context) (Info, error) {
	var (
		info     Info
		modified bool

		date       = "undefined"
		head       = "undefined"
		sdkVersion = "undefined"
	)
	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		for _, dep := range buildInfo.Deps {
			if cosmosver.CosmosSDKModulePathPattern.MatchString(dep.Path) {
				sdkVersion = dep.Version
				break
			}
		}

		for _, kv := range buildInfo.Settings {
			switch kv.Key {
			case "vcs.revision":
				head = kv.Value
			case "vcs.time":
				date = kv.Value
			case "vcs.modified":
				modified = kv.Value == "true"
			}
		}
		if modified {
			// add * suffix to head to indicate the sources have been modified.
			head += "*"
		}
	}

	goVersionBuf := &bytes.Buffer{}
	if err := exec.Exec(ctx, []string{"go", "version"}, exec.StepOption(step.Stdout(goVersionBuf))); err != nil {
		return info, err
	}

	var (
		unameCmd = "uname"
		uname    = ""
	)
	if xexec.IsCommandAvailable(unameCmd) {
		unameBuf := &bytes.Buffer{}
		unameBuf.Reset()
		if err := exec.Exec(ctx, []string{unameCmd, "-a"}, exec.StepOption(step.Stdout(unameBuf))); err != nil {
			return info, err
		}
		uname = strings.TrimSpace(unameBuf.String())
	}

	bufVersion, err := cosmosbuf.Version(ctx)
	if err != nil {
		return info, err
	}

	info.Uname = uname
	info.CLIVersion = resolveDevVersion(ctx)
	info.BuildDate = date
	info.BufVersion = bufVersion
	info.SourceHash = head
	info.ConfigVersion = fmt.Sprintf("v%d", chainconfig.LatestVersion)
	info.SDKVersion = sdkVersion
	info.OS = runtime.GOOS
	info.Arch = runtime.GOARCH
	info.GoVersion = strings.TrimSpace(goVersionBuf.String())
	info.BuildFromSource = fromSource()

	if cwd, err := os.Getwd(); err == nil {
		info.CWD = cwd
	}

	return info, nil
}

// AssertSupportedCosmosSDKVersion asserts that a Cosmos SDK version is supported by Ignite CLI.
func AssertSupportedCosmosSDKVersion(v cosmosver.Version) error {
	if v.LT(cosmosver.StargateFortySevenTwoVersion) {
		return errors.Errorf(errOldCosmosSDKVersionStr, v)
	}
	return nil
}
