package mapper

import (
	"testing"

	"github.com/dranikpg/dto-mapper"
	"github.com/jinzhu/copier"
)

func BenchmarkAutomapper_Struct(b *testing.B) {
	src := benchSourceValue

	b.ReportAllocs()

	for b.Loop() {
		result, err := ConvertFor[benchSource, benchTarget](&src)
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

func BenchmarkDTOMapper_Struct(b *testing.B) {
	src := benchSourceValue

	b.ReportAllocs()

	for b.Loop() {
		var dst benchTarget

		if err := dto.Map(&dst, &src); err != nil {
			b.Fatal(err)
		}

		_ = dst
	}
}

func BenchmarkAutomapper_NestedStruct(b *testing.B) {
	src := benchNestedValue

	b.ReportAllocs()

	for b.Loop() {
		var dst benchNestedTarget

		result, err := ConvertFor[benchNestedSource, benchNestedTarget](&src)
		if err != nil {
			b.Fatal(err)
		}

		dst = *result
		_ = dst
	}
}

func BenchmarkAutomapper_Slice(b *testing.B) {
	src := benchSliceValue

	b.ReportAllocs()

	for b.Loop() {
		var dst benchSliceTarget

		result, err := ConvertFor[benchSliceSource, benchSliceTarget](&src)
		if err != nil {
			b.Fatal(err)
		}

		dst = *result
		_ = dst
	}
}

func BenchmarkAutomapper_Map(b *testing.B) {
	src := benchMapValue

	b.ReportAllocs()

	for b.Loop() {
		var dst benchMapTarget

		result, err := ConvertFor[benchMapSource, benchMapTarget](&src)
		if err != nil {
			b.Fatal(err)
		}

		dst = *result
		_ = dst
	}
}

// Copier equivalents, kept here so both use the exact same benchmark setup.

func BenchmarkCopier_NestedStruct_Fair(b *testing.B) {
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

func BenchmarkCopier_Slice_Fair(b *testing.B) {
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

func BenchmarkCopier_Map_Fair(b *testing.B) {
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
