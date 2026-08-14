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
	ChanDir     = reflect.ChanDir
	StructField struct {
		Type    Type
		Name    string
		PkgPath string
		//Tag       StructTag
		Offset    uintptr
		Index     []int
		Anonymous bool
	}
	Method struct {
		Name    string
		PkgPath string

		Type  Type
		Func  reflect.Value
		Index int
	}

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
		MemoryLayout() MemoryLayout
		PointerCount() int
		ConcreteType() Type

		GoType() reflect.Type
	}

	MemoryLayout interface {
		Layout() []byte
		HashCode() string
		IdenticalTo(MemoryLayout) bool
	}

	rtype struct {
		t         reflect.Type
		ct        Type
		ptrCount  int
		signature MemoryLayout
		once      sync.Once
	}

	rmemoryLayout struct {
		layout []byte
		hash   string
	}
)

var (
	types map[reflect.Type]Type
	mutx  sync.RWMutex
)

func init() {
	types = make(map[reflect.Type]Type)
}

func newStructField(t *reflect.StructField) StructField {
	if t == nil {
		return StructField{}
	}

	return StructField{
		Name:      t.Name,
		PkgPath:   t.PkgPath,
		Offset:    t.Offset,
		Index:     t.Index,
		Anonymous: t.Anonymous,
		Type:      typeOf(t.Type),
	}
}

func newMethod(t *reflect.Method) Method {
	if t == nil {
		return Method{}
	}

	return Method{
		Name:    t.Name,
		PkgPath: t.PkgPath,
		Index:   t.Index,
		Type:    typeOf(t.Type),
	}
}

func typeOf(t reflect.Type) Type {
	if t == nil {
		return nil
	}

	mutx.RLock()
	if val, ok := types[t]; ok {
		mutx.RUnlock()
		return val
	}
	mutx.RUnlock()

	n, ct := DeReferenceType(t)
	out := &rtype{
		t:        t,
		ptrCount: n,
	}
	if n != 0 {
		out.ct = typeOf(ct)
	} else {
		out.ct = out
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

func TypeFor[T any]() Type {
	return typeOf(reflect.TypeFor[T]())
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
	in := rt.t.Method(i)
	return newMethod(&in)
}

func (rt *rtype) Methods() iter.Seq[Method] {
	return func(yield func(Method) bool) {
		for method := range rt.t.Methods() {
			if !yield(newMethod(&method)) {
				return
			}
		}
	}
}

func (rt *rtype) MethodByName(name string) (Method, bool) {
	out, ok := rt.t.MethodByName(name)
	return newMethod(&out), ok
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
	return newStructField(&in)
}

func (rt *rtype) Fields() iter.Seq[StructField] {
	return func(yield func(StructField) bool) {
		for i := 0; i < rt.t.NumField(); i++ {
			in := rt.t.Field(i)
			if !yield(newStructField(&in)) {
				return
			}
		}
	}
}

func (rt *rtype) FieldByIndex(index []int) StructField {
	in := rt.t.FieldByIndex(index)
	return newStructField(&in)
}

func (rt *rtype) FieldByName(name string) (StructField, bool) {
	out, ok := rt.t.FieldByName(name)
	return newStructField(&out), ok
}

func (rt *rtype) FieldByNameFunc(match func(string) bool) (StructField, bool) {
	out, ok := rt.t.FieldByNameFunc(match)
	return newStructField(&out), ok
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

func (rt *rtype) MemoryLayout() MemoryLayout {
	rt.once.Do(func() {
		out := &rmemoryLayout{}

		signature := bytes.NewBuffer(nil)
		buf := make([]byte, binary.MaxVarintLen64)

		signature.WriteByte(byte(rt.PointerCount()))
		signature.WriteByte(0x0)
		switch rt.ConcreteType().Kind() {
		case reflect.Slice:
			{
				signature := bytes.NewBuffer(nil)
				signature.WriteByte(byte(reflect.Slice))
				signature.WriteByte(0x0)

				signature.Write(rt.ConcreteType().Elem().ConcreteType().MemoryLayout().Layout())
				signature.WriteByte(0x0)

				bytes := signature.Bytes()
				sha256 := sha256.Sum256(bytes)
				hash := hex.EncodeToString(sha256[:])

				out.layout = bytes
				out.hash = hash

				rt.signature = out

			}
		case reflect.Array:
			{
				signature := bytes.NewBuffer(nil)
				signature.WriteByte(byte(reflect.Array))
				signature.WriteByte(0x0)

				n := binary.PutUvarint(buf, uint64(rt.ConcreteType().Len()))
				signature.Write(buf[:n])
				signature.WriteByte(0x0)

				signature.Write(rt.ConcreteType().Elem().ConcreteType().MemoryLayout().Layout())
				signature.WriteByte(0x0)

				bytes := signature.Bytes()
				sha256 := sha256.Sum256(bytes)
				hash := hex.EncodeToString(sha256[:])

				out.layout = bytes
				out.hash = hash

				rt.signature = out
			}
		case reflect.Map:
			{
				signature := bytes.NewBuffer(nil)
				signature.WriteByte(byte(reflect.Map))
				signature.WriteByte(0x0)

				signature.Write(rt.ConcreteType().Key().MemoryLayout().Layout())
				signature.WriteByte(0x0)

				signature.Write(rt.ConcreteType().Elem().MemoryLayout().Layout())
				signature.WriteByte(0x0)

				bytes := signature.Bytes()
				sha256 := sha256.Sum256(bytes)
				hash := hex.EncodeToString(sha256[:])

				out.layout = bytes
				out.hash = hash

				rt.signature = out
			}
		case reflect.Struct:
			{
				signature := bytes.NewBuffer(nil)
				signature.WriteByte(byte(reflect.Struct))
				signature.WriteByte(0x0)

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

					signature.Write(f.Type.MemoryLayout().Layout())
					signature.WriteByte(0x0)
					signature.WriteByte(0x0)
				}
				bytes := signature.Bytes()
				sha256 := sha256.Sum256(bytes)
				hash := hex.EncodeToString(sha256[:])

				out.layout = bytes
				out.hash = hash

				rt.signature = out
			}
		case reflect.Func:
			{
				signature := bytes.NewBuffer(nil)
				signature.WriteByte(byte(reflect.Func))
				signature.WriteByte(0x0)

				signature.WriteString(rt.Name())
				signature.WriteByte(0x0)

				signature.WriteString(rt.PkgPath())
				signature.WriteByte(0x0)

				for i := range rt.Ins() {
					signature.Write(i.MemoryLayout().Layout())
					signature.WriteByte(0x0)
					signature.WriteByte(0x0)
				}

				for i := range rt.Outs() {
					signature.Write(i.MemoryLayout().Layout())
					signature.WriteByte(0x0)
					signature.WriteByte(0x0)
				}
				bytes := signature.Bytes()
				sha256 := sha256.Sum256(bytes)
				hash := hex.EncodeToString(sha256[:])

				out.layout = bytes
				out.hash = hash

				rt.signature = out
			}
		case reflect.Interface:
			{
				signature := bytes.NewBuffer(nil)
				signature.WriteByte(byte(reflect.Interface))
				signature.WriteByte(0x0)

				for f := range rt.ConcreteType().Methods() {
					var n int
					n = binary.PutUvarint(buf, uint64(f.Index))
					signature.Write(buf[:n])
					signature.WriteByte(0x0)

					signature.WriteString(f.PkgPath)
					signature.WriteByte(0x0)

					signature.WriteString(f.Name)
					signature.WriteByte(0x0)

					for i := range f.Type.Ins() {
						signature.Write(i.MemoryLayout().Layout())
						signature.WriteByte(0x0)
						signature.WriteByte(0x0)
					}

					for i := range f.Type.Outs() {
						signature.Write(i.MemoryLayout().Layout())
						signature.WriteByte(0x0)
						signature.WriteByte(0x0)
					}

				}
				bytes := signature.Bytes()
				sha256 := sha256.Sum256(bytes)
				hash := hex.EncodeToString(sha256[:])

				out.layout = bytes
				out.hash = hash

				rt.signature = out
			}
		default:
			{
				if rt.ConcreteType().Comparable() {
					signature.WriteByte(byte(rt.ConcreteType().Kind()))
					signature.WriteByte(0x0)

					n := binary.PutUvarint(buf, uint64(rt.ConcreteType().Align()))
					signature.Write(buf[:n])
					signature.WriteByte(0x0)

					n = binary.PutUvarint(buf, uint64(rt.ConcreteType().Size()))
					signature.Write(buf[:n])
					signature.WriteByte(0x0)

					signature.WriteByte(byte(rt.ConcreteType().PointerCount()))
					signature.WriteByte(0x0)

					bytes := signature.Bytes()
					sha256 := sha256.Sum256(bytes)
					hash := hex.EncodeToString(sha256[:])

					out.layout = bytes
					out.hash = hash

					rt.signature = out

					return
				}
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

func (rml *rmemoryLayout) Layout() []byte {
	return rml.layout
}

func (rml *rmemoryLayout) HashCode() string {
	return rml.hash
}

func (rml *rmemoryLayout) IdenticalTo(t MemoryLayout) bool {
	if t == nil {
		return false
	}
	return rml.HashCode() == t.HashCode()
}
