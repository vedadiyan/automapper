package mapper

import (
	"testing"

	"github.com/jinzhu/copier"
)

func BenchmarkAutomapper_Struct(b *testing.B) {
	src := &benchSourceValue

	b.ReportAllocs()

	converter := CodecFor[*benchSource, benchTarget]()

	for b.Loop() {
		result, err := converter(&src)
		if err != nil {
			b.Fatal(err)
		}
		_ = result
	}
}

func BenchmarkCopier_Struct(b *testing.B) {
	src := benchSourceValue

	b.ReportAllocs()

	for b.Loop() {
		var dst benchTarget

		if err := copier.Copy(&dst, &src); err != nil {
			b.Fatal(err)
		}
		_ = dst
	}
}

func BenchmarkAutomapper_NestedStruct(b *testing.B) {
	src := benchNestedValue

	b.ReportAllocs()

	for b.Loop() {
		result, err := ConvertFor[benchNestedSource, benchNestedTarget](&src)
		if err != nil {
			b.Fatal(err)
		}
		_ = result
	}
}

func BenchmarkCopier_NestedStruct(b *testing.B) {
	src := benchNestedValue

	b.ReportAllocs()

	for b.Loop() {
		var dst benchNestedTarget

		if err := copier.Copy(&dst, &src); err != nil {
			b.Fatal(err)
		}
		_ = dst
	}
}

func BenchmarkAutomapper_Slice(b *testing.B) {
	src := benchSliceValue

	b.ReportAllocs()

	for b.Loop() {
		result, err := ConvertFor[benchSliceSource, benchSliceTarget](&src)
		if err != nil {
			b.Fatal(err)
		}
		_ = result
	}
}

func BenchmarkCopier_Slice(b *testing.B) {
	src := benchSliceValue

	b.ReportAllocs()

	for b.Loop() {
		var dst benchSliceTarget

		if err := copier.Copy(&dst, &src); err != nil {
			b.Fatal(err)
		}
		_ = dst
	}
}

func BenchmarkAutomapper_Map(b *testing.B) {
	src := benchMapValue

	b.ReportAllocs()

	for b.Loop() {
		result, err := ConvertFor[benchMapSource, benchMapTarget](&src)
		if err != nil {
			b.Fatal(err)
		}
		_ = result
	}
}

func BenchmarkCopier_Map(b *testing.B) {
	src := benchMapValue

	b.ReportAllocs()

	for b.Loop() {
		var dst benchMapTarget

		if err := copier.Copy(&dst, &src); err != nil {
			b.Fatal(err)
		}
		_ = dst
	}
}
