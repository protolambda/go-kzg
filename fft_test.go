package go_verkle

import (
	"fmt"
	"math/rand"
	"testing"
)

func TestFFTRoundtrip(t *testing.T) {
	fs := NewFFTSettings(4)
	data := make([]Big, fs.width, fs.width)
	for i := uint64(0); i < fs.width; i++ {
		asBig(&data[i], i)
	}
	coeffs, err := fs.FFT(data, false)
	if err != nil {
		t.Fatal(err)
	}
	res, err := fs.FFT(coeffs, true)
	if err != nil {
		t.Fatal(err)
	}
	for i := range res {
		if got, expected := &res[i], &data[i]; !equalBig(got, expected) {
			t.Errorf("difference: %d: got: %s  expected: %s", i, bigStr(got), bigStr(expected))
		}
	}
	t.Log("zero", bigStr(&ZERO))
	t.Log("zero", bigStr(&ONE))
}

func TestInvFFT(t *testing.T) {
	fs := NewFFTSettings(4)
	data := make([]Big, fs.width, fs.width)
	for i := uint64(0); i < fs.width; i++ {
		asBig(&data[i], i)
	}
	debugBigs("input data", data)
	res, err := fs.FFT(data, true)
	if err != nil {
		t.Fatal(err)
	}
	debugBigs("result", res)
	bigNumHelper := func(v string) (out Big) {
		bigNum(&out, v)
		return
	}
	expected := []Big{
		bigNumHelper("26217937587563095239723870254092982918845276250263818911301829349969290592264"),
		bigNumHelper("8864682297557565932517422087434646388650579555464978742404310425307854971414"),
		bigNumHelper("42397926345479656069499145686287671633657326275595206970800938736622240188372"),
		bigNumHelper("20829590431265536861492157516271359172322844207237904580180981500923098586768"),
		bigNumHelper("26217937587563095241456442667129809078233411015607690300436955584351971573760"),
		bigNumHelper("40905488090558605688319636812215252217941835718478251840326926365086504505065"),
		bigNumHelper("42397926345479656066034000860214019314881056744907464192530686267856878225364"),
		bigNumHelper("28940579956850634752414611731231234796717032005329840446009750351940536963695"),
		bigNumHelper("26217937587563095239723870254092982918845276250263818911301829349969290592256"),
		bigNumHelper("23495295218275555727033128776954731040973520495197797376593908347998044220817"),
		bigNumHelper("10037948829646534413413739647971946522809495755620173630072972432081702959148"),
		bigNumHelper("11530387084567584791128103695970713619748716782049385982276732334852076679447"),
		bigNumHelper("26217937587563095237991297841056156759457141484919947522166703115586609610752"),
		bigNumHelper("31606284743860653617955582991914606665367708293289733242422677199015482597744"),
		bigNumHelper("10037948829646534409948594821898294204033226224932430851802719963316340996140"),
		bigNumHelper("43571192877568624546930318420751319449039972945062659080199348274630726213098"),
	}
	for i := range res {
		if got := &res[i]; !equalBig(got, &expected[i]) {
			t.Errorf("difference: %d: got: %s  expected: %s", i, bigStr(got), bigStr(&expected[i]))
		}
	}
}

func TestDASFFTExtension(t *testing.T) {
	fs := NewFFTSettings(4)
	half := fs.width / 2
	data := make([]Big, half, half)
	for i := uint64(0); i < half; i++ {
		asBig(&data[i], i)
	}
	debugBigs("even data", data)
	fs.DASFFTExtension(data)
	debugBigs("odd data", data)
	bigNumHelper := func(v string) (out Big) {
		bigNum(&out, v)
		return
	}
	expected := []Big{
		bigNumHelper("35517140934261047308355351661356802312031268910108466120070952281657631518077"),
		bigNumHelper("46293835246856164064818777137000049805076132996160294782312647979750015529053"),
		bigNumHelper("16918734240865143167627244020755511206883014059731428924262453949515587703435"),
		bigNumHelper("11473449502290064142245761066479007451139502549599385854846611945573094960557"),
		bigNumHelper("16918734240865143167627244020755511206883014059731428924262453949515587703435"),
		bigNumHelper("46293835246856164064818777137000049805076132996160294782312647979750015529053"),
		bigNumHelper("35517140934261047308355351661356802312031268910108466120070952281657631518077"),
		bigNumHelper("810630354249988693942455328040129251641875520510785782275914432334760276393"),
	}
	for i := range data {
		if got := &data[i]; !equalBig(got, &expected[i]) {
			t.Errorf("difference: %d: got: %s  expected: %s", i, bigStr(got), bigStr(&expected[i]))
		}
	}
}

func TestParametrizedDASFFTExtension(t *testing.T) {
	testScale := func(seed int64, scale uint8, t *testing.T) {
		fs := NewFFTSettings(scale)
		evenData := make([]Big, fs.width/2, fs.width/2)
		rng := rand.New(rand.NewSource(seed))
		for i := uint64(0); i < fs.width/2; i++ {
			asBig(&evenData[i], rng.Uint64()) // TODO could be a full random F_r instead of uint64
		}
		debugBigs("input data", evenData)
		// we don't want to modify the original input, and the inner function would modify it in-place, so make a copy.
		oddData := make([]Big, fs.width/2, fs.width/2)
		for i := 0; i < len(oddData); i++ {
			CopyBigNum(&oddData[i], &evenData[i])
		}
		fs.DASFFTExtension(oddData)
		debugBigs("output data", oddData)

		// reconstruct data
		data := make([]Big, fs.width, fs.width)
		for i := uint64(0); i < fs.width; i += 2 {
			CopyBigNum(&data[i], &evenData[i>>1])
			CopyBigNum(&data[i+1], &oddData[i>>1])
		}
		debugBigs("reconstructed data", data)
		// get coefficients of reconstructed data with inverse FFT
		coeffs, err := fs.FFT(data, true)
		if err != nil {
			t.Fatal(err)
		}
		debugBigs("coeffs data", coeffs)
		// second half of all coefficients should be zero
		for i := fs.width / 2; i < fs.width; i++ {
			if !equalZero(&coeffs[i]) {
				t.Errorf("expected zero coefficient on index %d", i)
			}
		}
	}
	for scale := uint8(4); scale < 10; scale++ {
		for i := int64(0); i < 4; i++ {
			t.Run(fmt.Sprintf("scale_%d_i_%d", scale, i), func(t *testing.T) {
				testScale(i, scale, t)
			})
		}
	}
}

func bigRange(start uint64, end uint64) []Big {
	l := end - start
	out := make([]Big, l, l)
	for i := uint64(0); i < l; i++ {
		asBig(&out[i], start+i)
	}
	return out
}

func TestErasureCodeRecoverSimple(t *testing.T) {
	// Create some random data, with padding...
	fs := NewFFTSettings(5)
	data := make([]Big, fs.width, fs.width)
	for i := uint64(0); i < fs.width/2; i++ {
		asBig(&data[i], i)
	}
	for i := fs.width / 2; i < fs.width; i++ {
		data[i] = ZERO
	}
	debugBigs("data", data)
	// Get coefficients for polynomial SLOW_INDICES
	coeffs, err := fs.FFT(data, false)
	if err != nil {
		t.Fatal(err)
	}
	debugBigs("coeffs", coeffs)

	// copy over the 2nd half, leave the first half as nils
	subset := make([]*Big, fs.width, fs.width)
	half := fs.width / 2
	for i := half; i < fs.width; i++ {
		subset[i] = &coeffs[i]
	}

	debugBigPtrs("subset", subset)
	recovered, err := fs.ErasureCodeRecover(subset)
	if err != nil {
		t.Fatal(err)
	}
	debugBigs("recovered", recovered)
	for i := range recovered {
		if got := &recovered[i]; !equalBig(got, &coeffs[i]) {
			t.Errorf("recovery at index %d got %s but expected %s", i, bigStr(got), bigStr(&coeffs[i]))
		}
	}
	// And recover the original data for good measure
	back, err := fs.FFT(recovered, true)
	if err != nil {
		t.Fatal(err)
	}
	debugBigs("back", back)
	for i := uint64(0); i < half; i++ {
		if got := &back[i]; !equalBig(got, &data[i]) {
			t.Errorf("data at index %d got %s but expected %s", i, bigStr(got), bigStr(&data[i]))
		}
	}
	for i := half; i < fs.width; i++ {
		if got := &back[i]; !equalZero(got) {
			t.Errorf("expected zero padding in index %d", i)
		}
	}
}

func TestErasureCodeRecover(t *testing.T) {
	// Create some random data, with padding...
	fs := NewFFTSettings(7)
	data := make([]Big, fs.width, fs.width)
	for i := uint64(0); i < fs.width/2; i++ {
		asBig(&data[i], i)
	}
	for i := fs.width / 2; i < fs.width; i++ {
		data[i] = ZERO
	}
	debugBigs("data", data)
	// Get coefficients for polynomial SLOW_INDICES
	coeffs, err := fs.FFT(data, false)
	if err != nil {
		t.Fatal(err)
	}
	debugBigs("coeffs", coeffs)

	// Util to pick a random subnet of the values
	randomSubset := func(known uint64, rngSeed uint64) []*Big {
		withMissingValues := make([]*Big, fs.width, fs.width)
		for i := range coeffs {
			withMissingValues[i] = &coeffs[i]
		}
		rng := rand.New(rand.NewSource(int64(rngSeed)))
		missing := fs.width - known
		pruned := rng.Perm(int(fs.width))[:missing]
		for _, i := range pruned {
			withMissingValues[i] = nil
		}
		return withMissingValues
	}

	// Try different amounts of known indices, and try it in multiple random ways
	var lastKnown uint64 = 0
	for knownRatio := 0.7; knownRatio < 1.0; knownRatio += 0.05 {
		known := uint64(float64(fs.width) * knownRatio)
		if known == lastKnown {
			continue
		}
		lastKnown = known
		for i := 0; i < 3; i++ {
			t.Run(fmt.Sprintf("random_subset_%d_known_%d", i, known), func(t *testing.T) {
				subset := randomSubset(known, uint64(i))

				debugBigPtrs("subset", subset)
				recovered, err := fs.ErasureCodeRecover(subset)
				if err != nil {
					t.Fatal(err)
				}
				debugBigs("recovered", recovered)
				for i := range recovered {
					if got := &recovered[i]; !equalBig(got, &coeffs[i]) {
						t.Errorf("recovery at index %d got %s but expected %s", i, bigStr(got), bigStr(&coeffs[i]))
					}
				}
				// And recover the original data for good measure
				back, err := fs.FFT(recovered, true)
				if err != nil {
					t.Fatal(err)
				}
				debugBigs("back", back)
				half := uint64(len(back)) / 2
				for i := uint64(0); i < half; i++ {
					if got := &back[i]; !equalBig(got, &data[i]) {
						t.Errorf("data at index %d got %s but expected %s", i, bigStr(got), bigStr(&data[i]))
					}
				}
				for i := half; i < fs.width; i++ {
					if got := &back[i]; !equalZero(got) {
						t.Errorf("expected zero padding in index %d", i)
					}
				}
			})
		}
	}
}
