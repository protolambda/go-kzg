module github.com/protolambda/go-kzg

go 1.15

require (
	github.com/herumi/bls-eth-go-binary v0.0.0-20210302070600-dfaa902c7773
	github.com/holiman/uint256 v1.1.1
	github.com/kilic/bls12-381 v0.1.1-0.20210208205449-6045b0235e36
	github.com/supranational/blst v0.3.3-0.20210305121809-81cd381e23cd
	golang.org/x/sys v0.0.0-20210305034016-7844c3c200c3 // indirect
)

replace (
	github.com/supranational/blst v0.3.3-0.20210305121809-81cd381e23cd => ../blst
)
