package mapper

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"reflect"
	"sync"
)

type (
	Converter func(value reflect.Value, typ reflect.Type) (reflect.Value, error)
	Type      struct {
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
	mut.RLock()
	if value, ok := typeCache[typ]; ok {
		mut.RUnlock()
		return value, nil
	}
	mut.RUnlock()

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

func FastConvert(in reflect.Value, o reflect.Type) (out reflect.Value, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	return reflect.NewAt(o, in.UnsafePointer()), nil
}

func TryAssign(sourceField reflect.StructField, targetField reflect.StructField, sourceValue reflect.Value, targetValue reflect.Value, n int) bool {
	if !sourceField.Type.AssignableTo(targetField.Type) {
		return false
	}
	targetValue.FieldByIndex(targetField.Index).Set(Reference(n, sourceValue))
	return true
}

func TryConvert(sourceField reflect.StructField, targetField reflect.StructField, sourceValue reflect.Value, targetValue reflect.Value, n int) bool {
	if !sourceField.Type.ConvertibleTo(targetField.Type) {
		return false
	}
	targetValue.FieldByIndex(targetField.Index).Set(Reference(n, sourceValue.Convert(targetField.Type)))
	return true
}

func TryDereference(sourceField *reflect.StructField, targetField *reflect.StructField, sourceValue *reflect.Value) int {
	_, realSourceType := DeReferenceType(sourceField.Type)
	n, realTargetType := DeReferenceType(targetField.Type)

	sourceField.Type = realSourceType
	targetField.Type = realTargetType

	_, *sourceValue = DeReference(*sourceValue)
	return n
}

func TryChangeType(sourceField reflect.StructField, targetField reflect.StructField, sourceValue reflect.Value, targetValue reflect.Value, n int) bool {
	if sourceField.Type.Kind() != targetField.Type.Kind() {
		return false
	}
	val, err := Convert(sourceValue, targetValue.Type())
	if err != nil {
		return false
	}
	targetValue.FieldByIndex(targetField.Index).Set(Reference(n, val.Elem()))
	return true
}

func TryCustomConvert(sourceField reflect.StructField, targetField reflect.StructField, sourceValue reflect.Value, targetValue reflect.Value, n int) bool {
	conv, ok := FindConverter(sourceField.Type, targetField.Type)
	if !ok {
		return false
	}
	val, err := conv(sourceValue, targetValue.Type())
	if err != nil {
		return false
	}
	targetValue.FieldByIndex(targetField.Index).Set(Reference(n, val))
	return true
}

func SlowConvert(sourceType *Type, targetType *Type, source reflect.Value) (reflect.Value, error) {
	_, leftValue := DeReference(source)
	targetValue := reflect.New(targetType.typ).Elem()

	for sourceField := range sourceType.typ.Fields() {
		if !sourceField.IsExported() {
			continue
		}

		sourceValue := leftValue.FieldByIndex(sourceField.Index)

		targetField, ok := targetType.typ.FieldByName(sourceField.Name)
		if !ok {
			continue
		}

		n := 0
		for range 2 {
			if TryAssign(sourceField, targetField, sourceValue, targetValue, n) {
				break
			}
			if TryConvert(sourceField, targetField, sourceValue, targetValue, n) {
				break
			}

			if TryChangeType(sourceField, targetField, sourceValue, targetValue, n) {
				break
			}

			if TryCustomConvert(sourceField, targetField, sourceValue, targetValue, n) {
				break
			}

			if n = TryDereference(&sourceField, &targetField, &sourceValue); n == 0 {
				break
			}
		}

	}
	ref := Reference(targetType.refCount, targetValue)

	return ref.Addr(), nil
}
func Convert(source reflect.Value, target reflect.Type) (reflect.Value, error) {
	var zero reflect.Value

	leftType, err := Analyze(source.Type())
	if err != nil {
		return zero, err
	}
	rightType, err := Analyze(target)
	if err != nil {
		return zero, err
	}
	if leftType.hash == rightType.hash {
		return FastConvert(source, target)
	}
	return SlowConvert(leftType, rightType, source)
}
