package main

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const AppName = "COSMOS_EXPORTER"

var (
	log            = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()
	config *Config = NewConfig()
)

func main() {
	config.parseFlags(rootCmd)
	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Could not start application")
	}
}

var rootCmd = &cobra.Command{
	Use:  "cosmos-exporter",
	Long: "Scrape the data about the validators set, specific validators or wallets in the Cosmos network.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {

		if err := viper.BindPFlags(cmd.Flags()); err != nil {
			log.Info().Err(err).Msg("There was an error reading from the config file")
		}

		viper.SetEnvPrefix(AppName)
		viper.AutomaticEnv()

		viper.SetConfigName("config")
		viper.AddConfigPath(".")
		if err := viper.ReadInConfig(); err != nil {
			log.Info().Err(err).Msg("There was an error reading from the config file")
		}

		if err := viper.Unmarshal(config); err != nil {
			log.Fatal().Err(err).Msg("There was an issue parsing the values into the config file")
		}

		return nil
	},
	Run: Execute,
}

func Execute(cmd *cobra.Command, args []string) {
	logLevel, err := zerolog.ParseLevel(config.LogLevel)
	if err != nil {
		log.Info().Err(err).Msg("Could not parse log level, setting default to INFO")
		logLevel, err = zerolog.ParseLevel("INFO")
		if err != nil {
			panic(err) // This should never happen as we are manually setting the logLevel to 0
		}
	}
	fmt.Printf("%+v\n", config)

	zerolog.SetGlobalLevel(logLevel)

	// TODO: Investigate whether this needs doing
	// config := sdk.GetConfig()
	// config.SetBech32PrefixForAccount(AccountPrefix, AccountPubkeyPrefix)
	// config.SetBech32PrefixForValidator(ValidatorPrefix, ValidatorPubkeyPrefix)
	// config.SetBech32PrefixForConsensusNode(ConsensusNodePrefix, ConsensusNodePubkeyPrefix)
	// config.Seal()

	//TODO: Create multiple networks parsed from the config
	network, err := NewCosmosNetwork(config.NodeAddress)
	if err != nil {
		panic(err)
	}

	network.setChainID(config.TendermintRPC)
	network.setDenom()

	server := NewServer(config.ListenAddress, network)

	log.Info().Str("listen_address", config.ListenAddress).Msgf("Listening on %s", config.ListenAddress)
	if err = server.Start(); err != nil {
		log.Fatal().Err(err).Msg("Could not start application")
	}
}
