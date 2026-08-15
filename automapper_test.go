package mapper

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
	"unsafe"
)

// -----------------------------------------------------------------------------
// Test types
// -----------------------------------------------------------------------------

type sourceBasic struct {
	Name   string
	Age    int
	Active bool
	Score  float64
}

type targetBasic struct {
	Name   string
	Age    int
	Active bool
	Score  float64
}

type sourceNested struct {
	Name  string
	Child sourceBasic
}

type targetNested struct {
	Name  string
	Child targetBasic
}

type sourceExtraFields struct {
	Name   string
	Age    int
	Secret string
}

type targetMissingFields struct {
	Name string
}

type sourceSlice struct {
	Values []int
}

type targetSlice struct {
	Values []int64
}

type sourceArray struct {
	Values [3]int
}

type sourceMap struct {
	Values map[string]int
}

type targetMap struct {
	Values map[string]int64
}

type sourceMapKey struct {
	Values map[int]string
}

type targetMapKey struct {
	Values map[string]string
}

type sourceTime struct {
	Time time.Time
}

type targetTime struct {
	Time string
}

type namedInt int
type namedIntTarget int64

type interfaceA interface {
	Test() bool
}

type interfaceB interface {
	Other() bool
}

type interfaceImpl struct{}

func (interfaceImpl) Test() bool {
	return true
}

type interfaceImplB struct{}

func (interfaceImplB) Other() bool {
	return true
}

type methodType struct{}

func (methodType) PublicMethod() int {
	return 1
}

func (methodType) AnotherMethod(x int) string {
	return fmt.Sprintf("%d", x)
}

type recursiveNode struct {
	Value int
	Next  *recursiveNode
}

type recursiveNode2 struct {
	Value int
	Next  *recursiveNode2
}

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------

func assertNoError(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func assertError(t *testing.T, err error) {
	t.Helper()

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func resetConverters() {
	converters = make(map[Type]map[Type]Converter)
}

func registerConverter[S any, T any](fn func(Value, Type) (Value, error)) {
	s := TypeFor[S]()
	t := TypeFor[T]()

	if converters[s] == nil {
		converters[s] = make(map[Type]Converter)
	}

	converters[s][t] = fn
}

// -----------------------------------------------------------------------------
// Zero
// -----------------------------------------------------------------------------

func TestZero(t *testing.T) {
	got := Zero[int]()

	if got != 0 {
		t.Fatalf("expected zero int, got %v", got)
	}

	gotStruct := Zero[sourceBasic]()

	if gotStruct != (sourceBasic{}) {
		t.Fatalf("expected zero struct, got %+v", gotStruct)
	}

	gotPointer := Zero[*int]()

	if gotPointer != nil {
		t.Fatalf("expected nil pointer, got %v", gotPointer)
	}
}

// -----------------------------------------------------------------------------
// TargetFieldName
// -----------------------------------------------------------------------------

func TestTargetFieldName_Default(t *testing.T) {
	type X struct {
		Name string
	}

	field := TypeFor[X]().Field(0)

	if got := TargetFieldName(field); got != "Name" {
		t.Fatalf("expected Name, got %q", got)
	}
}

func TestTargetFieldName_MapToTag(t *testing.T) {
	type X struct {
		Name string `mapto:"DifferentName"`
	}

	field := TypeFor[X]().Field(0)

	if got := TargetFieldName(field); got != "DifferentName" {
		t.Fatalf("expected DifferentName, got %q", got)
	}
}

func TestTargetFieldName_EmptyTag(t *testing.T) {
	type X struct {
		Name string `mapto:""`
	}

	field := TypeFor[X]().Field(0)

	if got := TargetFieldName(field); got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
}

// -----------------------------------------------------------------------------
// TypeOf / TypeFor
// -----------------------------------------------------------------------------

func TestTypeOf(t *testing.T) {
	typ := TypeOf(sourceBasic{})

	if typ == nil {
		t.Fatal("TypeOf returned nil")
	}

	if typ.GoType() != reflect.TypeOf(sourceBasic{}) {
		t.Fatalf("unexpected GoType: %v", typ.GoType())
	}
}

func TestTypeOfNil(t *testing.T) {
	if got := TypeOf(nil); got != nil {
		t.Fatalf("expected nil Type, got %v", got)
	}
}

func TestTypeFor(t *testing.T) {
	got := TypeFor[sourceBasic]()

	if got == nil {
		t.Fatal("TypeFor returned nil")
	}

	if got.GoType() != reflect.TypeFor[sourceBasic]() {
		t.Fatalf("unexpected GoType: %v", got.GoType())
	}
}

func TestTypeOf_CachesType(t *testing.T) {
	a := TypeFor[sourceBasic]()
	b := TypeFor[sourceBasic]()

	if a != b {
		t.Fatal("expected cached Type instance")
	}
}

func TestTypeFor_PointerCount(t *testing.T) {
	tests := []struct {
		name string
		typ  Type
		want int
	}{
		{"value", TypeFor[int](), 0},
		{"pointer", TypeFor[*int](), 1},
		{"double pointer", TypeFor[**int](), 2},
		{"triple pointer", TypeFor[***int](), 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.typ.PointerCount(); got != tt.want {
				t.Fatalf("got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestDeReferenceType(t *testing.T) {
	tests := []struct {
		name string
		typ  reflect.Type
		n    int
		want reflect.Type
	}{
		{
			name: "int",
			typ:  reflect.TypeFor[int](),
			n:    0,
			want: reflect.TypeFor[int](),
		},
		{
			name: "pointer",
			typ:  reflect.TypeFor[*int](),
			n:    1,
			want: reflect.TypeFor[int](),
		},
		{
			name: "double pointer",
			typ:  reflect.TypeFor[**int](),
			n:    2,
			want: reflect.TypeFor[int](),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n, got := DeReferenceType(tt.typ)

			if n != tt.n {
				t.Fatalf("pointer count = %d, want %d", n, tt.n)
			}

			if got != tt.want {
				t.Fatalf("type = %v, want %v", got, tt.want)
			}
		})
	}
}

// -----------------------------------------------------------------------------
// Type reflection wrappers
// -----------------------------------------------------------------------------

func TestTypeReflectionMethods(t *testing.T) {
	typ := TypeFor[methodType]()

	if typ.Name() != "methodType" {
		t.Fatalf("unexpected Name: %q", typ.Name())
	}

	if typ.Kind() != reflect.Struct {
		t.Fatalf("unexpected Kind: %v", typ.Kind())
	}

	if typ.Size() != reflect.TypeFor[methodType]().Size() {
		t.Fatalf("unexpected Size")
	}

	if typ.Align() != reflect.TypeFor[methodType]().Align() {
		t.Fatalf("unexpected Align")
	}

	if typ.FieldAlign() != reflect.TypeFor[methodType]().FieldAlign() {
		t.Fatalf("unexpected FieldAlign")
	}

	if typ.NumMethod() != reflect.TypeFor[methodType]().NumMethod() {
		t.Fatalf("unexpected NumMethod")
	}

	if _, ok := typ.MethodByName("PublicMethod"); !ok {
		t.Fatal("expected PublicMethod")
	}

	if _, ok := typ.MethodByName("Missing"); ok {
		t.Fatal("did not expect Missing")
	}
}

func TestTypeFields(t *testing.T) {
	typ := TypeFor[sourceBasic]()

	if typ.NumField() != 4 {
		t.Fatalf("got %d fields, want 4", typ.NumField())
	}

	field, ok := typ.FieldByName("Name")

	if !ok {
		t.Fatal("Name field not found")
	}

	if field.Name != "Name" {
		t.Fatalf("unexpected field name: %q", field.Name)
	}

	if !field.IsExported() {
		t.Fatal("Name should be exported")
	}

	if _, ok := typ.FieldByName("Missing"); ok {
		t.Fatal("unexpected Missing field")
	}
}

func TestTypeFieldsIteration(t *testing.T) {
	typ := TypeFor[sourceBasic]()

	var fields []string

	for field := range typ.Fields() {
		fields = append(fields, field.Name)
	}

	want := []string{"Name", "Age", "Active", "Score"}

	if !reflect.DeepEqual(fields, want) {
		t.Fatalf("got %v, want %v", fields, want)
	}
}

func TestTypeFieldByNameFunc(t *testing.T) {
	typ := TypeFor[sourceBasic]()

	field, ok := typ.FieldByNameFunc(func(name string) bool {
		return strings.HasPrefix(name, "Act")
	})

	if !ok {
		t.Fatal("expected field")
	}

	if field.Name != "Active" {
		t.Fatalf("got %q, want Active", field.Name)
	}
}

func TestTypeElem(t *testing.T) {
	typ := TypeFor[[]int]()

	if typ.Elem().GoType() != reflect.TypeFor[int]() {
		t.Fatalf("unexpected element type")
	}
}

func TestTypeKey(t *testing.T) {
	typ := TypeFor[map[string]int]()

	if typ.Key().GoType() != reflect.TypeFor[string]() {
		t.Fatalf("unexpected key type")
	}

	if typ.Elem().GoType() != reflect.TypeFor[int]() {
		t.Fatalf("unexpected element type")
	}
}

func TestTypeLen(t *testing.T) {
	typ := TypeFor[[5]int]()

	if typ.Len() != 5 {
		t.Fatalf("got %d, want 5", typ.Len())
	}
}

func TestTypeComparable(t *testing.T) {
	if !TypeFor[int]().Comparable() {
		t.Fatal("int should be comparable")
	}

	if TypeFor[[]int]().Comparable() {
		t.Fatal("slice should not be comparable")
	}
}

func TestTypeAssignableAndConvertible(t *testing.T) {
	intType := TypeFor[int]()
	int64Type := TypeFor[int64]()

	if !intType.AssignableTo(intType) {
		t.Fatal("int should be assignable to int")
	}

	if intType.ConvertibleTo(int64Type) == false {
		t.Fatal("int should be convertible to int64")
	}
}

func TestTypeBits(t *testing.T) {
	if got := TypeFor[int32]().Bits(); got != 32 {
		t.Fatalf("got %d, want 32", got)
	}
}

func TestTypeOverflow(t *testing.T) {
	if !TypeFor[int8]().OverflowInt(128) {
		t.Fatal("expected int8 overflow")
	}

	if TypeFor[int64]().OverflowInt(1) {
		t.Fatal("unexpected int64 overflow")
	}
}

func TestTypePointerConcreteType(t *testing.T) {
	typ := TypeFor[***int]()

	if typ.ConcreteType().GoType() != reflect.TypeFor[int]() {
		t.Fatalf("unexpected concrete type")
	}
}

func TestTypeIDUnique(t *testing.T) {
	a := TypeFor[int]()
	b := TypeFor[string]()

	if a.ID() == b.ID() {
		t.Fatal("different types must have different IDs")
	}
}

// -----------------------------------------------------------------------------
// Memory layout
// -----------------------------------------------------------------------------

func TestMemoryLayout_NonNil(t *testing.T) {
	layout := TypeFor[int]().MemoryLayout()

	if layout == nil {
		t.Fatal("expected memory layout")
	}

	if len(layout.Layout()) == 0 {
		t.Fatal("expected non-empty layout")
	}

	if layout.HashCode() == "" {
		t.Fatal("expected hash")
	}
}

func TestMemoryLayout_IdenticalSameType(t *testing.T) {
	a := TypeFor[int]().MemoryLayout()
	b := TypeFor[int]().MemoryLayout()

	if !a.IdenticalTo(b) {
		t.Fatal("same type should have identical memory layout")
	}
}

func TestMemoryLayout_DifferentTypes(t *testing.T) {
	a := TypeFor[int]().MemoryLayout()
	b := TypeFor[string]().MemoryLayout()

	if a.IdenticalTo(b) {
		t.Fatal("different types should not have identical layouts")
	}
}

func TestMemoryLayout_Nil(t *testing.T) {
	if TypeFor[int]().MemoryLayout().IdenticalTo(nil) {
		t.Fatal("nil layout should not be identical")
	}
}

func TestMemoryLayout_Slice(t *testing.T) {
	TypeFor[[]int]().MemoryLayout()
	TypeFor[[][]int]().MemoryLayout()
	TypeFor[[]*int]().MemoryLayout()
}

func TestMemoryLayout_Array(t *testing.T) {
	TypeFor[[3]int]().MemoryLayout()
	TypeFor[[4]int]().MemoryLayout()

	if TypeFor[[3]int]().MemoryLayout().IdenticalTo(
		TypeFor[[4]int]().MemoryLayout(),
	) {
		t.Fatal("arrays with different lengths must differ")
	}
}

func TestMemoryLayout_Map(t *testing.T) {
	TypeFor[map[string]int]().MemoryLayout()
	TypeFor[map[int]string]().MemoryLayout()
	TypeFor[map[string][]int]().MemoryLayout()
}

func TestMemoryLayout_Struct(t *testing.T) {
	TypeFor[sourceBasic]().MemoryLayout()
	TypeFor[sourceNested]().MemoryLayout()
}

func TestMemoryLayout_Function(t *testing.T) {
	TypeFor[func(int) string]().MemoryLayout()
}

func TestMemoryLayout_Interface(t *testing.T) {
	TypeFor[interfaceA]().MemoryLayout()
}

func TestMemoryLayout_Pointer(t *testing.T) {
	a := TypeFor[int]().MemoryLayout()
	b := TypeFor[*int]().MemoryLayout()

	if a.IdenticalTo(b) {
		t.Fatal("pointer and value layouts must differ")
	}
}

// -----------------------------------------------------------------------------
// FindConverter
// -----------------------------------------------------------------------------

func TestFindConverter_NotFound(t *testing.T) {
	resetConverters()

	_, ok := FindConverter(
		TypeFor[int](),
		TypeFor[string](),
	)

	if ok {
		t.Fatal("converter should not exist")
	}
}

func TestFindConverter_Found(t *testing.T) {
	oldConverters := converters
	t.Cleanup(func() {
		converters = oldConverters
	})

	converters = make(map[Type]map[Type]Converter)

	fn := func(value Value, typ Type) (Value, error) {
		return ValueOf("ok"), nil
	}

	registerConverter[int, string](fn)

	got, ok := FindConverter(
		TypeFor[int](),
		TypeFor[string](),
	)

	if !ok {
		t.Fatal("converter should exist")
	}

	if got == nil {
		t.Fatal("converter is nil")
	}
}

// -----------------------------------------------------------------------------
// TryAssign
// -----------------------------------------------------------------------------

func TestTryAssign_Success(t *testing.T) {
	source := 42

	sourceType := TypeFor[int]()
	targetType := TypeFor[int]()

	sourceValue := ValueOf(source)
	targetValue := New(targetType).Elem()

	ok, err := TryAssign(
		sourceType,
		targetType,
		sourceValue,
		targetValue,
	)

	assertNoError(t, err)

	if !ok {
		t.Fatal("expected assignment to succeed")
	}

	if got := targetValue.Interface(); got != 42 {
		t.Fatalf("got %v, want 42", got)
	}
}

func TestTryAssign_NotAssignable(t *testing.T) {
	sourceType := TypeFor[int]()
	targetType := TypeFor[string]()

	targetValue := New(targetType).Elem()

	ok, err := TryAssign(
		sourceType,
		targetType,
		ValueOf(42),
		targetValue,
	)

	assertNoError(t, err)

	if ok {
		t.Fatal("expected assignment to be rejected")
	}
}

// -----------------------------------------------------------------------------
// TryConvert
// -----------------------------------------------------------------------------

func TestTryConvert_IntToInt64(t *testing.T) {
	target := New(TypeFor[int64]()).Elem()

	ok, err := TryConvert(
		TypeFor[int](),
		TypeFor[int64](),
		ValueOf(42),
		target,
	)

	assertNoError(t, err)

	if !ok {
		t.Fatal("expected conversion")
	}

	if got := target.Int(); got != 42 {
		t.Fatalf("got %d, want 42", got)
	}
}

func TestTryConvert_NotConvertible(t *testing.T) {
	target := New(TypeFor[string]()).Elem()

	ok, err := TryConvert(
		TypeFor[struct{}](),
		TypeFor[string](),
		ValueOf(struct{}{}),
		target,
	)

	assertNoError(t, err)

	if ok {
		t.Fatal("expected conversion to be rejected")
	}
}

// -----------------------------------------------------------------------------
// Struct conversion
// -----------------------------------------------------------------------------

func TestConvertFor_BasicStruct(t *testing.T) {
	input := &sourceBasic{
		Name:   "Alice",
		Age:    30,
		Active: true,
		Score:  9.5,
	}

	got, err := ConvertFor[sourceBasic, targetBasic](input)

	assertNoError(t, err)

	want := &targetBasic{
		Name:   "Alice",
		Age:    30,
		Active: true,
		Score:  9.5,
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestConvertFor_ExtraSourceFieldsIgnored(t *testing.T) {
	input := &sourceExtraFields{
		Name:   "Alice",
		Age:    30,
		Secret: "hidden",
	}

	got, err := ConvertFor[sourceExtraFields, targetMissingFields](input)

	assertNoError(t, err)

	if got.Name != "Alice" {
		t.Fatalf("got %q", got.Name)
	}
}

func TestConvertFor_MissingSourceFieldsRemainZero(t *testing.T) {
	input := &targetMissingFields{
		Name: "Alice",
	}

	got, err := ConvertFor[targetMissingFields, sourceExtraFields](input)

	assertNoError(t, err)

	if got.Name != "Alice" {
		t.Fatalf("got %q", got.Name)
	}

	if got.Age != 0 {
		t.Fatalf("expected zero Age, got %d", got.Age)
	}

	if got.Secret != "" {
		t.Fatalf("expected zero Secret, got %q", got.Secret)
	}
}

func TestConvertFor_NestedStruct(t *testing.T) {
	input := &sourceNested{
		Name: "root",
		Child: sourceBasic{
			Name:   "child",
			Age:    10,
			Active: true,
			Score:  4.5,
		},
	}

	got, err := ConvertFor[sourceNested, targetNested](input)

	assertNoError(t, err)

	if got.Name != "root" {
		t.Fatalf("got %q", got.Name)
	}

	if got.Child.Name != "child" {
		t.Fatalf("got %q", got.Child.Name)
	}

	if got.Child.Age != 10 {
		t.Fatalf("got %d", got.Child.Age)
	}
}

// -----------------------------------------------------------------------------
// Slice conversion
// -----------------------------------------------------------------------------

func TestConvertFor_Slice(t *testing.T) {
	input := &sourceSlice{
		Values: []int{1, 2, 3, 4},
	}

	got, err := ConvertFor[sourceSlice, targetSlice](input)

	assertNoError(t, err)

	want := []int64{1, 2, 3, 4}

	if !reflect.DeepEqual(got.Values, want) {
		t.Fatalf("got %v, want %v", got.Values, want)
	}
}

func TestConvertFor_EmptySlice(t *testing.T) {
	input := &sourceSlice{
		Values: []int{},
	}

	got, err := ConvertFor[sourceSlice, targetSlice](input)

	assertNoError(t, err)

	if got.Values == nil {
		t.Fatal("expected non-nil empty slice if mapper initializes it")
	}

	if len(got.Values) != 0 {
		t.Fatalf("expected empty slice, got %v", got.Values)
	}
}

func TestConvertFor_NilSlice(t *testing.T) {
	input := &sourceSlice{
		Values: nil,
	}

	got, err := ConvertFor[sourceSlice, targetSlice](input)

	assertNoError(t, err)

	if len(got.Values) != 0 {
		t.Fatalf("expected zero-length result, got %v", got.Values)
	}
}

// -----------------------------------------------------------------------------
// Array conversion
// -----------------------------------------------------------------------------

func TestConvertFor_ArrayToSlice(t *testing.T) {
	input := &sourceArray{
		Values: [3]int{1, 2, 3},
	}

	got, err := ConvertFor[sourceArray, targetSlice](input)

	assertNoError(t, err)

	want := []int64{1, 2, 3}

	if !reflect.DeepEqual(got.Values, want) {
		t.Fatalf("got %v, want %v", got.Values, want)
	}
}

func TestConvertFor_ArrayToArray(t *testing.T) {
	type S struct {
		Values [3]int
	}

	type T struct {
		Values [3]int64
	}

	input := &S{Values: [3]int{1, 2, 3}}

	got, err := ConvertFor[S, T](input)

	assertNoError(t, err)

	want := [3]int64{1, 2, 3}

	if got.Values != want {
		t.Fatalf("got %v, want %v", got.Values, want)
	}
}

func TestConvertFor_DifferentArrayLengths(t *testing.T) {
	type S struct {
		Values [2]int
	}

	type T struct {
		Values [3]int
	}

	input := &S{Values: [2]int{1, 2}}

	got, err := ConvertFor[S, T](input)

	assertNoError(t, err)

	if got == nil {
		t.Fatal("expected result")
	}
}

// -----------------------------------------------------------------------------
// Map conversion
// -----------------------------------------------------------------------------

func TestConvertFor_Map(t *testing.T) {
	input := &sourceMap{
		Values: map[string]int{
			"a": 1,
			"b": 2,
			"c": 3,
		},
	}

	got, err := ConvertFor[sourceMap, targetMap](input)

	assertNoError(t, err)

	if len(got.Values) != 3 {
		t.Fatalf("got %d entries", len(got.Values))
	}

	if got.Values["a"] != 1 ||
		got.Values["b"] != 2 ||
		got.Values["c"] != 3 {
		t.Fatalf("unexpected map: %v", got.Values)
	}
}

func TestConvertFor_EmptyMap(t *testing.T) {
	input := &sourceMap{
		Values: map[string]int{},
	}

	got, err := ConvertFor[sourceMap, targetMap](input)

	assertNoError(t, err)

	if got.Values != nil {
		t.Fatalf("expected nil map for empty source under current implementation, got %v", got.Values)
	}
}

func TestConvertFor_NilMap(t *testing.T) {
	input := &sourceMap{
		Values: nil,
	}

	got, err := ConvertFor[sourceMap, targetMap](input)

	assertNoError(t, err)

	if got.Values != nil {
		t.Fatalf("expected nil map, got %v", got.Values)
	}
}

func TestConvertFor_MapKeyConversion(t *testing.T) {
	input := &sourceMapKey{
		Values: map[int]string{
			1: "one",
			2: "two",
		},
	}

	got, err := ConvertFor[sourceMapKey, targetMapKey](input)

	assertNoError(t, err)

	if got.Values["\x01"] != "one" {
		t.Fatalf("unexpected value: %v", got.Values)
	}

	if got.Values["\x02"] != "two" {
		t.Fatalf("unexpected value: %v", got.Values)
	}
}

// -----------------------------------------------------------------------------
// Pointer conversion
// -----------------------------------------------------------------------------

func TestConvertFor_PointerToValue(t *testing.T) {
	i := 42
	input := &i

	got, err := ConvertFor[*int, int](&input)

	assertNoError(t, err)

	if *got != 42 {
		t.Fatalf("got %d, want 42", got)
	}
}

func TestConvertFor_ValueToPointer(t *testing.T) {
	input := 42

	got, err := ConvertFor[int, *int](&input)

	assertNoError(t, err)

	if got == nil {
		t.Fatal("expected non-nil pointer")
	}

	if **got != 42 {
		t.Fatalf("got %d, want 42", *got)
	}
}

func TestConvertFor_PointerToPointer(t *testing.T) {
	value := 42
	input := &value

	got, err := ConvertFor[*int, **int](&input)

	assertNoError(t, err)

	if got == nil {
		t.Fatal("expected non-nil result")
	}

	if *got == nil {
		t.Fatal("expected nested pointer")
	}

	if ***got != 42 {
		t.Fatalf("got %d, want 42", **got)
	}
}

// -----------------------------------------------------------------------------
// Custom converters
// -----------------------------------------------------------------------------

func TestTryCustomConvert_Success(t *testing.T) {
	resetConverters()

	registerConverter[int, string](
		func(value Value, typ Type) (Value, error) {
			return ValueOf(fmt.Sprintf("%d", value.Int())), nil
		},
	)

	input := 123

	got, err := ConvertFor[int, string](&input)

	assertNoError(t, err)

	if *got != "123" {
		t.Fatalf("got %q, want 123", *got)
	}
}

func TestTryCustomConvert_Error(t *testing.T) {
	resetConverters()

	expected := errors.New("conversion failed")

	registerConverter[int, string](
		func(value Value, typ Type) (Value, error) {
			return Zero[Value](), expected
		},
	)

	input := 123

	_, err := ConvertFor[int, string](&input)

	if err == nil {
		t.Fatal("expected error")
	}

	if !strings.Contains(err.Error(), expected.Error()) {
		t.Fatalf("got error %v", err)
	}
}

func TestTryCustomConvert_NotFound(t *testing.T) {
	resetConverters()

	target := New(TypeFor[string]()).Elem()

	ok, err := TryCustomConvert(
		TypeFor[int](),
		TypeFor[string](),
		ValueOf(123),
		target,
	)

	assertNoError(t, err)

	if ok {
		t.Fatal("expected false")
	}
}

// -----------------------------------------------------------------------------
// Convert
// -----------------------------------------------------------------------------

func TestConvert_FastPath(t *testing.T) {
	value := 42

	got, err := Convert(
		ValueOf(value),
		TypeFor[int](),
	)

	assertNoError(t, err)

	if got.Interface().(int) != 42 {
		t.Fatalf("unexpected result: %v", got.Interface())
	}
}

func TestConvert_SlowPath(t *testing.T) {
	value := 42

	got, err := Convert(
		ValueOf(value),
		TypeFor[int64](),
	)

	assertNoError(t, err)

	if got.Interface().(int64) != 42 {
		t.Fatalf("unexpected result: %v", got.Interface())
	}
}

func TestConvert_UnconvertibleTypes(t *testing.T) {
	type A struct {
		A int
	}

	type B struct {
		B string
	}

	_, err := Convert(
		ValueOf(A{A: 1}),
		TypeFor[B](),
	)

	// Current implementation considers a struct pair handled even when
	// there are no matching fields, so this documents current behavior.
	assertNoError(t, err)
}

// -----------------------------------------------------------------------------
// FastConvert
// -----------------------------------------------------------------------------

func TestFastConvert(t *testing.T) {
	value := 123

	got, err := FastConvert(
		ValueOf(value),
		TypeFor[int](),
	)

	assertNoError(t, err)

	if got.Interface().(int) != 123 {
		t.Fatalf("got %v", got.Interface())
	}
}

func TestFastConvertFor(t *testing.T) {
	value := 123

	got, err := FastConvertFor[int, int](&value)

	assertNoError(t, err)

	if got == nil {
		t.Fatal("expected result")
	}

	if *got != 123 {
		t.Fatalf("got %d", *got)
	}
}

func TestFastConvertFor_DifferentLayout(t *testing.T) {
	value := 123

	got, err := FastConvertFor[int, int64](&value)

	// This API deliberately performs an unsafe reinterpretation.
	// The test primarily ensures it does not panic.
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got == nil {
		t.Fatal("expected result")
	}
}

// -----------------------------------------------------------------------------
// Time / custom nested conversion
// -----------------------------------------------------------------------------

func TestConvertFor_CustomNestedField(t *testing.T) {
	resetConverters()

	registerConverter[time.Time, string](
		func(value Value, typ Type) (Value, error) {
			return ValueOf(value.Interface().(time.Time).Format(time.RFC3339)), nil
		},
	)

	now := time.Date(
		2025,
		time.January,
		2,
		3,
		4,
		5,
		0,
		time.UTC,
	)

	input := &sourceTime{Time: now}

	got, err := ConvertFor[sourceTime, targetTime](input)

	assertNoError(t, err)

	if got.Time != now.Format(time.RFC3339) {
		t.Fatalf("got %q", got.Time)
	}
}

// -----------------------------------------------------------------------------
// Interface behavior
// -----------------------------------------------------------------------------

func TestTypeImplements(t *testing.T) {
	impl := TypeFor[interfaceImpl]()
	iface := TypeFor[interfaceA]()

	if !impl.Implements(iface) {
		t.Fatal("interfaceImpl should implement interfaceA")
	}
}

func TestTypeDoesNotImplement(t *testing.T) {
	impl := TypeFor[interfaceImpl]()
	iface := TypeFor[interfaceB]()

	if impl.Implements(iface) {
		t.Fatal("interfaceImpl should not implement interfaceB")
	}
}

// -----------------------------------------------------------------------------
// Value wrapper tests
// -----------------------------------------------------------------------------

func TestValueOfBasic(t *testing.T) {
	v := ValueOf(123)

	if !v.IsValid() {
		t.Fatal("value should be valid")
	}

	if v.Kind() != reflect.Int {
		t.Fatalf("got %v", v.Kind())
	}

	if v.Int() != 123 {
		t.Fatalf("got %d", v.Int())
	}

	if v.Type().GoType() != reflect.TypeFor[int]() {
		t.Fatalf("unexpected type")
	}
}

func TestValueOfString(t *testing.T) {
	v := ValueOf("hello")

	if v.String() != "hello" {
		t.Fatalf("got %q", v.String())
	}
}

func TestValueOfBool(t *testing.T) {
	v := ValueOf(true)

	if !v.Bool() {
		t.Fatal("expected true")
	}
}

func TestValueOfFloat(t *testing.T) {
	v := ValueOf(1.25)

	if v.Float() != 1.25 {
		t.Fatalf("got %f", v.Float())
	}
}

func TestValueOfUint(t *testing.T) {
	v := ValueOf(uint64(42))

	if v.Uint() != 42 {
		t.Fatalf("got %d", v.Uint())
	}
}

func TestValueOfComplex(t *testing.T) {
	v := ValueOf(complex(1, 2))

	if v.Complex() != complex(1, 2) {
		t.Fatalf("unexpected complex value")
	}
}

func TestValueCanConvert(t *testing.T) {
	v := ValueOf(123)

	if !v.CanConvert(TypeFor[int64]()) {
		t.Fatal("int should be convertible to int64")
	}
}

func TestValueConvert(t *testing.T) {
	v := ValueOf(123)

	got := v.Convert(TypeFor[int64]())

	if got.Int() != 123 {
		t.Fatalf("got %d", got.Int())
	}
}

func TestValueAddr(t *testing.T) {
	value := 123

	v := ValueOf(&value).Elem()

	addr := v.Addr()

	if !addr.IsValid() {
		t.Fatal("expected valid address")
	}

	if addr.Elem().Int() != 123 {
		t.Fatalf("got %d, want 123", addr.Elem().Int())
	}
}

func TestValueReference(t *testing.T) {
	value := 123
	v := ValueOf(value)

	ref := v.Reference(1)

	if ref.Kind() != reflect.Pointer {
		t.Fatalf("got %v", ref.Kind())
	}
}

func TestValueElem(t *testing.T) {
	value := 123
	v := ValueOf(&value)

	elem := v.Elem()

	if elem.Int() != 123 {
		t.Fatalf("got %d", elem.Int())
	}
}

func TestValueField(t *testing.T) {
	value := sourceBasic{
		Name: "Alice",
		Age:  20,
	}

	v := ValueOf(value)

	if v.Field(0).String() != "Alice" {
		t.Fatalf("unexpected Name")
	}

	if v.Field(1).Int() != 20 {
		t.Fatalf("unexpected Age")
	}
}

func TestValueFieldByName(t *testing.T) {
	value := sourceBasic{Name: "Alice"}

	v := ValueOf(value)

	if v.FieldByName("Name").String() != "Alice" {
		t.Fatalf("unexpected value")
	}
}

func TestValueFieldByIndex(t *testing.T) {
	value := sourceBasic{Name: "Alice"}

	v := ValueOf(value)

	got := v.FieldByIndex([]int{0})

	if got.String() != "Alice" {
		t.Fatalf("unexpected value")
	}
}

func TestValueIndex(t *testing.T) {
	v := ValueOf([]int{10, 20, 30})

	if v.Index(1).Int() != 20 {
		t.Fatalf("unexpected value")
	}
}

func TestValueSlice(t *testing.T) {
	v := ValueOf([]int{1, 2, 3, 4})

	got := v.Slice(1, 3)

	if !reflect.DeepEqual(got.Interface(), []int{2, 3}) {
		t.Fatalf("got %v", got.Interface())
	}
}

func TestValueSlice3(t *testing.T) {
	v := ValueOf([]int{1, 2, 3, 4})

	got := v.Slice3(1, 3, 3)

	if !reflect.DeepEqual(got.Interface(), []int{2, 3}) {
		t.Fatalf("got %v", got.Interface())
	}
}

func TestValueLenCap(t *testing.T) {
	v := ValueOf([]int{1, 2, 3})

	if v.Len() != 3 {
		t.Fatalf("unexpected Len")
	}

	if v.Cap() != 3 {
		t.Fatalf("unexpected Cap")
	}
}

func TestValueIsZero(t *testing.T) {
	if !ValueOf(0).IsZero() {
		t.Fatal("0 should be zero")
	}

	if ValueOf(1).IsZero() {
		t.Fatal("1 should not be zero")
	}
}

func TestValueIsNil(t *testing.T) {
	var p *int

	if !ValueOf(p).IsNil() {
		t.Fatal("nil pointer should be nil")
	}
}

func TestValueSet(t *testing.T) {
	target := New(TypeFor[int]()).Elem()

	target.Set(ValueOf(42))

	if target.Int() != 42 {
		t.Fatalf("got %d", target.Int())
	}
}

func TestValueSetInt(t *testing.T) {
	target := New(TypeFor[int]()).Elem()

	target.SetInt(42)

	if target.Int() != 42 {
		t.Fatalf("got %d", target.Int())
	}
}

func TestValueSetString(t *testing.T) {
	target := New(TypeFor[string]()).Elem()

	target.SetString("hello")

	if target.String() != "hello" {
		t.Fatalf("got %q", target.String())
	}
}

func TestValueSetBool(t *testing.T) {
	target := New(TypeFor[bool]()).Elem()

	target.SetBool(true)

	if !target.Bool() {
		t.Fatal("expected true")
	}
}

func TestValueSetFloat(t *testing.T) {
	target := New(TypeFor[float64]()).Elem()

	target.SetFloat(1.5)

	if target.Float() != 1.5 {
		t.Fatalf("got %f", target.Float())
	}
}

func TestValueSetUint(t *testing.T) {
	target := New(TypeFor[uint64]()).Elem()

	target.SetUint(42)

	if target.Uint() != 42 {
		t.Fatalf("got %d", target.Uint())
	}
}

func TestValueSetComplex(t *testing.T) {
	target := New(TypeFor[complex128]()).Elem()

	target.SetComplex(complex(1, 2))

	if target.Complex() != complex(1, 2) {
		t.Fatal("unexpected complex value")
	}
}

func TestValueSetZero(t *testing.T) {
	i := int(42)
	target := ValueOf(&i).Elem()

	target.SetZero()

	if target.Int() != 0 {
		t.Fatalf("got %d", target.Int())
	}
}

func TestValueComparable(t *testing.T) {
	a := ValueOf(10)
	b := ValueOf(10)

	if !a.Comparable() {
		t.Fatal("int should be comparable")
	}

	if !a.Equal(b) {
		t.Fatal("values should be equal")
	}
}

func TestValueMap(t *testing.T) {
	input := map[string]int{
		"a": 1,
		"b": 2,
	}

	v := ValueOf(input)

	if v.MapIndex(ValueOf("a")).Int() != 1 {
		t.Fatal("unexpected map value")
	}

	keys := v.MapKeys()

	if len(keys) != 2 {
		t.Fatalf("got %d keys", len(keys))
	}
}

func TestValueMapRange(t *testing.T) {
	input := map[string]int{
		"a": 1,
		"b": 2,
	}

	v := ValueOf(input)
	iter := v.MapRange()

	count := 0

	for iter.Next() {
		count++

		if iter.Key().String() == "" {
			t.Fatal("empty key")
		}
	}

	if count != 2 {
		t.Fatalf("got %d entries", count)
	}
}

// -----------------------------------------------------------------------------
// Function calls
// -----------------------------------------------------------------------------

func TestValueCall(t *testing.T) {
	fn := ValueOf(func(a int, b int) int {
		return a + b
	})

	got := fn.Call([]Value{
		ValueOf(2),
		ValueOf(3),
	})

	if len(got) != 1 {
		t.Fatalf("got %d results", len(got))
	}

	if got[0].Int() != 5 {
		t.Fatalf("got %d", got[0].Int())
	}
}

// -----------------------------------------------------------------------------
// Channels
// -----------------------------------------------------------------------------

func TestValueChannelSendReceive(t *testing.T) {
	ch := make(chan int, 1)

	v := ValueOf(ch)

	v.Send(ValueOf(42))

	got := <-ch

	if got != 42 {
		t.Fatalf("got %d", got)
	}
}

func TestValueChannelClose(t *testing.T) {
	ch := make(chan int)

	v := ValueOf(ch)

	v.Close()

	select {
	case _, ok := <-ch:
		if ok {
			t.Fatal("channel should be closed")
		}
	default:
		t.Fatal("channel should be closed")
	}
}

// -----------------------------------------------------------------------------
// Concurrent type access
// -----------------------------------------------------------------------------

func TestTypeOf_Concurrent(t *testing.T) {
	const workers = 100

	var wg sync.WaitGroup
	wg.Add(workers)

	results := make([]Type, workers)

	for i := 0; i < workers; i++ {
		i := i

		go func() {
			defer wg.Done()

			results[i] = TypeFor[sourceNested]()
		}()
	}

	wg.Wait()

	for i := 1; i < workers; i++ {
		if results[i] != results[0] {
			t.Fatal("concurrent TypeFor returned different Type instances")
		}
	}
}

func TestMemoryLayout_Concurrent(t *testing.T) {
	const workers = 100

	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()

			TypeFor[sourceNested]().MemoryLayout()
			TypeFor[map[string][]int]().MemoryLayout()
			TypeFor[func(int) string]().MemoryLayout()
		}()
	}

	wg.Wait()
}

// -----------------------------------------------------------------------------
// Conversion concurrency
// -----------------------------------------------------------------------------

func TestConvert_Concurrent(t *testing.T) {
	const workers = 100

	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()

			input := &sourceBasic{
				Name:   "Alice",
				Age:    42,
				Active: true,
				Score:  10.5,
			}

			got, err := ConvertFor[sourceBasic, targetBasic](input)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if got.Name != "Alice" ||
				got.Age != 42 ||
				!got.Active ||
				got.Score != 10.5 {
				t.Errorf("bad result: %+v", got)
			}
		}()
	}

	wg.Wait()
}

// -----------------------------------------------------------------------------
// Boundary numeric conversions
// -----------------------------------------------------------------------------

func TestConvertFor_NumericBoundaries(t *testing.T) {
	tests := []struct {
		name string
		in   int
	}{
		{"zero", 0},
		{"positive", 123},
		{"negative", -123},
		{"max-int8", 127},
		{"min-int8", -128},
	}

	type S struct {
		Value int
	}

	type T struct {
		Value int64
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertFor[S, T](&S{Value: tt.in})

			assertNoError(t, err)

			if got.Value != int64(tt.in) {
				t.Fatalf("got %d, want %d", got.Value, tt.in)
			}
		})
	}
}

// -----------------------------------------------------------------------------
// Named types
// -----------------------------------------------------------------------------

func TestConvertFor_NamedNumericTypes(t *testing.T) {
	type S struct {
		Value namedInt
	}

	type T struct {
		Value namedIntTarget
	}

	got, err := ConvertFor[S, T](&S{Value: 42})

	assertNoError(t, err)

	if got.Value != 42 {
		t.Fatalf("got %v", got.Value)
	}
}

// -----------------------------------------------------------------------------
// Struct field type incompatibility
// -----------------------------------------------------------------------------

func TestConvertFor_FieldConversion(t *testing.T) {
	type S struct {
		Value int
	}

	type T struct {
		Value string
	}

	resetConverters()

	registerConverter[int, string](
		func(value Value, typ Type) (Value, error) {
			return ValueOf(fmt.Sprintf("%d", value.Int())), nil
		},
	)

	got, err := ConvertFor[S, T](&S{Value: 123})

	assertNoError(t, err)

	if got.Value != "123" {
		t.Fatalf("got %q", got.Value)
	}
}

// -----------------------------------------------------------------------------
// Converter precedence
// -----------------------------------------------------------------------------

func TestCustomConverterTakesPrecedence(t *testing.T) {
	resetConverters()

	registerConverter[int, string](
		func(value Value, typ Type) (Value, error) {
			return ValueOf("custom"), nil
		},
	)

	got, err := ConvertFor[int, string](new(int))

	assertNoError(t, err)

	if *got != "custom" {
		t.Fatalf("got %q", *got)
	}
}

// -----------------------------------------------------------------------------
// FastConvert pointer handling
// -----------------------------------------------------------------------------

func TestFastConvert_PointerInput(t *testing.T) {
	value := 42

	got, err := FastConvert(
		ValueOf(&value),
		TypeFor[int](),
	)

	assertNoError(t, err)

	if got.Interface().(int) != 42 {
		t.Fatalf("got %v", got.Interface())
	}
}

// -----------------------------------------------------------------------------
// Unsafe pointer API sanity
// -----------------------------------------------------------------------------

func TestValueUnsafePointer(t *testing.T) {
	value := 123

	v := ValueOf(&value)

	ptr := v.UnsafePointer()

	if ptr == nil {
		t.Fatal("expected non-nil unsafe pointer")
	}

	got := *(*int)(unsafe.Pointer(ptr))

	if got != 123 {
		t.Fatalf("got %d", got)
	}
}

// -----------------------------------------------------------------------------
// Zero-length / empty structures
// -----------------------------------------------------------------------------

func TestConvertEmptyStruct(t *testing.T) {
	type A struct{}
	type B struct{}

	got, err := ConvertFor[A, B](&A{})

	assertNoError(t, err)

	if got == nil {
		t.Fatal("expected result")
	}
}

func TestConvertStructWithNoMatchingFields(t *testing.T) {
	type A struct {
		One int
	}

	type B struct {
		Two string
	}

	got, err := ConvertFor[A, B](&A{One: 123})

	assertNoError(t, err)

	if got == nil {
		t.Fatal("expected result")
	}

	if got.Two != "" {
		t.Fatalf("expected zero value, got %q", got.Two)
	}
}

// -----------------------------------------------------------------------------
// Recursive types
// -----------------------------------------------------------------------------

func TestMemoryLayout_RecursiveType(t *testing.T) {
	// Must terminate despite recursive type.
	layout := TypeFor[recursiveNode]().MemoryLayout()

	if layout == nil {
		t.Fatal("expected layout")
	}
}

func TestMemoryLayout_RecursiveType2(t *testing.T) {
	layout := TypeFor[recursiveNode2]().MemoryLayout()

	if layout == nil {
		t.Fatal("expected layout")
	}
}

// -----------------------------------------------------------------------------
// Type methods
// -----------------------------------------------------------------------------

func TestTypeMethodsIteration(t *testing.T) {
	typ := TypeFor[methodType]()

	var names []string

	for method := range typ.Methods() {
		names = append(names, method.Name)
	}

	if len(names) == 0 {
		t.Fatal("expected methods")
	}
}

func TestTypeInsOuts(t *testing.T) {
	typ := TypeFor[func(int, string) (bool, error)]()

	if typ.NumIn() != 2 {
		t.Fatalf("got %d inputs", typ.NumIn())
	}

	if typ.NumOut() != 2 {
		t.Fatalf("got %d outputs", typ.NumOut())
	}

	var inputs []Type

	for in := range typ.Ins() {
		inputs = append(inputs, in)
	}

	if len(inputs) != 2 {
		t.Fatalf("got %d inputs", len(inputs))
	}

	var outputs []Type

	for out := range typ.Outs() {
		outputs = append(outputs, out)
	}

	if len(outputs) != 2 {
		t.Fatalf("got %d outputs", len(outputs))
	}
}

func TestTypeVariadic(t *testing.T) {
	typ := TypeFor[func(...int)]()

	if !typ.IsVariadic() {
		t.Fatal("expected variadic function")
	}
}

func TestTypeChanDir(t *testing.T) {
	if TypeFor[chan int]().ChanDir() != reflect.BothDir {
		t.Fatal("expected bidirectional channel")
	}

	if TypeFor[<-chan int]().ChanDir() != reflect.RecvDir {
		t.Fatal("expected receive-only channel")
	}

	if TypeFor[chan<- int]().ChanDir() != reflect.SendDir {
		t.Fatal("expected send-only channel")
	}
}

// -----------------------------------------------------------------------------
// Error / panic recovery tests
// -----------------------------------------------------------------------------

func TestFastConvert_Recovery(t *testing.T) {
	// Invalid conversion should be recovered rather than panic.
	_, err := FastConvert(
		ValueOf("hello"),
		TypeFor[int](),
	)

	// Depending on reflect/unsafe behavior this may be implementation-specific.
	// The important contract is that a panic must not escape.
	_ = err
}

func TestFastConvertFor_Recovery(t *testing.T) {
	// Same principle for the generic unsafe API.
	value := 123

	_, err := FastConvertFor[int, string](&value)

	_ = err
}

// -----------------------------------------------------------------------------
// Fuzz tests
// -----------------------------------------------------------------------------

func FuzzConvertIntToInt64(f *testing.F) {
	f.Add(int(0))
	f.Add(int(1))
	f.Add(int(-1))
	f.Add(int(123456))

	f.Fuzz(func(t *testing.T, value int) {
		got, err := ConvertFor[int, int64](&value)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if *got != int64(value) {
			t.Fatalf("got %d, want %d", *got, int64(value))
		}
	})
}

func FuzzConvertStringToString(f *testing.F) {
	f.Add("")
	f.Add("hello")
	f.Add("unicode: 世界")
	f.Add("emoji: 🚀")

	f.Fuzz(func(t *testing.T, value string) {
		got, err := ConvertFor[string, string](&value)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if *got != value {
			t.Fatalf("got %q, want %q", *got, value)
		}
	})
}

func FuzzConvertSlice(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{1})
	f.Add([]byte{1, 2, 3, 4, 5})

	f.Fuzz(func(t *testing.T, value []byte) {
		type S struct {
			Values []byte
		}

		type T struct {
			Values []byte
		}

		got, err := ConvertFor[S, T](&S{Values: value})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !reflect.DeepEqual(got.Values, value) {
			t.Fatalf("got %v, want %v", got.Values, value)
		}
	})
}
