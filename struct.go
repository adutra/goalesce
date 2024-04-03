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

// MergeStrategyTag is the struct tag used to specify the merge strategy to use for a struct field.
const MergeStrategyTag = "goalesce"

const (
	// MergeStrategyAtomic applies "atomic" semantics.
	MergeStrategyAtomic = "atomic"
	// MergeStrategyAppend applies "list-append" semantics.
	MergeStrategyAppend = "append"
	// MergeStrategyUnion applies "set-union" semantics.
	MergeStrategyUnion = "union"
	// MergeStrategyIndex applies "merge-by-index" semantics.
	MergeStrategyIndex = "index"
	// MergeStrategyID applies "merge-by-id" semantics.
	MergeStrategyID = "id"
)

func (c *coalescer) deepMergeStruct(v1, v2 reflect.Value) (reflect.Value, error) {
	// don't fallback to deepCopy if we have custom field mergers
	if value, done := checkZero(v1, v2); done && !c.hasFieldMergers(v1.Type()) {
		return c.deepCopy(value)
	}
	merged := reflect.New(v1.Type()).Elem()
	for i := 0; i < v1.NumField(); i++ {
		field := v1.Type().Field(i)
		if field.IsExported() {
			if fieldMerger, err := c.fieldMerger(v1.Type(), field); err != nil {
				return reflect.Value{}, err
			} else if mergedField, err := fieldMerger(v1.Field(i), v2.Field(i)); err != nil {
				return reflect.Value{}, err
			} else {
				merged.Field(i).Set(mergedField)
			}
		}
	}
	return merged, nil
}

func (c *coalescer) deepCopyStruct(v reflect.Value) (reflect.Value, error) {
	if v.IsZero() {
		return reflect.Zero(v.Type()), nil
	}
	copied := reflect.New(v.Type()).Elem()
	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		if field.IsExported() {
			copiedField, err := c.deepCopy(v.Field(i))
			if err != nil {
				return reflect.Value{}, err
			}
			copied.Field(i).Set(copiedField)
		}
	}
	return copied, nil
}

func (c *coalescer) hasFieldMergers(structType reflect.Type) bool {
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		if field.IsExported() {
			if _, foundTag := field.Tag.Lookup(MergeStrategyTag); foundTag {
				return true
			} else if fieldMergers, foundStruct := c.fieldMergers[structType]; foundStruct {
				if _, foundField := fieldMergers[field.Name]; foundField {
					return true
				}
			}
		}
	}
	return false
}

func (c *coalescer) fieldMerger(structType reflect.Type, field reflect.StructField) (DeepMergeFunc, error) {
	fieldMerger, err := c.fieldMergerFromTag(structType, field)
	if err != nil {
		return nil, err
	}
	if fieldMerger == nil {
		if fieldMergers, foundStruct := c.fieldMergers[structType]; foundStruct {
			if customFieldMerger, foundField := fieldMergers[field.Name]; foundField {
				fieldMerger = func(v1, v2 reflect.Value) (reflect.Value, error) {
					merged, err := customFieldMerger(v1, v2)
					if done, merged, err := checkCustomResult(merged, err, v1.Type()); done {
						return merged, err
					}
					return c.deepMerge(v1, v2)
				}
			}
		}
	}
	if fieldMerger == nil {
		fieldMerger = c.deepMerge
	}
	return fieldMerger, nil
}

func (c *coalescer) fieldMergerFromTag(structType reflect.Type, field reflect.StructField) (DeepMergeFunc, error) {
	mergeStrategy, found := field.Tag.Lookup(MergeStrategyTag)
	if !found {
		return nil, nil
	}
	switch {
	case mergeStrategy == MergeStrategyAtomic:
		return c.deepMergeAtomic, nil
	case mergeStrategy == MergeStrategyAppend:
		return c.appendFieldMerger(structType, field)
	case mergeStrategy == MergeStrategyUnion:
		return c.unionFieldMerger(structType, field)
	case mergeStrategy == MergeStrategyIndex:
		return c.indexFieldMerger(structType, field)
	case strings.HasPrefix(mergeStrategy, MergeStrategyID):
		return c.idFieldMerger(structType, field, mergeStrategy)
	}
	return nil, fmt.Errorf("field %s.%s: unknown merge strategy: %s", structType.String(), field.Name, mergeStrategy)
}

func (c *coalescer) appendFieldMerger(structType reflect.Type, field reflect.StructField) (DeepMergeFunc, error) {
	if field.Type.Kind() != reflect.Slice {
		return nil, fmt.Errorf("field %s.%s: %s strategy is only supported for slices", structType.String(), field.Name, MergeStrategyAppend)
	}
	return c.deepMergeSliceWithListAppend, nil
}

func (c *coalescer) unionFieldMerger(structType reflect.Type, field reflect.StructField) (DeepMergeFunc, error) {
	if field.Type.Kind() != reflect.Slice {
		return nil, fmt.Errorf("field %s.%s: %s strategy is only supported for slices", structType.String(), field.Name, MergeStrategyUnion)
	}
	return func(v1, v2 reflect.Value) (reflect.Value, error) {
		return c.deepMergeSliceWithMergeKey(v1, v2, SliceUnion)
	}, nil
}

func (c *coalescer) indexFieldMerger(structType reflect.Type, field reflect.StructField) (DeepMergeFunc, error) {
	switch field.Type.Kind() {
	case reflect.Slice:
		return func(v1, v2 reflect.Value) (reflect.Value, error) {
			return c.deepMergeSliceWithMergeKey(v1, v2, SliceIndex)
		}, nil
	case reflect.Array:
		return func(v1, v2 reflect.Value) (reflect.Value, error) {
			return c.deepMergeArrayByIndex(v1, v2)
		}, nil
	default:
		return nil, fmt.Errorf("field %s.%s: %s strategy is only supported for slices and arrays", structType.String(), field.Name, MergeStrategyIndex)
	}
}

func (c *coalescer) idFieldMerger(structType reflect.Type, field reflect.StructField, strategy string) (DeepMergeFunc, error) {
	if field.Type.Kind() != reflect.Slice {
		return nil, fmt.Errorf("field %s.%s: %s strategy is only supported for slices", structType.String(), field.Name, MergeStrategyID)
	}
	var key string
	if i := strings.IndexRune(strategy, ':'); i != -1 {
		key = strategy[i+1:]
	}
	if key == "" {
		return nil, fmt.Errorf("field %s.%s: %s strategy must be followed by a colon and the merge key", structType.String(), field.Name, MergeStrategyID)
	}
	elemType := indirect(field.Type.Elem())
	if elemType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("field %s.%s: expecting slice of struct or pointer thereto, got: %s", structType.String(), field.Name, field.Type.String())
	} else if _, found := elemType.FieldByName(key); !found {
		return nil, fmt.Errorf("field %s.%s: slice element type %s has no field named %s", structType.String(), field.Name, elemType.String(), key)
	}
	return func(v1, v2 reflect.Value) (reflect.Value, error) {
		return c.deepMergeSliceWithMergeKey(v1, v2, newMergeByField(key))
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
