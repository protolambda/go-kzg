package bls

import (
	"fmt"
	"math/big"
)

var Scale2RootOfUnity []Fr

var ZERO, ONE, TWO Fr
var MODULUS_MINUS1, MODULUS_MINUS1_DIV2, MODULUS_MINUS2 Fr
var INVERSE_TWO Fr

func ToFr(v string) (out Fr) {
	SetFr(&out, v)
	return
}

func initGlobals() {

	// MODULUS = 52435875175126190479447740508185965837690552500527637822603658699938581184513
	// PRIMITIVE_ROOT = 7
	// [pow(PRIMITIVE_ROOT, (MODULUS - 1) // (2**i), MODULUS) for i in range(32)]
	Scale2RootOfUnity = []Fr{
		/* k=0          r=1          */ ToFr("1"),
		/* k=1          r=2          */ ToFr("52435875175126190479447740508185965837690552500527637822603658699938581184512"),
		/* k=2          r=4          */ ToFr("3465144826073652318776269530687742778270252468765361963008"),
		/* k=3          r=8          */ ToFr("23674694431658770659612952115660802947967373701506253797663184111817857449850"),
		/* k=4          r=16         */ ToFr("14788168760825820622209131888203028446852016562542525606630160374691593895118"),
		/* k=5          r=32         */ ToFr("36581797046584068049060372878520385032448812009597153775348195406694427778894"),
		/* k=6          r=64         */ ToFr("31519469946562159605140591558550197856588417350474800936898404023113662197331"),
		/* k=7          r=128        */ ToFr("47309214877430199588914062438791732591241783999377560080318349803002842391998"),
		/* k=8          r=256        */ ToFr("36007022166693598376559747923784822035233416720563672082740011604939309541707"),
		/* k=9          r=512        */ ToFr("4214636447306890335450803789410475782380792963881561516561680164772024173390"),
		/* k=10         r=1024       */ ToFr("22781213702924172180523978385542388841346373992886390990881355510284839737428"),
		/* k=11         r=2048       */ ToFr("49307615728544765012166121802278658070711169839041683575071795236746050763237"),
		/* k=12         r=4096       */ ToFr("39033254847818212395286706435128746857159659164139250548781411570340225835782"),
		/* k=13         r=8192       */ ToFr("32731401973776920074999878620293785439674386180695720638377027142500196583783"),
		/* k=14         r=16384      */ ToFr("39072540533732477250409069030641316533649120504872707460480262653418090977761"),
		/* k=15         r=32768      */ ToFr("22872204467218851938836547481240843888453165451755431061227190987689039608686"),
		/* k=16         r=65536      */ ToFr("15076889834420168339092859836519192632846122361203618639585008852351569017005"),
		/* k=17         r=131072     */ ToFr("15495926509001846844474268026226183818445427694968626800913907911890390421264"),
		/* k=18         r=262144     */ ToFr("20439484849038267462774237595151440867617792718791690563928621375157525968123"),
		/* k=19         r=524288     */ ToFr("37115000097562964541269718788523040559386243094666416358585267518228781043101"),
		/* k=20         r=1048576    */ ToFr("1755840822790712607783180844474754741366353396308200820563736496551326485835"),
		/* k=21         r=2097152    */ ToFr("32468834368094611004052562760214251466632493208153926274007662173556188291130"),
		/* k=22         r=4194304    */ ToFr("4859563557044021881916617240989566298388494151979623102977292742331120628579"),
		/* k=23         r=8388608    */ ToFr("52167942466760591552294394977846462646742207006759917080697723404762651336366"),
		/* k=24         r=16777216   */ ToFr("18596002123094854211120822350746157678791770803088570110573239418060655130524"),
		/* k=25         r=33554432   */ ToFr("734830308204920577628633053915970695663549910788964686411700880930222744862"),
		/* k=26         r=67108864   */ ToFr("4541622677469846713471916119560591929733417256448031920623614406126544048514"),
		/* k=27         r=134217728  */ ToFr("15932505959375582308231798849995567447410469395474322018100309999481287547373"),
		/* k=28         r=268435456  */ ToFr("37480612446576615530266821837655054090426372233228960378061628060638903214217"),
		/* k=29         r=536870912  */ ToFr("5660829372603820951332104046316074966592589311213397907344198301300676239643"),
		/* k=30         r=1073741824 */ ToFr("20094891866007995289136270587723853997043774683345353712639419774914899074390"),
		/* k=31         r=2147483648 */ ToFr("34070893824967080313820779135880760772780807222436853681508667398599787661631"),
	}

	AsFr(&ZERO, 0)
	AsFr(&ONE, 1)
	AsFr(&TWO, 2)

	SubModFr(&MODULUS_MINUS1, &ZERO, &ONE)
	DivModFr(&MODULUS_MINUS1_DIV2, &MODULUS_MINUS1, &TWO)
	SubModFr(&MODULUS_MINUS2, &ZERO, &TWO)
	InvModFr(&INVERSE_TWO, &TWO)
}

func IsPowerOfTwo(v uint64) bool {
	return v&(v-1) == 0
}

func EvalPolyAtUnoptimized(dst *Fr, coeffs []Fr, x *Fr) {
	if len(coeffs) == 0 {
		CopyFr(dst, &ZERO)
		return
	}
	if EqualZero(x) {
		CopyFr(dst, &coeffs[0])
		return
	}
	// Horner's method: work backwards, avoid doing more than N multiplications
	// https://en.wikipedia.org/wiki/Horner%27s_method
	var last Fr
	CopyFr(&last, &coeffs[len(coeffs)-1])
	var tmp Fr
	for i := len(coeffs) - 2; i >= 0; i-- {
		MulModFr(&tmp, &last, x)
		AddModFr(&last, &tmp, &coeffs[i])
	}
	CopyFr(dst, &last)
}

// Evaluate a polynomial (in evaluation form) at an arbitrary point x using the barycentric formula:
//
//	f(x) = (1 - x**WIDTH) / WIDTH  *  sum_(i=0)^WIDTH  (f(DOMAIN[i]) * DOMAIN[i]) / (x - DOMAIN[i])
//
// Scale is used to indicate the power of 2 to use to stride through the domain values.
// Note: scale == 0 when the roots of unity length matches the polynomial.
func EvaluatePolyInEvaluationForm(yFr *Fr, poly []Fr, x *Fr, rootsOfUnity []Fr, scale uint8) {
	if len(poly) != len(rootsOfUnity)>>scale {
		panic(fmt.Errorf("expected roots of unity (len %d >> %d == %d) to match polynomial size (len %d)", len(rootsOfUnity), scale, len(rootsOfUnity)>>scale, len(poly)))
	}

	width := big.NewInt(int64(len(poly)))
	var widthFr Fr
	AsFr(&widthFr, uint64(len(poly)))
	var inverseWidth Fr
	InvModFr(&inverseWidth, &widthFr)

	// Precomputing the mod inverses as a batch is alot faster
	invDenom := make([]Fr, len(poly))
	for i := range invDenom {
		// (x - DOMAIN[i])
		SubModFr(&invDenom[i], x, &rootsOfUnity[i<<scale])
	}
	// now each value becomes 1 / (x - DOMAIN[i])
	BatchInvModFr(invDenom)

	// sum_(i=0)^WIDTH  (f(DOMAIN[i]) * DOMAIN[i]) / (x - DOMAIN[i])
	var y Fr
	for i := 0; i < len(poly); i++ {
		// f(DOMAIN[i]) * DOMAIN[i])
		var num Fr
		MulModFr(&num, &poly[i], &rootsOfUnity[i<<scale])

		// (f(DOMAIN[i]) * DOMAIN[i]) / (x - DOMAIN[i])
		var div Fr
		MulModFr(&div, &num, &invDenom[i])

		// sum
		var tmp Fr
		AddModFr(&tmp, &y, &div)
		CopyFr(&y, &tmp)
	}

	// (1 - x**WIDTH)
	var powB Fr
	ExpModFr(&powB, x, width)
	SubModFr(&powB, &powB, &ONE) // TODO: prysm does x**width - 1, and it passes the test, but old spec comment says 1 - x**width ?
	// (1 - x**WIDTH) / WIDTH
	var tmp Fr
	MulModFr(&tmp, &powB, &inverseWidth)

	// (1 - x**WIDTH) / WIDTH  *  sum_(i=0)^WIDTH  (f(DOMAIN[i]) * DOMAIN[i]) / (x - DOMAIN[i])
	MulModFr(yFr, &y, &tmp)
}
