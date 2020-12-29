// Original: https://github.com/ethereum/research/blob/master/mimc_stark/fft.py

package kate

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
