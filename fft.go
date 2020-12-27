// Experimental translation to Go.
// Original:
// - https://github.com/ethereum/research/blob/master/mimc_stark/fft.py
// - https://github.com/ethereum/research/blob/master/mimc_stark/recovery.py

package kate

import (
	"fmt"
	"strings"
)

var scale2RootOfUnity []Big

var ZERO_G1 G1

var ZERO, ONE, TWO Big
var MODULUS_MINUS1, MODULUS_MINUS1_DIV2, MODULUS_MINUS2 Big
var INVERSE_TWO Big

func initGlobals() {
	bigNumHelper := func(v string) (out Big) {
		bigNum(&out, v)
		return
	}
	// MODULUS = 52435875175126190479447740508185965837690552500527637822603658699938581184513
	// PRIMITIVE_ROOT = 5
	// [pow(PRIMITIVE_ROOT, (MODULUS - 1) // (2**i), MODULUS) for i in range(32)]
	scale2RootOfUnity = []Big{
		/* k=0          r=1          */ bigNumHelper("1"),
		/* k=1          r=2          */ bigNumHelper("52435875175126190479447740508185965837690552500527637822603658699938581184512"),
		/* k=2          r=4          */ bigNumHelper("3465144826073652318776269530687742778270252468765361963008"),
		/* k=3          r=8          */ bigNumHelper("28761180743467419819834788392525162889723178799021384024940474588120723734663"),
		/* k=4          r=16         */ bigNumHelper("35811073542294463015946892559272836998938171743018714161809767624935956676211"),
		/* k=5          r=32         */ bigNumHelper("32311457133713125762627935188100354218453688428796477340173861531654182464166"),
		/* k=6          r=64         */ bigNumHelper("6460039226971164073848821215333189185736442942708452192605981749202491651199"),
		/* k=7          r=128        */ bigNumHelper("3535074550574477753284711575859241084625659976293648650204577841347885064712"),
		/* k=8          r=256        */ bigNumHelper("21071158244812412064791010377580296085971058123779034548857891862303448703672"),
		/* k=9          r=512        */ bigNumHelper("12531186154666751577774347439625638674013361494693625348921624593362229945844"),
		/* k=10         r=1024       */ bigNumHelper("21328829733576761151404230261968752855781179864716879432436835449516750606329"),
		/* k=11         r=2048       */ bigNumHelper("30450688096165933124094588052280452792793350252342406284806180166247113753719"),
		/* k=12         r=4096       */ bigNumHelper("7712148129911606624315688729500842900222944762233088101895611600385646063109"),
		/* k=13         r=8192       */ bigNumHelper("4862464726302065505506688039068558711848980475932963135959468859464391638674"),
		/* k=14         r=16384      */ bigNumHelper("36362449573598723777784795308133589731870287401357111047147227126550012376068"),
		/* k=15         r=32768      */ bigNumHelper("30195699792882346185164345110260439085017223719129789169349923251189180189908"),
		/* k=16         r=65536      */ bigNumHelper("46605497109352149548364111935960392432509601054990529243781317021485154656122"),
		/* k=17         r=131072     */ bigNumHelper("2655041105015028463885489289298747241391034429256407017976816639065944350782"),
		/* k=18         r=262144     */ bigNumHelper("42951892408294048319804799042074961265671975460177021439280319919049700054024"),
		/* k=19         r=524288     */ bigNumHelper("26418991338149459552592774439099778547711964145195139895155358980955972635668"),
		/* k=20         r=1048576    */ bigNumHelper("23615957371642610195417524132420957372617874794160903688435201581369949179370"),
		/* k=21         r=2097152    */ bigNumHelper("50175287592170768174834711592572954584642344504509533259061679462536255873767"),
		/* k=22         r=4194304    */ bigNumHelper("1664636601308506509114953536181560970565082534259883289958489163769791010513"),
		/* k=23         r=8388608    */ bigNumHelper("36760611456605667464829527713580332378026420759024973496498144810075444759800"),
		/* k=24         r=16777216   */ bigNumHelper("13205172441828670567663721566567600707419662718089030114959677511969243860524"),
		/* k=25         r=33554432   */ bigNumHelper("10335750295308996628517187959952958185340736185617535179904464397821611796715"),
		/* k=26         r=67108864   */ bigNumHelper("51191008403851428225654722580004101559877486754971092640244441973868858562750"),
		/* k=27         r=134217728  */ bigNumHelper("24000695595003793337811426892222725080715952703482855734008731462871475089715"),
		/* k=28         r=268435456  */ bigNumHelper("18727201054581607001749469507512963489976863652151448843860599973148080906836"),
		/* k=29         r=536870912  */ bigNumHelper("50819341139666003587274541409207395600071402220052213520254526953892511091577"),
		/* k=30         r=1073741824 */ bigNumHelper("3811138593988695298394477416060533432572377403639180677141944665584601642504"),
		/* k=31         r=2147483648 */ bigNumHelper("43599901455287962219281063402626541872197057165786841304067502694013639882090"),
	}

	asBig(&ZERO, 0)
	asBig(&ONE, 1)
	asBig(&TWO, 2)

	subModBig(&MODULUS_MINUS1, &ZERO, &ONE)
	divModBig(&MODULUS_MINUS1_DIV2, &MODULUS_MINUS1, &TWO)
	subModBig(&MODULUS_MINUS2, &ZERO, &TWO)
	invModBig(&INVERSE_TWO, &TWO)

	ClearG1(&ZERO_G1)
}

// Expands the power circle for a given root of unity to WIDTH+1 values.
// The first entry will be 1, the last entry will also be 1,
// for convenience when reversing the array (useful for inverses)
func expandRootOfUnity(rootOfUnity *Big) []Big {
	rootz := make([]Big, 2)
	rootz[0] = ONE // some unused number in py code
	rootz[1] = *rootOfUnity
	for i := 1; !equalOne(&rootz[i]); {
		rootz = append(rootz, Big{})
		this := &rootz[i]
		i++
		mulModBig(&rootz[i], this, rootOfUnity)
	}
	return rootz
}

type FFTSettings struct {
	maxWidth uint64
	// the generator used to get all roots of unity
	rootOfUnity *Big
	// domain, starting and ending with 1 (duplicate!)
	expandedRootsOfUnity []Big
	// reverse domain, same as inverse values of domain. Also starting and ending with 1.
	reverseRootsOfUnity []Big
}

// TODO: generate some setup G1, G2 for testing purposes
// Secret point to evaluate polynomials at.
// Setup values are defined as [g * s**i for i in range(m)]  (correct?)

func NewFFTSettings(maxScale uint8) *FFTSettings {
	width := uint64(1) << maxScale
	root := &scale2RootOfUnity[maxScale]
	rootz := expandRootOfUnity(&scale2RootOfUnity[maxScale])
	// reverse roots of unity
	rootzReverse := make([]Big, len(rootz), len(rootz))
	copy(rootzReverse, rootz)
	for i, j := uint64(0), uint64(len(rootz)-1); i < j; i, j = i+1, j-1 {
		rootzReverse[i], rootzReverse[j] = rootzReverse[j], rootzReverse[i]
	}

	return &FFTSettings{
		maxWidth:             width,
		rootOfUnity:          root,
		expandedRootsOfUnity: rootz,
		reverseRootsOfUnity:  rootzReverse,
	}
}

func (fs *FFTSettings) zPoly(positions []uint64) []Big {
	n := uint64(len(positions))
	stride := fs.maxWidth / n
	return fs._zPoly(positions, stride)
}

func debugBigPtrs(msg string, values []*Big) {
	var out strings.Builder
	out.WriteString("---")
	out.WriteString(msg)
	out.WriteString("---\n")
	for i := range values {
		out.WriteString(fmt.Sprintf("#%4d: %s\n", i, bigStr(values[i])))
	}
	fmt.Println(out.String())
}

func debugBigs(msg string, values []Big) {
	var out strings.Builder
	out.WriteString("---")
	out.WriteString(msg)
	out.WriteString("---\n")
	for i := range values {
		out.WriteString(fmt.Sprintf("#%4d: %s\n", i, bigStr(&values[i])))
	}
	fmt.Println(out.String())
}

func debugBigsOffsetStride(msg string, values []Big, offset uint64, stride uint64) {
	var out strings.Builder
	out.WriteString("---")
	out.WriteString(msg)
	out.WriteString("---\n")
	j := uint64(0)
	for i := offset; i < uint64(len(values)); i += stride {
		out.WriteString(fmt.Sprintf("#%4d: %s\n", j, bigStr(&values[i])))
		j++
	}
	fmt.Println(out.String())
}

func (fs *FFTSettings) simpleFT(vals []Big, valsOffset uint64, valsStride uint64, rootsOfUnity []Big, rootsOfUnityStride uint64, out []Big) {
	l := uint64(len(out))
	var v Big
	var tmp Big
	var last Big
	for i := uint64(0); i < l; i++ {
		jv := &vals[valsOffset]
		r := &rootsOfUnity[0]
		mulModBig(&v, jv, r)
		CopyBigNum(&last, &v)

		for j := uint64(1); j < l; j++ {
			jv := &vals[valsOffset+j*valsStride]
			r := &rootsOfUnity[((i*j)%l)*rootsOfUnityStride]
			mulModBig(&v, jv, r)
			CopyBigNum(&tmp, &last)
			addModBig(&last, &tmp, &v)
		}
		CopyBigNum(&out[i], &last)
	}
}

func (fs *FFTSettings) simpleFTG1(vals []G1, valsOffset uint64, valsStride uint64, rootsOfUnity []Big, rootsOfUnityStride uint64, out []G1) {
	l := uint64(len(out))
	var v G1
	var tmp G1
	var last G1
	for i := uint64(0); i < l; i++ {
		jv := &vals[valsOffset]
		r := &rootsOfUnity[0]
		mulG1(&v, jv, r)
		CopyG1(&last, &v)

		for j := uint64(1); j < l; j++ {
			jv := &vals[valsOffset+j*valsStride]
			r := &rootsOfUnity[((i*j)%l)*rootsOfUnityStride]
			mulG1(&v, jv, r)
			CopyG1(&tmp, &last)
			addG1(&last, &tmp, &v)
		}
		CopyG1(&out[i], &last)
	}
}

func (fs *FFTSettings) _fft(vals []Big, valsOffset uint64, valsStride uint64, rootsOfUnity []Big, rootsOfUnityStride uint64, out []Big) {
	if len(out) <= 4 { // if the value count is small, run the unoptimized version instead. // TODO tune threshold.
		fs.simpleFT(vals, valsOffset, valsStride, rootsOfUnity, rootsOfUnityStride, out)
		return
	}

	half := uint64(len(out)) >> 1
	// L will be the left half of out
	fs._fft(vals, valsOffset, valsStride<<1, rootsOfUnity, rootsOfUnityStride<<1, out[:half])
	// R will be the right half of out
	fs._fft(vals, valsOffset+valsStride, valsStride<<1, rootsOfUnity, rootsOfUnityStride<<1, out[half:]) // just take even again

	var yTimesRoot Big
	var x, y Big
	for i := uint64(0); i < half; i++ {
		// temporary copies, so that writing to output doesn't conflict with input
		CopyBigNum(&x, &out[i])
		CopyBigNum(&y, &out[i+half])
		root := &rootsOfUnity[i*rootsOfUnityStride]
		mulModBig(&yTimesRoot, &y, root)
		addModBig(&out[i], &x, &yTimesRoot)
		subModBig(&out[i+half], &x, &yTimesRoot)
	}
}

func (fs *FFTSettings) _fftG1(vals []G1, valsOffset uint64, valsStride uint64, rootsOfUnity []Big, rootsOfUnityStride uint64, out []G1) {
	if len(out) <= 4 { // if the value count is small, run the unoptimized version instead. // TODO tune threshold. (can be different for G1)
		fs.simpleFTG1(vals, valsOffset, valsStride, rootsOfUnity, rootsOfUnityStride, out)
		return
	}

	half := uint64(len(out)) >> 1
	// L will be the left half of out
	fs._fftG1(vals, valsOffset, valsStride<<1, rootsOfUnity, rootsOfUnityStride<<1, out[:half])
	// R will be the right half of out
	fs._fftG1(vals, valsOffset+valsStride, valsStride<<1, rootsOfUnity, rootsOfUnityStride<<1, out[half:]) // just take even again

	var yTimesRoot G1
	var x, y G1
	for i := uint64(0); i < half; i++ {
		// temporary copies, so that writing to output doesn't conflict with input
		CopyG1(&x, &out[i])
		CopyG1(&y, &out[i+half])
		root := &rootsOfUnity[i*rootsOfUnityStride]
		mulG1(&yTimesRoot, &y, root)
		addG1(&out[i], &x, &yTimesRoot)
		subG1(&out[i+half], &x, &yTimesRoot)
	}
}

func isPowerOfTwo(v uint64) bool {
	return v&(v-1) == 0
}

func (fs *FFTSettings) FFT(vals []Big, inv bool) ([]Big, error) {
	n := uint64(len(vals))
	if n > fs.maxWidth {
		return nil, fmt.Errorf("got %d values but only have %d roots of unity", n, fs.maxWidth)
	}
	if !isPowerOfTwo(n) {
		return nil, fmt.Errorf("got %d values but not a power of two", n)
	}
	// We make a copy so we can mutate it during the work.
	valsCopy := make([]Big, n, n) // TODO: maybe optimize this away, and write back to original input array?
	for i := uint64(0); i < n; i++ {
		CopyBigNum(&valsCopy[i], &vals[i])
	}
	if inv {
		var invLen Big
		asBig(&invLen, n)
		invModBig(&invLen, &invLen)
		rootz := fs.reverseRootsOfUnity[:fs.maxWidth]
		stride := fs.maxWidth / n

		out := make([]Big, n, n)
		fs._fft(valsCopy, 0, 1, rootz, stride, out)
		debugBigs("inv fft without len invert", out)
		var tmp Big
		for i := 0; i < len(out); i++ {
			mulModBig(&tmp, &out[i], &invLen)
			CopyBigNum(&out[i], &tmp) // TODO: depending on bignum implementation, allow to directly write back to an input
		}
		return out, nil
	} else {
		out := make([]Big, n, n)
		rootz := fs.expandedRootsOfUnity[:fs.maxWidth]
		stride := fs.maxWidth / n
		// Regular FFT
		fs._fft(valsCopy, 0, 1, rootz, stride, out)
		return out, nil
	}
}

func (fs *FFTSettings) FFTG1(vals []G1, inv bool) ([]G1, error) {
	n := uint64(len(vals))
	if n > fs.maxWidth {
		return nil, fmt.Errorf("got %d values but only have %d roots of unity", n, fs.maxWidth)
	}
	if !isPowerOfTwo(n) {
		return nil, fmt.Errorf("got %d values but not a power of two", n)
	}
	// We make a copy so we can mutate it during the work.
	valsCopy := make([]G1, n, n)
	for i := 0; i < len(vals); i++ { // TODO: maybe optimize this away, and write back to original input array?
		CopyG1(&valsCopy[i], &vals[i])
	}
	if inv {
		var invLen Big
		asBig(&invLen, n)
		invModBig(&invLen, &invLen)
		rootz := fs.reverseRootsOfUnity[:fs.maxWidth]
		stride := fs.maxWidth / n

		out := make([]G1, n, n)
		fs._fftG1(valsCopy, 0, 1, rootz, stride, out)
		var tmp G1
		for i := 0; i < len(out); i++ {
			mulG1(&tmp, &out[i], &invLen)
			CopyG1(&out[i], &tmp)
		}
		return out, nil
	} else {
		out := make([]G1, n, n)
		rootz := fs.reverseRootsOfUnity[:fs.maxWidth]
		stride := fs.maxWidth / n
		// Regular FFT
		fs._fftG1(valsCopy, 0, 1, rootz, stride, out)
		return out, nil
	}
}

// warning: the values in `a` are modified in-place to become the outputs.
// Make a deep copy first if you need to use them later.
func (fs *FFTSettings) dASFFTExtension(ab []Big, domainStride uint64) {
	if len(ab) == 2 {
		aHalf0 := &ab[0]
		aHalf1 := &ab[1]
		var tmp Big
		addModBig(&tmp, aHalf0, aHalf1)
		var x Big
		mulModBig(&x, &tmp, &INVERSE_TWO)
		// y = (((a_half0 - x) % modulus) * inverse_domain[0]) % modulus     # inverse_domain[0] will always be 1
		var y Big
		subModBig(&y, aHalf0, &x)
		// re-use tmp for y_times_root
		mulModBig(&tmp, &y, &fs.expandedRootsOfUnity[domainStride])
		addModBig(&ab[0], &x, &tmp)
		subModBig(&ab[1], &x, &tmp)
		return
	}

	if len(ab) < 2 {
		panic("bad usage")
	}

	half := uint64(len(ab))
	halfHalf := half >> 1
	abHalf0s := ab[:halfHalf]
	abHalf1s := ab[halfHalf:half]
	// Instead of allocating L0 and L1, just modify a in-place.
	//L0[i] = (((a_half0 + a_half1) % modulus) * inv2) % modulus
	//R0[i] = (((a_half0 - L0[i]) % modulus) * inverse_domain[i * 2]) % modulus
	var tmp1, tmp2, tmp3 Big
	for i := uint64(0); i < halfHalf; i++ {
		aHalf0 := &abHalf0s[i]
		aHalf1 := &abHalf1s[i]
		addModBig(&tmp1, aHalf0, aHalf1)
		mulModBig(&tmp2, &tmp1, &INVERSE_TWO) // tmp2 holds later L0[i] result
		subModBig(&tmp3, aHalf0, &tmp2)
		mulModBig(aHalf1, &tmp3, &fs.reverseRootsOfUnity[i*2*domainStride])
		CopyBigNum(aHalf0, &tmp2)
	}

	// L will be the left half of out
	fs.dASFFTExtension(abHalf0s, domainStride<<1)
	// R will be the right half of out
	fs.dASFFTExtension(abHalf1s, domainStride<<1)

	// The odd deduced outputs are written to the output array already, but then updated in-place
	// L1 = b[:halfHalf]
	// R1 = b[halfHalf:]

	// Half the work of a regular FFT: only deal with uneven-index outputs
	var yTimesRoot Big
	var x, y Big
	for i := uint64(0); i < halfHalf; i++ {
		// Temporary copies, so that writing to output doesn't conflict with input.
		// Note that one hand is from L1, the other R1
		CopyBigNum(&x, &abHalf0s[i])
		CopyBigNum(&y, &abHalf1s[i])
		root := &fs.expandedRootsOfUnity[(1+2*i)*domainStride]
		mulModBig(&yTimesRoot, &y, root)
		// write outputs in place, avoid unnecessary list allocations
		addModBig(&abHalf0s[i], &x, &yTimesRoot)
		subModBig(&abHalf1s[i], &x, &yTimesRoot)
	}
}

// Takes vals as input, the values of the even indices.
// Then computes the values for the odd indices, which combined would make the right half of coefficients zero.
// Warning: the odd results are written back to the vals slice.
func (fs *FFTSettings) DASFFTExtension(vals []Big) {
	if uint64(len(vals))*2 > fs.maxWidth {
		panic("domain too small for extending requested values")
	}
	fs.dASFFTExtension(vals, 1)
}

func (fs *FFTSettings) mulPolys(a []Big, b []Big, rootsOfUnityStride uint64) []Big {
	size := fs.maxWidth / rootsOfUnityStride
	aVals := make([]Big, size, size)
	bVals := make([]Big, size, size)
	for i := 0; i < len(a); i++ {
		aVals[i] = a[i]
	}
	for i := len(a); i < len(aVals); i++ {
		aVals[i] = ZERO
	}
	for i := 0; i < len(b); i++ {
		bVals[i] = b[i]
	}
	for i := len(b); i < len(bVals); i++ {
		bVals[i] = ZERO
	}
	rootz := fs.expandedRootsOfUnity[:fs.maxWidth]
	// Get FFT of a and b
	x1 := make([]Big, len(aVals), len(aVals))
	fs._fft(aVals, 0, 1, rootz, rootsOfUnityStride, x1)

	x2 := make([]Big, len(bVals), len(bVals))
	fs._fft(bVals, 0, 1, rootz, rootsOfUnityStride, x2)

	// multiply the two. Hack: store results in x1
	var tmp Big
	for i := 0; i < len(x1); i++ {
		CopyBigNum(&tmp, &x1[i])
		mulModBig(&x1[i], &tmp, &x2[i])
	}
	revRootz := fs.reverseRootsOfUnity[:fs.maxWidth]

	out := make([]Big, len(x1), len(x1))
	// compute the FFT of the multiplied values.
	fs._fft(x1, 0, 1, revRootz, rootsOfUnityStride, out)
	return out
}

// Calculates modular inverses [1/values[0], 1/values[1] ...]
func multiInv(values []Big) []Big {
	partials := make([]Big, len(values)+1, len(values)+1)
	partials[0] = values[0]
	for i := 0; i < len(values); i++ {
		mulModBig(&partials[i+1], &partials[i], &values[i])
	}
	var inv Big
	var tmp Big
	invModBig(&inv, &partials[len(partials)-1])
	outputs := make([]Big, len(values), len(values))
	for i := len(values); i > 0; i-- {
		mulModBig(&outputs[i-1], &partials[i-1], &inv)
		CopyBigNum(&tmp, &inv)
		mulModBig(&inv, &tmp, &values[i-1])
	}
	return outputs
}

// Generates q(x) = poly(k * x)
func pOfKX(poly []Big, k *Big) []Big {
	out := make([]Big, len(poly), len(poly))
	powerOfK := ONE
	var tmp Big
	for i := range poly {
		mulModBig(&out[i], &poly[i], &powerOfK)
		CopyBigNum(&tmp, &powerOfK)
		mulModBig(&powerOfK, &tmp, k)
	}
	return out
}

func inefficientOddEvenDiv2(positions []uint64) (even []uint64, odd []uint64) { // TODO optimize away
	for _, p := range positions {
		if p&1 == 0 {
			even = append(even, p>>1)
		} else {
			odd = append(odd, p>>1)
		}
	}
	return
}

// Return (x - root**positions[0]) * (x - root**positions[1]) * ...
// possibly with a constant factor offset
func (fs *FFTSettings) _zPoly(positions []uint64, rootsOfUnityStride uint64) []Big {
	// If there are not more than 4 positions, use the naive
	// O(n^2) algorithm as it is faster
	if len(positions) <= 4 {
		/*
		   root = [1]
		   for pos in positions:
		       x = roots_of_unity[pos]
		       root.insert(0, 0)
		       for j in range(len(root)-1):
		           root[j] -= root[j+1] * x
		   return [x % modulus for x in root]
		*/
		root := make([]Big, len(positions)+1, len(positions)+1)
		root[0] = ONE
		i := 1
		var v Big
		var tmp Big
		for _, pos := range positions {
			x := &fs.expandedRootsOfUnity[pos*rootsOfUnityStride]
			root[i] = ZERO
			for j := i; j >= 1; j-- {
				mulModBig(&v, &root[j-1], x)
				CopyBigNum(&tmp, &root[j])
				subModBig(&root[j], &tmp, &v)
			}
			i++
		}
		// We did the reverse representation of 'root' as the python code, to not insert at the start all the time.
		// Now turn it back around.
		for i, j := 0, len(root)-1; i < j; i, j = i+1, j-1 {
			root[i], root[j] = root[j], root[i]
		}
		return root
	}
	// Recursively find the zpoly for even indices and odd
	// indices, operating over a half-size subgroup in each case
	evenPositions, oddPositions := inefficientOddEvenDiv2(positions)
	left := fs._zPoly(evenPositions, rootsOfUnityStride<<1)
	right := fs._zPoly(oddPositions, rootsOfUnityStride<<1)
	invRoot := &fs.reverseRootsOfUnity[rootsOfUnityStride]
	// Offset the result for the odd indices, and combine the two
	out := fs.mulPolys(left, pOfKX(right, invRoot), rootsOfUnityStride)
	// Deal with the special case where mul_polys returns zero
	// when it should return x ^ (2 ** k) - 1
	isZero := true
	for i := range out {
		if !equalZero(&out[i]) {
			isZero = false
			break
		}
	}
	if isZero {
		// TODO: it's [1] + [0] * (len(out) - 1) + [modulus - 1] in python, but strange it's 1 larger than out
		out[0] = ONE
		for i := 1; i < len(out); i++ {
			out[i] = ZERO
		}
		last := MODULUS_MINUS1
		out = append(out, last)
		return out
	} else {
		return out
	}
}

// TODO test unhappy case
const maxRecoverAttempts = 10

func (fs *FFTSettings) ErasureCodeRecover(vals []*Big) ([]Big, error) {
	// Generate the polynomial that is zero at the roots of unity
	// corresponding to the indices where vals[i] is None
	positions := make([]uint64, 0, len(vals))
	for i := uint64(0); i < uint64(len(vals)); i++ {
		if vals[i] == nil {
			positions = append(positions, i)
		}
	}
	z := fs._zPoly(positions, fs.maxWidth/uint64(len(vals)))
	//debugBigs("z", z)
	zVals, err := fs.FFT(z, false)
	if err != nil {
		return nil, err
	}
	//debugBigs("zvals", zVals)

	// Pointwise-multiply (vals filling in zero at missing spots) * z
	// By construction, this equals vals * z
	pTimesZVals := make([]Big, len(vals), len(vals))
	for i := uint(0); i < uint(len(vals)); i++ {
		if vals[i] == nil {
			// 0 * zVals[i] == 0
			pTimesZVals[i] = ZERO
		} else {
			mulModBig(&pTimesZVals[i], vals[i], &zVals[i])
		}
	}
	//debugBigs("p_times_z_vals", pTimesZVals)
	pTimesZ, err := fs.FFT(pTimesZVals, true)
	if err != nil {
		return nil, err
	}
	//debugBigs("p_times_z", pTimesZ)

	// Keep choosing k values until the algorithm does not fail
	// Check only with primitive roots of unity
	attempts := 0
	var kBig Big
	var tmp Big
	for k := uint64(2); attempts < maxRecoverAttempts; k++ {
		asBig(&kBig, k)
		// // TODO: implement this, translation of 'if pow(k, (modulus - 1) // 2, modulus) == 1:'
		//someOp(&tmp, &kBig)
		//if equalOne(&tmp) {
		//	continue
		//}
		var invk Big
		invModBig(&invk, &kBig)
		// Convert p_times_z(x) and z(x) into new polynomials
		// q1(x) = p_times_z(k*x) and q2(x) = z(k*x)
		// These are likely to not be 0 at any of the evaluation points.
		pTimesZOfKX := pOfKX(pTimesZ, &kBig)
		//debugBigs("p_times_z_of_kx", pTimesZOfKX)
		pTimesZOfKXVals, err := fs.FFT(pTimesZOfKX, false)
		if err != nil {
			return nil, err
		}
		//debugBigs("p_times_z_of_kx_vals", pTimesZOfKXVals)
		zOfKX := pOfKX(z, &kBig)
		//debugBigs("z_of_kx", zOfKX)
		zOfKXVals, err := fs.FFT(zOfKX, false)
		if err != nil {
			return nil, err
		}
		//debugBigs("z_of_kx_vals", zOfKXVals)

		// Compute q1(x) / q2(x) = p(k*x)
		invZOfKXVals := multiInv(zOfKXVals)
		//debugBigs("inv_z_of_kv_vals", invZOfKXVals)
		pOfKxVals := make([]Big, len(pTimesZOfKXVals), len(pTimesZOfKXVals))
		for i := 0; i < len(pOfKxVals); i++ {
			mulModBig(&pOfKxVals[i], &pTimesZOfKXVals[i], &invZOfKXVals[i])
		}
		//debugBigs("p_of_kx_vals", pOfKxVals)
		pOfKx, err := fs.FFT(pOfKxVals, true)
		if err != nil {
			return nil, err
		}
		//debugBigs("p_of_kx", pOfKx)

		// Given q3(x) = p(k*x), recover p(x)
		pOfX := make([]Big, len(pOfKx), len(pOfKx))
		if len(pOfKx) >= 1 {
			pOfX[0] = pOfKx[0]
		}
		if len(pOfKx) >= 2 {
			mulModBig(&pOfX[1], &pOfKx[1], &invk)
			invKPowI := invk
			for i := 2; i < len(pOfKx); i++ {
				CopyBigNum(&tmp, &invKPowI)
				mulModBig(&invKPowI, &tmp, &invk)
				mulModBig(&pOfX[i], &pOfKx[i], &invKPowI)
			}
		}
		output, err := fs.FFT(pOfX, false)
		if err != nil {
			return nil, err
		}

		// Check that the output matches the input
		success := true
		for i, inpd := range vals {
			if inpd == nil {
				continue
			}
			if !equalBig(inpd, &output[i]) {
				success = false
				break
			}
		}

		if !success {
			attempts += 1
			continue
		}
		// Output the evaluations if all good
		return output, nil
	}
	return nil, fmt.Errorf("max attempts reached: %d", attempts)
}
