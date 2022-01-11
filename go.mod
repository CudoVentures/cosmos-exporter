module main

go 1.16

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1

replace google.golang.org/grpc => google.golang.org/grpc v1.33.2

replace github.com/cosmos/cosmos-sdk => github.com/CudoVentures/cosmos-sdk v0.0.0-20220111092913-4117cd46b688

require (
	github.com/cosmos/cosmos-sdk v0.44.3
	github.com/google/uuid v1.2.0
	github.com/prometheus/client_golang v1.11.0
	github.com/rs/zerolog v1.23.0
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.8.1
	github.com/tendermint/tendermint v0.34.14
	google.golang.org/grpc v1.42.0
)
