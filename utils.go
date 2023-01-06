package main

import "math/big"

func ToNativeBalance(balance *big.Int) float64 {
	tokensRatioBig := new(big.Float).Quo(new(big.Float).SetInt(balance), new(big.Float).SetFloat64(config.DenomCoefficient))
	value, _ := tokensRatioBig.Float64()
	return value
}

func BigIntToFloat(value *big.Int) float64 {
	f, _ := new(big.Float).SetInt(value).Float64()
	return f
}
