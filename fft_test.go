package go_verkle

import (
	"testing"
)

func TestFFTRoundtrip(t *testing.T) {
	data := make([]Big, WIDTH, WIDTH)
	for i := uint64(0); i < WIDTH; i++ {
		data[i] = asBig(i)
	}
	coeffs := FFT(data, MODULUS, ROOT_OF_UNITY, false)
	res := FFT(coeffs, MODULUS, ROOT_OF_UNITY, true)
	for i, got := range res {
		if cmpBig(got, data[i]) != 0 {
			t.Errorf("difference: %d: got: %s  expected: %s", i, bigStr(got), bigStr(data[i]))
		}
	}
}

func TestInvFFT(t *testing.T) {
	data := make([]Big, WIDTH, WIDTH)
	for i := uint64(0); i < WIDTH; i++ {
		data[i] = asBig(i)
	}
	debugBigs("input data", data)
	res := FFT(data, MODULUS, ROOT_OF_UNITY, true)
	debugBigs("result", res)
	expected := []Big{
		bigNum("26217937587563095239723870254092982918845276250263818911301829349969290592264"),
		bigNum("40905488090558605688319636812215252217941835718478251840326926365086504505065"),
		bigNum("10037948829646534413413739647971946522809495755620173630072972432081702959148"),
		bigNum("43571192877568624546930318420751319449039972945062659080199348274630726213098"),
		bigNum("26217937587563095241456442667129809078233411015607690300436955584351971573760"),
		bigNum("23495295218275555727033128776954731040973520495197797376593908347998044220817"),
		bigNum("10037948829646534409948594821898294204033226224932430851802719963316340996140"),
		bigNum("20829590431265536861492157516271359172322844207237904580180981500923098586768"),
		bigNum("26217937587563095239723870254092982918845276250263818911301829349969290592256"),
		bigNum("31606284743860653617955582991914606665367708293289733242422677199015482597744"),
		bigNum("42397926345479656069499145686287671633657326275595206970800938736622240188372"),
		bigNum("28940579956850634752414611731231234796717032005329840446009750351940536963695"),
		bigNum("26217937587563095237991297841056156759457141484919947522166703115586609610752"),
		bigNum("8864682297557565932517422087434646388650579555464978742404310425307854971414"),
		bigNum("42397926345479656066034000860214019314881056744907464192530686267856878225364"),
		bigNum("11530387084567584791128103695970713619748716782049385982276732334852076679447"),
	}
	for i, got := range res {
		if cmpBig(got, expected[i]) != 0 {
			t.Errorf("difference: %d: got: %s  expected: %s", i, bigStr(got), bigStr(expected[i]))
		}
	}
}

func bigRange(start uint64, end uint64) []Big {
	l := end - start
	out := make([]Big, l, l)
	for i := uint64(0); i < l; i++ {
		out[i] = asBig(start + i)
	}
	return out
}

func evalPolyRange(coeffs []Big, xs []Big) []Big {
	out := make([]Big, len(xs), len(xs))
	for i, x := range xs {
		out[i] = evalPolyAt(coeffs, x)
	}
	return out
}

func evalPolyAt(coeffs []Big, x Big) Big {
	var out = ZERO
	var powerOfX = ONE
	for _, c := range coeffs {
		v := mulModBig(c, powerOfX, MODULUS)
		out = addModBig(out, v, MODULUS)
		powerOfX = mulModBig(powerOfX, x, MODULUS)
	}
	return out
}

// Test the test util, sanity check
func TestPolyRange(t *testing.T) {
	coeffs := bigRange(4, 8)
	out := evalPolyAt(coeffs, asBig(3))
	// 4*(3^0) + 5*(3^1) + 6*(3^2) + 7*(3^3) = 262
	if cmpBig(out, asBig(262)) != 0 {
		t.Fatalf("bad result: %s", bigStr(out))
	}
}
