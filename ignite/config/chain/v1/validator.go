package v1

import (
	"github.com/ignite/cli/v29/ignite/pkg/xyaml"
)

// Validator holds info related to validator settings.
type Validator struct {
	// Name is the name of the validator.
	Name string `yaml:"name" doc:"Name of the validator."`

	// Bonded is how much the validator has staked.
	Bonded string `yaml:"bonded" doc:"Amount staked by the validator."`

	// App overwrites appd's config/app.toml configs.
	App xyaml.Map `yaml:"app,omitempty" doc:"Overwrites the appd's config/app.toml configurations."`

	// Config overwrites appd's config/config.toml configs.
	Config xyaml.Map `yaml:"config,omitempty" doc:"Overwrites the appd's config/config.toml configurations."`

	// Client overwrites appd's config/client.toml configs.
	Client xyaml.Map `yaml:"client,omitempty" doc:"Overwrites the appd's config/client.toml configurations."`

	// Home overwrites default home directory used for the app.
	Home string `yaml:"home,omitempty" doc:"Overwrites the default home directory used for the application."`

	// Gentx overwrites appd's config/gentx.toml configs.
	Gentx *Gentx `yaml:"gentx,omitempty" doc:"Overwrites the appd's config/gentx.toml configurations."`
}

// Gentx holds info related to Gentx settings.
type Gentx struct {
	// Amount is the amount for the current Gentx.
	Amount string `yaml:"amount" doc:"Amount for the current Gentx."`

	// Moniker is the validator's (optional) moniker.
	Moniker string `yaml:"moniker" doc:"Optional moniker for the validator."`

	// KeyringBackend is keyring's backend.
	KeyringBackend string `yaml:"keyring-backend" doc:"Backend for the keyring."`

	// ChainID is the network chain ID.
	ChainID string `yaml:"chain-id" doc:"Network chain ID."`

	// CommissionMaxChangeRate is the maximum commission change rate percentage (per day).
	CommissionMaxChangeRate string `yaml:"commission-max-change-rate" doc:"Maximum commission change rate percentage per day."`

	// CommissionMaxRate is the maximum commission rate percentage.
	CommissionMaxRate string `yaml:"commission-max-rate" doc:"Maximum commission rate percentage (e.g., 0.01 = 1%)."`

	// CommissionRate is the initial commission rate percentage.
	CommissionRate string `yaml:"commission-rate" doc:"Initial commission rate percentage (e.g., 0.01 = 1%)."`

	// Details is the validator's (optional) details.
	Details string `yaml:"details" doc:"Optional details about the validator."`

	// SecurityContact is the validator's (optional) security contact email.
	SecurityContact string `yaml:"security-contact" doc:"Optional security contact email for the validator."`

	// Website is the validator's (optional) website.
	Website string `yaml:"website" doc:"Optional website for the validator."`

	// AccountNumber is the account number of the signing account (offline mode only).
	AccountNumber int `yaml:"account-number" doc:"Account number of the signing account (offline mode only)."`

	// BroadcastMode is the transaction broadcasting mode (sync|async|block) (default "sync").
	BroadcastMode string `yaml:"broadcast-mode" doc:"Transaction broadcasting mode (sync|async|block) (default is 'sync')."`

	// DryRun is a boolean determining whether to ignore the --gas flag and perform a simulation of a transaction.
	DryRun bool `yaml:"dry-run" doc:"Simulates the transaction without actually performing it, ignoring the --gas flag."`

	// FeeAccount is the fee account pays fees for the transaction instead of deducting from the signer.
	FeeAccount string `yaml:"fee-account" doc:"Account that pays the transaction fees instead of the signer."`

	// Fee is the fee to pay along with transaction; eg: 10uatom.
	Fee string `yaml:"fee" doc:"Fee to pay with the transaction (e.g.: 10uatom)."`

	// From is the name or address of private key with which to sign.
	From string `yaml:"from" doc:"Name or address of the private key used to sign the transaction."`

	// From is the gas limit to set per-transaction; set to "auto" to calculate sufficient gas automatically (default 200000).
	Gas string `yaml:"gas" doc:"Gas limit per transaction; set to 'auto' to calculate sufficient gas automatically (default is 200000)."`

	// GasAdjustment is the adjustment factor to be multiplied against the estimate returned by the tx simulation; if the gas limit is set manually this flag is ignored  (default 1).
	GasAdjustment string `yaml:"gas-adjustment" doc:"Factor to multiply against the estimated gas (default is 1)."`

	// GasPrices is the gas prices in decimal format to determine the transaction fee (e.g. 0.1uatom).
	GasPrices string `yaml:"gas-prices" doc:"Gas prices in decimal format to determine the transaction fee (e.g., 0.1uatom)."`

	// GenerateOnly is a boolean determining whether to build an unsigned transaction and write it to STDOUT.
	GenerateOnly bool `yaml:"generate-only" doc:"Creates an unsigned transaction and writes it to STDOUT."`

	// Identity is the (optional) identity signature (ex. UPort or Keybase).
	Identity string `yaml:"identity" doc:"Identity signature (e.g., UPort or Keybase)."`

	// IP is the node's public IP (default "192.168.1.64").
	IP string `yaml:"ip" doc:"Node's public IP address (default is '192.168.1.64')."`

	// KeyringDir is the client Keyring directory; if omitted, the default 'home' directory will be used.
	KeyringDir string `yaml:"keyring-dir" doc:"Directory for the client keyring; defaults to the 'home' directory if omitted."`

	// Ledger is a boolean determining whether to use a connected Ledger device.
	Ledger bool `yaml:"ledger" doc:"Uses a connected Ledger device if true."`

	// KeyringDir is the minimum self delegation required on the validator.
	MinSelfDelegation string `yaml:"min-self-delegation" doc:"Minimum self-delegation required for the validator."`

	// Node is <host>:<port> to tendermint rpc interface for this chain (default "tcp://localhost:26657").
	Node string `yaml:"node" doc:"<host>:<port> for the Tendermint RPC interface (default 'tcp://localhost:26657')"`

	// NodeID is the node's NodeID.
	NodeID string `yaml:"node-id" doc:"Node's NodeID"`

	// Note is the note to add a description to the transaction (previously --memo).
	Note string `yaml:"note" doc:"Adds a description to the transaction (formerly --memo)."`

	// Offline is a boolean determining the offline mode (does not allow any online functionality).
	Offline bool `yaml:"offline" doc:"Operates in offline mode, disallowing any online functionality."`

	// Output is the output format (text|json) (default "json").
	Output string `yaml:"output" doc:"Output format (text|json) (default 'json')."`

	// OutputDocument writes the genesis transaction JSON document to the given file instead of the default location.
	OutputDocument string `yaml:"output-document" doc:"Writes the genesis transaction JSON document to the specified file instead of the default location."`

	// PubKey is the validator's Protobuf JSON encoded public key.
	PubKey string `yaml:"pubkey" doc:"Protobuf JSON encoded public key of the validator."`

	// Sequence is the sequence number of the signing account (offline mode only).
	Sequence uint `yaml:"sequence" doc:"Sequence number of the signing account (offline mode only)."`

	// SignMode is the choose sign mode (direct|amino-json), this is an advanced feature.
	SignMode string `yaml:"sign-mode" doc:"Chooses sign mode (direct|amino-json), an advanced feature."`

	// TimeoutHeight sets a block timeout height to prevent the tx from being committed past a certain height.
	TimeoutHeight uint `yaml:"timeout-height" doc:"Sets a block timeout height to prevent the transaction from being committed past a certain height."`
}
