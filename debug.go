package kzg

import (
	"fmt"
	"github.com/protolambda/go-kzg/bls"
	"strings"
)

func debugFrPtrs(msg string, values []*bls.Fr) {
	var out strings.Builder
	out.WriteString("---")
	out.WriteString(msg)
	out.WriteString("---\n")
	for i := range values {
		out.WriteString(fmt.Sprintf("#%4d: %s\n", i, bls.FrStr(values[i])))
	}
	fmt.Println(out.String())
}

func debugFrs(msg string, values []bls.Fr) {
	fmt.Println("---------------------------")
	var out strings.Builder
	for i := range values {
		out.WriteString(fmt.Sprintf("%s %d: %s\n", msg, i, bls.FrStr(&values[i])))
	}
	fmt.Print(out.String())
}
