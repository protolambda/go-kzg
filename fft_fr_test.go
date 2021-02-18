package kzg

import (
	"github.com/protolambda/go-kzg/bls"
	"testing"
)

func TestFFTRoundtrip(t *testing.T) {
	fs := NewFFTSettings(4)
	data := make([]bls.Big, fs.maxWidth, fs.maxWidth)
	for i := uint64(0); i < fs.maxWidth; i++ {
		bls.AsBig(&data[i], i)
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
		if got, expected := &res[i], &data[i]; !bls.EqualBig(got, expected) {
			t.Errorf("difference: %d: got: %s  expected: %s", i, bls.BigStr(got), bls.BigStr(expected))
		}
	}
	t.Log("zero", bls.BigStr(&bls.ZERO))
	t.Log("zero", bls.BigStr(&bls.ONE))
}

func TestInvFFT(t *testing.T) {
	fs := NewFFTSettings(4)
	data := make([]bls.Big, fs.maxWidth, fs.maxWidth)
	for i := uint64(0); i < fs.maxWidth; i++ {
		bls.AsBig(&data[i], i)
	}
	debugBigs("input data", data)
	res, err := fs.FFT(data, true)
	if err != nil {
		t.Fatal(err)
	}
	debugBigs("result", res)
	ToBigNum := func(v string) (out bls.Big) {
		bls.BigNum(&out, v)
		return
	}
	expected := []bls.Big{
		ToBigNum("26217937587563095239723870254092982918845276250263818911301829349969290592264"),
		ToBigNum("40905488090558605688319636812215252217941835718478251840326926365086504505065"),
		ToBigNum("10037948829646534413413739647971946522809495755620173630072972432081702959148"),
		ToBigNum("43571192877568624546930318420751319449039972945062659080199348274630726213098"),
		ToBigNum("26217937587563095241456442667129809078233411015607690300436955584351971573760"),
		ToBigNum("23495295218275555727033128776954731040973520495197797376593908347998044220817"),
		ToBigNum("10037948829646534409948594821898294204033226224932430851802719963316340996140"),
		ToBigNum("20829590431265536861492157516271359172322844207237904580180981500923098586768"),
		ToBigNum("26217937587563095239723870254092982918845276250263818911301829349969290592256"),
		ToBigNum("31606284743860653617955582991914606665367708293289733242422677199015482597744"),
		ToBigNum("42397926345479656069499145686287671633657326275595206970800938736622240188372"),
		ToBigNum("28940579956850634752414611731231234796717032005329840446009750351940536963695"),
		ToBigNum("26217937587563095237991297841056156759457141484919947522166703115586609610752"),
		ToBigNum("8864682297557565932517422087434646388650579555464978742404310425307854971414"),
		ToBigNum("42397926345479656066034000860214019314881056744907464192530686267856878225364"),
		ToBigNum("11530387084567584791128103695970713619748716782049385982276732334852076679447"),
	}
	for i := range res {
		if got := &res[i]; !bls.EqualBig(got, &expected[i]) {
			t.Errorf("difference: %d: got: %s  expected: %s", i, bls.BigStr(got), bls.BigStr(&expected[i]))
		}
	}
}
