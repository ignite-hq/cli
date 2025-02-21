# Changelog

## Unreleased

### Features

- [#3707](https://github.com/ignite/cli/pull/3707) and [#4094](https://github.com/ignite/cli/pull/4094) Add collections support.
- [#3977](https://github.com/ignite/cli/pull/3977) Add `chain lint` command to lint the chain's codebase using `golangci-lint`
- [#3770](https://github.com/ignite/cli/pull/3770) Add `scaffold configs` and `scaffold params` commands
- [#4001](https://github.com/ignite/cli/pull/4001) Improve `xgenny` dry run
- [#3967](https://github.com/ignite/cli/issues/3967) Add HD wallet parameters `address index` and `account number` to the chain account config
- [#4004](https://github.com/ignite/cli/pull/4004) Remove all import placeholders using the `xast` pkg
- [#4071](https://github.com/ignite/cli/pull/4071) Support custom proto path
- [#3718](https://github.com/ignite/cli/pull/3718) Add `gen-mig-diffs` tool app to compare scaffold output of two versions of ignite
- [#4100](https://github.com/ignite/cli/pull/4100) Set the `proto-dir` flag only for the `scaffold chain` command and use the proto path from the config
- [#4111](https://github.com/ignite/cli/pull/4111) Remove vuex generation
- [#4113](https://github.com/ignite/cli/pull/4113) Generate chain config documentation automatically
- [#4131](https://github.com/ignite/cli/pull/4131) Support `bytes` as data type in the `scaffold` commands
- [#4300](https://github.com/ignite/cli/pull/4300) Only panics the module in the most top function level
- [#4327](https://github.com/ignite/cli/pull/4327) Use the TxConfig from simState instead create a new one
- [#4326](https://github.com/ignite/cli/pull/4326) Add `buf.build` version to `ignite version` command
- [#4436](https://github.com/ignite/cli/pull/4436) Return tx hash to the faucet API
- [#4437](https://github.com/ignite/cli/pull/4437) Remove module placeholders
- [#4289](https://github.com/ignite/cli/pull/4289), [#4423](https://github.com/ignite/cli/pull/4423), [#4432](https://github.com/ignite/cli/pull/4432), [#4507](https://github.com/ignite/cli/pull/4507) Cosmos SDK v0.52 support and downgrade back to 0.50, while keeping latest improvements.
- [#4480](https://github.com/ignite/cli/pull/4480) Add field max length
- [#4477](https://github.com/ignite/cli/pull/4477) IBC v10 support
- [#4166](https://github.com/ignite/cli/issues/4166) Migrate buf config files to v2
- [#4494](https://github.com/ignite/cli/pull/4494) Automatic migrate the buf configs to v2

### Changes

- [#4094](https://github.com/ignite/cli/pull/4094) Scaffolding a multi-index map using `ignite s map foo bar baz --index foobar,foobaz` is no longer supported. Use one index instead of use `collections.IndexedMap`.
- [#4058](https://github.com/ignite/cli/pull/4058) Simplify scaffolded modules by including `ValidateBasic()` logic in message handler.
- [#4058](https://github.com/ignite/cli/pull/4058) Use `address.Codec` instead of `AccAddressFromBech32`.
- [#3993](https://github.com/ignite/cli/pull/3993) Oracle scaffolding was deprecated and has been removed
- [#3962](https://github.com/ignite/cli/pull/3962) Rename all RPC endpoints and autocli commands generated for `map`/`list`/`single` types
- [#3976](https://github.com/ignite/cli/pull/3976) Remove error checks for Cobra command value get calls
- [#4002](https://github.com/ignite/cli/pull/4002) Bump buf build
- [#4008](https://github.com/ignite/cli/pull/4008) Rename `pkg/yaml` to `pkg/xyaml`
- [#4075](https://github.com/ignite/cli/pull/4075) Use `gopkg.in/yaml.v3` instead `gopkg.in/yaml.v2`
- [#4118](https://github.com/ignite/cli/pull/4118) Version scaffolded protos as `v1` to follow SDK structure.
- [#4167](https://github.com/ignite/cli/pull/4167) Scaffold `int64` instead of `int32` when a field type is `int`
- [#4159](https://github.com/ignite/cli/pull/4159) Enable gci linter
- [#4160](https://github.com/ignite/cli/pull/4160) Enable copyloopvar linter
- [#4162](https://github.com/ignite/cli/pull/4162) Enable errcheck linter
- [#4189](https://github.com/ignite/cli/pull/4189) Deprecate `ignite node` for `ignite connect` app
- [#4290](https://github.com/ignite/cli/pull/4290) Remove ignite ics logic from ignite cli (this functionality will be in the `consumer` app)
- [#4295](https://github.com/ignite/cli/pull/4295) Stop scaffolding `pulsar` files
- [#4317](https://github.com/ignite/cli/pull/4317) Remove xchisel dependency
- [#4361](https://github.com/ignite/cli/pull/4361) Remove unused `KeyPrefix` method
- [#4384](https://github.com/ignite/cli/pull/4384) Compare genesis params into chain genesis tests
- [#4463](https://github.com/ignite/cli/pull/4463) Run `chain simulation` with any simulation test case

### Fixes

- [#4000](https://github.com/ignite/cli/pull/4000) Run all dry runners before the wet run in the `xgenny` pkg
- [#4091](https://github.com/ignite/cli/pull/4091) Fix race conditions in the plugin logic
- [#4128](https://github.com/ignite/cli/pull/4128) Check for duplicate proto fields in config
- [#4402](https://github.com/ignite/cli/pull/4402) Fix gentx parser into the cosmosutil package

## [`v28.8.0`](https://github.com/ignite/cli/releases/tag/v28.8.0)

### Features

- [#4513](https://github.com/ignite/cli/pull/4513) Allow to pass tx fees to faucet server

### Changes

- [#4439](https://github.com/ignite/cli/pull/4439) Simplify Ignite CLI dependencies by removing `moby` and `gorilla` dependencies.
- [#4471](https://github.com/ignite/cli/pull/4471) Bump CometBFT to v0.38.15.
- [#4471](https://github.com/ignite/cli/pull/4471) Bump Ignite & chain minimum Go version to 1.23.
- [#4529](https://github.com/ignite/cli/pull/4531) Bump Cosmos SDK to v0.50.12.

### Bug Fixes

- [#4474](https://github.com/ignite/cli/pull/4474) Fix issue in `build --release` command
- [#4479](https://github.com/ignite/cli/pull/4479) Scaffold an `uint64 type crashs Ignite
- [#4483](https://github.com/ignite/cli/pull/4483) Fix default flag parser for apps
- [#4532](https://github.com/ignite/cli/pull/4532) Fix non working _shortcuts_ in validator home config

## [`v28.7.0`](https://github.com/ignite/cli/releases/tag/v28.7.0)

### Features

- [#4457](https://github.com/ignite/cli/pull/4457) Add `skip-build` flag to `chain serve` command to avoid (re)building the chain
- [#4413](https://github.com/ignite/cli/pull/4413) Add `ignite s chain-registry` command

## [`v28.6.1`](https://github.com/ignite/cli/releases/tag/v28.6.1)

### Changes

- [#4449](https://github.com/ignite/cli/pull/4449) Bump scaffolded chain to Cosmos SDK `v0.50.11`. Previous version have a high security vulnerability.

## [`v28.6.0`](https://github.com/ignite/cli/releases/tag/v28.6.0)

### Features

- [#4377](https://github.com/ignite/cli/pull/4377) Add multi node (validator) testnet
- [#4362](https://github.com/ignite/cli/pull/4362) Scaffold `Makefile`

### Changes

- [#4376](https://github.com/ignite/cli/pull/4376) Set different chain-id for in place testnet

### Bug Fixes

- [#4421](https://github.com/ignite/cli/pull/4422) Fix typo in simulation template

## [`v28.5.3`](https://github.com/ignite/cli/releases/tag/v28.5.3)

### Changes

- [#4372](https://github.com/ignite/cli/pull/4372) Bump Cosmos SDK to `v0.50.10`
- [#4357](https://github.com/ignite/cli/pull/4357) Bump chain dependencies (store, ics, log, etc)
- [#4328](https://github.com/ignite/cli/pull/4328) Send ignite bug report to sentry. Opt out the same way as for usage analytics

## [`v28.5.2`](https://github.com/ignite/cli/releases/tag/v28.5.2)

### Features

- [#4297](https://github.com/ignite/cli/pull/4297) Add in-place testnet creation command for apps.

### Changes

- [#4292](https://github.com/ignite/cli/pull/4292) Bump Cosmos SDK to `v0.50.9`
- [#4341](https://github.com/ignite/cli/pull/4341) Bump `ibc-go` to `8.5.0`
- [#4345](https://github.com/ignite/cli/pull/4345) Added survey link

### Fixes

- [#4319](https://github.com/ignite/cli/pull/4319) Remove fee abstraction module from open api code generation
- [#4309](https://github.com/ignite/cli/pull/4309) Fix chain id for chain simulations
- [#4322](https://github.com/ignite/cli/pull/4322) Create a message for authenticate buf for generate ts-client
- [#4323](https://github.com/ignite/cli/pull/4323) Add missing `--config` handling in the `chain` commands
- [#4350](https://github.com/ignite/cli/pull/4350) Skip upgrade prefix for sim tests

## [`v28.5.1`](https://github.com/ignite/cli/releases/tag/v28.5.1)

### Features

- [#4276](https://github.com/ignite/cli/pull/4276) Add `cosmosclient.CreateTxWithOptions` method to facilite more custom tx creation

### Changes

- [#4262](https://github.com/ignite/cli/pull/4262) Bring back relayer command
- [#4269](https://github.com/ignite/cli/pull/4269) Add custom flag parser for extensions
- [#4270](https://github.com/ignite/cli/pull/4270) Add flags to the extension hooks commands
- [#4286](https://github.com/ignite/cli/pull/4286) Add missing verbose mode flags

## [`v28.5.0`](https://github.com/ignite/cli/releases/tag/v28.5.0)

### Features

- [#4183](https://github.com/ignite/cli/pull/4183) Set `chain-id` in the client.toml
- [#4090](https://github.com/ignite/cli/pull/4090) Remove `protoc` pkg and also nodetime helpers `ts-proto` and `sta`
- [#4076](https://github.com/ignite/cli/pull/4076) Remove the ignite `relayer` and `tools` commands with all ts-relayer logic
- [#4133](https://github.com/ignite/cli/pull/4133) Improve buf rate limit

### Changes

- [#4095](https://github.com/ignite/cli/pull/4095) Migrate to matomo analytics
- [#4149](https://github.com/ignite/cli/pull/4149) Bump cometbft to `v0.38.7`
- [#4168](https://github.com/ignite/cli/pull/4168) Bump IBC to `v8.3.1`
  If you are upgrading manually from `v8.2.0` to `v8.3.1`, add the following to your `ibc.go` file:

  ```diff
  app.ICAHostKeeper = ...
  + app.ICAHostKeeper.WithQueryRouter(app.GRPCQueryRouter())`
  app.ICAControllerKeeper = ...
  ```

- [#4178](https://github.com/ignite/cli/pull/4178) Bump cosmos-sdk to `v0.50.7`
- [#4194](https://github.com/ignite/cli/pull/4194) Bump client/v2 to `v2.0.0-beta.2`
  If you are uprading manually, check out the recommended changes in `root.go` from the above PR.
- [#4210](https://github.com/ignite/cli/pull/4210) Improve default home wiring
- [#4077](https://github.com/ignite/cli/pull/4077) Merge the swagger files manually instead use nodetime `swagger-combine`
- [#4249](https://github.com/ignite/cli/pull/4249) Prevent creating a chain with number in the name
- [#4253](https://github.com/ignite/cli/pull/4253) Bump cosmos-sdk to `v0.50.8`

### Fixes

- [#4184](https://github.com/ignite/cli/pull/4184) Set custom `InitChainer` because of manually registered modules
- [#4198](https://github.com/ignite/cli/pull/4198) Set correct prefix overwriting in `buf.gen.pulsar.yaml`
- [#4199](https://github.com/ignite/cli/pull/4199) Set and seal SDK global config in `app/config.go`
- [#4212](https://github.com/ignite/cli/pull/4212) Set default values for extension flag to dont crash ignite
- [#4216](https://github.com/ignite/cli/pull/4216) Avoid create duplicated scopedKeppers
- [#4242](https://github.com/ignite/cli/pull/4242) Use buf build binary from the gobin path
- [#4250](https://github.com/ignite/cli/pull/4250) Set gas adjustment before calculating

## [`v28.4.0`](https://github.com/ignite/cli/releases/tag/v28.4.0)

### Features

- [#4108](https://github.com/ignite/cli/pull/4108) Add `xast` package (cherry-picked from [#3770](https://github.com/ignite/cli/pull/3770))
- [#4110](https://github.com/ignite/cli/pull/4110) Scaffold a consumer chain with `interchain-security` v5.0.0.
- [#4117](https://github.com/ignite/cli/pull/4117), [#4125](https://github.com/ignite/cli/pull/4125) Support relative path when installing local plugins

### Changes

- [#3959](https://github.com/ignite/cli/pull/3959) Remove app name prefix from the `.gitignore` file
- [#4103](https://github.com/ignite/cli/pull/4103) Bump cosmos-sdk to `v0.50.6`

### Fixes

- [#3969](https://github.com/ignite/cli/pull/3969) Get first config validator using a getter to avoid index errors
- [#4033](https://github.com/ignite/cli/pull/4033) Fix cobra completion using `fishshell`
- [#4062](https://github.com/ignite/cli/pull/4062) Avoid nil `scopedKeeper` in `TransmitXXX` functions
- [#4086](https://github.com/ignite/cli/pull/4086) Retry to get the IBC balance if it fails the first time
- [#4096](https://github.com/ignite/cli/pull/4096) Add new reserved names module and remove duplicated genesis order
- [#4112](https://github.com/ignite/cli/pull/4112) Remove duplicate SetCmdClientContextHandler
- [#4219](https://github.com/ignite/cli/pull/4219) Remove deprecated `sdk.MustSortJSON`

## [`v28.3.0`](https://github.com/ignite/cli/releases/tag/v28.3.0)

### Features

- [#4019](https://github.com/ignite/cli/pull/4019) Add `skip-proto` flag to `s chain` command
- [#3985](https://github.com/ignite/cli/pull/3985) Make some `cmd` pkg functions public
- [#3956](https://github.com/ignite/cli/pull/3956) Prepare for wasm app
- [#3660](https://github.com/ignite/cli/pull/3660) Add ability to scaffold ICS consumer chain

### Changes

- [#4035](https://github.com/ignite/cli/pull/4035) Bump `cometbft` to `v0.38.6` and `ibc-go/v8` to `v8.1.1`
- [#4031](https://github.com/ignite/cli/pull/4031) Bump `cli-plugin-network` to `v0.2.2` due to dependencies issue.
- [#4013](https://github.com/ignite/cli/pull/4013) Bump `cosmos-sdk` to `v0.50.5`
- [#4010](https://github.com/ignite/cli/pull/4010) Use `AppName` instead `ModuleName` for scaffold a new App
- [#3972](https://github.com/ignite/cli/pull/3972) Skip Ignite app loading for some base commands that don't allow apps
- [#3983](https://github.com/ignite/cli/pull/3983) Bump `cosmos-sdk` to `v0.50.4` and `ibc-go` to `v8.1.0`

### Fixes

- [#4021](https://github.com/ignite/cli/pull/4021) Set correct custom signer in `s list --signer <signer>`
- [#3995](https://github.com/ignite/cli/pull/3995) Fix interface check for ibc modules
- [#3953](https://github.com/ignite/cli/pull/3953) Fix apps `Stdout` is redirected to `Stderr`
- [#3863](https://github.com/ignite/cli/pull/3963) Fix breaking issue for app client API when reading app chain info

## [`v28.2.0`](https://github.com/ignite/cli/releases/tag/v28.2.0)

### Features

- [#3924](https://github.com/ignite/cli/pull/3924) Scaffold NFT module by default
- [#3839](https://github.com/ignite/cli/pull/3839) New structure for app scaffolding
- [#3835](https://github.com/ignite/cli/pull/3835) Add `--minimal` flag to `scaffold chain` to scaffold a chain with the least amount of sdk modules
- [#3820](https://github.com/ignite/cli/pull/3820) Add integration tests for IBC chains
- [#3956](https://github.com/ignite/cli/pull/3956) Prepare for wasm app

### Changes

- [#3899](https://github.com/ignite/cli/pull/3899) Introduce `plugin.Execute` function
- [#3903](https://github.com/ignite/cli/pull/3903) Don't specify a default build tag and deprecate notion of app version

### Fixes

- [#3905](https://github.com/ignite/cli/pull/3905) Fix `ignite completion`
- [#3931](https://github.com/ignite/cli/pull/3931) Fix `app update` command and duplicated apps

## [`v28.1.1`](https://github.com/ignite/cli/releases/tag/v28.1.1)

### Fixes

- [#3878](https://github.com/ignite/cli/pull/3878) Support local forks of Cosmos SDK in scaffolded chain.
- [#3869](https://github.com/ignite/cli/pull/3869) Fix .git in parent dir
- [#3867](https://github.com/ignite/cli/pull/3867) Fix genesis export for ibc modules.
- [#3850](https://github.com/ignite/cli/pull/3871) Fix app.go file detection in apps scaffolded before v28.0.0

### Changes

- [#3885](https://github.com/ignite/cli/pull/3885) Scaffold chain with Cosmos SDK `v0.50.3`
- [#3877](https://github.com/ignite/cli/pull/3877) Change Ignite App extension to "ign"
- [#3897](https://github.com/ignite/cli/pull/3897) Introduce alternative folder in templates

## [`v28.1.0`](https://github.com/ignite/cli/releases/tag/v28.1.0)

### Features

- [#3786](https://github.com/ignite/cli/pull/3786) Add artifacts for publishing Ignite to FlatHub and Snapcraft
- [#3830](https://github.com/ignite/cli/pull/3830) Remove gRPC info from Ignite Apps errors
- [#3861](https://github.com/ignite/cli/pull/3861) Send to the analytics if the user is using a GitPod

### Changes

- [#3822](https://github.com/ignite/cli/pull/3822) Improve default scaffolded AutoCLI config
- [#3838](https://github.com/ignite/cli/pull/3838) Scaffold chain with Cosmos SDK `v0.50.2`, and bump confix and x/upgrade to latest
- [#3829](https://github.com/ignite/cli/pull/3829) Support version prefix for cached values
- [#3723](https://github.com/ignite/cli/pull/3723) Create a wrapper for errors

### Fixes

- [#3827](https://github.com/ignite/cli/pull/3827) Change ignite apps to be able to run in any directory
- [#3831](https://github.com/ignite/cli/pull/3831) Correct ignite app gRPC server stop memory issue
- [#3825](https://github.com/ignite/cli/pull/3825) Fix a minor Keplr type-checking bug in TS client
- [#3836](https://github.com/ignite/cli/pull/3836), [#3858](https://github.com/ignite/cli/pull/3858) Add missing IBC commands for scaffolded chain
- [#3833](https://github.com/ignite/cli/pull/3833) Improve Cosmos SDK detection to support SDK forks
- [#3849](https://github.com/ignite/cli/pull/3849) Add missing `tx.go` file by default and enable cli if autocli does not exist
- [#3851](https://github.com/ignite/cli/pull/3851) Add missing ibc interfaces to chain client
- [#3860](https://github.com/ignite/cli/pull/3860) Fix analytics event name

## [`v28.0.0`](https://github.com/ignite/cli/releases/tag/v28.0.0)

### Features

- [#3659](https://github.com/ignite/cli/pull/3659) cosmos-sdk `v0.50.x` upgrade
- [#3694](https://github.com/ignite/cli/pull/3694) Query and Tx AutoCLI support
- [#3536](https://github.com/ignite/cli/pull/3536) Change app.go to v2 and add AppWiring feature
- [#3544](https://github.com/ignite/cli/pull/3544) Add bidirectional communication to app (plugin) system
- [#3756](https://github.com/ignite/cli/pull/3756) Add faucet compatibility for latest sdk chains
- [#3476](https://github.com/ignite/cli/pull/3476) Use `buf.build` binary to code generate from proto files
- [#3724](https://github.com/ignite/cli/pull/3724) Add or vendor proto packages from Go dependencies
- [#3561](https://github.com/ignite/cli/pull/3561) Add GetChainInfo method to plugin system API
- [#3626](https://github.com/ignite/cli/pull/3626) Add logging levels to relayer
- [#3614](https://github.com/ignite/cli/pull/3614) feat: use DefaultBaseappOptions for app.New method
- [#3715](https://github.com/ignite/cli/pull/3715) Add test suite for the cli tests

### Changes

- [#3793](https://github.com/ignite/cli/pull/3793) Refactor Ignite to follow semantic versioning (prepares v28.0.0). If you are using packages, do not forget to import the `/v28` version of the packages.
- [#3529](https://github.com/ignite/cli/pull/3529) Refactor plugin system to use gRPC
- [#3751](https://github.com/ignite/cli/pull/3751) Rename label to skip changelog check
- [#3745](https://github.com/ignite/cli/pull/3745) Set tx fee amount as option
- [#3748](https://github.com/ignite/cli/pull/3748) Change default rpc endpoint to a working one
- [#3621](https://github.com/ignite/cli/pull/3621) Change `pkg/availableport` to allow custom parameters in `Find` function and handle duplicated ports
- [#3810](https://github.com/ignite/cli/pull/3810) Bump network app version to `v0.2.1`
- [#3581](https://github.com/ignite/cli/pull/3581) Bump cometbft and cometbft-db in the template
- [#3522](https://github.com/ignite/cli/pull/3522) Remove indentation from `chain serve` output
- [#3346](https://github.com/ignite/cli/issues/3346) Improve scaffold query --help
- [#3601](https://github.com/ignite/cli/pull/3601) Update ts-relayer version to `0.10.0`
- [#3658](https://github.com/ignite/cli/pull/3658) Rename Marshaler to Codec in EncodingConfig
- [#3653](https://github.com/ignite/cli/pull/3653) Add "app" extension to plugin binaries
- [#3656](https://github.com/ignite/cli/pull/3656) Disable Go toolchain download
- [#3662](https://github.com/ignite/cli/pull/3662) Refactor CLI "plugin" command to "app"
- [#3669](https://github.com/ignite/cli/pull/3669) Rename `plugins` config file to `igniteapps`
- [#3683](https://github.com/ignite/cli/pull/3683) Resolve `--dep auth` as `--dep account` in `scaffold module`
- [#3795](https://github.com/ignite/cli/pull/3795) Bump cometbft to `v0.38.2`
- [#3599](https://github.com/ignite/cli/pull/3599) Add analytics as an option
- [#3670](https://github.com/ignite/cli/pull/3670) Remove binaries

### Fixes

- [#3386](https://github.com/ignite/cli/issues/3386) Prevent scaffolding of default module called "ibc"
- [#3592](https://github.com/ignite/cli/pull/3592) Fix `pkg/protoanalysis` to support HTTP rule parameter arguments
- [#3598](https://github.com/ignite/cli/pull/3598) Fix consensus param keeper constructor key in `app.go`
- [#3610](https://github.com/ignite/cli/pull/3610) Fix overflow issue of cosmos faucet in `pkg/cosmosfaucet/transfer.go` and `pkg/cosmosfaucet/cosmosfaucet.go`
- [#3618](https://github.com/ignite/cli/pull/3618) Fix TS client generation import path issue
- [#3631](https://github.com/ignite/cli/pull/3631) Fix unnecessary vue import in hooks/composables template
- [#3661](https://github.com/ignite/cli/pull/3661) Change `pkg/cosmosanalysis` to find Cosmos SDK runtime app registered modules
- [#3716](https://github.com/ignite/cli/pull/3716) Fix invalid plugin hook check
- [#3725](https://github.com/ignite/cli/pull/3725) Fix flaky TS client generation issues on linux
- [#3726](https://github.com/ignite/cli/pull/3726) Update TS client dependencies. Bump vue/react template versions
- [#3728](https://github.com/ignite/cli/pull/3728) Fix wrong parser for proto package names
- [#3729](https://github.com/ignite/cli/pull/3729) Fix broken generator due to caching /tmp include folders
- [#3767](https://github.com/ignite/cli/pull/3767) Fix `v0.50` ibc genesis issue
- [#3808](https://github.com/ignite/cli/pull/3808) Correct TS code generation to generate paginated fields

## [`v0.27.2`](https://github.com/ignite/cli/releases/tag/v0.27.2)

### Changes

- [#3701](https://github.com/ignite/cli/pull/3701) Bump `go` version to 1.21

## [`v0.27.1`](https://github.com/ignite/cli/releases/tag/v0.27.1)

### Features

- [#3505](https://github.com/ignite/cli/pull/3505) Auto migrate dependency tools
- [#3538](https://github.com/ignite/cli/pull/3538) bump sdk to `v0.47.3` and ibc to `v7.1.0`
- [#2736](https://github.com/ignite/cli/issues/2736) Add `--skip-git` flag to skip git repository initialization.
- [#3381](https://github.com/ignite/cli/pull/3381) Add `ignite doctor` command
- [#3446](https://github.com/ignite/cli/pull/3446) Add `gas-adjustment` flag to the cosmos client.
- [#3439](https://github.com/ignite/cli/pull/3439) Add `--build.tags` flag for `chain serve` and `chain build` commands.
- [#3524](https://github.com/ignite/cli/pull/3524) Apply auto tools migration to other commands
- Added compatibility check and auto migration features and interactive guidelines for the latest versions of the SDK

### Changes

- [#3444](https://github.com/ignite/cli/pull/3444) Add support for ICS chains in ts-client generation
- [#3494](https://github.com/ignite/cli/pull/3494) bump `cosmos-sdk` and `cometbft` versions
- [#3434](https://github.com/ignite/cli/pull/3434) Detect app wiring implementation

### Fixes

- [#3497](https://github.com/ignite/cli/pull/3497) Use corret bank balance query url in faucet openapi
- [#3481](https://github.com/ignite/cli/pull/3481) Use correct checksum format in release checksum file
- [#3470](https://github.com/ignite/cli/pull/3470) Prevent overriding minimum-gas-prices with default value
- [#3523](https://github.com/ignite/cli/pull/3523) Upgrade Cosmos SDK compatibility check for scaffolded apps
- [#3441](https://github.com/ignite/cli/pull/3441) Correct wrong client context for cmd query methods
- [#3487](https://github.com/ignite/cli/pull/3487) Handle ignired error in package `cosmosaccount` `Account.PubKey`

## [`v0.26.1`](https://github.com/ignite/cli/releases/tag/v0.26.1)

### Features

- [#3238](https://github.com/ignite/cli/pull/3238) Add `Sharedhost` plugin option
- [#3214](https://github.com/ignite/cli/pull/3214) Global plugins config.
- [#3142](https://github.com/ignite/cli/pull/3142) Add `ignite network request param-change` command.
- [#3181](https://github.com/ignite/cli/pull/3181) Addition of `add` and `remove` commands for `plugins`
- [#3184](https://github.com/ignite/cli/pull/3184) Separate `plugins.yml` config file.
- [#3038](https://github.com/ignite/cli/pull/3038) Addition of Plugin Hooks in Plugin System
- [#3056](https://github.com/ignite/cli/pull/3056) Add `--genesis-config` flag option to `ignite network chain publish`
- [#2892](https://github.com/ignite/cli/pull/2982/) Add `ignite scaffold react` command.
- [#2892](https://github.com/ignite/cli/pull/2982/) Add `ignite generate composables` command.
- [#2892](https://github.com/ignite/cli/pull/2982/) Add `ignite generate hooks` command.
- [#2955](https://github.com/ignite/cli/pull/2955/) Add `ignite network request add-account` command.
- [#2877](https://github.com/ignite/cli/pull/2877) Plugin system
- [#3060](https://github.com/ignite/cli/pull/3060) Plugin system flag support
- [#3105](https://github.com/ignite/cli/pull/3105) Addition of `ignite plugin describe <path>` command
- [#2995](https://github.com/ignite/cli/pull/2995/) Add `ignite network request remove-validator` command.
- [#2999](https://github.com/ignite/cli/pull/2999/) Add `ignite network request remove-account` command.
- [#2458](https://github.com/ignite/cli/issues/2458) New `chain serve` command UI.
- [#2992](https://github.com/ignite/cli/issues/2992) Add `ignite chain debug` command.

### Changes

- [#3369](https://github.com/ignite/cli/pull/3369) Update `ibc-go` to `v6.1.0`.
- [#3306](https://github.com/ignite/cli/pull/3306) Move network command into a plugin
- [#3305](https://github.com/ignite/cli/pull/3305) Bump Cosmos SDK version to `v0.46.7`.
- [#3068](https://github.com/ignite/cli/pull/3068) Add configs to generated TS code for working with JS projects
- [#3071](https://github.com/ignite/cli/pull/3071) Refactor `ignite/templates` package.
- [#2892](https://github.com/ignite/cli/pull/2982/) `ignite scaffold vue` and `ignite scaffold react` use v0.4.2 templates
- [#2892](https://github.com/ignite/cli/pull/2982/) `removeSigner()` method added to generated `ts-client`
- [#3035](https://github.com/ignite/cli/pull/3035) Bump Cosmos SDK to `v0.46.4`.
- [#3037](https://github.com/ignite/cli/pull/3037) Bump `ibc-go` to `v5.0.1`.
- [#2957](https://github.com/ignite/cli/pull/2957) Change generate commands to print the path to the generated code.
- [#2981](https://github.com/ignite/cli/issues/2981) Change CLI to also search chain binary in Go binary path.
- [#2958](https://github.com/ignite/cli/pull/2958) Support absolute paths for client code generation config paths.
- [#2993](https://github.com/ignite/cli/pull/2993) Hide `ignite scaffold band` command and deprecate functionality.
- [#2986](https://github.com/ignite/cli/issues/2986) Remove `--proto-all-modules` flag because it is now the default behaviour.
- [#2986](https://github.com/ignite/cli/issues/2986) Remove automatic Vue code scaffolding from `scaffold chain` command.
- [#2986](https://github.com/ignite/cli/issues/2986) Add `--generate-clients` to `chain serve` command for optional client code (re)generation.
- [#2998](https://github.com/ignite/cli/pull/2998) Hide `ignite generate dart` command and remove functionality.
- [#2991](https://github.com/ignite/cli/pull/2991) Hide `ignite scaffold flutter` command and remove functionality.
- [#2944](https://github.com/ignite/cli/pull/2944) Add a new event "update" status option to `pkg/cliui`.
- [#3030](https://github.com/ignite/cli/issues/3030) Remove colon syntax from module scaffolding `--dep` flag.
- [#3025](https://github.com/ignite/cli/issues/3025) Improve config version error handling.
- [#3084](https://github.com/ignite/cli/pull/3084) Add Ignite Chain documentation.
- [#3109](https://github.com/ignite/cli/pull/3109) Refactor scaffolding for proto files to not rely on placeholders.
- [#3106](https://github.com/ignite/cli/pull/3106) Add zoom image plugin.
- [#3194](https://github.com/ignite/cli/issues/3194) Move config validators check to validate only when required.
- [#3183](https://github.com/ignite/cli/pull/3183/) Make config optional for init phase.
- [#3224](https://github.com/ignite/cli/pull/3224) Remove `grpc_*` prefix from query files in scaffolded chains
- [#3229](https://github.com/ignite/cli/pull/3229) Rename `campaign` to `project` in ignite network set of commands
- [#3122](https://github.com/ignite/cli/issues/3122) Change `generate ts-client` to ignore the cache by default.
- [#3244](https://github.com/ignite/cli/pull/3244) Update `actions.yml` for resolving deprecation message
- [#3337](https://github.com/ignite/cli/pull/3337) Remove `pkg/openapiconsole` import from scaffold template.
- [#3337](https://github.com/ignite/cli/pull/3337) Register`nodeservice` gRPC in `app.go` template.
- [#3455](https://github.com/ignite/cli/pull/3455) Bump `cosmos-sdk` to `v0.47.1`
- [#3434](https://github.com/ignite/cli/pull/3434) Detect app wiring implementation.
- [#3445](https://github.com/ignite/cli/pull/3445) refactor: replace `github.com/ghodss/yaml` with `sigs.k8s.io/yaml`

### Breaking Changes

- [#3033](https://github.com/ignite/cli/pull/3033) Remove Cosmos SDK Launchpad version support.

### Fixes

- [#3114](https://github.com/ignite/cli/pull/3114) Fix out of gas issue when approving many requests
- [#3068](https://github.com/ignite/cli/pull/3068) Fix REST codegen method casing bug
- [#3031](https://github.com/ignite/cli/pull/3031) Move keeper hooks to after all keepers initialized in `app.go` template.
- [#3098](https://github.com/ignite/cli/issues/3098) Fix config upgrade issue that left config empty on error.
- [#3129](https://github.com/ignite/cli/issues/3129) Remove redundant `keyring-backend` config option.
- [#3187](https://github.com/ignite/cli/issues/3187) Change prompt text to fit within 80 characters width.
- [#3203](https://github.com/ignite/cli/issues/3203) Fix relayer to work with multiple paths.
- [#3320](https://github.com/ignite/cli/pull/3320) Allow `id` and `creator` as names when scaffolding a type.
- [#3327](https://github.com/ignite/cli/issues/3327) Scaffolding messages with same name leads to aliasing.
- [#3383](https://github.com/ignite/cli/pull/3383) State error and info are now displayed when using serve UI.
- [#3379](https://github.com/ignite/cli/issues/3379) Fix `ignite docs` issue by disabling mouse support.
- [#3435](https://github.com/ignite/cli/issues/3435) Fix wrong client context for cmd query methods.

## [`v0.25.2`](https://github.com/ignite/cli/releases/tag/v0.25.1)

### Changes

- [#3145](https://github.com/ignite/cli/pull/3145) Security fix upgrading Cosmos SDK to `v0.46.6`

## [`v0.25.1`](https://github.com/ignite/cli/releases/tag/v0.25.1)

### Changes

- [#2968](https://github.com/ignite/cli/pull/2968) Dragonberry security fix upgrading Cosmos SDK to `v0.46.3`

## [`v0.25.0`](https://github.com/ignite/cli/releases/tag/v0.25.0)

### Features

- Add `pkg/cosmostxcollector` package with support to query and save TXs and events.
- Add `ignite network coordinator` command set.
- Add `ignite network validator` command set.
- Deprecate `cosmoscmd` pkg and add cmd templates for scaffolding.
- Add generated TS client test support to integration tests.

### Changes

- Updated `pkg/cosmosanalysis` to discover the list of app modules when defined in variables or functions.
- Improve genesis parser for `network` commands
- Integration tests build their own ignite binary.
- Updated `pkg/cosmosanalysis` to discover the list of app modules when defined in variables.
- Switch to broadcast mode sync in `cosmosclient`
- Updated `nodetime`: `ts-proto` to `v1.123.0`, `protobufjs` to `v7.1.1`, `swagger-typescript-api` to `v9.2.0`
- Switched codegen client to use `axios` instead of `fetch`
- Added `useKeplr()` and `useSigner()` methods to TS client. Allowed query-only instantiation.
- `nodetime` built with `vercel/pkg@5.6.0`
- Change CLI to use an events bus to print to stdout.
- Move generated proto files to `proto/{appname}/{module}`
- Update `pkg/cosmosanalysis` to detect when proto RPC services are using pagination.
- Add `--peer-address` flag to `network chain join` command.
- Change nightly tag format
- Add cosmos-sdk version in `version` command
- [#2935](https://github.com/ignite/cli/pull/2935) Update `gobuffalo/plush` templating tool to `v4`

### Fixes

- Fix ICA controller wiring.
- Change vuex generation to use a default TS client path.
- Fix cli action org in templates.
- Seal the capability keeper in the `app.go` template.
- Change faucet to allow CORS preflight requests.
- Fix config file migration to void leaving end of file content chunks.
- Change session print loop to block until all events are handled.
- Handle "No records were found in keyring" message when checking keys.
- [#2941](https://github.com/ignite/cli/issues/2941) Fix session to use the same spinner referece.
- [#2922](https://github.com/ignite/cli/pull/2922) Network commands check for latest config version before building the chain binary.

## [`v0.24.1`](https://github.com/ignite/cli/releases/tag/v0.24.1)

### Features

- Upgraded Cosmos SDK to `v0.46.2`.

## [`v0.24.0`](https://github.com/ignite/cli/releases/tag/v0.24.0)

### Features

- Upgraded Cosmos SDK to `v0.46.0` and IBC to `v5` in CLI and scaffolding templates
- Change chain init to check that no gentx are present in the initial genesis
- Add `network rewards release` command
- Add "make mocks" target to Makefile
- Add `--skip-proto` flag to `build`, `init` and `serve` commands to build the chain without building proto files
- Add `node query tx` command to query a transaction in any chain.
- Add `node query bank` command to query an account's bank balance in any chain.
- Add `node tx bank send` command to send funds from one account to another in any chain.
- Add migration system for the config file to allow config versioning
- Add `node tx bank send` command to send funds from one account to another in any chain.
- Implement `network profile` command
- Add `generate ts-client` command to generate a stand-alone modular TypeScript client.

### Changes

- Add changelog merge strategy in `.gitattributes` to avoid conflicts.
- Refactor `templates/app` to remove `monitoringp` module from the default template
- Updated keyring dependency to match Cosmos SDK
- Speed up the integration tests
- Refactor ignite network and fix genesis generation bug
- Make Go dependency verification optional during build by adding the `--check-dependencies` flag
  so Ignite CLI can work in a Go workspace context.
- Temporary SPN address change for nightly
- Rename `simapp.go.plush` simulation file template to `helpers.go.plush`
- Remove campaign creation from the `network chain publish` command
- Optimized JavaScript generator to use a single typescript API generator binary
- Improve documentation and add support for protocol buffers and Go modules syntax
- Add inline documentation for CLI commands
- Change `cmd/account` to skip passphrase prompt when importing from mnemonic
- Add nodejs version in the output of ignite version
- Removed `handler.go` from scaffolded module template
- Migrated to `cosmossdk.io` packages for and `math`
- Vuex stores from the `generate vuex` command use the new TypeScript client
- Upgraded frontend Vue template to v0.3.10

### Fixes

- Improved error handling for crypto wrapper functions
- Fix `pkg/cosmosclient` to call the faucet prior to creating the tx.
- Test and refactor `pkg/comosclient`.
- Change templates to add missing call to `RegisterMsgServer` in the default module's template to match what's specified
  in the docs
- Fix cosmoscmd appID parameter value to sign a transaction correctly
- Fix `scaffold query` command to use `GetClientQueryContext` instead of `GetClientTxContext`
- Fix flaky integration tests issue that failed with "text file busy"
- Fix default chain ID for publish
- Replace `os.Rename` with `xos.Rename`
- Fix CLI reference generation to add `ignite completion` documentation
- Remove usage of deprecated `io/ioutil` package

## [`v0.23.0`](https://github.com/ignite/cli/releases/tag/v0.23.0)

### Features

- Apps can now use generics

### Fixes

- Fix `pkg/cosmosanalysis` to support apps with generics
- Remove `ignite-hq/cli` from dependency list in scaffolded chains

### Changes

- Change `pkg/cosmosgen` to allow importing IBC proto files
- Improve docs for Docker related commands
- Improve and fix documentation issues in developer tutorials
- Add migration docs for v0.22.2
- Improve `go mod download` error report in `pkg/cosmosgen`

## [`v0.22.2`](https://github.com/ignite/cli/releases/tag/v0.22.2)

### Features

- Enable Darwin ARM 64 target for chain binary releases in CI templates

### Changes

- Rename `ignite-hq` to `ignite`

## [`v0.22.1`](https://github.com/ignite/cli/releases/tag/v0.22.1)

### Fixes

- Fix IBC module scaffolding interface in templates

## [`v0.22.0`](https://github.com/ignite/cli/releases/tag/v0.22.0)

### Features

- Optimized the build system. The `chain serve`, `chain build`, `chain generate` commands and other variants are way
  faster now
- Upgraded CLI and templates to use IBC v3

### Fixes

- Add a fix in code generation to avoid user's NodeJS configs to break TS client generation routine

## [`v0.21.2`](https://github.com/ignite/cli/releases/tag/v0.21.2)

### Fixes

- Set min. gas to zero when running `chain` command set

## [`v0.21.1`](https://github.com/ignite/cli/releases/tag/v0.21.1)

### Features

- Add compatibility to run chains built with Cosmos-SDK `v0.46.0-alpha1` and above
- Scaffold chains now will have `auth` module enabled by default

### Fixes

- Fixed shell completion generation
- Make sure proto package names are valid when using simple app names

## [`v0.21.0`](https://github.com/ignite/cli/releases/tag/v0.21.0)

### Features

- Support simple app names when scaffolding chains. e.g.: `ignite scaffold chain mars`
- Ask confirmation when scaffolding over changes that are not committed yet

## [`v0.20.4`](https://github.com/ignite/cli/releases/tag/v0.20.4)

### Fixes

- Use `protoc` binary compiled in an older version of macOS AMD64 for backwards compatibility in code generation

## [`v0.20.3`](https://github.com/ignite/cli/releases/tag/v0.20.3)

### Fixes

- Use the latest version of CLI in templates to fix Linux ARM support _(It's now possible to develop chains in Linux ARM
  machines and since the chain depends on the CLI in its `go.mod`, it needs to use the latest version that support ARM
  targets)_

## [`v0.20.2`](https://github.com/ignite/cli/releases/tag/v0.20.2)

### Fixes

- Use `unsafe-reset-all` cmd under `tendermint` cmd for chains that use `=> v0.45.3` version of Cosmos SDK

## [`v0.20.1`](https://github.com/ignite/cli/releases/tag/v0.20.1)

### Features

- Release the CLI with Linux ARM and native M1 binaries

## [`v0.20.0`](https://github.com/ignite/cli/releases/tag/v0.20.0)

Our new name is **Ignite CLI**!

**IMPORTANT!** This upgrade renames `starport` command to `ignite`. From now on, use `ignite` command to access the CLI.

### Features

- Upgraded Cosmos SDK version to `v0.45.2`
- Added support for in memory backend in `pkg/cosmosclient` package
- Improved our tutorials and documentation

## [`v0.19.5`](https://github.com/ignite/cli/pull/2158/commits)

### Features

- Enable client code and Vuex code generation for query only modules as well.
- Upgraded the Vue template to `v0.3.5`.

### Fixes

- Fixed snake case in code generation.
- Fixed plugin installations for Go =>v1.18.

### Changes

- Dropped transpilation of TS to JS. Code generation now only produces TS files.

## `v0.19.4`

### Features

- Upgraded Vue template to `v0.3.0`.

## `v0.19.3`

### Features

- Upgraded Flutter template to `v2.0.3`

## [`v0.19.2`](https://github.com/ignite/cli/milestone/14)

### Fixes

- Fixed race condition during faucet transfer
- Fixed account sequence mismatch issue on faucet and relayer
- Fixed templates for IBC code scaffolding

### Features

- Upgraded blockchain templates to use IBC v2.0.2

### Breaking Changes

- Deprecated the Starport Modules [tendermint/spm](https://github.com/tendermint/spm) repo and moved the contents to the
  Ignite CLI repo [`ignite/pkg/`](https://github.com/ignite/cli/tree/main/ignite/pkg/)
  in [PR 1971](https://github.com/ignite/cli/pull/1971/files)

  Updates are required if your chain uses these packages:

  - `spm/ibckeeper` is now `pkg/cosmosibckeeper`
  - `spm/cosmoscmd` is now `pkg/cosmoscmd`
  - `spm/openapiconsole` is now `pkg/openapiconsole`
  - `testutil/sample` is now `cosmostestutil/sample`

- Updated the faucet HTTP API schema. See API changes
  in [fix: improve faucet reliability #1974](https://github.com/ignite/cli/pull/1974/files#diff-0e157f4f60d6fbd95e695764df176c8978d85f1df61475fbfa30edef62fe35cd)

## `v0.19.1`

### Fixes

- Enabled the `scaffold flutter` command

## `v0.19.0`

### Features

- `starport scaffold` commands support `ints`, `uints`, `strings`, `coin`, `coins` as field types (#1579)
- Added simulation testing with `simapp` to the default template (#1731)
- Added `starport generate dart` to generate a Dart client from protocol buffer files
- Added `starport scaffold flutter` to scaffold a Flutter mobile app template
- Parameters can be specified with a new `--params` flag when scaffolding modules (#1716)
- Simulations can be run with `starport chain simulate`
- Set `cointype` for accounts in `config.yml` (#1663)

### Fixes

- Allow using a `creator` field when scaffolding a model with a `--no-message` flag (#1730)
- Improved error handling when generating code (#1907)
- Ensure account has funds after faucet transfer when using `cosmosclient` (#1846)
- Move from `io/ioutil` to `io` and `os` package (refactoring) (#1746)

## `v0.18.0`

### Breaking Changes

- Starport v0.18 comes with Cosmos SDK v0.44 that introduced changes that are not compatible with chains that were
  scaffolded with Starport versions lower than v0.18. After upgrading from Starport v0.17.3 to Starport v0.18, you must
  update the default blockchain template to use blockchains that were scaffolded with earlier versions.
  See [Migration](https://docs.ignite.com/migration).

### Features

- Scaffold commands allow using previously scaffolded types as fields
- Added `--signer` flag to `message`, `list`, `map`, and `single` scaffolding to allow customizing the name of the
  signer of the message
- Added `--index` flag to `scaffold map` to provide a custom list of indices
- Added `scaffold type` to scaffold a protocol buffer definition of a type
- Automatically check for new Starport versions
- Added `starport tools completions` to generate CLI completions
- Added `starport account` commands to manage accounts (key pairs)
- `starport version` now prints detailed information about OS, Go version, and more
- Modules are scaffolded with genesis validation tests
- Types are scaffolded with tests for `ValidateBasic` methods
- `cosmosclient` has been refactored and can be used as a library for interacting with Cosmos SDK chains
- `starport relayer` uses `starport account`
- Added `--path` flag for all `scaffold`, `generate` and `chain` commands
- Added `--output` flag to the `build` command
- Configure port of gRPC web in `config.yml` with the `host.grpc-web` property
- Added `build.main` field to `config.yml` for apps to specify the path of the chain's main package. This property is
  required to be set only when an app contains multiple main packages.

### Fixes

- Scaffolding a message no longer prevents scaffolding a map, list, or single that has the same type name when using
  the `--no-message` flag
- Generate Go code from proto files only from default directories or directories specified in `config.yml`
- Fixed faucet token transfer calculation
- Removed `creator` field for types scaffolded with the `--no-message` flag
- Encode the count value in the store with `BigEndian`

## `v0.17.3`

### Fixes

- oracle: add a specific BandChain pkg version to avoid Cosmos SDK version conflicts

## `v0.17.2`

### Features

- `client.toml` is initialized and used by node's CLI, can be configured through `config.yml` with the `init.client`
  property
- Support serving Cosmos SDK `v0.43.x` based chains

## `v0.17.1`

### Fixes

- Set visibility to `public` on Gitpod's port 7575 to enable peer discovery for SPN
- Fixed GitHub action that releases blockchain node's binary
- Fixed an error in chain scaffolding due to "unknown revision"
- Fixed an error in `starport chain serve` by limiting the scope where proto files are searched for

## `v0.17`

### Features

- Added GitHub action that automatically builds and releases a binary
- The `--release` flag for the `build` command adds the ability to release binaries in a tarball with a checksum file.
- Added the flag `--no-module` to the command `starport app` to prevent scaffolding a default module when creating a new
  app
- Added `--dep` flag to specify module dependency when scaffolding a module
- Added support for multiple naming conventions for component names and field names
- Print created and modified files when scaffolding a new component
- Added `starport generate` namespace with commands to generate Go, Vuex and OpenAPI
- Added `starport chain init` command to initialize a chain without starting a node
- Scaffold a type that contains a single instance in the store
- Introduced `starport tools` command for advanced users. Existing `starport relayer lowlevel *` commands are also moved
  under `tools`
- Added `faucet.rate_limit_window` property to `config.yml`
- Simplified the `cmd` package in the template
- Added `starport scaffold band` oracle query scaffolding
- Updated TypeScript relayer to 0.2.0
- Added customizable gas limits for the relayer

### Fixes

- Use snake case for generated files
- Prevent using incorrect module name
- Fixed permissions issue when using Starport in Docker
- Ignore hidden directories when building a chain
- Fix error when scaffolding an IBC module in non-Starport chains

## `v0.16.2`

### Fix

- Prevent indirect Buf dependency

## `v0.16.1`

### Features

- Ensure that CLI operates fine even if the installation directory (bin) of Go programs is not configured properly

## `v0.16.0`

### Features

- The new `join` flag adds the ability to pass a `--genesis` file and `--peers` address list
  with `starport network chain join`
- The new `show` flag adds the ability to show `--genesis` and `--peers` list with `starport network chain show`
- `protoc` is now bundled with Ignite CLI. You don't need to install it anymore.
- Starport is now published automatically on the Docker Hub
- `starport relayer` `configure` and `connect` commands now use
  the [confio/ts-relayer](https://github.com/confio/ts-relayer) under the hood. Also, checkout the
  new `starport relayer lowlevel` command
- An OpenAPI spec for your chain is now automatically generated with `serve` and `build` commands: a console is
  available at `localhost:1317` and spec at `localhost:1317/static/openapi.yml` by default for the newly scaffolded
  chains
- Keplr extension is supported on web apps created with Starport
- Added tests to the scaffold
- Improved reliability of scaffolding by detecting placeholders
- Added ability to scaffold modules in chains not created with Starport
- Added the ability to scaffold Cosmos SDK queries
- IBC relayer support is available on web apps created with Starport
- New types without CRUD operations can be added with the `--no-message` flag in the `type` command
- New packet without messages can be added with the `--no-message` flag in the `packet` command
- Added `docs` command to read Starport documentation on the CLI
- Published documentation on <https://docs.starport.network>
- Added `mnemonic` property to account in the `accounts` list to generate a key from a mnemonic

### Fixes

- `starport network chain join` hanging issue when creating an account
- Error when scaffolding a chain with an underscore in the repo name (thanks @bensooraj!)

### Changes

- `starport serve` no longer starts the web app in the `vue` directory (use `npm` to start it manually)
- Default scaffold no longer includes legacy REST API endpoints (thanks @bensooraj!)
- Removed support for Cosmos SDK v0.39 Launchpad

## `v0.15.0`

### Features

- IBC module scaffolding
- IBC packet scaffolding with acknowledgements
- JavaScript and Vuex client code generation for Cosmos SDK and custom modules
- Standalone relayer with `configure` and `connect` commands
- Advanced relayer options for configuring ports and versions
- Scaffold now follows `MsgServer` convention
- Message scaffolding
- Added `starport type ... --indexed` to scaffold indexed types
- Custom config file support with `starport serve -c custom.yml`
- Detailed terminal output for created accounts: name, address, mnemonic
- Added spinners to indicate progress for long-running commands
- Updated to Cosmos SDK v0.42.1

### Changes

- Replaced `packr` with Go 1.16 `embed`
- Renamed `servers` top-level property to `host`

## `v0.14.0`

### Features

- Chain state persistence between `starport serve` launches
- Integrated Stargate app's `scripts/protocgen` into Starport as a native feature. Running `starport build/serve` will
  automatically take care of building proto files without a need of script in the app's source code.
- Integrated third-party proto-files used by Cosmos SDK modules into Ignite CLI
- Added ability to customize binary name with `build.binary` in `config.yml`
- Added ability to change path to home directory with `.home` in `config.yml`
- Added ability to add accounts by `address` with in `config.yml`
- Added faucet functionality available on port 4500 and configurable with `faucet` in `config.yml`
- Added `starport faucet [address] [coins]` command
- Updated scaffold to Cosmos SDK v0.41.0
- Distroless multiplatform docker containers for starport that can be used for `starport serve`
- UI containers for chains scaffolded with Starport
- Use SOS-lite and Docker instead of systemD
- Arch PKGBUILD in `scripts`

### Fixes

- Support for CosmWasm on Stargate
- Bug with dashes in GitHub username breaking proto package name
- Bug with custom address prefix
- use docker buildx as a single command with multiple platforms to make multi-manifest work properly

## `v0.13.0`

### Features

- Added `starport network` commands for launching blockchains
- Added proxy (Chisel) to support launching blockchains from Gitpod
- Upgraded the template (Stargate) to Cosmos SDK v0.40.0-rc3
- Added a gRPC-Web proxy that is available under <http://localhost:12345/grpc>
- Added chain id configurability by recognizing `chain_id` from `genesis` section of `config.yml`.
- Added `config/app.toml` and `config/config.toml` configurability for appd under new `init.app` and `init.config`
  sections of `config.yml`
- Point to Stargate as default SDK version for scaffolding
- Covered CRUD operations for Stargate scaffolding
- Added docs on gopath to build from source directions
- Arch Linux Based Raspberry Pi development environment
- Calculate the necessary gas for sending transactions to SPN

### Fixes

- Routing REST API endpoints of querier on Stargate
- Evaluate `--address-prefix` option when scaffolding for Stargate
- Use a deterministic method to generate scaffolded type IDs
- Modify scaffolded type's creator type from address to string
- Copy built starport arm64 binary from tendermintdevelopment/starport:arm64 for device images
- Added git to amd64 docker image
- Comment out Gaia's seeds in the systemd unit template for downstream chains

## `v0.12.0`

### Features

- Added GitHub CLI to gitpod environment for greater ease of use
- Added `starport build` command to build and install app binaries
- Improved the first-time experience for readers of the Starport readme and parts of the Starport Handbook
- Added `starport module create` command to scaffold custom modules
- Raspberry Pi now installs, builds, and serves the Vue UI
- Improved documentation for Raspberry Pi Device Images
- Added IBC and some other modules
- Added an option to configure server addresses under `servers` section in `config.yml`

### Fixes

- `--address-prefix` will always be translated to lowercase while scaffolding with `app` command
- HTTP API: accept strings in JSON and cast them to int and bool
- Update @tendermint/vue to `v0.1.7`
- Removed "Starport Pi"
- Removed Makefile from Downstream Pi
- Fixed Downstream Pi image GitHub Action
- Prevent duplicated fields with `type` command
- Fixed handling of protobuf profiler: prof_laddr -> pprof_laddr
- Fix an error, when a Stargate `serve` cmd doesn't start if a user doesn't have a relayer installed

## `v0.11.1`

### Features

- Published on Snapcraft

## `v0.11.0`

### Features

- Added experimental [Stargate](https://stargate.cosmos.network/) scaffolding option with `--sdk-version stargate` flag
  on `starport app` command
- Pi Image Generation for chains generated with Starport
- GitHub action with capture of binary artifacts for chains generated with Starport
- Gitpod: added guidelines and changed working directory into `docs`
- Updated web scaffold with an improved sign in, balance list and a simple wallet
- Added CRUD actions for scaffolded types: delete, update, and get

## `v0.0.10`

### Features

- Add ARM64 releases
- OS Image Generation for Raspberry Pi 3 and 4
- Added `version` command
- Added support for _validator_ configuration in _config.yml_.
- Starport can be launched on Gitpod
- Added `make clean`

### Fixes

- Compile with go1.15
- Running `starport add type...` multiple times no longer breaks the app
- Running `appcli tx app create-x` now checks for all required args
- Removed unused `--denom` flag from the `app` command. It previously has moved as a prop to the `config.yml`
  under `accounts` section
- Disabled proxy server in the Vue app (this was causing to some compatibility issues) and enabled CORS
  for `appcli rest-server` instead
- `type` command supports dashes in app names

## `v0.0.10-rc.3`

### Features

- Configure `genesis.json` through `genesis` field in `config.yml`
- Initialize git repository on `app` scaffolding
- Check Go and GOPATH when running `serve`

### Changes

- verbose is --verbose, not -v, in the cli
- Renamed `frontend` directory to `vue`
- Added first E2E tests (for `app` and `add wasm` subcommands)

### Fixes

- No longer crashes when git is initialized but doesn't have commits
- Failure to start the frontend doesn't prevent Starport from running
- Changes to `config.yml` trigger reinitialization of the app
- Running `starport add wasm` multiple times no longer breaks the app

## `v0.0.10-rc.X`

### Features

- Initialize with accounts defined `config.yml`
- `starport serve --verbose` shows detailed output from every process
- Custom address prefixes with `--address-prefix` flag
- Cosmos SDK Launchpad support
- Rebuild and reinitialize on file change

## `v0.0.9`

Initial release.
