package testing

import (
	"fmt"
	"runtime"
	"time"
)

func header(s string) {
	_, file, line, _ := runtime.Caller(2)
	println(" ") // blank line
	println(" ") // blank line
	fmt.Println(s, file, line)
}

type T struct {
	common
}

type B struct {
	common
	N int
}

func (b *B) ReportAllocs()              {}
func (b *B) RunParallel(body func(*PB)) {}
func (b *B) SetBytes(n int64)           {}
func (b *B) SetParallelism(p int)       {}
func (b *B) StartTimer()                {}
func (b *B) StopTimer()                 {}
func (b *B) ResetTimer()                {}

type PB struct {
	// contains filtered or unexported fields
}

func (pb *PB) Next() bool { return false }

type common struct{}

func (c *common) Error(args ...interface{}) {
	header("Error")
	fmt.Println(args...)
	runtime.Breakpoint()
}

func (c *common) Errorf(format string, args ...interface{}) {
	header("Errorf")
	fmt.Printf(format, args...)
	runtime.Breakpoint()
}
func (c *common) Fatalf(format string, args ...interface{}) {
	header("Fatalf")
	fmt.Printf(format, args...)
	runtime.Breakpoint()
}
func (c *common) Logf(format string, args ...interface{}) {
	header("Logf")
	fmt.Printf(format, args...)
	runtime.Breakpoint()
}
func (c *common) Fail()        { header("Fail"); runtime.Breakpoint() }
func (c *common) FailNow()     { header("FailNow"); runtime.Breakpoint() }
func (c *common) Failed() bool { return false }
func (c *common) Fatal(args ...interface{}) {
	header("Fatal")
	fmt.Println(args...)
	runtime.Breakpoint()
}
func (c *common) Log(args ...interface{})  { header("Log"); fmt.Println(args...); runtime.Breakpoint() }
func (t *common) Parallel()                {}
func (c *common) Skip(args ...interface{}) { header("Skip"); fmt.Println(args...); runtime.Breakpoint() }
func (c *common) SkipNow()                 { header("SkipNow"); runtime.Breakpoint() }
func (c *common) Skipf(format string, args ...interface{}) {
	header("Skipf")
	fmt.Printf(format, args...)
	runtime.Breakpoint()
}
func (c *common) Skipped() bool { return false }

func Short() bool   { return true }
func Verbose() bool { return false }

type InternalTest struct {
	Name string
	F    func(*T)
}
type InternalBenchmark InternalTest
type InternalExample InternalTest

func AllocsPerRun(runs int, f func()) (avg float64) { return 0 }

type BenchmarkResult struct {
	N         int           // The number of iterations.
	T         time.Duration // The total time taken.
	Bytes     int64         // Bytes processed in one iteration.
	MemAllocs uint64        // The total number of memory allocations.
	MemBytes  uint64        // The total number of bytes allocated.
}

func Benchmark(f func(b *B)) BenchmarkResult { return BenchmarkResult{} }

func (r BenchmarkResult) AllocedBytesPerOp() int64 { return 0 }

func (r BenchmarkResult) AllocsPerOp() int64 { return 0 }

func (r BenchmarkResult) MemString() string { return "" }

func (r BenchmarkResult) NsPerOp() int64 { return 0 }

func (r BenchmarkResult) String() string { return "" }

// An internal function but exported because it is cross-package; part of the implementation
// of the "go test" command.
func Main(matchString func(pat, str string) (bool, error), tests []InternalTest, benchmarks []InternalBenchmark, examples []InternalExample) {
	//os.Exit(MainStart(matchString, tests, benchmarks, examples).Run())
	fmt.Println("testing.Main")
	runtime.UnzipTestFS()
	var t T
	for _, f := range tests {
		fmt.Println(f.Name)
		f.F(&t)
	}
}

func init() {
	//var t = T{}
	//if false {
	//	Main(nil, nil, nil, nil)
	//	t.Error("...")
	//	t.Errorf("format", "...")
	//}
}
