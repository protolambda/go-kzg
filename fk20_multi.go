package kate

// FK20 Method to compute all proofs
// Toeplitz multiplication via http://www.netlib.org/utk/people/JackDongarra/etemplates/node384.html
// Multi proof method

// For a polynomial of size n, let w be a n-th root of unity. Then this method will return
// k=n/l KZG proofs for the points
//
// 	   proof[0]: w^(0*l + 0), w^(0*l + 1), ... w^(0*l + l - 1)
// 	   proof[0]: w^(0*l + 0), w^(0*l + 1), ... w^(0*l + l - 1)
// 	   ...
// 	   proof[i]: w^(i*l + 0), w^(i*l + 1), ... w^(i*l + l - 1)
// 	   ...
func (ks *KateSettings) FK20Multi() []G1 {
	// TODO
	return nil
}

// FK20 multi-proof method, optimized for dava availability where the top half of polynomial
// coefficients == 0
func (ks *KateSettings) FK20MultiDAOptimized() []G1 {
	// TODO
	return nil
}

// Computes all the KZG proofs for data availability checks. This involves sampling on the double domain
// and reordering according to reverse bit order
func (ks *KateSettings) DAUsingFK20Multi() []G1 {
	// TODO
	return nil
}
