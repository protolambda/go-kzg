package kzg

import (
	"fmt"
	"github.com/protolambda/go-kzg/bls"
	"strings"
)

func debugBigPtrs(msg string, values []*bls.Big) {
	var out strings.Builder
	out.WriteString("---")
	out.WriteString(msg)
	out.WriteString("---\n")
	for i := range values {
		out.WriteString(fmt.Sprintf("#%4d: %s\n", i, bls.BigStr(values[i])))
	}
	fmt.Println(out.String())
}

func debugBigs(msg string, values []bls.Big) {
	fmt.Println("---------------------------")
	var out strings.Builder
	for i := range values {
		out.WriteString(fmt.Sprintf("%s %d: %s\n", msg, i, bls.BigStr(&values[i])))
	}
	fmt.Print(out.String())
}
