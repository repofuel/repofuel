The result of the benchmarks as the following:

```
goos: darwin
goarch: amd64
pkg: https://github.com/repofuel/repofuel/ingest/pkg/classify/stemmers
BenchmarkStemmerMatch
BenchmarkStemmerMatch-8              	  484072	      2323 ns/op
BenchmarkStemmerMatchDchest
BenchmarkStemmerMatchDchest-8        	   16779	     67921 ns/op
BenchmarkStemmerMatchKljensen
BenchmarkStemmerMatchKljensen-8      	    7458	    153165 ns/op
BenchmarkStemmerMatchSurgebase
BenchmarkStemmerMatchSurgebase-8     	  106540	     11529 ns/op
BenchmarkStemmerMatchCaneroj1
BenchmarkStemmerMatchCaneroj1-8      	   39782	     27831 ns/op
BenchmarkStemmerMatchAgonopol
BenchmarkStemmerMatchAgonopol-8      	   44338	     26373 ns/op
BenchmarkStemmerMatchReiver
BenchmarkStemmerMatchReiver-8        	   53984	     22982 ns/op
BenchmarkStemmerOneWordDchest
BenchmarkStemmerOneWordDchest-8      	  542220	      2226 ns/op
BenchmarkStemmerOneWordKljensen
BenchmarkStemmerOneWordKljensen-8    	  451542	      2749 ns/op
BenchmarkStemmerOneWordSurgebase
BenchmarkStemmerOneWordSurgebase-8   	 4549485	       274 ns/op
BenchmarkStemmerOneWordCaneroj1
BenchmarkStemmerOneWordCaneroj1-8    	  679083	      1647 ns/op
BenchmarkStemmerOneWordAgonopol
BenchmarkStemmerOneWordAgonopol-8    	 2227963	       499 ns/op
BenchmarkStemmerOneWordReiver
BenchmarkStemmerOneWordReiver-8      	 3639223	       332 ns/op
PASS
```
