package kate

// For each subset in `subsets` (provided as a list of indices into `numbers`),
// compute the sum of that subset of `numbers`. More efficient than the naive method.
func MultiSubsetBig(numbers []Big, subsets []Big) {
	// TODO
}

// Alternative algorithm. Less optimal than the above, but much lower bit twiddling
// overhead and much simpler.
func MultiSubset2Big(numbers []Big, subsets []Big) {
	// TODO
}

// For each subset in `subsets` (provided as a list of indices into `numbers`),
// compute the sum of that subset of `numbers`. More efficient than the naive method.
func MultiSubsetG1(numbers []G1, subsets []Big) {
	// TODO
}

// Alternative algorithm. Less optimal than the above, but much lower bit twiddling
// overhead and much simpler.
func MultiSubset2G1(numbers []G1, subsets []Big) {
	// TODO
}

// Reduces a linear combination `numbers[0] * factors[0] + numbers[1] * factors[1] + ...`
// into a multi-subset problem, and computes the result efficiently
func LinCombBigWithSubsets(numbers []Big, factors []Big) *Big {
	// TODO nil
	return nil
}

func LinCombG1WithSubsets(numbers []G1, factors []Big) *G1 {
	// TODO
	return nil
}
