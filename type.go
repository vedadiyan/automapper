package mapper

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"iter"
	"reflect"
	"sync"
)

type (
	Kind        = reflect.Kind
	ChanDir     reflect.ChanDir
	StructField struct {
		Type    Type
		Name    string
		PkgPath string
		//Tag       StructTag
		Offset    uintptr
		Index     []int
		Anonymous bool
	}
	Method reflect.Method

	Type interface {
		Align() int
		FieldAlign() int
		Method(int) Method
		Methods() iter.Seq[Method]
		MethodByName(string) (Method, bool)
		NumMethod() int
		Name() string
		PkgPath() string
		Size() uintptr
		String() string
		Kind() Kind
		Implements(Type) bool
		AssignableTo(Type) bool
		ConvertibleTo(Type) bool
		Comparable() bool
		Bits() int
		ChanDir() ChanDir
		IsVariadic() bool
		Elem() Type
		Field(i int) StructField
		Fields() iter.Seq[StructField]
		FieldByIndex(index []int) StructField
		FieldByName(name string) (StructField, bool)
		FieldByNameFunc(match func(string) bool) (StructField, bool)
		In(i int) Type
		Ins() iter.Seq[Type]
		Key() Type
		Len() int
		NumField() int
		NumIn() int
		NumOut() int
		Out(i int) Type
		Outs() iter.Seq[Type]
		OverflowComplex(x complex128) bool
		OverflowFloat(x float64) bool
		OverflowInt(x int64) bool
		OverflowUint(x uint64) bool
		CanSeq() bool
		CanSeq2() bool
		IdenticalTo(Type) bool
		Signature() string
		PointerCount() int
		ConcreteType() Type

		GoType() reflect.Type
	}

	rtype struct {
		t         reflect.Type
		ct        Type
		ptrCount  int
		signature string
		once      sync.Once
	}
)

var (
	types map[reflect.Type]Type
	mutx  sync.RWMutex
)

func init() {
	types = make(map[reflect.Type]Type)
}

func newStructField(t *reflect.StructField) *StructField {
	return &StructField{
		Name:      t.Name,
		PkgPath:   t.PkgPath,
		Offset:    t.Offset,
		Index:     t.Index,
		Anonymous: t.Anonymous,
		Type:      typeOf(t.Type),
	}
}

func derefType(t reflect.Type) (int, Type) {
	n, ct := DeReferenceType(t)
	out := &rtype{}
	out.t = ct
	out.ct = out
	return n, out
}

func typeOf(t reflect.Type) Type {
	mutx.RLock()
	if val, ok := types[t]; ok {
		mutx.RUnlock()
		return val
	}
	mutx.RUnlock()

	n, ct := derefType(t)
	out := &rtype{
		t:        t,
		ct:       ct,
		ptrCount: n,
	}

	mutx.Lock()
	if val, ok := types[t]; ok {
		mutx.Unlock()
		return val
	}
	types[t] = out
	mutx.Unlock()
	return out
}

func TypeOf(i any) Type {
	return typeOf(reflect.TypeOf(i))
}

func (rt *rtype) GoType() reflect.Type {
	return rt.t
}

func (rt *rtype) Align() int {
	return rt.t.Align()
}

func (rt *rtype) FieldAlign() int {
	return rt.t.FieldAlign()
}

func (rt *rtype) Method(i int) Method {
	return Method(rt.t.Method(i))
}

func (rt *rtype) Methods() iter.Seq[Method] {
	return func(yield func(Method) bool) {
		for method := range rt.t.Methods() {
			if !yield(Method(method)) {
				return
			}
		}
	}
}

func (rt *rtype) MethodByName(name string) (Method, bool) {
	out, ok := rt.t.MethodByName(name)
	return Method(out), ok
}

func (rt *rtype) NumMethod() int {
	return rt.t.NumMethod()
}

func (rt *rtype) Name() string {
	return rt.t.Name()
}

func (rt *rtype) PkgPath() string {
	return rt.t.PkgPath()
}

func (rt *rtype) Size() uintptr {
	return rt.t.Size()
}

func (rt *rtype) String() string {
	return rt.t.String()
}

func (rt *rtype) Kind() Kind {
	return Kind(rt.t.Kind())
}

func (rt *rtype) Implements(t Type) bool {
	return rt.t.Implements(t.GoType())
}

func (rt *rtype) AssignableTo(t Type) bool {
	return rt.t.AssignableTo(t.GoType())
}

func (rt *rtype) ConvertibleTo(t Type) bool {
	return rt.t.ConvertibleTo(t.GoType())
}

func (rt *rtype) Comparable() bool {
	return rt.t.Comparable()
}

func (rt *rtype) Bits() int {
	return rt.t.Bits()
}

func (rt *rtype) ChanDir() ChanDir {
	return ChanDir(rt.t.ChanDir())
}

func (rt *rtype) IsVariadic() bool {
	return rt.t.IsVariadic()
}

func (rt *rtype) Elem() Type {
	return typeOf(rt.t.Elem())
}

func (rt *rtype) Field(i int) StructField {
	in := rt.t.Field(i)
	return *newStructField(&in)
}

func (rt *rtype) Fields() iter.Seq[StructField] {
	return func(yield func(StructField) bool) {
		for i := 0; i < rt.t.NumField(); i++ {
			in := rt.t.Field(i)
			if !yield(*newStructField(&in)) {
				return
			}
		}
	}
}

func (rt *rtype) FieldByIndex(index []int) StructField {
	in := rt.t.FieldByIndex(index)
	return *newStructField(&in)
}

func (rt *rtype) FieldByName(name string) (StructField, bool) {
	out, ok := rt.t.FieldByName(name)
	return *newStructField(&out), ok
}

func (rt *rtype) FieldByNameFunc(match func(string) bool) (StructField, bool) {
	out, ok := rt.t.FieldByNameFunc(match)
	return *newStructField(&out), ok
}

func (rt *rtype) In(i int) Type {
	return typeOf(rt.t.In(i))
}

func (rt *rtype) Ins() iter.Seq[Type] {
	return func(yield func(Type) bool) {
		for in := range rt.t.Ins() {
			if !yield(typeOf(in)) {
				return
			}
		}
	}
}

func (rt *rtype) Key() Type {
	return typeOf(rt.t.Key())
}

func (rt *rtype) Len() int {
	return rt.t.Len()
}

func (rt *rtype) NumField() int {
	return rt.t.NumField()
}

func (rt *rtype) NumIn() int {
	return rt.t.NumIn()
}

func (rt *rtype) NumOut() int {
	return rt.t.NumOut()
}

func (rt *rtype) Out(i int) Type {
	return typeOf(rt.t.Out(i))
}

func (rt *rtype) Outs() iter.Seq[Type] {
	return func(yield func(Type) bool) {
		for i := 0; i < rt.t.NumOut(); i++ {
			if !yield(typeOf(rt.t.Out(i))) {
				return
			}
		}
	}
}

func (rt *rtype) OverflowComplex(x complex128) bool {
	return rt.t.OverflowComplex(x)
}

func (rt *rtype) OverflowFloat(x float64) bool {
	return rt.t.OverflowFloat(x)
}

func (rt *rtype) OverflowInt(x int64) bool {
	return rt.t.OverflowInt(x)
}

func (rt *rtype) OverflowUint(x uint64) bool {
	return rt.t.OverflowUint(x)
}

func (rt *rtype) CanSeq() bool {
	return rt.t.CanSeq()
}

func (rt *rtype) CanSeq2() bool {
	return rt.t.CanSeq2()
}

func (rt *rtype) IdenticalTo(t Type) bool {
	return rt.Signature() == t.Signature() && rt.PointerCount() == t.PointerCount()
}

func (rt *rtype) Signature() string {
	rt.once.Do(func() {
		signature := bytes.NewBuffer(nil)
		signature.WriteByte(byte(rt.PointerCount()))
		signature.WriteByte(0x0)
		if rt.ConcreteType().Comparable() {
			signature.WriteByte(byte(rt.ConcreteType().Kind()))
			signature.WriteByte(0x0)

			signature.WriteByte(byte(rt.ConcreteType().PointerCount()))
			signature.WriteByte(0x0)

			sha256 := sha256.Sum256(signature.Bytes())
			rt.signature = hex.EncodeToString(sha256[:])
			return
		}
		switch rt.ConcreteType().Kind() {
		case reflect.Array, reflect.Slice:
			{
				signature := bytes.NewBuffer(nil)
				signature.WriteByte(byte(reflect.Array))
				signature.WriteByte(0x0)

				signature.WriteByte(byte(rt.ConcreteType().Elem().PointerCount()))
				signature.WriteByte(0x0)

				signature.WriteString(rt.ConcreteType().Elem().ConcreteType().Signature())
				signature.WriteByte(0x0)

				sha256 := sha256.Sum256(signature.Bytes())
				rt.signature = hex.EncodeToString(sha256[:])
			}
		case reflect.Map:
			{
				signature := bytes.NewBuffer(nil)
				signature.WriteByte(byte(reflect.Map))
				signature.WriteByte(0x0)

				signature.WriteByte(byte(rt.ConcreteType().Elem().PointerCount()))
				signature.WriteByte(0x0)

				signature.WriteString(rt.ConcreteType().Key().ConcreteType().Signature())
				signature.WriteByte(0x0)

				signature.WriteString(rt.ConcreteType().Elem().ConcreteType().Signature())
				signature.WriteByte(0x0)

				sha256 := sha256.Sum256(signature.Bytes())
				rt.signature = hex.EncodeToString(sha256[:])
			}
		case reflect.Struct:
			{
				signature := bytes.NewBuffer(nil)
				signature.WriteByte(byte(reflect.Struct))
				signature.WriteByte(0x0)

				buf := make([]byte, binary.MaxVarintLen64)
				for f := range rt.ConcreteType().Fields() {
					var n int
					n = binary.PutUvarint(buf, uint64(f.Offset))
					signature.Write(buf[:n])
					signature.WriteByte(0x0)

					for _, i := range f.Index {
						n = binary.PutUvarint(buf, uint64(i))
						signature.Write(buf[:n])
						signature.WriteByte(0x0)
					}

					n = binary.PutUvarint(buf, uint64(f.Type.ConcreteType().Align()))
					signature.Write(buf[:n])
					signature.WriteByte(0x0)

					n = binary.PutUvarint(buf, uint64(f.Type.ConcreteType().Size()))
					signature.Write(buf[:n])
					signature.WriteByte(0x0)

					n = binary.PutUvarint(buf, uint64(f.Type.PointerCount()))
					signature.Write(buf[:n])
					signature.WriteByte(0x0)
					signature.WriteByte(0x0)
				}
				sha256 := sha256.Sum256(signature.Bytes())
				rt.signature = hex.EncodeToString(sha256[:])
			}
		}

	})

	return rt.signature
}

func (rt *rtype) PointerCount() int {
	return rt.ptrCount
}

func (rt *rtype) ConcreteType() Type {
	return rt.ct
}
