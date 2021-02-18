package kzg

import (
	"fmt"
	"github.com/protolambda/go-kzg/bls"
	"math/rand"
	"testing"
)

func TestFFTSettings_reduceLeaves(t *testing.T) {
	fs := NewFFTSettings(4)

	var fromTreeReduction []bls.Big

	{
		// prepare some leaves
		leaves := [][]bls.Big{make([]bls.Big, 3), make([]bls.Big, 3), make([]bls.Big, 3), make([]bls.Big, 3)}
		leafIndices := [][]uint64{{1, 3}, {7, 8}, {9, 10}, {12, 13}}
		for i := 0; i < 4; i++ {
			fs.makeZeroPolyMulLeaf(leaves[i], leafIndices[i], 1)
		}

		dst := make([]bls.Big, 16, 16)
		scratch := make([]bls.Big, 16*3, 16*3)
		fs.reduceLeaves(scratch, dst, leaves)
		fromTreeReduction = dst[:2*4+1]
	}

	var fromDirect []bls.Big
	{
		dst := make([]bls.Big, 9, 9)
		indices := []uint64{1, 3, 7, 8, 9, 10, 12, 13}
		fs.makeZeroPolyMulLeaf(dst, indices, 1)
		fromDirect = dst
	}

	if len(fromDirect) != len(fromTreeReduction) {
		t.Fatal("length mismatch")
	}
	for i := 0; i < len(fromDirect); i++ {
		a, b := &fromDirect[i], &fromTreeReduction[i]
		if !bls.EqualBig(a, b) {
			t.Errorf("zero poly coeff %d is different. direct: %s, tree: %s", i, bls.BigStr(a), bls.BigStr(b))
		}
	}
	//debugBigs("zero poly (tree reduction)", fromTreeReduction)
	//debugBigs("zero poly (direct slow)", fromDirect)
}

func TestFFTSettings_reduceLeaves_parametrized(t *testing.T) {
	ratios := []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7}
	for scale := uint8(5); scale < 13; scale++ {
		t.Run(fmt.Sprintf("scale_%d", scale), func(t *testing.T) {
			for i, ratio := range ratios {
				t.Run(fmt.Sprintf("ratio_%.3f", ratio), func(t *testing.T) {
					seed := int64(1000*int(scale) + i)
					testReduceLeaves(scale, ratio, seed, t)
				})
			}
		})
	}
}

func testReduceLeaves(scale uint8, missingRatio float64, seed int64, t *testing.T) {
	fs := NewFFTSettings(scale)
	rng := rand.New(rand.NewSource(seed))
	pointCount := uint64(1) << scale
	missingCount := uint64(int(float64(pointCount) * missingRatio))

	// select the missing points
	missing := make([]uint64, pointCount, pointCount)
	for i := uint64(0); i < pointCount; i++ {
		missing[i] = i
	}
	rng.Shuffle(int(pointCount), func(i, j int) {
		missing[i], missing[j] = missing[j], missing[i]
	})
	missing = missing[:missingCount]

	// build the leaves
	pointsPerLeaf := uint64(63)
	leafCount := (missingCount + pointsPerLeaf - 1) / pointsPerLeaf
	leaves := make([][]bls.Big, leafCount, leafCount)
	for i := uint64(0); i < leafCount; i++ {
		start := i * pointsPerLeaf
		end := start + pointsPerLeaf
		if end > missingCount {
			end = missingCount
		}
		leafSize := end - start
		leaf := make([]bls.Big, leafSize+1, leafSize+1)
		indices := make([]uint64, leafSize, leafSize)
		for j := uint64(0); j < leafSize; j++ {
			indices[j] = missing[i*pointsPerLeaf+j]
		}
		fs.makeZeroPolyMulLeaf(leaf, indices, 1)
		leaves[i] = leaf
	}

	var fromTreeReduction []bls.Big

	{
		dst := make([]bls.Big, pointCount, pointCount)
		scratch := make([]bls.Big, pointCount*3, pointCount*3)
		fs.reduceLeaves(scratch, dst, leaves)
		fromTreeReduction = dst[:missingCount+1]
	}

	var fromDirect []bls.Big
	{
		dst := make([]bls.Big, missingCount+1, missingCount+1)
		fs.makeZeroPolyMulLeaf(dst, missing, fs.maxWidth/pointCount)
		fromDirect = dst
	}

	if len(fromDirect) != len(fromTreeReduction) {
		t.Fatal("length mismatch")
	}
	for i := 0; i < len(fromDirect); i++ {
		a, b := &fromDirect[i], &fromTreeReduction[i]
		if !bls.EqualBig(a, b) {
			t.Errorf("zero poly coeff %d is different. direct: %s, tree: %s", i, bls.BigStr(a), bls.BigStr(b))
		}
	}
	//debugBigs("zero poly (tree reduction)", fromTreeReduction)
	//debugBigs("zero poly (direct slow)", fromDirect)
}

func TestFFTSettings_ZeroPolyViaMultiplication_Python(t *testing.T) {
	fs := NewFFTSettings(4)

	exists := []bool{
		true, false, false, true,
		false, true, true, false,
		false, false, true, true,
		false, true, false, true,
	}
	var missingIndices []uint64
	for i, v := range exists {
		if !v {
			missingIndices = append(missingIndices, uint64(i))
		}
	}

	zeroEval, zeroPoly := fs.ZeroPolyViaMultiplication(missingIndices, uint64(len(exists)))

	// produced from python implementation, check it's exactly correct.
	expectedEval := []bls.Big{
		bls.ToBigNum("14588039771402811141309184187446855981335438080893546259057924963590957391610"),
		bls.ToBigNum("0"),
		bls.ToBigNum("0"),
		bls.ToBigNum("25282314916481609559521954076339682473205592322878335865825728051159013479404"),
		bls.ToBigNum("0"),
		bls.ToBigNum("9734294374130760583715448090686447252507379360428151468094660312309164340954"),
		bls.ToBigNum("46174059940592560972885266237294437331033682990367334129313899533918398326759"),
		bls.ToBigNum("0"),
		bls.ToBigNum("0"),
		bls.ToBigNum("0"),
		bls.ToBigNum("19800438175532257114364592658377771559959372488282871075375645402573059163542"),
		bls.ToBigNum("51600792158839053735333095261675086809225297622863271039022341472045686698468"),
		bls.ToBigNum("0"),
		bls.ToBigNum("30826826002656394595578928901119179510733149506206612508564887111870905005459"),
		bls.ToBigNum("0"),
		bls.ToBigNum("15554185610546001233857357261484634664347627247695018404843511652636226542123"),
	}
	for i := range zeroEval {
		if !bls.EqualBig(&expectedEval[i], &zeroEval[i]) {
			t.Errorf("at eval %d, expected: %s, got: %s", i, bls.BigStr(&expectedEval[i]), bls.BigStr(&zeroEval[i]))
		}
	}
	expectedPoly := []bls.Big{
		bls.ToBigNum("16624801632831727463500847948913128838752380757508923660793891075002624508302"),
		bls.ToBigNum("657600938076390596890050185197950209451778703253960215879283709261059409858"),
		bls.ToBigNum("3323305725086409462431021445881322078102454991213853012292210556336005043908"),
		bls.ToBigNum("28834633028751086963335689622252225417970192887686504864119125368464893106943"),
		bls.ToBigNum("13240145897582070561550318352041568075426755012978281815272419515864405431856"),
		bls.ToBigNum("29207346592337407428161116115756746704727357067233245260187026881605970530301"),
		bls.ToBigNum("26541641805327388562620144855073374836076680779273352463774100034531024896251"),
		bls.ToBigNum("1030314501662711061715476678702471496208942882800700611947185222402136833216"),
		bls.ToBigNum("1"),
		bls.ToBigNum("0"),
		bls.ToBigNum("0"),
		bls.ToBigNum("0"),
		bls.ToBigNum("0"),
		bls.ToBigNum("0"),
		bls.ToBigNum("0"),
		bls.ToBigNum("0"),
	}
	for i := range zeroPoly {
		if !bls.EqualBig(&expectedPoly[i], &zeroPoly[i]) {
			t.Errorf("at poly %d, expected: %s, got: %s", i, bls.BigStr(&expectedPoly[i]), bls.BigStr(&zeroPoly[i]))
		}
	}
}

func testZeroPoly(t *testing.T, scale uint8, seed int64) {
	fs := NewFFTSettings(scale)

	rng := rand.New(rand.NewSource(seed))

	exists := make([]bool, fs.maxWidth, fs.maxWidth)
	var missingIndices []uint64
	missingStr := ""
	for i := 0; i < len(exists); i++ {
		if rng.Intn(2) == 0 {
			exists[i] = true
		} else {
			missingIndices = append(missingIndices, uint64(i))
			missingStr += fmt.Sprintf(" %d", i)
		}
	}
	//t.Logf("missing indices:%s", missingStr)

	zeroEval, zeroPoly := fs.ZeroPolyViaMultiplication(missingIndices, uint64(len(exists)))

	//debugBigs("zero eval", zeroEval)
	//debugBigs("zero poly", zeroPoly)

	for i, v := range exists {
		if !v {
			var at bls.Big
			bls.CopyBigNum(&at, &fs.expandedRootsOfUnity[i])
			var out bls.Big
			bls.EvalPolyAt(&out, zeroPoly, &at)
			if !bls.EqualZero(&out) {
				t.Errorf("expected zero at %d, but got: %s", i, bls.BigStr(&out))
			}
		}
	}

	p, err := fs.FFT(zeroEval, true)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < len(zeroPoly); i++ {
		if !bls.EqualBig(&p[i], &zeroPoly[i]) {
			t.Errorf("fft not correct, i: %v, a: %s, b: %s", i, bls.BigStr(&p[i]), bls.BigStr(&zeroPoly[i]))
		}
	}
	for i := len(zeroPoly); i < len(p); i++ {
		if !bls.EqualZero(&p[i]) {
			t.Errorf("fft not correct, i: %v, a: %s, b: 0", i, bls.BigStr(&p[i]))
		}
	}
}

func TestFFTSettings_ZeroPolyViaMultiplication_Parametrized(t *testing.T) {
	for i := uint8(3); i < 12; i++ {
		t.Run(fmt.Sprintf("scale_%d", i), func(t *testing.T) {
			for j := int64(0); j < 3; j++ {
				t.Run(fmt.Sprintf("case_%d", j), func(t *testing.T) {
					testZeroPoly(t, i, int64(i)*1000+j)
				})
			}
		})
	}
}
