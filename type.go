package mapper

import (
	"bytes"
	"encoding/binary"
	"hash/maphash"
	"iter"
	"reflect"
	"sync"

	"github.com/google/uuid"
)

type (
	Kind        = reflect.Kind
	ChanDir     = reflect.ChanDir
	StructTag   = reflect.StructTag
	StructField struct {
		Type      Type
		Name      string
		PkgPath   string
		Tag       StructTag
		Offset    uintptr
		Index     []int
		Anonymous bool

		structField *reflect.StructField
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
		memoryLayout(map[string]Type) MemoryLayout
		PointerCount() int
		ConcreteType() Type

		GoType() reflect.Type
		ID() string
	}

	MemoryLayout interface {
		Layout() []byte
		HashCode() uint64
		IdenticalTo(MemoryLayout) bool
	}

	rtype struct {
		id        string
		t         reflect.Type
		ct        reflect.Type
		ptrCount  int
		signature MemoryLayout
		once      sync.Once

		fieldsOnce   sync.Once
		fields       []StructField
		fieldsByName map[string]StructField

		concreteOnce sync.Once
		concrete     Type

		elemOnce sync.Once
		elem     Type

		keyOnce sync.Once
		key     Type
	}

	rmemoryLayout struct {
		layout []byte
		hash   uint64
	}
)

var (
	types map[reflect.Type]func() Type
	mutx  sync.RWMutex
	seed  maphash.Seed
)

func init() {
	types = make(map[reflect.Type]func() Type)
	seed = maphash.MakeSeed()
}

func newStructField(t *reflect.StructField) StructField {
	if t == nil {
		return StructField{}
	}

	return StructField{
		Name:        t.Name,
		PkgPath:     t.PkgPath,
		Tag:         t.Tag,
		Offset:      t.Offset,
		Index:       t.Index,
		Anonymous:   t.Anonymous,
		Type:        typeOf(t.Type),
		structField: t,
	}
}

func (st StructField) IsExported() bool {
	return st.structField.IsExported()
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
		return val()
	}
	mutx.RUnlock()
	mutx.Lock()
	if val, ok := types[t]; ok {
		mutx.Unlock()
		return val()
	}
	fn := sync.OnceValue(func() Type {
		n, ct := DeReferenceType(t)
		out := &rtype{
			id:       uuid.New().String(),
			t:        t,
			ptrCount: n,
			ct:       ct,
		}
		return out
	})
	types[t] = fn
	mutx.Unlock()
	return fn()
}

func TypeOf(i any) Type {
	return typeOf(reflect.TypeOf(i))
}

func TypeFor[T any]() Type {
	return typeOf(reflect.TypeFor[T]())
}

func (rt *rtype) buildFields() {
	rt.fieldsOnce.Do(func() {
		rt.fields = make([]StructField, rt.NumField())
		rt.fieldsByName = make(map[string]StructField)
		for i := range rt.NumField() {
			f := rt.t.Field(i)
			ref := newStructField(&f)
			rt.fields[i] = ref
			rt.fieldsByName[f.Name] = ref
		}
	})
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
	rt.elemOnce.Do(func() {
		rt.elem = typeOf(rt.t.Elem())
	})
	return rt.elem
}

func (rt *rtype) Field(i int) StructField {
	rt.buildFields()
	return rt.fields[i]
}

func (rt *rtype) Fields() iter.Seq[StructField] {
	rt.buildFields()
	return func(yield func(StructField) bool) {
		for i := 0; i < rt.NumField(); i++ {
			in := rt.fields[i]
			if !yield(in) {
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
	rt.buildFields()
	val, ok := rt.fieldsByName[name]
	return val, ok
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
	rt.keyOnce.Do(func() {
		rt.key = typeOf(rt.t.Key())
	})
	return rt.key
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

func (rt *rtype) ID() string {
	return rt.id
}

func (rt *rtype) MemoryLayout() MemoryLayout {
	return rt.memoryLayout(make(map[string]Type))
}

func (rt *rtype) memoryLayout(lt map[string]Type) MemoryLayout {
	rt.once.Do(func() {
		lt[rt.id] = rt
		defer delete(lt, rt.id)
		out := &rmemoryLayout{}

		signature := bytes.NewBuffer(nil)
		buf := make([]byte, binary.MaxVarintLen64)

		signature.WriteByte(byte(rt.PointerCount()))
		signature.WriteByte(0x0)
		ct := rt.ConcreteType()
		switch ct.Kind() {
		case reflect.Slice:
			{
				signature.WriteByte(byte(reflect.Slice))
				signature.WriteByte(0x0)

				if val, ok := lt[ct.Elem().ID()]; ok {
					signature.WriteString(val.ID())
					signature.WriteByte(0x0)
					break
				}

				signature.Write(ct.Elem().memoryLayout(lt).Layout())
				signature.WriteByte(0x0)
			}
		case reflect.Array:
			{
				signature.WriteByte(byte(reflect.Array))
				signature.WriteByte(0x0)

				n := binary.PutUvarint(buf, uint64(ct.Len()))
				signature.Write(buf[:n])
				signature.WriteByte(0x0)

				if val, ok := lt[ct.Elem().ID()]; ok {
					signature.WriteString(val.ID())
					signature.WriteByte(0x0)
					break
				}
				signature.Write(ct.Elem().memoryLayout(lt).Layout())
				signature.WriteByte(0x0)
			}
		case reflect.Map:
			{
				signature.WriteByte(byte(reflect.Map))
				signature.WriteByte(0x0)

				if val, ok := lt[ct.Key().ID()]; ok {
					signature.WriteString(val.ID())
					signature.WriteByte(0x0)

				} else {
					signature.Write(ct.Key().memoryLayout(lt).Layout())
					signature.WriteByte(0x0)
				}
				if val, ok := lt[ct.Elem().ID()]; ok {
					signature.WriteString(val.ID())
					signature.WriteByte(0x0)

				} else {

					signature.Write(ct.Elem().memoryLayout(lt).Layout())
					signature.WriteByte(0x0)
				}
			}
		case reflect.Struct:
			{
				signature.WriteByte(byte(reflect.Struct))
				signature.WriteByte(0x0)

				for i := range ct.NumField() {
					f := ct.Field(i)
					var n int
					n = binary.PutUvarint(buf, uint64(f.Offset))
					signature.Write(buf[:n])
					signature.WriteByte(0x0)

					for _, i := range f.Index {
						n = binary.PutUvarint(buf, uint64(i))
						signature.Write(buf[:n])
						signature.WriteByte(0x0)
					}

					if val, ok := lt[f.Type.ID()]; ok {
						signature.WriteString(val.ID())
						signature.WriteByte(0x0)
						signature.WriteByte(0x0)
						continue

					}
					signature.Write(f.Type.memoryLayout(lt).Layout())
					signature.WriteByte(0x0)
					signature.WriteByte(0x0)
				}
			}
		case reflect.Func:
			{
				signature.WriteByte(byte(reflect.Func))
				signature.WriteByte(0x0)

				signature.WriteString(rt.Name())
				signature.WriteByte(0x0)

				signature.WriteString(rt.PkgPath())
				signature.WriteByte(0x0)

				for i := range rt.Ins() {
					if val, ok := lt[i.ID()]; ok {
						signature.WriteString(val.ID())
						signature.WriteByte(0x0)
						signature.WriteByte(0x0)
						continue

					}

					signature.Write(i.memoryLayout(lt).Layout())
					signature.WriteByte(0x0)
					signature.WriteByte(0x0)
				}

				for i := range rt.Outs() {
					if val, ok := lt[i.ID()]; ok {
						signature.WriteString(val.ID())
						signature.WriteByte(0x0)
						signature.WriteByte(0x0)
						continue

					}

					signature.Write(i.memoryLayout(lt).Layout())
					signature.WriteByte(0x0)
					signature.WriteByte(0x0)
				}
			}
		case reflect.Interface:
			{
				signature.WriteByte(byte(reflect.Interface))
				signature.WriteByte(0x0)

				for f := range ct.Methods() {
					var n int
					n = binary.PutUvarint(buf, uint64(f.Index))
					signature.Write(buf[:n])
					signature.WriteByte(0x0)

					signature.WriteString(f.PkgPath)
					signature.WriteByte(0x0)

					signature.WriteString(f.Name)
					signature.WriteByte(0x0)

					for i := range f.Type.Ins() {
						if val, ok := lt[i.ID()]; ok {
							signature.WriteString(val.ID())
							signature.WriteByte(0x0)
							signature.WriteByte(0x0)
							continue

						}

						signature.Write(i.memoryLayout(lt).Layout())
						signature.WriteByte(0x0)
						signature.WriteByte(0x0)
					}

					for i := range f.Type.Outs() {
						if val, ok := lt[i.ID()]; ok {
							signature.WriteString(val.ID())
							signature.WriteByte(0x0)
							signature.WriteByte(0x0)
							continue

						}

						signature.Write(i.memoryLayout(lt).Layout())
						signature.WriteByte(0x0)
						signature.WriteByte(0x0)
					}
				}
			}
		default:
			{
				signature.WriteByte(byte(ct.Kind()))
				signature.WriteByte(0x0)

				n := binary.PutUvarint(buf, uint64(ct.Align()))
				signature.Write(buf[:n])
				signature.WriteByte(0x0)

				n = binary.PutUvarint(buf, uint64(ct.Size()))
				signature.Write(buf[:n])
				signature.WriteByte(0x0)

				signature.WriteByte(byte(ct.PointerCount()))
				signature.WriteByte(0x0)
			}
		}
		bytes := signature.Bytes()
		var h maphash.Hash
		h.SetSeed(seed)
		h.Write(bytes)
		out.layout = bytes
		out.hash = h.Sum64()

		rt.signature = out
	})

	return rt.signature
}

func (rt *rtype) PointerCount() int {
	return rt.ptrCount
}

func (rt *rtype) ConcreteType() Type {
	rt.concreteOnce.Do(func() {
		rt.concrete = typeOf(rt.ct)
	})
	return rt.concrete
}

func (rml *rmemoryLayout) Layout() []byte {
	return rml.layout
}

func (rml *rmemoryLayout) HashCode() uint64 {
	return rml.hash
}

func (rml *rmemoryLayout) IdenticalTo(t MemoryLayout) bool {
	if t == nil {
		return false
	}
	return rml.HashCode() == t.HashCode()
}

func DeReferenceType(v reflect.Type) (int, reflect.Type) {
	i := 0
	for ; v.Kind() == reflect.Pointer; i++ {
		v = v.Elem()
	}
	return i, v
}
