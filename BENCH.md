
```shell
mkdir -p out/benches
go test -bench=. -run=^Bench -tags=bignum_hbls -count=1 -benchtime=2s ./... > out/benches/hbls.txt
go test -bench=. -run=^Bench -tags=bignum_kilic -count=1 -benchtime=2s ./... > out/benches/kilic.txt
go test -bench=. -run=^Bench -tags=bignum_pure -count=1 -benchtime=2s ./... > out/benches/pure.txt
go test -bench=. -run=^Bench -tags=bignum_hol256 -count=1 -benchtime=2s ./... > out/benches/hol256.txt
```

Benchmarks:
- FFT: A single FFT per operation, of `2**scale` elements. Random numbers of `0...modulus` as inputs.
- FFTG1: A single FFT per operation, of `2**scale` elements. Random full G1 points, generator times random scalar.
- FFTExtension: Do an extension of even values to odd values (even values are half the domain of the total).
  Then next round applies the same function again, but to the output of the previous round.
- RecoverPolyFromSamples: recover a polynomial with `N/2` missing points, to get back `N` points, where `N = 2**scale`.
- ZeroPolyViaMultiplication: compute a zero polynomial for `N/2` points in a domain of `N` points, where `N = 2**scale`.

Benchmarks on a `AMD Ryzen 9 5950X 16-Core @ 32x 3.4GHz`:

**all numbers in `ns/op`, lower is better**
```
Benchmark                                   Herumi BLS     Kilic BLS  math/big.Int  Hol. uint256
FFTExtension/scale_4                              6943          6436         18120          6808
FFTExtension/scale_5                             16354          9426         46208         13154
FFTExtension/scale_6                             36731         14649        114624         26863
FFTExtension/scale_7                             79517         27099        268413         53251
FFTExtension/scale_8                            187907         50896        599219        121535
FFTExtension/scale_9                            448536        108707       1430588        263396
FFTExtension/scale_10                           933580        231713       3142765        584304
FFTExtension/scale_11                          1842469        516664       6680015       1304376
FFTExtension/scale_12                          4351600       1169011      15065928       2879859
FFTExtension/scale_13                          9669196       2569475      32051722       6245514
FFTExtension/scale_14                         21868114       5421951      69237883      13680838
FFTExtension/scale_15                         44655458      11377382     131074970      29098188
FFT/scale_4                                      12176          3991         35851         12505
FFT/scale_5                                      27535          8383         83080         28774
FFT/scale_6                                      66410         19445        186216         66960
FFT/scale_7                                     140656         40446        413013        140888
FFT/scale_8                                     287225         87280        942126        288167
FFT/scale_9                                     646755        197884       1998237        673450
FFT/scale_10                                   1404484        417888       4450796       1384810
FFT/scale_11                                   2735691        881049       9392908       3009116
FFT/scale_12                                   6183629       1911871      20022278       6784383
FFT/scale_13                                  14290446       3891241      41155551      14660930
FFT/scale_14                                  28762232       8331212      83200388      28077789
FFT/scale_15                                  41157878      15442864     168285554      45337024
FFTG1/scale_4                                  1900117       4592074             -             -
FFTG1/scale_5                                  4524605      11496429             -             -
FFTG1/scale_6                                 10362223      28781767             -             -
FFTG1/scale_7                                 25088604      64030299             -             -
FFTG1/scale_8                                 57022201     148535217             -             -
FFTG1/scale_9                                125807653     305014341             -             -
FFTG1/scale_10                               281827965     760028615             -             -
FFTG1/scale_11                               624146434    1684792328             -             -
FFTG1/scale_12                              1379302210    3745748396             -             -
FFTG1/scale_13                              2921717660    7411640027             -             -
FFTG1/scale_14                              6351155786   12106541411             -             -
FFTG1/scale_15                             13513868449   21183031053             -             -
RecoverPolyFromSamples/scale_5                  243687        255459        698860        333387
RecoverPolyFromSamples/scale_6                  585075        571952       1594815        722082
RecoverPolyFromSamples/scale_7                 1774233       1299551       4763826       2008941
RecoverPolyFromSamples/scale_8                 4013121       2815613      11241467       4467860
RecoverPolyFromSamples/scale_9                 9442010       5835441      26532183      10327366
RecoverPolyFromSamples/scale_10               18958002      12809586      56971963      21640572
RecoverPolyFromSamples/scale_11               43917746      25176204     140140528      48933953
RecoverPolyFromSamples/scale_12              109026283      50779730     300112653      99573750
RecoverPolyFromSamples/scale_13              227823950     114070337     674811329     209266978
RecoverPolyFromSamples/scale_14              422747090     199684810    1355134266     419589619
RecoverPolyFromSamples/scale_15              729376022     425497194    2727187583     792483322
ZeroPolyViaMultiplication/scale_5                37923         12411        124084         42023
ZeroPolyViaMultiplication/scale_6               102724         34445        436629        127479
ZeroPolyViaMultiplication/scale_7               660084        222416       2244591        611469
ZeroPolyViaMultiplication/scale_8              1696890        564170       5508568       1354510
ZeroPolyViaMultiplication/scale_9              3843737       1418044      14285246       3228628
ZeroPolyViaMultiplication/scale_10            10975127       3394786      32161403       7799145
ZeroPolyViaMultiplication/scale_11            24890603       7670790      81190701      17558297
ZeroPolyViaMultiplication/scale_12            53912184      18257011     183405378      42517099
ZeroPolyViaMultiplication/scale_13           146484921      41847868     440848135      96153861
ZeroPolyViaMultiplication/scale_14           288853279      83304452     872245124     217806182
ZeroPolyViaMultiplication/scale_15           515863108     172534656    1772161612     434451293
```
