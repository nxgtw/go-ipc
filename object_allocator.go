// Copyright 2015 Aleksandr Demakin. All rights reserved.

package ipc

import (
	"fmt"
	"reflect"
	"unsafe"
)

const maxObjectSize = 128 * 1024 * 1024

// returns an address of the object stored continuously in the memory
// the object must not contain any references
func valueObjectAddress(v interface{}) uintptr {
	const (
		interfaceSize = unsafe.Sizeof(v)
		pointerSize   = unsafe.Sizeof(uintptr(0))
	)
	interfaceBytes := *((*[interfaceSize]byte)(unsafe.Pointer(&v)))
	objRawPointer := *(*uintptr)(unsafe.Pointer(&(interfaceBytes[interfaceSize-pointerSize])))
	return objRawPointer
}

// returns the address of the given object
// if a slice is passed, it will returns a pointer to the actual data
func objectAddress(object interface{}, kind reflect.Kind) uintptr {
	var addr uintptr
	addr = valueObjectAddress(object)
	if kind == reflect.Slice {
		header := *(*reflect.SliceHeader)(unsafe.Pointer(addr))
		addr = header.Data
	}
	return addr
}

func objectSize(object reflect.Value) int {
	t := object.Type()
	size := t.Size()
	if object.Kind() == reflect.Slice {
		size = uintptr(object.Len()) * t.Elem().Size()
	}
	return int(size)
}

// copies value's data into a byte slice.
// if a slice is passed, it will copy data it references to
func copyObjectData(value reflect.Value, memory []byte) {
	addr := objectAddress(value.Interface(), value.Kind())
	size := objectSize(value)
	objectData := *((*[maxObjectSize]byte)(unsafe.Pointer(addr)))
	copy(memory, objectData[:size])
}

// copies value's data into a byte slice performing soem sanity checks.
// the object either must be a slice, or should be a sort of an object,
// which does not contain any references inside, i.e. should be placed
// in the memory continuously.
// if the object is a slice, only actual data is stored. the calling site
// must save object's lenght and capacity
func alloc(memory []byte, object interface{}) error {
	value := reflect.ValueOf(object)
	if !value.IsValid() {
		return fmt.Errorf("inavlid object")
	}
	size := objectSize(value)
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

func byteSliceAddress(memory []byte) uintptr {
	return uintptr(unsafe.Pointer(&(memory[0])))
}

func intSliceFromMemory(memory []byte, lenght, capacity int) []int {
	sl := reflect.SliceHeader{
		Len:  lenght,
		Cap:  capacity,
		Data: byteSliceAddress(memory),
	}
	return *(*[]int)(unsafe.Pointer(&sl))
}

// checks if an object of type can be safely copied by byte.
// the object must not contain any reference types like
// maps, strings, pointers and so on.
// slices can be at the top level only
func checkObject(object interface{}) error {
	return checkType(reflect.ValueOf(object).Type(), 0)
}

func checkType(t reflect.Type, depth int) error {
	kind := t.Kind()
	if kind == reflect.Array {
		return checkType(t.Elem(), depth+1)
	}
	if kind == reflect.Slice {
		if depth != 0 {
			return fmt.Errorf("slices as array elems or struct fields are not supported")
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
