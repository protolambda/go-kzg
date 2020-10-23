package go_verkle

import (
	"math/big"
	"testing"
)

func TestFFT(t *testing.T) {
	data := make([]Big, WIDTH, WIDTH)
	for i := 0; i < WIDTH; i++ {
		data[i] = Big(big.NewInt(int64(i)))
	}
	debugBigs("input data", data)
	coeffs := FFT(data, MODULUS, ROOT_OF_UNITY, true)
	debugBigs("coeffs", coeffs)
	//evaluations := evalPolyRange(coeffs, bigRange(0, WIDTH))
	//debugBigs("eval 0...N", evaluations)
	//extension := evalPolyRange(coeffs, bigRange(WIDTH, WIDTH*2))
	//debugBigs("eval N...2N", extension)
	//
	//back := FFT(evaluations, MODULUS, ROOT_OF_UNITY, true)
	//debugBigs("back", back)
	//out2 := FFT(out, MODULUS, ROOT_OF_UNITY, true)
	//debugBigs("inv", out2)
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

func TestPolyRange(t *testing.T) {
	coeffs := bigRange(4, 8)
	out := evalPolyAt(coeffs, asBig(3))
	// 4*(3^0) + 5*(3^1) + 6*(3^2) + 7*(3^3) = 262
	if cmpBig(out, asBig(262)) != 0 {
		t.Fatalf("bad result: %s", bigStr(out))
	}
}
