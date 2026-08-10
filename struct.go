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
	Converter                      func(value reflect.Value, typ reflect.Type) (reflect.Value, error)
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

var (
	converters map[reflect.Type]map[reflect.Type]Converter
	typeCache  map[reflect.Type]*Type
	mut        sync.RWMutex
)

func init() {
	converters = make(map[reflect.Type]map[reflect.Type]Converter)
	typeCache = make(map[reflect.Type]*Type)
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

func Analyze(t reflect.Type) (*Type, error) {
	n, typ := DeReferenceType(t)
	mut.RLocker()
	if value, ok := typeCache[typ]; ok {
		mut.RUnlock()
		return value, nil
	}

	mut.Lock()
	defer mut.Unlock()
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

func AnalyzeFor[T any]() (*Type, error) {
	return Analyze(reflect.TypeFor[T]())
}

func FastConvert(in reflect.Value, o reflect.Type) (out reflect.Value, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	return reflect.NewAt(o, in.UnsafePointer()), nil
}

func FastConvertFor[T any, R any](in *T) (r *R, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	return (*R)(unsafe.Pointer(in)), nil
}

func Convert(in reflect.Value, out reflect.Type) (reflect.Value, error) {
	var zero reflect.Value

	lt, err := Analyze(in.Type())
	if err != nil {
		return zero, err
	}
	rt, err := Analyze(out)
	if err != nil {
		return zero, err
	}
	if lt.hash == rt.hash {
		return FastConvert(in, out)
	}

	_, lv := DeReference(in)
	rv := reflect.New(rt.typ).Elem()

	for field := range lt.typ.Fields() {
		if !field.IsExported() {
			continue
		}

		srcValue := lv.FieldByIndex(field.Index)

		targetType, ok := rt.typ.FieldByName(field.Name)
		if !ok {
			continue
		}

		if field.Type.AssignableTo(targetType.Type) {
			rv.FieldByIndex(targetType.Index).Set(srcValue)
			continue
		}

		if field.Type.ConvertibleTo(targetType.Type) {
			rv.FieldByIndex(targetType.Index).Set(srcValue.Convert(targetType.Type))
			continue
		}

		_, realSourceType := DeReferenceType(field.Type)
		n, realTargetType := DeReferenceType(targetType.Type)

		_, realSourceValue := DeReference(srcValue)

		if realSourceType.AssignableTo(realTargetType) {
			rv.FieldByIndex(targetType.Index).Set(Reference(n, realSourceValue))
			continue
		}

		if realSourceType.ConvertibleTo(realTargetType) {
			rv.FieldByIndex(targetType.Index).Set(Reference(n, realSourceValue.Convert(realTargetType)))
			continue
		}

		if realSourceType.Kind() == realTargetType.Kind() {
			val, err := Convert(realSourceValue, realTargetType)
			if err != nil {
				return zero, err
			}
			rv.FieldByIndex(targetType.Index).Set(Reference(n, val.Elem()))
		}

		conv, ok := getConverter(realSourceType, realTargetType)
		if !ok {
			return zero, fmt.Errorf("no conversion possible between %s and %s", realSourceType.Name(), realTargetType.Name())
		}
		val, err := conv(realSourceValue, realTargetType)
		if err != nil {
			return zero, err
		}
		rv.FieldByIndex(targetType.Index).Set(Reference(n, val))
	}
	ref := Reference(rt.refCount, rv)

	return ref.Addr(), nil
}

func ConvertFor[T any, R any](in *T) (*R, error) {
	out, err := Convert(reflect.ValueOf(in), reflect.TypeFor[R]())
	if err != nil {
		return nil, err
	}
	return out.Interface().(*R), nil
}

func getConverter(l reflect.Type, r reflect.Type) (Converter, bool) {
	lVal, ok := converters[l]
	if !ok {
		return nil, false
	}
	rVal, ok := lVal[r]
	if !ok {
		return nil, false
	}
	return rVal, true
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
