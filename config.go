package main

import "github.com/spf13/cobra"

type Config struct {
	ConfigPath string

	Denom              string            `mapstructure:"denom"`
	ListenAddress      string            `mapstructure:"listen-address"`
	NodeAddress        string            `mapstructure:"node"`
	TendermintRPC      string            `mapstructure:"tendermint-rpc"`
	OsmosisAPI         string            `mapstructure:"osmosis-api"`
	EthRPC             string            `mapstructure:"eth-rpc"`
	EthTokenContract   string            `mapstructure:"eth-token-contract"`
	EthGravityContract string            `mapstructure:"eth-gravity-contract"`
	OptionalNetworks   map[string]string `mapstructure:"option-networks"`
	LogLevel           string            `mapstructure:"log-level"`
	Limit              uint64            `mapstructure:"limit"`

	Prefix                    string `mapstructure:"bech-prefix"`
	AccountPrefix             string `mapstructure:"bech-account-prefix"`
	AccountPubkeyPrefix       string `mapstructure:"bech-account-pubkey-prefix"`
	ValidatorPrefix           string `mapstructure:"bech-validator-prefix"`
	ValidatorPubkeyPrefix     string `mapstructure:"bech-validator-pubkey-prefix"`
	ConsensusNodePrefix       string `mapstructure:"bech-consensus-node-prefix"`
	ConsensusNodePubkeyPrefix string `mapstructure:"bech-consensus-node-pubkey-prefix"`

	ChainID          string `mapstructure:"chain-id"`
	ConstLabels      map[string]string
	DenomCoefficient float64 `mapstructure:"denom-coefficient"`

	TokenPrices []string `mapstructure:"token-prices"`
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) parseFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&c.ConfigPath, "config", "/var/lib/cosmos/config.json", "Config file path")
	cmd.PersistentFlags().StringVar(&c.Denom, "denom", "", "Cosmos coin denom")
	cmd.PersistentFlags().Float64Var(&c.DenomCoefficient, "denom-coefficient", 0, "Denom coefficient")
	cmd.PersistentFlags().StringVar(&c.ListenAddress, "listen-address", ":9300", "The address this exporter would listen on")
	cmd.PersistentFlags().StringVar(&c.NodeAddress, "node", "localhost:9090", "RPC node address")
	cmd.PersistentFlags().StringVar(&c.LogLevel, "log-level", "info", "Logging level")
	cmd.PersistentFlags().Uint64Var(&c.Limit, "limit", 1000, "Pagination limit for gRPC requests")
	cmd.PersistentFlags().StringVar(&c.TendermintRPC, "tendermint-rpc", "http://localhost:26657", "Tendermint RPC address")
	cmd.PersistentFlags().StringToStringVar(&c.OptionalNetworks, "optional-networks", nil, "Optional grpc networks")
	cmd.PersistentFlags().StringVar(&c.EthRPC, "eth-rpc", "http://localhost:8545", "Ethereum RPC address")
	cmd.PersistentFlags().StringVar(&c.EthTokenContract, "eth-token-contract", "", "Ethereum token contract")
	cmd.PersistentFlags().StringVar(&c.EthGravityContract, "eth-gravity-contract", "", "Ethereum gravity contract")
	cmd.PersistentFlags().StringSliceVar(&c.TokenPrices, "token-prices", nil, "List of CoinGecko token ids to retrieve current prices")

	// some networks, like Iris, have the different prefixes for address, validator and consensus node
	cmd.PersistentFlags().StringVar(&c.Prefix, "bech-prefix", "persistence", "Bech32 global prefix")
	cmd.PersistentFlags().StringVar(&c.AccountPrefix, "bech-account-prefix", "", "Bech32 account prefix")
	cmd.PersistentFlags().StringVar(&c.AccountPubkeyPrefix, "bech-account-pubkey-prefix", "", "Bech32 pubkey account prefix")
	cmd.PersistentFlags().StringVar(&c.ValidatorPrefix, "bech-validator-prefix", "", "Bech32 validator prefix")
	cmd.PersistentFlags().StringVar(&c.ValidatorPubkeyPrefix, "bech-validator-pubkey-prefix", "", "Bech32 pubkey validator prefix")
	cmd.PersistentFlags().StringVar(&c.ConsensusNodePrefix, "bech-consensus-node-prefix", "", "Bech32 consensus node prefix")
	cmd.PersistentFlags().StringVar(&c.ConsensusNodePubkeyPrefix, "bech-consensus-node-pubkey-prefix", "", "Bech32 pubkey consensus node prefix")
}
