// Copyright 2015 Aleksandr Demakin. All rights reserved.

package ipc

import (
	"fmt"
	"reflect"
	"unsafe"
)

func objectAddress(v interface{}) uintptr {
	const (
		interfaceSize = unsafe.Sizeof(v)
		pointerSize   = unsafe.Sizeof(uintptr(0))
	)
	interfaceBytes := *((*[interfaceSize]byte)(unsafe.Pointer(&v)))
	objRawPointer := *(*uintptr)(unsafe.Pointer(&(interfaceBytes[interfaceSize-pointerSize])))
	return objRawPointer
}

func alloc(memory []byte, object interface{}) error {
	const maxObjectSize = 128 * 1024 * 1024
	size := reflect.ValueOf(object).Type().Size()
	if size > maxObjectSize {
		return fmt.Errorf("the object exceeds max object size of %d", maxObjectSize)
	}
	if int(size) > len(memory) {
		return fmt.Errorf("the object is too large for the buffer")
	}
	value := reflect.ValueOf(object)
	if !value.IsValid() {
		return fmt.Errorf("inavlid object")
	}
	if err := checkType(value.Type(), 0); err != nil {
		return err
	}
	var addr uintptr
	if value.Kind() != reflect.Slice {
		addr = objectAddress(object)
	} else {
		addr = value.Pointer()
	}
	vvalue := *((*[maxObjectSize]byte)(unsafe.Pointer(addr)))
	copy(memory, vvalue[:])
	return nil
}

func checkObject(object interface{}) error {
	return checkType(reflect.ValueOf(object).Type(), 0)
}

// the object must not contain any reference types like
// maps, strings, pointers and so on
func checkType(t reflect.Type, depth int) error {
	kind := t.Kind()
	if kind == reflect.Array {
		return checkType(t.Elem(), depth+1)
	}
	if kind == reflect.Slice {
		if depth != 0 {
			return fmt.Errorf("unsupported slice in a struct")
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
