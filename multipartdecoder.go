// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schema

import (
	"errors"
	"fmt"
	"mime/multipart"
	"reflect"
)

// NewDecoder returns a new Decoder.
func NewMultipartDecoder() *MultipartDecoder {
	return &MultipartDecoder{Decoder: Decoder{cache: newCache()}}
}

// Decoder decodes values from a map[string][]string to a struct.
type MultipartDecoder struct {
	Decoder
}

// Decode decodes a map[string][]string to a struct.
//
// The first parameter must be a pointer to a struct.
//
// The second parameter is a map, typically url.Values from an HTTP request.
// Keys are "paths" in dotted notation to the struct fields and nested structs.
//
// See the package documentation for a full explanation of the mechanics.
func (d *MultipartDecoder) Decode(dst interface{}, src *multipart.Form) error {
	v := reflect.ValueOf(dst)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return errors.New("schema: interface must be a pointer to struct")
	}

	//normal decode
	errors := MultiError{}
	err := d.Decoder.Decode(dst, src.Value)
	if err != nil {
		switch e := err.(type) {
		case MultiError:
			errors = e
		case ConversionError:
			return e
		case error:
			return e
		}
	}

	//multipart decode
	v = v.Elem()
	t := v.Type()
	for path, multiparts := range src.File {
		if parts, err := d.cache.parsePath(path, t); err == nil {
			if err = d.decodeMultipart(v, path, parts, multiparts); err != nil {
				errors[path] = err
			}
		} else if !d.ignoreUnknownKeys {
			errors[path] = fmt.Errorf("schema: invalid path %q", path)
		}
	}

	if len(errors) > 0 {
		return errors
	}
	return nil
}

// decode fills a struct field using a parsed path.
func (d *MultipartDecoder) decodeMultipart(v reflect.Value, path string, parts []pathPart, values []*multipart.FileHeader) error {
	// Get the field walking the struct fields by index.
	for _, name := range parts[0].path {
		if v.Type().Kind() == reflect.Ptr {
			if v.IsNil() {
				v.Set(reflect.New(v.Type().Elem()))
			}
			v = v.Elem()
		}
		v = v.FieldByName(name)
	}

	// Don't even bother for unexported fields.
	if !v.CanSet() {
		return nil
	}

	// Dereference if needed.
	t := v.Type()
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		if v.IsNil() {
			v.Set(reflect.New(t))
		}
		v = v.Elem()
	}

	// Slice of structs. Let's go recursive.
	if len(parts) > 1 {
		idx := parts[0].index
		if v.IsNil() || v.Len() < idx+1 {
			value := reflect.MakeSlice(t, idx+1, idx+1)
			if v.Len() < idx+1 {
				// Resize it.
				reflect.Copy(value, v)
			}
			v.Set(value)
		}
		return d.decodeMultipart(v.Index(idx), path, parts[1:], values)
	}

	// Simple case.
	if t.Kind() == reflect.Slice {
		var items []reflect.Value
		for _, value := range values {
			items = append(items, reflect.ValueOf(value))
		}
		value := reflect.Append(reflect.MakeSlice(t, 0, 0), items...)
		v.Set(value)
	} else {
		value := (*multipart.FileHeader)(nil)
		// Use the last value provided if any values were provided
		if len(values) > 0 {
			value = values[len(values)-1]
		}
		v.Set(reflect.ValueOf(value).Elem())
	}
	return nil
}
