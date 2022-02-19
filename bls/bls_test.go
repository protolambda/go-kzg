// +build !bignum_pure,!bignum_hol256

package bls

import (
	"bytes"
	"testing"
)

func TestPointCompression(t *testing.T) {
	var x Fr
	SetFr(&x, "44689111813071777962210527909085028157792767057343609826799812096627770269092")
	var point G1Point
	MulG1(&point, &GenG1, &x)
	got := ToCompressedG1(&point)

	expected := []byte{134, 87, 163, 76, 148, 138, 55, 228, 171, 85, 80, 116, 242, 13, 169, 151, 167, 6, 219, 183, 108, 254, 214, 99, 184, 231, 210, 201, 39, 69, 184, 188, 105, 194, 22, 32, 9, 57, 220, 81, 82, 164, 97, 236, 201, 116, 2, 83}

	if !bytes.Equal(expected, got) {
		t.Fatalf("Invalid compression result, %v != %x", got, expected)
	}
}

func TestPointG1Marshalling(t *testing.T) {
	var x Fr
	SetFr(&x, "44689111813071777962210527909085028157792767057343609826799812096627770269092")
	var point G1Point
	MulG1(&point, &GenG1, &x)

	bytes, err := point.MarshalText()
	if err != nil {
		t.Fatal(err)
	}

	var anotherPoint G1Point
	err = anotherPoint.UnmarshalText(bytes)
	if err != nil {
		t.Fatal(err)
	}

	if point != anotherPoint {
		t.Fatal("G1 points did not match")
	}
}

func TestPointG2Marshalling(t *testing.T) {
	var x Fr
	SetFr(&x, "44689111813071777962210527909085028157792767057343609826799812096627770269092")
	var point G2Point
	MulG2(&point, &GenG2, &x)

	bytes, err := point.MarshalText()
	if err != nil {
		t.Fatal(err)
	}

	var anotherPoint G2Point
	err = anotherPoint.UnmarshalText(bytes)
	if err != nil {
		t.Fatal(err)
	}

	if point != anotherPoint {
		t.Fatal("G2 points did not match")
	}
}
