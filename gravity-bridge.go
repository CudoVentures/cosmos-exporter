package main

import (
	"context"
	"net/http"
	"sync"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type gravityMetrics struct {
	cudoOrchBalance     *prometheus.GaugeVec
	ethOrchBalance      *prometheus.GaugeVec
	ethERC20OrchBalance *prometheus.GaugeVec
	ethContractBalance  *prometheus.GaugeVec
}

func NewGravityMetrics(r *prometheus.Registry) *gravityMetrics {
	cudoOrchBalance := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "gravity_cudos_orchestrator_balance",
			Help:        "Balance of the cudos orchestrator wallet",
			ConstLabels: config.ConstLabels,
		},
		[]string{"cudos_orchestrator_address", "ethereum_orchestrator_address"},
	)

	ethOrchBalance := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "gravity_ethereum_orchestrator_balance",
			Help:        "Balance of the ethereum orchestrator wallet",
			ConstLabels: config.ConstLabels,
		},
		[]string{"cudos_orchestrator_address", "ethereum_orchestrator_address"},
	)

	ethERC20OrchBalance := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "gravity_ethereum_orchestrator_erc20_balance",
			Help:        "ERC20 balance of the ethereum orchestrator wallet",
			ConstLabels: config.ConstLabels,
		},
		[]string{"cudos_orchestrator_address", "ethereum_orchestrator_address"},
	)

	ethContractBalance := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "gravity_ethereum_contract_balance",
			Help:        "Balance of the ethereum gravity contract",
			ConstLabels: config.ConstLabels,
		},
		[]string{},
	)

	r.MustRegister(
		cudoOrchBalance,
		ethOrchBalance,
		ethERC20OrchBalance,
		ethContractBalance,
	)

	return &gravityMetrics{
		cudoOrchBalance,
		ethOrchBalance,
		ethERC20OrchBalance,
		ethContractBalance,
	}
}

func (s *Server) GravityBridgeHandler(w http.ResponseWriter, r *http.Request) {
	requestStart := time.Now()

	sublogger := log.With().
		Str("request_id", uuid.New().String()).
		Logger()

	cudosOrchestratorAddressParam := r.URL.Query().Get("cudos_orchestrator_address")
	cudosOrchestratorAddress, err := sdk.AccAddressFromBech32(cudosOrchestratorAddressParam)
	if err != nil {
		sublogger.Error().
			Str("cudos_orchestrator_address", cudosOrchestratorAddressParam).
			Err(err).
			Msg("Could not get cudos orchestrator address")
		return
	}

	ethConn, err := ethclient.Dial(config.EthRPC)
	if err != nil {
		sublogger.Error().
			Err(err).
			Msg("Could not connect to Ethereum node")
		return
	}
	ethOrchestratorAddressParam := r.URL.Query().Get("ethereum_orchestrator_address")
	ethOrchestratorAddress := common.HexToAddress(ethOrchestratorAddressParam)

	registry := prometheus.NewRegistry()
	metrics := NewGravityMetrics(registry)

	var wg sync.WaitGroup

	go func() {
		defer wg.Done()
		sublogger.Debug().
			Str("cudos_orchestrator_address", cudosOrchestratorAddress.String()).
			Msg("Started querying orchestrator wallet balance")
		queryStart := time.Now()

		bankRes, err := s.Networks[0].bank.AllBalances(
			context.Background(),
			&banktypes.QueryAllBalancesRequest{Address: cudosOrchestratorAddress.String()},
		)
		if err != nil {
			sublogger.Error().
				Str("cudos_orchestrator_address", cudosOrchestratorAddress.String()).
				Err(err).
				Msg("Could not get orchestrator balance")
			return
		}

		sublogger.Debug().
			Str("cudos_orchestrator_address", cudosOrchestratorAddress.String()).
			Float64("request_time", time.Since(queryStart).Seconds()).
			Msg("Finished querying orchestrator balance")

		for _, balance := range bankRes.Balances {
			tokensRatio := ToNativeBalance(balance.Amount.BigInt())
			metrics.cudoOrchBalance.With(prometheus.Labels{
				"cudos_orchestrator_address":    cudosOrchestratorAddress.String(),
				"ethereum_orchestrator_address": ethOrchestratorAddress.String(),
			}).Set(tokensRatio)

		}
	}()
	wg.Add(1)

	go func() {
		defer wg.Done()
		sublogger.Debug().
			Str("ethereum_orchestrator_address", ethOrchestratorAddress.String()).
			Msg("Started querying ethereum wallet balance")
		queryStart := time.Now()

		ethBal, err := ethConn.BalanceAt(context.Background(), ethOrchestratorAddress, nil)
		if err != nil {
			sublogger.Error().
				Str("ethereum_orchestrator_address", ethOrchestratorAddress.String()).
				Err(err).
				Msg("Could not get ethereum balance")
			return
		}

		sublogger.Debug().
			Str("ethereum_orchestrator_address", ethOrchestratorAddress.String()).
			Float64("request_time", time.Since(queryStart).Seconds()).
			Uint64("balance", ethBal.Uint64()).
			Msg("Finished querying balance")

		tokensRatio := ToNativeBalance(ethBal)

		metrics.ethOrchBalance.With(prometheus.Labels{
			"cudos_orchestrator_address":    cudosOrchestratorAddress.String(),
			"ethereum_orchestrator_address": ethOrchestratorAddress.String(),
		}).Set(tokensRatio)
	}()
	wg.Add(1)

	go func() {
		defer wg.Done()
		sublogger.Debug().
			Str("ethereum_orchestrator_address", ethOrchestratorAddress.String()).
			Msg("Started querying ethereum erc20 wallet balance")
		queryStart := time.Now()

		ethTokenAddress := common.HexToAddress(config.EthTokenContract)
		instance, err := NewMain(ethTokenAddress, ethConn)

		if err != nil {
			sublogger.Error().
				Err(err).
				Msg("Could not retrieve token contract")
			return
		}

		ethBal, err := instance.BalanceOf(&bind.CallOpts{}, ethOrchestratorAddress)
		if err != nil {
			sublogger.Error().
				Str("ethereum_token_address", ethTokenAddress.String()).
				Err(err).
				Msg("Could not get ethereum token balance")
			return
		}

		sublogger.Debug().
			Str("ethereum_orchestrator_address", ethOrchestratorAddress.String()).
			Float64("request_time", time.Since(queryStart).Seconds()).
			Uint64("balance", ethBal.Uint64()).
			Msg("Finished querying erc20 balance")

		tokensRatio := ToNativeBalance(ethBal)

		metrics.ethERC20OrchBalance.With(prometheus.Labels{
			"cudos_orchestrator_address":    cudosOrchestratorAddress.String(),
			"ethereum_orchestrator_address": ethOrchestratorAddress.String(),
		}).Set(tokensRatio)
	}()
	wg.Add(1)

	ethTokenAddress := common.HexToAddress(config.EthTokenContract)
	instance, err := NewMain(ethTokenAddress, ethConn)
	if err != nil {
		sublogger.Error().
			Err(err).
			Msg("Could not retrieve token contract")
		return
	}

	go func() {
		defer wg.Done()
		sublogger.Debug().
			Str("ethereum_gravity_contract", ethTokenAddress.String()).
			Msg("Started querying gravity ethereum gravity contract balance")
		queryStart := time.Now()
		gravityAddress := common.HexToAddress(config.EthGravityContract)
		ethBal, err := instance.BalanceOf(&bind.CallOpts{}, gravityAddress)
		if err != nil {
			sublogger.Error().
				Str("ethereum_token_address", ethTokenAddress.String()).
				Err(err).
				Msg("Could not get ethereum token balance")
			return
		}

		sublogger.Debug().
			Str("ethereum_gravity_contract", ethTokenAddress.String()).
			Float64("request_time", time.Since(queryStart).Seconds()).
			Msg("Finished querying gravity ethereum contract token balance")

		tokensRatio := ToNativeBalance(ethBal)
		metrics.ethContractBalance.With(nil).Set(tokensRatio)
	}()
	wg.Add(1)

	wg.Wait()

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
	sublogger.Info().
		Str("method", "GET").
		Str("endpoint", "/metrics/gravity-bridge/wallet?cudos_orchestrator_address="+cudosOrchestratorAddress.String()+"&ethereum_orchestrator_address="+ethOrchestratorAddress.String()).
		Float64("request_time", time.Since(requestStart).Seconds()).
		Msg("Request processed")
}
