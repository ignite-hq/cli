package integration_test

import (
	"testing"

	"github.com/tendermint/starport/starport/pkg/cmdrunner/step"
)

func TestGenerateAnAppWithTypeAndVerify(t *testing.T) {
	t.Parallel()

	var (
		env  = newEnv(t)
		path = env.Scaffold("blog", Launchpad)
	)

	env.Exec("create a type",
		step.New(
			step.Exec("starport", "type", "user", "email"),
			step.Workdir(path),
		),
	)

	env.Exec("should prevent creating an existing type",
		step.New(
			step.Exec("starport", "type", "user", "email"),
			step.Workdir(path),
		),
		ExecShouldError(),
	)

	env.EnsureAppIsSteady(path)
}

func TestGenerateAnAppWithStargateWithTypeAndVerify(t *testing.T) {
	t.Parallel()

	var (
		env  = newEnv(t)
		path = env.Scaffold("blog", Stargate)
	)

	env.Exec("create a type",
		step.New(
			step.Exec("starport", "type", "user", "email"),
			step.Workdir(path),
		),
	)

	env.Exec("should prevent creating an existing type",
		step.New(
			step.Exec("starport", "type", "user", "email"),
			step.Workdir(path),
		),
		ExecShouldError(),
	)

	env.EnsureAppIsSteady(path)
}

func TestCreateTypeInCustomModule(t *testing.T) {
	t.Parallel()

	var (
		env  = newEnv(t)
		path = env.Scaffold("blog", Launchpad)
	)

	env.Exec("create a module",
		step.New(
			step.Exec("starport", "module", "create", "example"),
			step.Workdir(path),
		),
	)

	env.Exec("create a type",
		step.New(
			step.Exec("starport", "type", "user", "email", "--module", "example"),
			step.Workdir(path),
		),
	)

	env.Exec("create a type in the app's module",
		step.New(
			step.Exec("starport", "type", "user", "email"),
			step.Workdir(path),
		),
	)

	env.Exec("should prevent creating a type in a non existant module",
		step.New(
			step.Exec("starport", "type", "user", "email", "--module", "idontexist"),
			step.Workdir(path),
		),
		ExecShouldError(),
	)

	env.Exec("should prevent creating an existing type",
		step.New(
			step.Exec("starport", "type", "user", "email", "--module", "example"),
			step.Workdir(path),
		),
		ExecShouldError(),
	)

	env.EnsureAppIsSteady(path)
}

func TestCreateTypeInCustomModuleWithStargate(t *testing.T) {
	t.Parallel()

	var (
		env  = newEnv(t)
		path = env.Scaffold("blog", Launchpad)
	)

	env.Exec("create a module",
		step.New(
			step.Exec("starport", "module", "create", "example"),
			step.Workdir(path),
		),
	)

	env.Exec("create a type",
		step.New(
			step.Exec("starport", "type", "user", "email", "--module", "example"),
			step.Workdir(path),
		),
	)

	env.Exec("create a type in the app's module",
		step.New(
			step.Exec("starport", "type", "user", "email"),
			step.Workdir(path),
		),
	)

	env.Exec("should prevent creating a type in a non existant module",
		step.New(
			step.Exec("starport", "type", "user", "email", "--module", "idontexist"),
			step.Workdir(path),
		),
		ExecShouldError(),
	)

	env.Exec("should prevent creating an existing type",
		step.New(
			step.Exec("starport", "type", "user", "email", "--module", "example"),
			step.Workdir(path),
		),
		ExecShouldError(),
	)

	env.EnsureAppIsSteady(path)
}
