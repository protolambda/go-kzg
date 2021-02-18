# KZG and FFT utils

This repo is *super experimental*.

This is an implementation in Go, initially aimed at chunkification and extension of data, 
and building/verifying KZG proofs for the output data.
The KZG proofs, or Kate proofs, are built on top of BLS12-381.

Part of a low-latency data-availability sampling network prototype for Eth2 Phase 1.
See https://github.com/protolambda/eth2-das

Code is based on:
- [KZG Data availability code by Dankrad](https://github.com/ethereum/research/tree/master/kzg_data_availability)
- [Verkle and FFT code by Dankrad and Vitalik](https://github.com/ethereum/research/tree/master/verkle)
- [Reed solomon erasure code recovery with FFTs by Vitalik](https://ethresear.ch/t/reed-solomon-erasure-code-recovery-in-n-log-2-n-time-with-ffts/3039)
- [FFT explainer by Vitalik](https://vitalik.ca/general/2019/05/12/fft.html)
- [Kate explainer by Dankrad](https://dankradfeist.de/ethereum/2020/06/16/kate-polynomial-commitments.html)
- [Kate amortized paper by Dankrad and Dmitry](https://github.com/khovratovich/Kate/blob/master/Kate_amortized.pdf)

Features:
- (I)FFT on `F_r`
- (I)FFT on `G1`
- Specialized FFT for extension of `F_r` data
- KZG
  - commitments
  - generate/verify proof for single point
  - generate/verify proofs for multiple points
  - generate/verify proofs for all points, using FK20
  - generate/verify proofs for ranges (cosets) of points, using FK20
- Data recovery: given an arbitrary subset of data (at least half), recover the rest
- Optimized for Data-availability usage
- Change Fr / BLS with build tags.

## BLS

TODO: working with Herumi BLS currently as it exposes more functionality in Go API than BLST does. Still very limited compared to python.
The BLS functionality is generalized, and simple `G1` and `G2` types are exposed to use it. In the future different BLS libraries could be supported.

## Frs

The BLS curve order is used for the modulo math, different libraries could be used to provide this functionality.
Note: some of these libraries do not have full BLS functionality, only fr-num / uint256. The KZG code will be excluded when compiling with a non-BLS build tag.

Build tag options:
- ` ` (default, empty): use Herumi BLS library. Previously used by `Fr_hbls` build tag. [`herumi/bls-eth-go-binary`](https://github.com/herumi/bls-eth-go-binary/)
- `-tags Fr_kilic`: Use Kilic BLS library. [`kilic/bls12-381`](https://github.com/kilic/bls12-381)
- `-tags Fr_hol256`: Use the uint256 code that Geth uses, [`holiman/uint256`](https://github.com/holiman/uint256)
- `-tags Fr_pure`: Use the native Go Fr implementation.


## Benchmarks

### FFT benchmarks

Benchmarks on a `Intel Core i7-6700HQ @ 8x 3.5GHz` with `-test.benchtime=10s`:

#### FFT with F_r elements

Using Herumi BLS. A single FFT per operation, of `2**scale` elements. Random numbers of `0...modulus` as inputs.

```
BenchmarkFFTSettings_FFT/scale_4-8         	  777571	     14872 ns/op
BenchmarkFFTSettings_FFT/scale_5-8         	  365899	     35068 ns/op
BenchmarkFFTSettings_FFT/scale_6-8         	  153195	     78093 ns/op
BenchmarkFFTSettings_FFT/scale_7-8         	   69822	    173340 ns/op
BenchmarkFFTSettings_FFT/scale_8-8         	   32451	    372674 ns/op
BenchmarkFFTSettings_FFT/scale_9-8         	   14539	    814604 ns/op
BenchmarkFFTSettings_FFT/scale_10-8        	    6512	   1804612 ns/op
BenchmarkFFTSettings_FFT/scale_11-8        	    3088	   3739216 ns/op
BenchmarkFFTSettings_FFT/scale_12-8        	    1549	   8132423 ns/op
BenchmarkFFTSettings_FFT/scale_13-8        	     687	  17562389 ns/op
BenchmarkFFTSettings_FFT/scale_14-8        	     319	  38035214 ns/op
BenchmarkFFTSettings_FFT/scale_15-8        	     146	  79236893 ns/op
```

Note: about 1/3 of the computation time is CGO call overhead to the native Herumi BLS code.
When moving the KZG code into the native library, or reducing the overhead otherwise, the performance can improve significantly.

#### FFT with G1 points

Operation: A single FFT per operation, of `2**scale` elements. Random full G1 points, generator times random scalar.

Herumi BLS:
```
BenchmarkFFTSettings_FFTG1/scale_4-8               3612     3214052 ns/op
BenchmarkFFTSettings_FFTG1/scale_5-8               1524     7814302 ns/op
BenchmarkFFTSettings_FFTG1/scale_6-8                652    18545635 ns/op
BenchmarkFFTSettings_FFTG1/scale_7-8                282    43225687 ns/op
BenchmarkFFTSettings_FFTG1/scale_8-8                121    97727803 ns/op
BenchmarkFFTSettings_FFTG1/scale_9-8                 52   220462276 ns/op
BenchmarkFFTSettings_FFTG1/scale_10-8                24   486038684 ns/op
BenchmarkFFTSettings_FFTG1/scale_11-8                10  1069652806 ns/op
BenchmarkFFTSettings_FFTG1/scale_12-8                 5  2337882518 ns/op
BenchmarkFFTSettings_FFTG1/scale_13-8                 2  5058462998 ns/op
BenchmarkFFTSettings_FFTG1/scale_14-8                 1  10894399107 ns/op
BenchmarkFFTSettings_FFTG1/scale_15-8                 1  23526792397 ns/op
```

Kilic BLS:
```
BenchmarkFFTSettings_FFTG1/scale_4-8         	    2662	   4404800 ns/op
BenchmarkFFTSettings_FFTG1/scale_5-8         	    1173	  10214626 ns/op
BenchmarkFFTSettings_FFTG1/scale_6-8         	     501	  23857393 ns/op
BenchmarkFFTSettings_FFTG1/scale_7-8         	     218	  54990199 ns/op
BenchmarkFFTSettings_FFTG1/scale_8-8         	      93	 123677161 ns/op
BenchmarkFFTSettings_FFTG1/scale_9-8         	      42	 275718756 ns/op
BenchmarkFFTSettings_FFTG1/scale_10-8        	      19	 609369038 ns/op
BenchmarkFFTSettings_FFTG1/scale_11-8        	       8	1337226078 ns/op
BenchmarkFFTSettings_FFTG1/scale_12-8        	       4	2917276546 ns/op
BenchmarkFFTSettings_FFTG1/scale_13-8        	       2	6227439226 ns/op
BenchmarkFFTSettings_FFTG1/scale_14-8        	       1	12745157278 ns/op
BenchmarkFFTSettings_FFTG1/scale_15-8        	       1	27044222289 ns/op
```

#### Alternative FFT with F_r elements

Operation: A single FFT per operation, of `2**scale` elements. Random numbers of `0...modulus` as inputs.

With Kilic BLS:
```
BenchmarkFFTSettings_FFT/scale_4-8         	 2052406	      5829 ns/op
BenchmarkFFTSettings_FFT/scale_5-8         	  928363	     12689 ns/op
BenchmarkFFTSettings_FFT/scale_6-8         	  441000	     27934 ns/op
BenchmarkFFTSettings_FFT/scale_7-8         	  194581	     58963 ns/op
BenchmarkFFTSettings_FFT/scale_8-8         	   93363	    126508 ns/op
BenchmarkFFTSettings_FFT/scale_9-8         	   44704	    272013 ns/op
BenchmarkFFTSettings_FFT/scale_10-8        	   20990	    576005 ns/op
BenchmarkFFTSettings_FFT/scale_11-8        	    9052	   1227210 ns/op
BenchmarkFFTSettings_FFT/scale_12-8        	    4526	   2648435 ns/op
BenchmarkFFTSettings_FFT/scale_13-8        	    2076	   5731532 ns/op
BenchmarkFFTSettings_FFT/scale_14-8        	     979	  12402697 ns/op
BenchmarkFFTSettings_FFT/scale_15-8        	     445	  25884444 ns/op
```

With Holiman u256:

```
BenchmarkFFTSettings_FFT/scale_4-8         	  723585	     16249 ns/op
BenchmarkFFTSettings_FFT/scale_5-8         	  328123	     37372 ns/op
BenchmarkFFTSettings_FFT/scale_6-8         	  147979	     81936 ns/op
BenchmarkFFTSettings_FFT/scale_7-8         	   65496	    184643 ns/op
BenchmarkFFTSettings_FFT/scale_8-8         	   29562	    407406 ns/op
BenchmarkFFTSettings_FFT/scale_9-8         	   13485	    912428 ns/op
BenchmarkFFTSettings_FFT/scale_10-8        	    5881	   1986363 ns/op
BenchmarkFFTSettings_FFT/scale_11-8        	    2884	   4259827 ns/op
BenchmarkFFTSettings_FFT/scale_12-8        	    1320	   9258334 ns/op
BenchmarkFFTSettings_FFT/scale_13-8        	     558	  19882835 ns/op
BenchmarkFFTSettings_FFT/scale_14-8        	     288	  43007504 ns/op
BenchmarkFFTSettings_FFT/scale_15-8        	     130	  89666704 ns/op
```

And with Go native fr numbers (`fr.Int`):
```
BenchmarkFFTSettings_FFT/scale_4-8         	  276931	     42424 ns/op
BenchmarkFFTSettings_FFT/scale_5-8         	  116278	     96975 ns/op
BenchmarkFFTSettings_FFT/scale_6-8         	   54937	    229615 ns/op
BenchmarkFFTSettings_FFT/scale_7-8         	   23569	    494532 ns/op
BenchmarkFFTSettings_FFT/scale_8-8         	   10000	   1125187 ns/op
BenchmarkFFTSettings_FFT/scale_9-8         	    5050	   2468875 ns/op
BenchmarkFFTSettings_FFT/scale_10-8        	    2275	   5260237 ns/op
BenchmarkFFTSettings_FFT/scale_11-8        	    1048	  11334248 ns/op
BenchmarkFFTSettings_FFT/scale_12-8        	     480	  24877020 ns/op
BenchmarkFFTSettings_FFT/scale_13-8        	     223	  53423275 ns/op
BenchmarkFFTSettings_FFT/scale_14-8        	      99	 115540193 ns/op
BenchmarkFFTSettings_FFT/scale_15-8        	      48	 244955527 ns/op
```

#### Extension

Operation: do an extension of even values to odd values (even values are half the domain of the total).
Then next round applies the same function again, but to the output of the previous round.

`2**scale` is the *extended* width.

With Herumi BLS `Fr`:
```
BenchmarkFFTExtension/scale_4-8         	 2901286	      8197 ns/op
BenchmarkFFTExtension/scale_5-8         	 1231710	     19712 ns/op
BenchmarkFFTExtension/scale_6-8         	  532922	     46486 ns/op
BenchmarkFFTExtension/scale_7-8         	  227506	    104641 ns/op
BenchmarkFFTExtension/scale_8-8         	  100636	    246322 ns/op
BenchmarkFFTExtension/scale_9-8         	   43652	    557844 ns/op
BenchmarkFFTExtension/scale_10-8        	   19508	   1250852 ns/op
BenchmarkFFTExtension/scale_11-8        	    8983	   2725149 ns/op
BenchmarkFFTExtension/scale_12-8        	    4099	   6055065 ns/op
BenchmarkFFTExtension/scale_13-8        	    1818	  12882515 ns/op
BenchmarkFFTExtension/scale_14-8        	     838	  28595864 ns/op
BenchmarkFFTExtension/scale_15-8        	     385	  61239177 ns/op
```

With Kilic BLS `Fr`:
```
BenchmarkFFTExtension/scale_4-8         	 3357422	      7116 ns/op
BenchmarkFFTExtension/scale_5-8         	 2341398	     10264 ns/op
BenchmarkFFTExtension/scale_6-8         	 1368536	     17573 ns/op
BenchmarkFFTExtension/scale_7-8         	  709798	     33829 ns/op
BenchmarkFFTExtension/scale_8-8         	  336696	     70498 ns/op
BenchmarkFFTExtension/scale_9-8         	  156084	    152893 ns/op
BenchmarkFFTExtension/scale_10-8        	   70930	    337709 ns/op
BenchmarkFFTExtension/scale_11-8        	   32154	    751008 ns/op
BenchmarkFFTExtension/scale_12-8        	   14248	   1639920 ns/op
BenchmarkFFTExtension/scale_13-8        	    6691	   3616834 ns/op
BenchmarkFFTExtension/scale_14-8        	    3027	   7919893 ns/op
BenchmarkFFTExtension/scale_15-8        	    1384	  17248972 ns/op
```

With Holiman u256:
```
BenchmarkFFTExtension/scale_4-8         	 1548870	      7614 ns/op
BenchmarkFFTExtension/scale_5-8         	  637507	     18480 ns/op
BenchmarkFFTExtension/scale_6-8         	  273312	     44227 ns/op
BenchmarkFFTExtension/scale_7-8         	  114825	    102308 ns/op
BenchmarkFFTExtension/scale_8-8         	   50733	    233961 ns/op
BenchmarkFFTExtension/scale_9-8         	   22381	    532758 ns/op
BenchmarkFFTExtension/scale_10-8        	    9105	   1190005 ns/op
BenchmarkFFTExtension/scale_11-8        	    4131	   2655445 ns/op
BenchmarkFFTExtension/scale_12-8        	    1953	   5835835 ns/op
BenchmarkFFTExtension/scale_13-8        	     937	  12809577 ns/op
BenchmarkFFTExtension/scale_14-8        	     430	  27925924 ns/op
BenchmarkFFTExtension/scale_15-8        	     198	  60263842 ns/op
```

**Note**: extending using regular FFTs costs more than a single FFT (1 to convert to coeffs, then pad with zeroes, then one 2x the size inverse).
To extend to `2**15` with normal FFTs, with Herumi: `79236893*1.5 = ~ 0.118 seconds`. With specialized extension: `61239177 = ~ 0.0612 seconds`, cutting 48% of the cost.
Note that Kilic BLS is even faster: `0.0172` seconds! However, its performance with G1 FFTs is a little worse, so it may not always be worth switching to Kilic BLS. 

A lot of the difference in speed between Herumi and Kilic BLS comes from native Go vs. CGO: 
Each Mul/Add/Sub in Herumi comes with a CGO overhead cost. Porting the FFT code to Herumi internals, and only inducing the call-overhead once, would likely alleviate this difference. 

#### Zero polynomial

Using Kilic BLS, computing a zero polynomial for `N/2` points in a domain of `N` points, where `N = 2**scale`.

```
BenchmarkFFTSettings_ZeroPolyViaMultiplication/scale_5-8         	  582222	     19849 ns/op
BenchmarkFFTSettings_ZeroPolyViaMultiplication/scale_6-8         	  212700	     56463 ns/op
BenchmarkFFTSettings_ZeroPolyViaMultiplication/scale_7-8         	   32444	    370005 ns/op
BenchmarkFFTSettings_ZeroPolyViaMultiplication/scale_8-8         	   13376	    899227 ns/op
BenchmarkFFTSettings_ZeroPolyViaMultiplication/scale_9-8         	    5282	   2264945 ns/op
BenchmarkFFTSettings_ZeroPolyViaMultiplication/scale_10-8        	    2251	   5317543 ns/op
BenchmarkFFTSettings_ZeroPolyViaMultiplication/scale_11-8        	     922	  12907627 ns/op
BenchmarkFFTSettings_ZeroPolyViaMultiplication/scale_12-8        	     405	  29571684 ns/op
BenchmarkFFTSettings_ZeroPolyViaMultiplication/scale_13-8        	     170	  69992466 ns/op
BenchmarkFFTSettings_ZeroPolyViaMultiplication/scale_14-8        	      75	 160242911 ns/op
BenchmarkFFTSettings_ZeroPolyViaMultiplication/scale_15-8        	      32	 367837490 ns/op
```

#### Recover from samples

Using Kilic BLS, recovering a polynomial with `N/2` missing points, to get back `N` points, where `N = 2**scale`.

```
BenchmarkFFTSettings_RecoverPolyFromSamples/scale_5-8         	   33908	    345835 ns/op
BenchmarkFFTSettings_RecoverPolyFromSamples/scale_6-8         	   17180	    700476 ns/op
BenchmarkFFTSettings_RecoverPolyFromSamples/scale_7-8         	    7290	   1655556 ns/op
BenchmarkFFTSettings_RecoverPolyFromSamples/scale_8-8         	    3410	   3486237 ns/op
BenchmarkFFTSettings_RecoverPolyFromSamples/scale_9-8         	    1596	   7516292 ns/op
BenchmarkFFTSettings_RecoverPolyFromSamples/scale_10-8        	     748	  15906447 ns/op
BenchmarkFFTSettings_RecoverPolyFromSamples/scale_11-8        	     346	  34549928 ns/op
BenchmarkFFTSettings_RecoverPolyFromSamples/scale_12-8        	     162	  73824466 ns/op
BenchmarkFFTSettings_RecoverPolyFromSamples/scale_13-8        	      74	 160337719 ns/op
BenchmarkFFTSettings_RecoverPolyFromSamples/scale_14-8        	      33	 344798427 ns/op
BenchmarkFFTSettings_RecoverPolyFromSamples/scale_15-8        	      15	 747324287 ns/op
```

## License

MIT, see [`LICENSE`](./LICENSE) file.

