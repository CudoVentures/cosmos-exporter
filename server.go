package main

import "net/http"

type Server struct {
	ListenAddr string
	Networks   []*CosmosNetwork
}

func NewServer(listenAddr string, networks ...*CosmosNetwork) *Server {
	return &Server{
		ListenAddr: listenAddr,
		Networks:   networks,
	}
}

func (s *Server) Start() error {
	http.HandleFunc("/metrics/wallet", s.WalletHandler)

	http.HandleFunc("/metrics/validator", s.ValidatorHandler)

	http.HandleFunc("/metrics/validators", s.ValidatorsHandler)

	http.HandleFunc("/metrics/params", s.ParamsHandler)

	http.HandleFunc("/metrics/general", s.GeneralHandler)

	http.HandleFunc("/metrics/gravity-bridge", s.GravityBridgeHandler)

	http.HandleFunc("/metrics/status", s.StatusHandler)

	http.HandleFunc("/metrics/osmosis", s.OsmosisHandler)

	return http.ListenAndServe(s.ListenAddr, nil)
}
