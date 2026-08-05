package mapper

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"iter"
	"reflect"
	"sync"
	"unsafe"
)

type (
	Converter                      func(src reflect.Value, target reflect.Value) error
	SimplifiedKind                 = reflect.Kind
	LinkedMap[T comparable, R any] struct {
		keys   []T
		values map[T]R
	}
	Type struct {
		id       string
		typ      reflect.Type
		refCount int
		size     uint
		hash     string
	}
)

const (
	KindInt     SimplifiedKind = reflect.Int
	KindUint    SimplifiedKind = reflect.Uint
	KindFloat   SimplifiedKind = reflect.Float64
	KindComplex SimplifiedKind = reflect.Complex128
	KindString  SimplifiedKind = reflect.String
	KindBool    SimplifiedKind = reflect.Bool
)

var (
	mappers   map[SimplifiedKind]map[SimplifiedKind]Converter
	typeCache map[reflect.Type]*Type
	mut       sync.RWMutex
)

func init() {
	mappers = make(map[SimplifiedKind]map[SimplifiedKind]Converter)
	typeCache = make(map[reflect.Type]*Type)
	mappers[KindInt] = map[SimplifiedKind]Converter{
		KindInt:     sameType,
		KindUint:    IntToUint,
		KindFloat:   IntToFloat,
		KindComplex: IntToComplex,
		KindString:  IntToString,
		KindBool:    IntToBool,
	}
	mappers[KindUint] = map[SimplifiedKind]Converter{
		KindUint:    sameType,
		KindInt:     UintToInt,
		KindFloat:   UintToFloat,
		KindComplex: UintToComplex,
		KindString:  UintToString,
		KindBool:    UintToBool,
	}
	mappers[KindFloat] = map[SimplifiedKind]Converter{
		KindFloat:   sameType,
		KindUint:    FloatToUint,
		KindInt:     FloatToInt,
		KindComplex: FloatToComplex,
		KindString:  FloatToString,
	}
	mappers[KindComplex] = map[SimplifiedKind]Converter{
		KindComplex: sameType,
		KindUint:    ComplexToUint,
		KindInt:     ComplexToInt,
		KindFloat:   ComplexToFloat,
		KindString:  ComplexToString,
	}
	mappers[KindString] = map[SimplifiedKind]Converter{
		KindString:  sameType,
		KindUint:    StringToUint,
		KindInt:     StringToInt,
		KindFloat:   StringToFloat,
		KindComplex: StringToComplex,
		KindBool:    StringToBool,
	}
	mappers[reflect.Struct] = map[SimplifiedKind]Converter{
		reflect.Struct: func(src, target reflect.Value) error {
			return nil
		},
	}
}

func (x *LinkedMap[T, R]) Set(key T, value R) {
	if _, ok := x.values[key]; ok {
		x.Delete(key)
		return
	}
	x.keys = append(x.keys, key)
	x.values[key] = value

}

func (x *LinkedMap[T, R]) Delete(key T) {
	for i, value := range x.keys {
		if value == key {
			x.keys = append(x.keys[:i], x.keys[i+1:]...)
			delete(x.values, value)
			break
		}
	}
}

func (x *LinkedMap[T, R]) Get(key T) (R, bool) {
	value, ok := x.values[key]
	return value, ok
}

func (x *LinkedMap[T, R]) Range() iter.Seq2[T, R] {
	return func(yield func(T, R) bool) {
		for _, key := range x.keys {
			if !yield(key, x.values[key]) {
				return
			}
		}
	}
}

func NewLinkedMap[T comparable, R any]() *LinkedMap[T, R] {
	return &LinkedMap[T, R]{
		keys:   make([]T, 0),
		values: make(map[T]R),
	}
}

func NewType(id string, typ reflect.Type, refCount int, size uint, hash string) *Type {
	return &Type{
		id:       id,
		typ:      typ,
		refCount: refCount,
		size:     size,
		hash:     hash,
	}
}

func SimplifyKind(kind reflect.Kind) SimplifiedKind {
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		{
			return KindInt
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		{
			return KindUint
		}
	case reflect.Float32, reflect.Float64:
		{
			return KindFloat
		}
	case reflect.Complex64, reflect.Complex128:
		{
			return KindComplex
		}
	default:
		{
			return kind
		}
	}
}

func sameType(src reflect.Value, target reflect.Value) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	target.Set(src)
	return nil
}

func Analyze[T any]() (*Type, error) {
	n, typ := DeReferenceType(reflect.TypeFor[T]())
	mut.RLocker()
	if value, ok := typeCache[typ]; ok {
		mut.RUnlock()
		return value, nil
	}

	mut.Lock()
	defer mut.Unlock()
	// Double read to block duplicate calls
	if value, ok := typeCache[typ]; ok {
		return value, nil
	}
	id := fmt.Sprintf("%s.%s", typ.PkgPath(), typ.Name())
	signature := bytes.NewBuffer(nil)
	for field := range typ.Fields() {
		n, typ := DeReferenceType(field.Type)
		bytes := make([]byte, 8)
		binary.PutUvarint(bytes, uint64(field.Type.Size()))
		signature.WriteByte(byte(n))
		signature.WriteString("\r\n")
		signature.WriteByte(byte(typ.Kind()))
		signature.WriteString("\r\n")
		signature.Write(bytes)
		signature.WriteString("\r\n")
	}
	sha256 := sha256.New()
	if _, err := sha256.Write(signature.Bytes()); err != nil {
		return nil, err
	}
	hash := sha256.Sum(nil)
	val := NewType(id, typ, n, uint(typ.Size()), hex.EncodeToString(hash))
	typeCache[typ] = val
	return val, nil
}

func FastConvert[T any, R any](in *T) (r *R, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	return (*R)(unsafe.Pointer(in)), nil
}

func Convert[T any, R any](in *T) (*R, error) {
	lt, err := Analyze[T]()
	if err != nil {
		return nil, err
	}
	rt, err := Analyze[R]()
	if err != nil {
		return nil, err
	}
	if lt.hash == rt.hash {
		return FastConvert[T, R](in)
	}

	_, lv := DeReference(reflect.ValueOf(in))
	rv := reflect.New(rt.typ).Elem()

	for field := range lt.typ.Fields() {
		srcValue := lv.FieldByIndex(field.Index)

		targetType, ok := rt.typ.FieldByName(field.Name)
		if !ok {
			continue
		}

		if field.Type.AssignableTo(targetType.Type) {
			rv.FieldByIndex(targetType.Index).Set(srcValue)
			continue
		}

		_, realSourceType := DeReferenceType(field.Type)
		n, realTargetType := DeReferenceType(targetType.Type)

		_, realSourceValue := DeReference(srcValue)

		if realSourceType.AssignableTo(realTargetType) {
			rv.FieldByIndex(targetType.Index).Set(Reference(n, realSourceValue))
			continue
		}

		conv, ok := mappers[SimplifyKind(realSourceType.Kind())]
		if !ok {
			return nil, fmt.Errorf("no converters found for %s to %s", realSourceType.Kind(), realTargetType.Kind())
		}
		fn, ok := conv[SimplifiedKind(realTargetType.Kind())]
		if !ok {
			return nil, fmt.Errorf("no converters found for %s to %s", realSourceType.Kind(), realTargetType.Kind())
		}
		value := reflect.New(realTargetType).Elem()
		fn(realSourceValue, value)
		rv.FieldByIndex(targetType.Index).Set(Reference(n, value))
	}
	ref := Reference(rt.refCount, rv)

	return ref.Addr().Interface().(*R), nil

}

func DeReferenceType(v reflect.Type) (int, reflect.Type) {
	i := 0
	for ; v.Kind() == reflect.Pointer; i++ {
		v = v.Elem()
	}
	return i, v
}

func DeReference(v reflect.Value) (int, reflect.Value) {
	i := 0
	for ; v.Kind() == reflect.Pointer; i++ {
		v = v.Elem()
	}
	return i, v
}

func Reference(n int, v reflect.Value) reflect.Value {
	if n == 0 {
		return v
	}

	ref := reflect.New(v.Type())
	ref.Elem().Set(v)

	for range n - 1 {
		next := reflect.New(ref.Type())
		next.Elem().Set(ref)
		ref = next
	}

	return ref
}

func Map[T any, R any](in T) (R, error) {
	tValue := reflect.ValueOf(in)
	tType := tValue.Type()
	zero := Zero[R]()

	if tType.Kind() != reflect.Struct {
		return Zero[R](), fmt.Errorf("unsupported type")
	}

	rValue := reflect.ValueOf(zero)
	rType := rValue.Type()

	for tField, tValue := range tValue.Fields() {
		name := TargetFieldName(tField)
		rField, ok := rType.FieldByName(name)
		if !ok {
			continue
		}
		if !rField.IsExported() {
			continue
		}
		left, ok := mappers[tField.Type.Kind()]
		if !ok {
			return zero, fmt.Errorf("")
		}
		right, ok := left[rField.Type.Kind()]
		if !ok {
			return zero, fmt.Errorf("")
		}
		if err := right(tValue, rValue.FieldByIndex(rField.Index)); err != nil {
			return zero, err
		}
	}

	return rValue.Interface().(R), nil
}

func TargetFieldName(field reflect.StructField) string {
	if val, ok := field.Tag.Lookup("mapto"); ok {
		return val
	}
	return field.Name
}

func Zero[T any]() T {
	var out T
	return out
}
