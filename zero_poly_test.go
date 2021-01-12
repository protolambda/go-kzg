package kate

import (
	"fmt"
	"math/rand"
	"testing"
)

func TestFFTSettings_reduceLeaves(t *testing.T) {
	fs := NewFFTSettings(4)

	var fromTreeReduction []Big

	{
		// prepare some leaves
		leaves := [][]Big{make([]Big, 3), make([]Big, 3), make([]Big, 3), make([]Big, 3)}
		leafIndices := [][]uint64{{1, 3}, {7, 8}, {9, 10}, {12, 13}}
		for i := 0; i < 4; i++ {
			fs.makeZeroPolyMulLeaf(leaves[i], leafIndices[i], 1)
		}

		dst := make([]Big, 16, 16)
		scratch := make([]Big, 32, 32)
		fs.reduceLeaves(scratch, dst, leaves)
		fromTreeReduction = dst[:2*4+1]
	}

	var fromDirect []Big
	{
		dst := make([]Big, 9, 9)
		indices := []uint64{1, 3, 7, 8, 9, 10, 12, 13}
		fs.makeZeroPolyMulLeaf(dst, indices, 1)
		fromDirect = dst
	}

	if len(fromDirect) != len(fromTreeReduction) {
		t.Fatal("length mismatch")
	}
	for i := 0; i < len(fromDirect); i++ {
		a, b := &fromDirect[i], &fromTreeReduction[i]
		if !equalBig(a, b) {
			t.Errorf("zero poly coeff %d is different. direct: %s, tree: %s", i, bigStr(a), bigStr(b))
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
	expectedEval := []Big{
		bigNumHelper("14588039771402811141309184187446855981335438080893546259057924963590957391610"),
		bigNumHelper("0"),
		bigNumHelper("0"),
		bigNumHelper("25282314916481609559521954076339682473205592322878335865825728051159013479404"),
		bigNumHelper("0"),
		bigNumHelper("9734294374130760583715448090686447252507379360428151468094660312309164340954"),
		bigNumHelper("46174059940592560972885266237294437331033682990367334129313899533918398326759"),
		bigNumHelper("0"),
		bigNumHelper("0"),
		bigNumHelper("0"),
		bigNumHelper("19800438175532257114364592658377771559959372488282871075375645402573059163542"),
		bigNumHelper("51600792158839053735333095261675086809225297622863271039022341472045686698468"),
		bigNumHelper("0"),
		bigNumHelper("30826826002656394595578928901119179510733149506206612508564887111870905005459"),
		bigNumHelper("0"),
		bigNumHelper("15554185610546001233857357261484634664347627247695018404843511652636226542123"),
	}
	for i := range zeroEval {
		if !equalBig(&expectedEval[i], &zeroEval[i]) {
			t.Errorf("at eval %d, expected: %s, got: %s", i, bigStr(&expectedEval[i]), bigStr(&zeroEval[i]))
		}
	}
	expectedPoly := []Big{
		bigNumHelper("16624801632831727463500847948913128838752380757508923660793891075002624508302"),
		bigNumHelper("657600938076390596890050185197950209451778703253960215879283709261059409858"),
		bigNumHelper("3323305725086409462431021445881322078102454991213853012292210556336005043908"),
		bigNumHelper("28834633028751086963335689622252225417970192887686504864119125368464893106943"),
		bigNumHelper("13240145897582070561550318352041568075426755012978281815272419515864405431856"),
		bigNumHelper("29207346592337407428161116115756746704727357067233245260187026881605970530301"),
		bigNumHelper("26541641805327388562620144855073374836076680779273352463774100034531024896251"),
		bigNumHelper("1030314501662711061715476678702471496208942882800700611947185222402136833216"),
		bigNumHelper("1"),
		bigNumHelper("0"),
		bigNumHelper("0"),
		bigNumHelper("0"),
		bigNumHelper("0"),
		bigNumHelper("0"),
		bigNumHelper("0"),
		bigNumHelper("0"),
	}
	for i := range zeroPoly {
		if !equalBig(&expectedPoly[i], &zeroPoly[i]) {
			t.Errorf("at poly %d, expected: %s, got: %s", i, bigStr(&expectedPoly[i]), bigStr(&zeroPoly[i]))
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
	t.Logf("missing indices:%s", missingStr)

	zeroEval, zeroPoly := fs.ZeroPolyViaMultiplication(missingIndices, uint64(len(exists)))

	for i, v := range exists {
		if i >= len(zeroEval) {
			continue
		}
		if !v && !equalZero(&zeroEval[i]) {
			t.Errorf("bad zero eval at: %d, got: %s", i, bigStr(&zeroEval[i]))
		}
	}

	debugBigs("zero eval", zeroEval)
	debugBigs("zero poly", zeroPoly)

	for i, v := range exists {
		if !v {
			var at Big
			CopyBigNum(&at, &fs.expandedRootsOfUnity[i])
			var out Big
			EvalPolyAt(&out, zeroPoly, &at)
			if !equalZero(&out) {
				t.Errorf("expected zero at %d, but got: %s", i, bigStr(&out))
			}
		}
	}

	p, err := fs.FFT(zeroEval, true)
	if err != nil {
		t.Fatal(err)
	}
	for i := range p {
		if !equalBig(&p[i], &zeroPoly[i]) {
			t.Errorf("fft not correct, i: %v, a: %s, b: %s", i, bigStr(&p[i]), bigStr(&zeroPoly[i]))
		}
	}
}

func TestFFTSettings_ZeroPolyViaMultiplication_Parametrized(t *testing.T) {
	for i := uint8(3); i < 10; i++ {
		t.Run(fmt.Sprintf("scale_%d", i), func(t *testing.T) {
			for j := int64(0); j < 3; j++ {
				t.Run(fmt.Sprintf("case_%d", j), func(t *testing.T) {
					testZeroPoly(t, i, j)
				})
			}
		})
	}
}
