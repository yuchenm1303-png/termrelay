package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCangyuanCanvasPriceExpectationMatchesFrozenBillingSnapshot(t *testing.T) {
	snapshot := &CangyuanMediaPriceSnapshot{
		PricingVersion:      "admin-v1",
		Currency:            "USD",
		Tier:                "default",
		Units:               1,
		UnitSalePriceUSD:    "0.01",
		UnitProviderCostUSD: "0.005",
		VerifiedByAdmin:     true,
	}
	expected := CangyuanMediaPriceExpectation{TotalSaleUSD: "0.01000000", PricingVersion: "admin-v1"}
	require.True(t, cangyuanMediaPriceMatchesExpectation(snapshot, expected))
	require.False(t, cangyuanMediaPriceMatchesExpectation(snapshot,
		CangyuanMediaPriceExpectation{TotalSaleUSD: "0.009", PricingVersion: "admin-v1"}))
	require.False(t, cangyuanMediaPriceMatchesExpectation(snapshot,
		CangyuanMediaPriceExpectation{TotalSaleUSD: "0.01", PricingVersion: "admin-v2"}))
	require.False(t, cangyuanMediaPriceMatchesExpectation(snapshot,
		CangyuanMediaPriceExpectation{TotalSaleUSD: "0.01", PricingVersion: ""}))
	require.False(t, cangyuanMediaPriceMatchesExpectation(snapshot,
		CangyuanMediaPriceExpectation{TotalSaleUSD: "bad", PricingVersion: "admin-v1"}))
	snapshot.VerifiedByAdmin = false
	require.False(t, cangyuanMediaPriceMatchesExpectation(snapshot, expected))
	snapshot.VerifiedByAdmin = true
	snapshot.UnitSalePriceUSD = "0.02"
	require.False(t, cangyuanMediaPriceMatchesExpectation(snapshot, expected),
		"a changed approved sale price must never bypass the reserved snapshot guard")
	require.False(t, cangyuanMediaPriceMatchesExpectation(nil, expected))
}
