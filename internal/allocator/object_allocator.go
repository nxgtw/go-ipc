// Copyright 2015 Aleksandr Demakin. All rights reserved.

package allocator

import (
	"fmt"
	"reflect"
	"unsafe"
)

const maxObjectSize = 128 * 1024 * 1024

// returns an address of the object stored continuously in the memory
// the object must not contain any references
func valueObjectAddress(v interface{}) unsafe.Pointer {
	const (
		interfaceSize = unsafe.Sizeof(v)
		pointerSize   = unsafe.Sizeof(uintptr(0))
	)
	interfaceBytes := *((*[interfaceSize]byte)(unsafe.Pointer(&v)))
	objRawPointer := *(*unsafe.Pointer)(unsafe.Pointer(&(interfaceBytes[interfaceSize-pointerSize])))
	return objRawPointer
}

// ObjectAddress returns the address of the given object
// if a slice or a pointer is passed, it will returns a pointer to the actual data
func ObjectAddress(object reflect.Value) unsafe.Pointer {
	var addr unsafe.Pointer
	kind := object.Kind()
	if kind == reflect.Slice || kind == reflect.Ptr {
		addr = unsafe.Pointer(object.Pointer())
	} else {
		addr = valueObjectAddress(object.Interface())
	}
	return addr
}

// ObjectSize returns the size of the object.
// If an object is a slice, it returns the size of the entire slice
// If an object is a pointer, it dereferences the pointer and
// returns the size of the underlying object.
func ObjectSize(object reflect.Value) int {
	var size int
	if object.Kind() == reflect.Slice {
		size = object.Len() * int(object.Type().Elem().Size())
	} else if object.Kind() == reflect.Ptr {
		size = int(object.Elem().Type().Size())
	} else {
		size = int(object.Type().Size())
	}
	return size
}

// copyObjectData copies value's data into a byte slice.
// If a slice is passed, it will copy the data it references to.
func copyObjectData(value reflect.Value, memory []byte) {
	addr := ObjectAddress(value)
	size := ObjectSize(value)
	objectData := byteSliceFromUnsafePointer(addr, size, size)
	copy(memory, objectData)
	use(addr)
}

// Alloc copies value's data into a byte slice performing some sanity checks.
// The object either must be a slice, or should be a sort of an object,
// which does not contain any references inside, i.e. should be placed
// in the memory continuously.
// If the object is a pointer it will be dereferenced. To alloc a pointer as is,
// use uintptr or unsafe.Pointer.
// If the object is a slice, only actual data is stored. the calling site
// must save object's lenght and capacity.
func Alloc(memory []byte, object interface{}) error {
	value := reflect.ValueOf(object)
	if !value.IsValid() {
		return fmt.Errorf("inavlid object")
	}
	for value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	size := ObjectSize(value)
	if size > maxObjectSize {
		return fmt.Errorf("the object exceeds max object size of %d", maxObjectSize)
	}
	if size > len(memory) {
		return fmt.Errorf("the object is too large for the buffer")
	}
	if err := checkType(value.Type(), 0); err != nil {
		return err
	}
	copyObjectData(value, memory)
	return nil
}

func intSliceFromMemory(memory []byte, lenght, capacity int) []int {
	sl := reflect.SliceHeader{
		Len:  lenght,
		Cap:  capacity,
		Data: uintptr(unsafe.Pointer(&memory[0])),
	}
	return *(*[]int)(unsafe.Pointer(&sl))
}

func byteSliceFromUnsafePointer(memory unsafe.Pointer, lenght, capacity int) []byte {
	sl := reflect.SliceHeader{
		Len:  lenght,
		Cap:  capacity,
		Data: uintptr(memory),
	}
	return *(*[]byte)(unsafe.Pointer(&sl))
}

// objectByteSlice returns objects underlying byte representation
// the object must stored continuously in the memory, ie must not contain any references.
// slices of plain objects are allowed.
func objectByteSlice(object interface{}) ([]byte, error) {
	value := reflect.ValueOf(object)
	if err := checkType(value.Type(), 0); err != nil {
		return nil, err
	}
	var data []byte
	objSize := ObjectSize(value)
	addr := ObjectAddress(value)
	defer use(unsafe.Pointer(addr))
	data = byteSliceFromUnsafePointer(addr, objSize, objSize)
	return data, nil
}

// CheckObjectReferences checks if an object of type can be safely copied byte by byte.
// the object must not contain any reference types like
// maps, strings, and so on.
// slices or pointers can be at the top level only
func CheckObjectReferences(object interface{}) error {
	return checkType(reflect.ValueOf(object).Type(), 0)
}

func checkType(t reflect.Type, depth int) error {
	kind := t.Kind()
	if kind == reflect.Array {
		return checkType(t.Elem(), depth+1)
	}
	if kind == reflect.Slice {
		if depth != 0 {
			return fmt.Errorf("unexpected slice type")
		}
		return checkType(t.Elem(), depth+1)
	}
	if kind == reflect.Ptr {
		if depth != 0 {
			return fmt.Errorf("unexpected pointer type")
		}
		return checkType(t.Elem(), depth+1)
	}
	if kind == reflect.Struct {
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if err := checkType(field.Type, depth+1); err != nil {
				return fmt.Errorf("field %s: %v", field.Name, err)
			}
		}
		return nil
	}
	return checkNumericType(kind)
}

func checkNumericType(kind reflect.Kind) error {
	if kind >= reflect.Bool && kind <= reflect.Complex128 {
		return nil
	}
	if kind == reflect.UnsafePointer {
		return nil
	}
	return fmt.Errorf("unsupported type %q", kind.String())
}

func use(unsafe.Pointer)
