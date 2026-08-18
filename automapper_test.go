package automapper

import (
	"errors"
	"reflect"
	"sync"
	"testing"
)

type User struct {
	ID   int
	Name string
}

type UserDTO struct {
	ID   int
	Name string
}

type Nested struct {
	User User
	Age  int
}

type NestedDTO struct {
	User UserDTO
	Age  int
}

func TestCreateCodecFor_Struct(t *testing.T) {
	mapper := CreateCodecFor[User, UserDTO]()

	got, err := mapper(&User{
		ID:   42,
		Name: "alice",
	})
	if err != nil {
		t.Fatal(err)
	}

	want := &UserDTO{
		ID:   42,
		Name: "alice",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestCreateCodecFor_NestedStruct(t *testing.T) {
	mapper := CreateCodecFor[Nested, NestedDTO]()

	got, err := mapper(&Nested{
		User: User{
			ID:   42,
			Name: "alice",
		},
		Age: 30,
	})
	if err != nil {
		t.Fatal(err)
	}

	want := &NestedDTO{
		User: UserDTO{
			ID:   42,
			Name: "alice",
		},
		Age: 30,
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestCreateCodecFor_Slice(t *testing.T) {
	mapper := CreateCodecFor[[]User, []UserDTO]()

	got, err := mapper(&[]User{
		{ID: 1, Name: "a"},
		{ID: 2, Name: "b"},
	})
	if err != nil {
		t.Fatal(err)
	}

	want := &[]UserDTO{
		{ID: 1, Name: "a"},
		{ID: 2, Name: "b"},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestCreateCodecFor_Map(t *testing.T) {
	mapper := CreateCodecFor[map[string]User, map[string]UserDTO]()

	got, err := mapper(&map[string]User{
		"one": {ID: 1, Name: "a"},
		"two": {ID: 2, Name: "b"},
	})
	if err != nil {
		t.Fatal(err)
	}

	want := &map[string]UserDTO{
		"one": {ID: 1, Name: "a"},
		"two": {ID: 2, Name: "b"},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestCreateCodecFor_Pointers(t *testing.T) {
	mapper := CreateCodecFor[User, *UserDTO]()

	got, err := mapper(&User{
		ID:   42,
		Name: "alice",
	})
	if err != nil {
		t.Fatal(err)
	}

	want := &UserDTO{
		ID:   42,
		Name: "alice",
	}

	if !reflect.DeepEqual(got, &want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestCreateCodecFor_Convertible(t *testing.T) {
	mapper := CreateCodecFor[int, int64]()

	got, err := mapper(new(int))
	if err != nil {
		t.Fatal(err)
	}

	if *got != 0 {
		t.Fatalf("got %d, want 0", *got)
	}
}

func TestCreateCodecFor_UnmappedFieldsRemainZero(t *testing.T) {
	type Source struct {
		Name string
	}

	type Target struct {
		Name string
		Age  int
	}

	mapper := CreateCodecFor[Source, Target]()

	got, err := mapper(&Source{Name: "alice"})
	if err != nil {
		t.Fatal(err)
	}

	if got.Name != "alice" {
		t.Fatalf("Name = %q", got.Name)
	}

	if got.Age != 0 {
		t.Fatalf("Age = %d, want 0", got.Age)
	}
}

func TestCreateCodecFor_Unsupported(t *testing.T) {

	mapper := CreateCodecFor[complex128, int]()

	if mapper != nil {
		t.Fatal("expected nil codec")
	}
}

func TestCustomCodec(t *testing.T) {
	SetCustomCodecs([]CodecFactory{
		CreateCustomCodec[User, UserDTO](func(v RValue) (RValue, error) {
			src := v.ConcreteValue().Interface().(User)

			return ValueOf(UserDTO{
				ID:   src.ID * 10,
				Name: src.Name + "!",
			}), nil
		}),
	})

	t.Cleanup(func() {
		SetCustomCodecs(nil)
	})

	mapper := CreateCodecFor[User, UserDTO]()

	got, err := mapper(&User{
		ID:   5,
		Name: "alice",
	})
	if err != nil {
		t.Fatal(err)
	}

	want := &UserDTO{
		ID:   50,
		Name: "alice!",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestCustomCodec_Error(t *testing.T) {
	expected := errors.New("conversion failed")

	SetCustomCodecs([]CodecFactory{
		CreateCustomCodec[User, UserDTO](func(RValue) (RValue, error) {
			return RValue{}, expected
		}),
	})

	t.Cleanup(func() {
		SetCustomCodecs(nil)
	})

	mapper := CreateCodecFor[User, UserDTO]()

	_, err := mapper(&User{})
	if !errors.Is(err, expected) {
		t.Fatalf("got %v, want %v", err, expected)
	}
}

func TestSetCustomCodecs_CopiesSlice(t *testing.T) {
	custom := []CodecFactory{
		CreateCustomCodec[User, UserDTO](func(v RValue) (RValue, error) {
			return ValueOf(UserDTO{ID: 1}), nil
		}),
	}

	SetCustomCodecs(custom)

	// Mutating the caller's slice must not affect the installed codecs.
	custom[0] = nil

	mapper := CreateCodecFor[User, UserDTO]()

	got, err := mapper(&User{})
	if err != nil {
		t.Fatal(err)
	}

	if got.ID != 1 {
		t.Fatalf("got ID %d, want 1", got.ID)
	}

	SetCustomCodecs(nil)
}

func TestCustomCodecs_ConcurrentReads(t *testing.T) {
	SetCustomCodecs([]CodecFactory{
		CreateCustomCodec[User, UserDTO](func(v RValue) (RValue, error) {
			src := v.ConcreteValue().Interface().(User)
			return ValueOf(UserDTO{
				ID:   src.ID,
				Name: src.Name,
			}), nil
		}),
	})

	t.Cleanup(func() {
		SetCustomCodecs(nil)
	})

	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			mapper := CreateCodecFor[User, UserDTO]()
			got, err := mapper(&User{ID: i})
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if got.ID != i {
				t.Errorf("got ID %d, want %d", got.ID, i)
			}
		}(i)
	}

	wg.Wait()
}
