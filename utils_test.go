package main

import (
	"math/big"
	"testing"
)

func TestToNativeBalance(t *testing.T) {
	value := big.NewInt(int64(10000000000))
	expected := 1.0

	formattedValue := ToNativeBalance(value)

	if formattedValue != expected {
		t.Fatalf(`Expected %f, got %f`, expected, (float64)(formattedValue))
	}
}
