package mapper

import (
	"testing"
)

type benchSource struct {
	ID        int
	Name      string
	Age       int
	Active    bool
	Score     float64
	CreatedAt int64
}

type benchTarget struct {
	ID        int64
	Name      string
	Age       int64
	Active    bool
	Score     float64
	CreatedAt int64
}

type benchNestedSource struct {
	ID    int
	Name  string
	Child benchSource
}

type benchNestedTarget struct {
	ID    int64
	Name  string
	Child benchTarget
}

type benchSliceSource struct {
	Values []int
}

type benchSliceTarget struct {
	Values []int64
}

type benchMapSource struct {
	Values map[string]int
}

type benchMapTarget struct {
	Values map[string]int64
}

var (
	benchSourceValue = benchSource{
		ID:        123,
		Name:      "benchmark",
		Age:       30,
		Active:    true,
		Score:     99.5,
		CreatedAt: 1234567890,
	}

	benchSourcePointer = &benchSourceValue

	benchSliceValue = benchSliceSource{
		Values: []int{
			1, 2, 3, 4, 5,
			6, 7, 8, 9, 10,
		},
	}

	benchMapValue = benchMapSource{
		Values: map[string]int{
			"a": 1,
			"b": 2,
			"c": 3,
			"d": 4,
			"e": 5,
		},
	}

	benchNestedValue = benchNestedSource{
		ID:   123,
		Name: "nested",
		Child: benchSource{
			ID:        456,
			Name:      "child",
			Age:       20,
			Active:    true,
			Score:     88.8,
			CreatedAt: 987654321,
		},
	}
)

// -----------------------------------------------------------------------------
// Baseline
// -----------------------------------------------------------------------------

func BenchmarkBaselineStructCopy(b *testing.B) {
	src := benchSourceValue

	b.ReportAllocs()

	for b.Loop() {
		dst := benchTarget{
			ID:        int64(src.ID),
			Name:      src.Name,
			Age:       int64(src.Age),
			Active:    src.Active,
			Score:     src.Score,
			CreatedAt: src.CreatedAt,
		}

		_ = dst
	}
}

// -----------------------------------------------------------------------------
// ConvertFor
// -----------------------------------------------------------------------------

func BenchmarkConvertFor_Struct(b *testing.B) {
	src := benchSourcePointer

	b.ReportAllocs()

	for b.Loop() {
		dst, err := ConvertFor[benchSource, benchTarget](src)
		if err != nil {
			b.Fatal(err)
		}

		_ = dst
	}
}

func BenchmarkConvertFor_Struct_Value(b *testing.B) {
	src := benchSourceValue

	b.ReportAllocs()

	for b.Loop() {
		dst, err := ConvertFor[benchSource, benchTarget](&src)
		if err != nil {
			b.Fatal(err)
		}

		_ = dst
	}
}

func BenchmarkConvertFor_NestedStruct(b *testing.B) {
	src := &benchNestedValue

	b.ReportAllocs()

	for b.Loop() {
		dst, err := ConvertFor[benchNestedSource, benchNestedTarget](src)
		if err != nil {
			b.Fatal(err)
		}

		_ = dst
	}
}

// -----------------------------------------------------------------------------
// Convert
// -----------------------------------------------------------------------------

func BenchmarkConvert_Struct(b *testing.B) {
	src := ValueOf(benchSourceValue)

	b.ReportAllocs()

	for b.Loop() {
		dst, err := Convert(src, TypeFor[benchTarget]())
		if err != nil {
			b.Fatal(err)
		}

		_ = dst
	}
}

func BenchmarkConvert_SameType(b *testing.B) {
	src := ValueOf(benchSourceValue)

	b.ReportAllocs()

	for b.Loop() {
		dst, err := Convert(src, TypeFor[benchSource]())
		if err != nil {
			b.Fatal(err)
		}

		_ = dst
	}
}

// -----------------------------------------------------------------------------
// FastConvert
// -----------------------------------------------------------------------------

func BenchmarkFastConvert_SameLayout(b *testing.B) {
	src := ValueOf(benchSourceValue)

	b.ReportAllocs()

	for b.Loop() {
		dst, err := FastConvert(src, TypeFor[benchSource]())
		if err != nil {
			b.Fatal(err)
		}

		_ = dst
	}
}

func BenchmarkFastConvertFor_SameLayout(b *testing.B) {
	src := &benchSourceValue

	b.ReportAllocs()

	for b.Loop() {
		dst, err := FastConvertFor[benchSource, benchSource](src)
		if err != nil {
			b.Fatal(err)
		}

		_ = dst
	}
}

func BenchmarkFastConvertFor_DifferentLayout(b *testing.B) {
	src := &benchSourceValue

	b.ReportAllocs()

	for b.Loop() {
		dst, err := FastConvertFor[benchSource, benchTarget](src)
		if err != nil {
			b.Fatal(err)
		}

		_ = dst
	}
}

// -----------------------------------------------------------------------------
// Slice
// -----------------------------------------------------------------------------

func BenchmarkConvertFor_Slice(b *testing.B) {
	src := &benchSliceValue

	b.ReportAllocs()

	for b.Loop() {
		dst, err := ConvertFor[benchSliceSource, benchSliceTarget](src)
		if err != nil {
			b.Fatal(err)
		}

		_ = dst
	}
}

func BenchmarkConvertFor_Slice_1(b *testing.B) {
	src := &benchSliceSource{
		Values: []int{1},
	}

	b.ReportAllocs()

	for b.Loop() {
		dst, err := ConvertFor[benchSliceSource, benchSliceTarget](src)
		if err != nil {
			b.Fatal(err)
		}

		_ = dst
	}
}

func BenchmarkConvertFor_Slice_100(b *testing.B) {
	src := &benchSliceSource{
		Values: make([]int, 100),
	}

	for i := range src.Values {
		src.Values[i] = i
	}

	b.ReportAllocs()

	for b.Loop() {
		dst, err := ConvertFor[benchSliceSource, benchSliceTarget](src)
		if err != nil {
			b.Fatal(err)
		}

		_ = dst
	}
}

func BenchmarkConvertFor_Slice_1000(b *testing.B) {
	src := &benchSliceSource{
		Values: make([]int, 1000),
	}

	for i := range src.Values {
		src.Values[i] = i
	}

	b.ReportAllocs()

	for b.Loop() {
		dst, err := ConvertFor[benchSliceSource, benchSliceTarget](src)
		if err != nil {
			b.Fatal(err)
		}

		_ = dst
	}
}

// -----------------------------------------------------------------------------
// Map
// -----------------------------------------------------------------------------

func BenchmarkConvertFor_Map(b *testing.B) {
	src := &benchMapValue

	b.ReportAllocs()

	for b.Loop() {
		dst, err := ConvertFor[benchMapSource, benchMapTarget](src)
		if err != nil {
			b.Fatal(err)
		}

		_ = dst
	}
}

func BenchmarkConvertFor_Map_100(b *testing.B) {
	src := &benchMapSource{
		Values: make(map[string]int, 100),
	}

	for i := 0; i < 100; i++ {
		src.Values[string(rune('a'+i%26))+string(rune(i))] = i
	}

	b.ReportAllocs()

	for b.Loop() {
		dst, err := ConvertFor[benchMapSource, benchMapTarget](src)
		if err != nil {
			b.Fatal(err)
		}

		_ = dst
	}
}

// -----------------------------------------------------------------------------
// Reflection primitives
// -----------------------------------------------------------------------------

func BenchmarkValueOf(b *testing.B) {
	value := 123

	b.ReportAllocs()

	for b.Loop() {
		v := ValueOf(value)
		_ = v
	}
}

func BenchmarkTypeFor(b *testing.B) {
	b.ReportAllocs()

	for b.Loop() {
		typ := TypeFor[benchSource]()
		_ = typ
	}
}

func BenchmarkMemoryLayout(b *testing.B) {
	b.ReportAllocs()

	for b.Loop() {
		layout := TypeFor[benchSource]().MemoryLayout()
		_ = layout
	}
}

func BenchmarkFindConverter(b *testing.B) {
	b.ReportAllocs()

	for b.Loop() {
		converter, ok := FindConverter(
			TypeFor[int](),
			TypeFor[int64](),
		)

		_, _ = converter, ok
	}
}

// -----------------------------------------------------------------------------
// Assignment / conversion primitives
// -----------------------------------------------------------------------------

func BenchmarkTryAssign(b *testing.B) {
	source := ValueOf(123)
	target := New(TypeFor[int]()).Elem()

	b.ReportAllocs()

	for b.Loop() {
		ok, err := TryAssign(
			TypeFor[int](),
			TypeFor[int](),
			source,
			target,
		)

		if err != nil {
			b.Fatal(err)
		}

		if !ok {
			b.Fatal("assignment failed")
		}
	}
}

func BenchmarkTryConvert(b *testing.B) {
	source := ValueOf(123)
	target := New(TypeFor[int64]()).Elem()

	b.ReportAllocs()

	for b.Loop() {
		ok, err := TryConvert(
			TypeFor[int](),
			TypeFor[int64](),
			source,
			target,
		)

		if err != nil {
			b.Fatal(err)
		}

		if !ok {
			b.Fatal("conversion failed")
		}
	}
}

// -----------------------------------------------------------------------------
// Parallel benchmarks
// -----------------------------------------------------------------------------

func BenchmarkConvertFor_Struct_Parallel(b *testing.B) {
	src := &benchSourceValue

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			dst, err := ConvertFor[benchSource, benchTarget](src)
			if err != nil {
				b.Error(err)
				return
			}

			_ = dst
		}
	})
}

func BenchmarkConvertFor_Slice_Parallel(b *testing.B) {
	src := &benchSliceValue

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			dst, err := ConvertFor[benchSliceSource, benchSliceTarget](src)
			if err != nil {
				b.Error(err)
				return
			}

			_ = dst
		}
	})
}

func BenchmarkTypeFor_Parallel(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = TypeFor[benchSource]()
		}
	})
}

func BenchmarkMemoryLayout_Parallel(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = TypeFor[benchSource]().MemoryLayout()
		}
	})
}
