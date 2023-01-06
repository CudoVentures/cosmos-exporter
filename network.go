package main

import (
	"context"
	"errors"
	"math"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	tmrpc "github.com/tendermint/tendermint/rpc/client/http"
	"google.golang.org/grpc"
)

type CosmosNetwork struct {
	ChainId      string
	Denom        string
	grpcClient   *grpc.ClientConn
	bank         banktypes.QueryClient
	staking      stakingtypes.QueryClient
	distribution distributiontypes.QueryClient
	mint         minttypes.QueryClient
	slashing     slashingtypes.QueryClient
}

func NewCosmosNetwork(grpcAddr string) (*CosmosNetwork, error) {
	grpcClient, err := grpc.Dial(grpcAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return &CosmosNetwork{
		grpcClient:   grpcClient,
		bank:         banktypes.NewQueryClient(grpcClient),
		staking:      stakingtypes.NewQueryClient(grpcClient),
		distribution: distributiontypes.NewQueryClient(grpcClient),
		mint:         minttypes.NewQueryClient(grpcClient),
		slashing:     slashingtypes.NewQueryClient(grpcClient),
	}, nil
}

func (cn *CosmosNetwork) setChainID(rpcAddr string) error {
	client, err := tmrpc.New(config.TendermintRPC, "/websocket")
	if err != nil {
		log.Fatal().Err(err).Msg("Could not create Tendermint client")
	}

	status, err := client.Status(context.Background())
	if err != nil {
		return err
	}

	cn.ChainId = status.NodeInfo.Network
	return nil
}

func (cn *CosmosNetwork) setDenom() error {
	// if --denom and --denom-coefficient are both provided, use them
	// instead of fetching them via gRPC. Can be useful for networks like osmosis.
	if config.Denom != "" && config.DenomCoefficient != 0 {
		log.Info().
			Str("denom", config.Denom).
			Float64("coefficient", config.DenomCoefficient).
			Msg("Using provided denom and coefficient.")
		return nil
	}

	denoms, err := cn.bank.DenomsMetadata(
		context.Background(),
		&banktypes.QueryDenomsMetadataRequest{},
	)
	if err != nil {
		return err
	}

	if len(denoms.Metadatas) == 0 {
		return errors.New("No denom infos. Try running the binary with --denom and --denom-coefficient to set them manually.")
	}

	metadata := denoms.Metadatas[0] // always using the first one
	if config.Denom == "" {         // using display currency
		config.Denom = metadata.Display
	}

	for _, unit := range metadata.DenomUnits {
		log.Debug().
			Str("denom", unit.Denom).
			Uint32("exponent", unit.Exponent).
			Msg("Denom info")
		if unit.Denom == config.Denom {
			config.DenomCoefficient = math.Pow10(int(unit.Exponent))
			log.Info().
				Str("denom", config.Denom).
				Float64("coefficient", config.DenomCoefficient).
				Msg("Got denom info")
			return nil
		}
	}

	return errors.New("could not find denom info")
}
