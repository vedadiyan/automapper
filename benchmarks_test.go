package automapper

import "testing"

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

func BenchmarkAutomapper_Struct(b *testing.B) {
	src := benchSourceValue

	converter := CreateCodecFor[benchSource, benchTarget]()

	b.ReportAllocs()

	for b.Loop() {
		result, err := converter(&src)
		if err != nil {
			b.Fatal(err)
		}
		_ = result
	}
}

func BenchmarkAutomapper_NestedStruct(b *testing.B) {
	src := benchNestedValue

	converter := CreateCodecFor[benchNestedSource, benchNestedTarget]()

	b.ReportAllocs()

	for b.Loop() {
		result, err := converter(&src)
		if err != nil {
			b.Fatal(err)
		}
		_ = result
	}
}

func BenchmarkAutomapper_Slice(b *testing.B) {
	src := benchSliceValue

	converter := CreateCodecFor[benchSliceSource, benchSliceTarget]()

	b.ReportAllocs()

	for b.Loop() {
		result, err := converter(&src)
		if err != nil {
			b.Fatal(err)
		}
		_ = result
	}
}

func BenchmarkAutomapper_Map(b *testing.B) {
	src := benchMapValue

	converter := CreateCodecFor[benchMapSource, benchMapTarget]()

	b.ReportAllocs()

	for b.Loop() {
		result, err := converter(&src)
		if err != nil {
			b.Fatal(err)
		}
		_ = result
	}
}
