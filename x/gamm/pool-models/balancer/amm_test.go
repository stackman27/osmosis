package balancer_test

import (
	"fmt"
	"math/rand"
	"testing"
	time "time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v7/osmoutils"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

// allowedErrRatio is the maximal multiplicative difference in either
// direction (positive or negative) that we accept to tolerate in
// unit tests for calcuating the number of shares to be returned by
// joining a pool. The comparison is done between Wolfram estimates and our AMM logic.
const allowedErrRatio = "0.0000001"

// This test sets up 2 asset pools, and then checks the spot price on them.
// It uses the pools spot price method, rather than the Gamm keepers spot price method.
func (suite *KeeperTestSuite) TestBalancerSpotPrice() {
	baseDenom := "uosmo"
	quoteDenom := "uion"

	tests := []struct {
		name                string
		baseDenomPoolInput  sdk.Coin
		quoteDenomPoolInput sdk.Coin
		expectError         bool
		expectedOutput      sdk.Dec
	}{
		{
			name:                "equal value",
			baseDenomPoolInput:  sdk.NewInt64Coin(baseDenom, 100),
			quoteDenomPoolInput: sdk.NewInt64Coin(quoteDenom, 100),
			expectError:         false,
			expectedOutput:      sdk.MustNewDecFromStr("1"),
		},
		{
			name:                "1:2 ratio",
			baseDenomPoolInput:  sdk.NewInt64Coin(baseDenom, 100),
			quoteDenomPoolInput: sdk.NewInt64Coin(quoteDenom, 200),
			expectError:         false,
			expectedOutput:      sdk.MustNewDecFromStr("0.500000000000000000"),
		},
		{
			name:                "2:1 ratio",
			baseDenomPoolInput:  sdk.NewInt64Coin(baseDenom, 200),
			quoteDenomPoolInput: sdk.NewInt64Coin(quoteDenom, 100),
			expectError:         false,
			expectedOutput:      sdk.MustNewDecFromStr("2.000000000000000000"),
		},
		{
			name:                "rounding after sigfig ratio",
			baseDenomPoolInput:  sdk.NewInt64Coin(baseDenom, 220),
			quoteDenomPoolInput: sdk.NewInt64Coin(quoteDenom, 115),
			expectError:         false,
			expectedOutput:      sdk.MustNewDecFromStr("1.913043480000000000"), // ans is 1.913043478260869565, rounded is 1.91304348
		},
		{
			name:                "check number of sig figs",
			baseDenomPoolInput:  sdk.NewInt64Coin(baseDenom, 100),
			quoteDenomPoolInput: sdk.NewInt64Coin(quoteDenom, 300),
			expectError:         false,
			expectedOutput:      sdk.MustNewDecFromStr("0.333333330000000000"),
		},
		{
			name:                "check number of sig figs high sizes",
			baseDenomPoolInput:  sdk.NewInt64Coin(baseDenom, 343569192534),
			quoteDenomPoolInput: sdk.NewCoin(quoteDenom, sdk.MustNewDecFromStr("186633424395479094888742").TruncateInt()),
			expectError:         false,
			expectedOutput:      sdk.MustNewDecFromStr("0.000000000001840877"),
		},
	}

	for _, tc := range tests {
		suite.SetupTest()

		poolId := suite.PrepareUni2PoolWithAssets(
			tc.baseDenomPoolInput,
			tc.quoteDenomPoolInput,
		)

		pool, err := suite.App.GAMMKeeper.GetPoolAndPoke(suite.Ctx, poolId)
		suite.Require().NoError(err, "test: %s", tc.name)
		balancerPool, isPool := pool.(*balancer.Pool)
		suite.Require().True(isPool, "test: %s", tc.name)

		spotPrice, err := balancerPool.SpotPrice(
			suite.Ctx,
			tc.baseDenomPoolInput.Denom,
			tc.quoteDenomPoolInput.Denom)

		if tc.expectError {
			suite.Require().Error(err, "test: %s", tc.name)
		} else {
			suite.Require().NoError(err, "test: %s", tc.name)
			suite.Require().True(spotPrice.Equal(tc.expectedOutput),
				"test: %s\nSpot price wrong, got %s, expected %s\n", tc.name,
				spotPrice, tc.expectedOutput)
		}
	}
}

// TestCalculateAmountOutAndIn_InverseRelationship tests that the same amount of token is guaranteed upon
// sequential operation of CalcInAmtGivenOut and CalcOutAmtGivenIn.
func TestCalculateAmountOutAndIn_InverseRelationship(t *testing.T) {
	type testcase struct {
		denomOut         string
		initialPoolOut   int64
		initialWeightOut int64
		initialCalcOut   int64

		denomIn         string
		initialPoolIn   int64
		initialWeightIn int64
	}

	// For every test case in testcases, apply a swap fee in swapFeeCases.
	testcases := []testcase{
		{
			denomOut:         "uosmo",
			initialPoolOut:   1_000_000_000_000,
			initialWeightOut: 100,
			initialCalcOut:   100,

			denomIn:         "ion",
			initialPoolIn:   1_000_000_000_000,
			initialWeightIn: 100,
		},
		{
			denomOut:         "uosmo",
			initialPoolOut:   1_000,
			initialWeightOut: 100,
			initialCalcOut:   100,

			denomIn:         "ion",
			initialPoolIn:   1_000_000,
			initialWeightIn: 100,
		},
		{
			denomOut:         "uosmo",
			initialPoolOut:   1_000,
			initialWeightOut: 100,
			initialCalcOut:   100,

			denomIn:         "ion",
			initialPoolIn:   1_000_000,
			initialWeightIn: 100,
		},
		{
			denomOut:         "uosmo",
			initialPoolOut:   1_000,
			initialWeightOut: 200,
			initialCalcOut:   100,

			denomIn:         "ion",
			initialPoolIn:   1_000_000,
			initialWeightIn: 50,
		},
		{
			denomOut:         "uosmo",
			initialPoolOut:   1_000_000,
			initialWeightOut: 200,
			initialCalcOut:   100000,

			denomIn:         "ion",
			initialPoolIn:   1_000_000_000,
			initialWeightIn: 50,
		},
	}

	swapFeeCases := []string{"0", "0.001", "0.1", "0.5", "0.99"}

	getTestCaseName := func(tc testcase, swapFeeCase string) string {
		return fmt.Sprintf("tokenOutInitial: %d, tokenInInitial: %d, initialOut: %d, swapFee: %s",
			tc.initialPoolOut,
			tc.initialPoolIn,
			tc.initialCalcOut,
			swapFeeCase,
		)
	}

	for _, tc := range testcases {
		for _, swapFee := range swapFeeCases {
			t.Run(getTestCaseName(tc, swapFee), func(t *testing.T) {
				ctx := createTestContext(t)

				poolAssetOut := balancer.PoolAsset{
					Token:  sdk.NewInt64Coin(tc.denomOut, tc.initialPoolOut),
					Weight: sdk.NewInt(tc.initialWeightOut),
				}

				poolAssetIn := balancer.PoolAsset{
					Token:  sdk.NewInt64Coin(tc.denomIn, tc.initialPoolIn),
					Weight: sdk.NewInt(tc.initialWeightIn),
				}

				swapFeeDec, err := sdk.NewDecFromStr(swapFee)
				require.NoError(t, err)

				exitFeeDec, err := sdk.NewDecFromStr("0")
				require.NoError(t, err)

				pool := createTestPool(t, swapFeeDec, exitFeeDec, poolAssetOut, poolAssetIn)
				require.NotNil(t, pool)

				initialOut := sdk.NewInt64Coin(poolAssetOut.Token.Denom, tc.initialCalcOut)
				initialOutCoins := sdk.NewCoins(initialOut)

				actualTokenIn, err := pool.CalcInAmtGivenOut(ctx, initialOutCoins, poolAssetIn.Token.Denom, swapFeeDec)
				require.NoError(t, err)

				inverseTokenOut, err := pool.CalcOutAmtGivenIn(ctx, sdk.NewCoins(actualTokenIn), poolAssetOut.Token.Denom, swapFeeDec)
				require.NoError(t, err)

				require.Equal(t, initialOut.Denom, inverseTokenOut.Denom)

				expected := initialOut.Amount.ToDec()
				actual := inverseTokenOut.Amount.ToDec()

				// allow a rounding error of up to 1 for this relation
				tol := sdk.NewDec(1)
				require.True(osmoutils.DecApproxEq(t, expected, actual, tol))
			})
		}
	}
}

func TestCalcSingleAssetInAndOut_InverseRelationship(t *testing.T) {
	type testcase struct {
		initialPoolOut   int64
		initialPoolIn    int64
		initialWeightOut int64
		tokenOut         int64
		initialWeightIn  int64
	}

	// For every test case in testcases, apply a swap fee in swapFeeCases.
	testcases := []testcase{
		{
			initialPoolOut:   1_000_000_000_000,
			tokenOut:         100,
			initialWeightOut: 100,
			initialWeightIn:  100,
		},
		{
			initialPoolOut:   1_000_000_000_000,
			tokenOut:         100,
			initialWeightOut: 50,
			initialWeightIn:  100,
		},
		{
			initialPoolOut:   1_000_000_000_000,
			tokenOut:         50,
			initialWeightOut: 100,
			initialWeightIn:  100,
		},
		{
			initialPoolOut:   1_000_000_000_000,
			tokenOut:         100,
			initialWeightOut: 100,
			initialWeightIn:  50,
		},
		{
			initialPoolOut:   1_000_000,
			tokenOut:         100,
			initialWeightOut: 100,
			initialWeightIn:  100,
		},
		{
			initialPoolOut:   2_351_333,
			tokenOut:         7,
			initialWeightOut: 148,
			initialWeightIn:  57,
		},
		{
			initialPoolOut:   1_000,
			tokenOut:         25,
			initialWeightOut: 100,
			initialWeightIn:  100,
		},
		{
			initialPoolOut:   1_000,
			tokenOut:         26,
			initialWeightOut: 100,
			initialWeightIn:  100,
		},
	}

	swapFeeCases := []string{"0", "0.001", "0.1", "0.5", "0.99"}

	getTestCaseName := func(tc testcase, swapFeeCase string) string {
		return fmt.Sprintf("initialPoolOut: %d, initialCalcOut: %d, initialWeightOut: %d, initialWeightIn: %d, swapFee: %s",
			tc.initialPoolOut,
			tc.tokenOut,
			tc.initialWeightOut,
			tc.initialWeightIn,
			swapFeeCase,
		)
	}

	for _, tc := range testcases {
		for _, swapFee := range swapFeeCases {
			t.Run(getTestCaseName(tc, swapFee), func(t *testing.T) {
				swapFeeDec, err := sdk.NewDecFromStr(swapFee)
				require.NoError(t, err)

				initialPoolBalanceOut := sdk.NewInt(tc.initialPoolOut)

				initialWeightOut := sdk.NewInt(tc.initialWeightOut)
				initialWeightIn := sdk.NewInt(tc.initialWeightIn)

				initialTotalShares := types.InitPoolSharesSupply.ToDec()
				initialCalcTokenOut := sdk.NewInt(tc.tokenOut)

				actualSharesOut := balancer.CalcPoolSharesOutGivenSingleAssetIn(
					initialPoolBalanceOut.ToDec(),
					initialWeightOut.ToDec().Quo(initialWeightOut.Add(initialWeightIn).ToDec()),
					initialTotalShares,
					initialCalcTokenOut.ToDec(),
					swapFeeDec,
				)

				inverseCalcTokenOut := balancer.CalcSingleAssetInGivenPoolSharesOut(
					initialPoolBalanceOut.Add(initialCalcTokenOut).ToDec(),
					initialWeightOut.ToDec().Quo(initialWeightOut.Add(initialWeightIn).ToDec()),
					initialTotalShares.Add(actualSharesOut),
					actualSharesOut,
					swapFeeDec,
				)

				tol := sdk.NewDec(1)
				require.True(osmoutils.DecApproxEq(t, initialCalcTokenOut.ToDec(), inverseCalcTokenOut, tol))
			})
		}
	}
}

func TestCalcJoinPoolShares(t *testing.T) {
	testCases := []struct {
		name         string
		swapFee      sdk.Dec
		poolAssets   []balancer.PoolAsset
		tokensIn     sdk.Coins
		expectErr    bool
		expectShares sdk.Int
		expectLiq    sdk.Coins
	}{
		{
			name:    "equal weights with zero swap fee",
			swapFee: sdk.MustNewDecFromStr("0"),
			poolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewInt64Coin("uosmo", 1_000_000_000_000),
					Weight: sdk.NewInt(100),
				},
				{
					Token:  sdk.NewInt64Coin("uatom", 1_000_000_000_000),
					Weight: sdk.NewInt(100),
				},
			},
			tokensIn: sdk.NewCoins(sdk.NewInt64Coin("uosmo", 50_000)),

			expectShares: sdk.NewInt(2_499_999_968_800),
			expectLiq:    sdk.NewCoins(sdk.NewInt64Coin("uosmo", 50_000)),
		},
		{
			// For uosmos
			// Expected output from Balancer paper (https://balancer.fi/whitepaper.pdf) using equation (25) on page 10:
			// P_issued = P_supply * ((1 + (A_t * swapFeeRatio  / B_t))^W_t - 1)
			// 625_000_000_000 = 100 * 10^18 * (( 1 + (25000 * (1 - (1 - 0.5)) / 1000000000000))^0.5 - 1) (without fee)
			// 621_874_980_000 = 100 * 10^18 *  (( 1 + (12500 * (1 - (1 - 0.5) * 0.01) / 1000000025000))^0.5 - 1)  (with fee)
			//
			// where:
			// 	P_supply = initial pool supply = 100 * 10^18 (set at pool creation, same for all new pools)
			//	A_t = amount of deposited asset = 50,000
			//	B_t = existing balance of deposited asset in the pool prior to deposit = 1,000,000,000,000
			//	W_t = normalized weight of deposited asset in pool = 0.5 (equally weighted two-asset pool)
			// 	swapFeeRatio = (1 - (1 - W_t) * swapFee)
			// Plugging all of this in, we get:
			// 	Full solution with fees: https://www.wolframalpha.com/input?i=100+*10%5E18*%28%281+%2B+%282500*%281+-+%281-0.5%29+*+0.01%29%2F1000000000000%29%29%5E0.5+-+1%29
			// 	Full Solution without fees: https://www.wolframalpha.com/input?i=100+*+10%5E18+*+%28%28+1+%2B+%2825000+*+%281+-+%281+-+0.5%29%29+%2F+1000000000000%29%29%5E0.5+-+1%29
			// 	Simplified:  P_issued = 625_000_000_000 * 2 + 621_874_980_000 * 2

			name:    "Multi-tokens In: unequal amounts, equal weights with 0.01 swap fee",
			swapFee: sdk.MustNewDecFromStr("0.01"),
			poolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewInt64Coin("uosmo", 1_000_000_000_000),
					Weight: sdk.NewInt(100),
				},
				{
					Token:  sdk.NewInt64Coin("uatom", 1_000_000_000_000),
					Weight: sdk.NewInt(100),
				},
			},
			tokensIn: sdk.NewCoins(
				sdk.NewInt64Coin("uosmo", 25_000),
				sdk.NewInt64Coin("uatom", 50_000),
			),

			expectShares: sdk.NewInt((625000000000 * 2) + (621_874_980_000 * 2)),
			expectLiq: sdk.NewCoins(
				sdk.NewInt64Coin("uosmo", 25_000),
				sdk.NewInt64Coin("uatom", 50_000),
			),
		},
		{
			// For uosmo:
			//
			// Expected output from Balancer paper (https://balancer.fi/whitepaper.pdf) using equation (25) on page 10:
			// P_issued = P_supply * ((1 + (A_t * swapFeeRatio  / B_t))^W_t - 1)
			//
			// 2_072_912_400_000_000 = 100 * 10^18 * (( 1 + (50,000 * (1 - (1 - 0.83) * 0.03) / 2_000_000_000))^0.83 - 1)
			//
			// where:
			// 	P_supply = initial pool supply = 100 * 10^18 (set at pool creation, same for all new pools)
			//	A_t = amount of deposited asset = 50,000
			//	B_t = existing balance of deposited asset in the pool prior to deposit = 2_000_000_000
			//	W_t = normalized weight of deposited asset in pool = 500 / 500 + 100 = 0.83
			// 	swapFeeRatio = (1 - (1 - W_t) * swapFee)
			// Plugging all of this in, we get:
			// 	Full solution: https://www.wolframalpha.com/input?i=100+*10%5E18*%28%281+%2B+%2850000*%281+-+%281-%28500+%2F+%28500+%2B+100%29%29%29+*+0.03%29%2F2000000000%29%29%5E%28500+%2F+%28500+%2B+100%29%29+-+1%29
			// 	Simplified:  P_issued = 2_072_912_400_000_000
			//
			//
			// For uatom:
			//
			// Expected output from Balancer paper (https://balancer.fi/whitepaper.pdf) using equation (25) on page 10:
			// P_issued = P_supply * ((1 + (A_t * swapFeeRatio  / B_t))^W_t - 1)
			//
			// 1_624_999_900_000 = 100 * 10^18 * (( 1 + (100_000 * (1 - (1 - 0.167) * 0.03) / 1_000_000_000_000))^0.167 - 1)
			//
			// where:
			// 	P_supply = initial pool supply = 100 * 10^18 (set at pool creation, same for all new pools)
			//	A_t = amount of deposited asset = 50,000
			//	B_t = existing balance of deposited asset in the pool prior to deposit = 1,000,000,000,000
			//	W_t = normalized weight of deposited asset in pool = 100 / 500 + 100 = 0.167
			// 	swapFeeRatio = (1 - (1 - W_t) * swapFee)
			// Plugging all of this in, we get:
			// 	Full solution: https://www.wolframalpha.com/input?i=100+*10%5E18*%28%281+%2B+%28100000*%281+-+%281-%28100+%2F+%28500+%2B+100%29%29%29+*+0.03%29%2F1000000000000%29%29%5E%28100+%2F+%28500+%2B+100%29%29+-+1%29
			// 	Simplified:  P_issued = 1_624_999_900_000
			name:    "Multi-tokens In: unequal amounts, with unequal weights with 0.03 swap fee",
			swapFee: sdk.MustNewDecFromStr("0.03"),
			poolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewInt64Coin("uosmo", 2_000_000_000_000),
					Weight: sdk.NewInt(500),
				},
				{
					Token:  sdk.NewInt64Coin("uatom", 1_000_000_000_000),
					Weight: sdk.NewInt(100),
				},
			},
			tokensIn: sdk.NewCoins(
				sdk.NewInt64Coin("uosmo", 50_000),
				sdk.NewInt64Coin("uatom", 100_000),
			),

			expectShares: sdk.NewInt(2072912400000000 + 1624999900000),
			expectLiq: sdk.NewCoins(
				sdk.NewInt64Coin("uosmo", 50_000),
				sdk.NewInt64Coin("uatom", 100_000),
			),
		},
		{
			name:    "equal weights with 0.001 swap fee",
			swapFee: sdk.MustNewDecFromStr("0.001"),
			poolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewInt64Coin("uosmo", 1_000_000_000_000),
					Weight: sdk.NewInt(100),
				},
				{
					Token:  sdk.NewInt64Coin("uatom", 1_000_000_000_000),
					Weight: sdk.NewInt(100),
				},
			},
			tokensIn:     sdk.NewCoins(sdk.NewInt64Coin("uosmo", 50_000)),
			expectShares: sdk.NewInt(2498749968800),
			expectLiq:    sdk.NewCoins(sdk.NewInt64Coin("uosmo", 50_000)),
		},
		{
			name:    "equal weights with 0.1 swap fee",
			swapFee: sdk.MustNewDecFromStr("0.1"),
			poolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewInt64Coin("uosmo", 1_000_000_000_000),
					Weight: sdk.NewInt(100),
				},
				{
					Token:  sdk.NewInt64Coin("uatom", 1_000_000_000_000),
					Weight: sdk.NewInt(100),
				},
			},
			tokensIn:     sdk.NewCoins(sdk.NewInt64Coin("uosmo", 50_000)),
			expectShares: sdk.NewInt(2374999971800),
			expectLiq:    sdk.NewCoins(sdk.NewInt64Coin("uosmo", 50_000)),
		},
		{
			name:    "equal weights with 0.99 swap fee",
			swapFee: sdk.MustNewDecFromStr("0.99"),
			poolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewInt64Coin("uosmo", 1_000_000_000_000),
					Weight: sdk.NewInt(100),
				},
				{
					Token:  sdk.NewInt64Coin("uatom", 1_000_000_000_000),
					Weight: sdk.NewInt(100),
				},
			},
			tokensIn:     sdk.NewCoins(sdk.NewInt64Coin("uosmo", 50_000)),
			expectShares: sdk.NewInt(1262499992100), // 1262500000000
			expectLiq:    sdk.NewCoins(sdk.NewInt64Coin("uosmo", 50_000)),
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			pool := createTestPool(t, tc.swapFee, sdk.MustNewDecFromStr("0"), tc.poolAssets...)

			shares, liquidity, err := pool.CalcJoinPoolShares(sdk.Context{}, tc.tokensIn, tc.swapFee)

			require.NoError(t, err)
			require.Equal(t, tc.expectLiq, liquidity)

			if tc.expectShares.Int64() == 0 {
				require.Equal(t, tc.expectShares, shares)
				return
			}

			assertExpectedSharesErrRatio(t, tc.expectShares, shares)

		})
	}
}

func TestRandomizedJoinPoolExitPoolInvariants(t *testing.T) {
	type testCase struct {
		initialTokensDenomIn  int64
		initialTokensDenomOut int64

		percentRatio int64

		numShares sdk.Int
	}

	const (
		denomOut = "denomOut"
		denomIn  = "denomIn"
	)

	now := time.Now().Unix()
	rng := rand.NewSource(now)
	t.Logf("Using random source of %d\n", now)

	// generate test case with randomized initial assets and join/exit ratio
	newCase := func() (tc *testCase) {
		tc = new(testCase)
		tc.initialTokensDenomIn = rng.Int63() % 1_000_000
		tc.initialTokensDenomOut = rng.Int63() % 1_000_000

		// 1%~100% of initial assets
		tc.percentRatio = rng.Int63()%100 + 1

		return tc
	}

	swapFeeDec, err := sdk.NewDecFromStr("0")
	require.NoError(t, err)

	exitFeeDec, err := sdk.NewDecFromStr("0")
	require.NoError(t, err)

	// create pool with randomized initial token amounts
	// and randomized ratio of join/exit
	createPool := func(tc *testCase) (pool *balancer.Pool) {
		poolAssetOut := balancer.PoolAsset{
			Token:  sdk.NewInt64Coin(denomOut, tc.initialTokensDenomOut),
			Weight: sdk.NewInt(5),
		}

		poolAssetIn := balancer.PoolAsset{
			Token:  sdk.NewInt64Coin(denomIn, tc.initialTokensDenomIn),
			Weight: sdk.NewInt(5),
		}

		pool = createTestPool(t, swapFeeDec, exitFeeDec, poolAssetOut, poolAssetIn).(*balancer.Pool)
		require.NotNil(t, pool)

		return pool
	}

	// joins with predetermined ratio
	joinPool := func(pool types.PoolI, tc *testCase) {
		tokensIn := sdk.Coins{
			sdk.NewInt64Coin(denomIn, tc.initialTokensDenomIn*tc.percentRatio/100),
			sdk.NewInt64Coin(denomOut, tc.initialTokensDenomOut*tc.percentRatio/100),
		}
		numShares, err := pool.JoinPool(sdk.Context{}, tokensIn, swapFeeDec)
		require.NoError(t, err)
		tc.numShares = numShares
	}

	// exits for same amount of shares minted
	exitPool := func(pool types.PoolI, tc *testCase) {
		_, err := pool.ExitPool(sdk.Context{}, tc.numShares, exitFeeDec)
		require.NoError(t, err)
	}

	invariantJoinExitInversePreserve := func(
		beforeCoins, afterCoins sdk.Coins,
		beforeShares, afterShares sdk.Int,
	) {
		// test token amount has been preserved
		require.True(t,
			!beforeCoins.IsAnyGT(afterCoins),
			"Coins has not been preserved before and after join-exit\nbefore:\t%s\nafter:\t%s",
			beforeCoins, afterCoins,
		)
		// test share amount has been preserved
		require.True(t,
			beforeShares.Equal(afterShares),
			"Shares has not been preserved before and after join-exit\nbefore:\t%s\nafter:\t%s",
			beforeShares, afterShares,
		)
	}

	testPoolInvariants := func() {
		tc := newCase()
		pool := createPool(tc)
		originalCoins, originalShares := pool.GetTotalPoolLiquidity(sdk.Context{}), pool.GetTotalShares()
		joinPool(pool, tc)
		exitPool(pool, tc)
		invariantJoinExitInversePreserve(
			originalCoins, pool.GetTotalPoolLiquidity(sdk.Context{}),
			originalShares, pool.GetTotalShares(),
		)
	}

	for i := 0; i < 1000; i++ {
		testPoolInvariants()
	}
}

func assertExpectedSharesErrRatio(t *testing.T, expectedShares, actualShares sdk.Int) {
	allowedErrRatioDec, err := sdk.NewDecFromStr(allowedErrRatio)
	require.NoError(t, err)

	errTolerance := osmoutils.ErrTolerance{
		MultiplicativeTolerance: allowedErrRatioDec,
	}

	require.Equal(
		t,
		0,
		errTolerance.Compare(expectedShares, actualShares),
		fmt.Sprintf("expectedShares: %d, actualShares: %d", expectedShares.Int64(), actualShares.Int64()))
}
