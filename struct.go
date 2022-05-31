// Copyright 2022 Alexandre Dutra
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package goalesce

import (
	"fmt"
	"reflect"
	"strings"
)

// CoalesceStrategyTag is the struct tag used to specify the coalescing strategy to use for a struct field.
const CoalesceStrategyTag = "goalesce"

const (
	// CoalesceStrategyAtomic applies "atomic" semantics.
	CoalesceStrategyAtomic = "atomic"
	// CoalesceStrategyAppend applies "list-append" semantics.
	CoalesceStrategyAppend = "append"
	// CoalesceStrategyUnion applies "set-union" semantics.
	CoalesceStrategyUnion = "union"
	// CoalesceStrategyIndex applies "merge-by-index" semantics.
	CoalesceStrategyIndex = "index"
	// CoalesceStrategyMerge applies "merge-by-key" semantics.
	CoalesceStrategyMerge = "merge"
)

func (c *mainCoalescer) coalesceStruct(v1, v2 reflect.Value) (reflect.Value, error) {
	coalesced := reflect.New(v1.Type()).Elem()
	for i := 0; i < v1.NumField(); i++ {
		field := v1.Type().Field(i)
		if field.IsExported() {
			if fieldCoalescer, err := c.fieldCoalescer(v1.Type(), field); err != nil {
				return reflect.Value{}, err
			} else if coalescedField, err := fieldCoalescer(v1.Field(i), v2.Field(i)); err != nil {
				return reflect.Value{}, err
			} else {
				coalesced.Field(i).Set(coalescedField)
			}
		}
	}
	return coalesced, nil
}

func (c *mainCoalescer) fieldCoalescer(structType reflect.Type, field reflect.StructField) (Coalescer, error) {
	fieldCoalescer, err := c.fieldCoalescerFromTag(structType, field)
	if err != nil {
		return nil, err
	}
	if fieldCoalescer == nil {
		if fieldCoalescers, found := c.fieldCoalescers[structType]; found {
			fieldCoalescer = fieldCoalescers[field.Name]
		}
	}
	if fieldCoalescer == nil {
		fieldCoalescer = c.coalesce
	}
	return fieldCoalescer, nil
}

func (c *mainCoalescer) fieldCoalescerFromTag(structType reflect.Type, field reflect.StructField) (Coalescer, error) {
	coalesceStrategy, found := field.Tag.Lookup(CoalesceStrategyTag)
	if !found {
		return nil, nil
	}
	switch {
	case coalesceStrategy == CoalesceStrategyAtomic:
		return coalesceAtomic, nil
	case coalesceStrategy == CoalesceStrategyAppend:
		return c.appendFieldCoalescer(structType, field)
	case coalesceStrategy == CoalesceStrategyUnion:
		return c.unionFieldCoalescer(structType, field)
	case coalesceStrategy == CoalesceStrategyIndex:
		return c.indexFieldCoalescer(structType, field)
	case strings.HasPrefix(coalesceStrategy, CoalesceStrategyMerge):
		return c.mergeFieldCoalescer(structType, field, coalesceStrategy)
	}
	return nil, fmt.Errorf("field %s.%s: unknown coalesce strategy: %s", structType.String(), field.Name, coalesceStrategy)
}

func (c *mainCoalescer) appendFieldCoalescer(structType reflect.Type, field reflect.StructField) (Coalescer, error) {
	if field.Type.Kind() != reflect.Slice {
		return nil, fmt.Errorf("field %s.%s: append strategy is only supported for slices", structType.String(), field.Name)
	}
	return coalesceSliceAppend, nil
}

func (c *mainCoalescer) unionFieldCoalescer(structType reflect.Type, field reflect.StructField) (Coalescer, error) {
	if field.Type.Kind() != reflect.Slice {
		return nil, fmt.Errorf("field %s.%s: union strategy is only supported for slices", structType.String(), field.Name)
	}
	return func(v1, v2 reflect.Value) (reflect.Value, error) {
		return c.coalesceSliceMerge(v1, v2, SliceUnion)
	}, nil
}

func (c *mainCoalescer) indexFieldCoalescer(structType reflect.Type, field reflect.StructField) (Coalescer, error) {
	if field.Type.Kind() != reflect.Slice {
		return nil, fmt.Errorf("field %s.%s: index strategy is only supported for slices", structType.String(), field.Name)
	}
	return func(v1, v2 reflect.Value) (reflect.Value, error) {
		return c.coalesceSliceMerge(v1, v2, SliceIndex)
	}, nil
}

func (c *mainCoalescer) mergeFieldCoalescer(structType reflect.Type, field reflect.StructField, strategy string) (Coalescer, error) {
	if field.Type.Kind() != reflect.Slice {
		return nil, fmt.Errorf("field %s.%s: merge strategy is only supported for slices", structType.String(), field.Name)
	}
	var key string
	if i := strings.IndexRune(strategy, ','); i != -1 {
		key = strategy[i+1:]
	}
	if key == "" {
		return nil, fmt.Errorf("field %s.%s: %s strategy must be followed by a comma and the merge key", structType.String(), field.Name, CoalesceStrategyMerge)
	}
	elemType := indirect(field.Type.Elem())
	if elemType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("field %s.%s: expecting slice of struct or pointer thereto, got: %s", structType.String(), field.Name, field.Type.String())
	} else if _, found := elemType.FieldByName(key); !found {
		return nil, fmt.Errorf("field %s.%s: slice element type %s has no field named %s", structType.String(), field.Name, elemType.String(), key)
	}
	return func(v1, v2 reflect.Value) (reflect.Value, error) {
		return c.coalesceSliceMerge(v1, v2, newMergeByField(key))
	}, nil
}

// newMergeByField returns a SliceMergeKeyFunc that returns the value of the given struct field for each slice element.
// This function is designed to work on slices of structs, and slices of pointers to structs. When this function
// encounters a pointer while extracting the merge key, it dereferences the pointer; if the pointer was nil, a zero
// value will be used instead, but beware that this may result in nondeterministic merge results.
func newMergeByField(key string) SliceMergeKeyFunc {
	return func(_ int, elem reflect.Value) (reflect.Value, error) {
		// the slice element itself may be a pointer; we want to dereference it and return a zero-value if it's nil.
		deref := safeIndirect(elem)
		if deref.Type().Kind() != reflect.Struct {
			return reflect.Value{}, fmt.Errorf("expecting struct or pointer thereto, got: %s", elem.Type().String())
		}
		field := deref.FieldByName(key)
		if !field.IsValid() {
			return reflect.Value{}, fmt.Errorf("struct type %s has no field named %s", deref.Type().String(), key)
		}
		// the slice element's field may also be a pointer; again, we want to dereference it and return a zero-value
		// if it's nil.
		return safeIndirect(field), nil
	}
}
