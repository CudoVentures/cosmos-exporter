package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type generalMetrics struct {
	bondedTokens   prometheus.Gauge
	unbondedTokens prometheus.Gauge
	communityPool  *prometheus.GaugeVec
	totalSupply    *prometheus.GaugeVec
	tokenPrice     *prometheus.GaugeVec
}

func newGeneralMetrics(r *prometheus.Registry) *generalMetrics {
	bondedTokens := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name:        "cosmos_general_bonded_tokens",
			Help:        "Bonded tokens",
			ConstLabels: config.ConstLabels,
		},
	)

	unbondedTokens := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name:        "cosmos_general_not_bonded_tokens",
			Help:        "Not bonded tokens",
			ConstLabels: config.ConstLabels,
		},
	)

	communityPool := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "cosmos_general_community_pool",
			Help:        "Community pool",
			ConstLabels: config.ConstLabels,
		},
		[]string{"denom"},
	)

	totalSupply := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "cosmos_general_supply_total",
			Help:        "Total supply",
			ConstLabels: config.ConstLabels,
		},
		[]string{"denom"},
	)

	tokenPrice := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "cosmos_token_price",
			Help:        "Token Price",
			ConstLabels: config.ConstLabels,
		},
		[]string{"token", "currency"},
	)

	r.MustRegister(
		bondedTokens,
		unbondedTokens,
		communityPool,
		totalSupply,
		tokenPrice,
	)

	return &generalMetrics{
		bondedTokens,
		unbondedTokens,
		communityPool,
		totalSupply,
		tokenPrice,
	}
}

func (s *Server) GeneralHandler(w http.ResponseWriter, r *http.Request) {
	requestStart := time.Now()

	sublogger := log.With().
		Str("request-id", uuid.New().String()).
		Logger()

	registry := prometheus.NewRegistry()
	metrics := newGeneralMetrics(registry)

	var wg sync.WaitGroup

	go func() {
		defer wg.Done()
		sublogger.Debug().Msg("Started querying staking pool")
		queryStart := time.Now()

		response, err := s.Networks[0].staking.Pool(
			context.Background(),
			&stakingtypes.QueryPoolRequest{},
		)
		if err != nil {
			sublogger.Error().Err(err).Msg("Could not get staking pool")
			return
		}

		sublogger.Debug().
			Float64("request-time", time.Since(queryStart).Seconds()).
			Msg("Finished querying staking pool")

		metrics.bondedTokens.Set(BigIntToFloat(response.Pool.BondedTokens.BigInt()))
		metrics.unbondedTokens.Set(BigIntToFloat(response.Pool.NotBondedTokens.BigInt()))
	}()
	wg.Add(1)

	go func() {
		defer wg.Done()
		sublogger.Debug().Msg("Started querying distribution community pool")
		queryStart := time.Now()

		response, err := s.Networks[0].distribution.CommunityPool(
			context.Background(),
			&distributiontypes.QueryCommunityPoolRequest{},
		)
		if err != nil {
			sublogger.Error().Err(err).Msg("Could not get distribution community pool")
			return
		}

		sublogger.Debug().
			Float64("request-time", time.Since(queryStart).Seconds()).
			Msg("Finished querying distribution community pool")

		for _, coin := range response.Pool {
			if value, err := strconv.ParseFloat(coin.Amount.String(), 64); err != nil {
				sublogger.Error().
					Err(err).
					Msg("Could not get community pool coin")
			} else {
				metrics.communityPool.With(prometheus.Labels{
					"denom": config.Denom,
				}).Set(value / config.DenomCoefficient)
			}
		}
	}()
	wg.Add(1)

	go func() {
		defer wg.Done()
		sublogger.Debug().Msg("Started querying bank total supply")
		queryStart := time.Now()

		response, err := s.Networks[0].bank.TotalSupply(
			context.Background(),
			&banktypes.QueryTotalSupplyRequest{},
		)
		if err != nil {
			sublogger.Error().Err(err).Msg("Could not get bank total supply")
			return
		}

		sublogger.Debug().
			Float64("request-time", time.Since(queryStart).Seconds()).
			Msg("Finished querying bank total supply")

		for _, coin := range response.Supply {
			if value, err := strconv.ParseFloat(coin.Amount.String(), 64); err != nil {
				sublogger.Error().
					Str("Denom", coin.Denom).
					Err(err).
					Msg("Could not get total supply for coin")
			} else {

				metrics.totalSupply.With(prometheus.Labels{
					"denom": coin.Denom,
				}).Set(value)
			}
		}
	}()
	wg.Add(1)

	go func() {
		defer wg.Done()

		sublogger.Debug().Msg("Started querying token prices")
		queryStart := time.Now()

		for _, token := range config.TokenPrices {
			println(token)
			response, err := http.Get("https://api.coingecko.com/api/v3/coins/" + token)

			if err != nil {
				sublogger.Error().Err(err).Str("Token", token).Msg("Could not get token price")
				return
			}

			var coinGeckoResponse struct {
				MarketData struct {
					CurrentPrice struct {
						Usd float64 `json:"usd"`
						Gbp float64 `json:"gbp"`
					} `json:"current_price"`
				} `json:"market_data"`
			}
			responseBytes, err := io.ReadAll(response.Body)

			if err != nil {
				sublogger.Error().Err(err).Str("Token", token).Msg("Could not read response body")
				return
			}
			err = json.Unmarshal(responseBytes, &coinGeckoResponse)
			if err != nil {
				sublogger.Error().Err(err).Str("Token", token).Msg("Could not umarshal json")
				return
			}

			metrics.tokenPrice.With(prometheus.Labels{
				"token":    token,
				"currency": "usd",
			}).Set(coinGeckoResponse.MarketData.CurrentPrice.Usd)

			metrics.tokenPrice.With(prometheus.Labels{
				"token":    token,
				"currency": "gbp",
			}).Set(coinGeckoResponse.MarketData.CurrentPrice.Gbp)
		}
		sublogger.Debug().
			Float64("request-time", time.Since(queryStart).Seconds()).
			Msg("Finished querying token prices")
	}()
	wg.Add(1)

	wg.Wait()

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
	sublogger.Info().
		Str("method", "GET").
		Str("endpoint", "/metrics/general").
		Float64("request-time", time.Since(requestStart).Seconds()).
		Msg("Request processed")
}
