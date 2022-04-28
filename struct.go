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

const CoalesceStrategyTag = "goalesce"

const (
	CoalesceStrategyAtomic = "atomic"
	CoalesceStrategyAppend = "append"
	CoalesceStrategyUnion  = "union"
	CoalesceStrategyIndex  = "index"
	CoalesceStrategyMerge  = "merge"
)

// StructCoalescerOption is an option to be passed to NewStructCoalescer.
type StructCoalescerOption func(c *structCoalescer)

// NewStructCoalescer creates a new Coalescer for structs.
func NewStructCoalescer(opts ...StructCoalescerOption) Coalescer {
	c := &structCoalescer{
		fallback: &atomicCoalescer{},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// WithFieldCoalescer uses the given Coalescer to coalesce the given struct field.
func WithFieldCoalescer(t reflect.Type, field string, coalescer Coalescer) StructCoalescerOption {
	return func(c *structCoalescer) {
		if c.fieldCoalescers == nil {
			c.fieldCoalescers = make(map[reflect.Type]map[string]Coalescer)
		}
		if c.fieldCoalescers[t] == nil {
			c.fieldCoalescers[t] = make(map[string]Coalescer)
		}
		c.fieldCoalescers[t][field] = coalescer
	}
}

// WithAtomicField causes the given field to be coalesced atomically, that is, with  "atomic" semantics, instead of its
// default coalesce semantics. When 2 non-zero-values of this field are coalesced, the second value is returned as is.
func WithAtomicField(t reflect.Type, field string) StructCoalescerOption {
	return WithFieldCoalescer(t, field, NewAtomicCoalescer())
}

type structCoalescer struct {
	fallback        Coalescer
	fieldCoalescers map[reflect.Type]map[string]Coalescer
}

func (c *structCoalescer) WithFallback(fallback Coalescer) {
	c.fallback = fallback
	for _, coalescers := range c.fieldCoalescers {
		for _, coalescer := range coalescers {
			coalescer.WithFallback(fallback)
		}
	}
}

func (c *structCoalescer) Coalesce(v1, v2 reflect.Value) (reflect.Value, error) {
	if err := checkTypesMatchWithKind(v1, v2, reflect.Struct); err != nil {
		return reflect.Value{}, err
	}
	if value, done := checkZero(v1, v2); done {
		return value, nil
	}
	coalesced := reflect.New(v1.Type()).Elem()
	for i := 0; i < v1.NumField(); i++ {
		field := v1.Type().Field(i)
		if field.IsExported() {
			if fieldCoalescer, err := c.fieldCoalescer(v1.Type(), field); err != nil {
				return reflect.Value{}, err
			} else if coalescedField, err := fieldCoalescer.Coalesce(v1.Field(i), v2.Field(i)); err != nil {
				return reflect.Value{}, err
			} else {
				coalesced.Field(i).Set(coalescedField)
			}
		}
	}
	return coalesced, nil
}

func (c *structCoalescer) fieldCoalescer(structType reflect.Type, field reflect.StructField) (Coalescer, error) {
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
		fieldCoalescer = c.fallback
	}
	return fieldCoalescer, nil
}

func (c *structCoalescer) fieldCoalescerFromTag(structType reflect.Type, field reflect.StructField) (Coalescer, error) {
	coalesceStrategy, found := field.Tag.Lookup(CoalesceStrategyTag)
	if !found {
		return nil, nil
	}
	if coalesceStrategy == CoalesceStrategyAtomic {
		return NewAtomicCoalescer(), nil
	} else if coalesceStrategy == CoalesceStrategyAppend {
		return c.appendFieldCoalescer(structType, field)
	} else if coalesceStrategy == CoalesceStrategyUnion {
		return c.unionFieldCoalescer(structType, field)
	} else if coalesceStrategy == CoalesceStrategyIndex {
		return c.indexFieldCoalescer(structType, field)
	} else if strings.HasPrefix(coalesceStrategy, CoalesceStrategyMerge) {
		return c.mergeFieldCoalescer(structType, field, coalesceStrategy)
	}
	return nil, fmt.Errorf("field %s.%s: unknown coalesce strategy: %s", structType.String(), field.Name, coalesceStrategy)
}

func (c *structCoalescer) appendFieldCoalescer(structType reflect.Type, field reflect.StructField) (Coalescer, error) {
	if field.Type.Kind() != reflect.Slice {
		return nil, fmt.Errorf("field %s.%s: append strategy is only supported for slices", structType.String(), field.Name)
	}
	return &sliceAppendCoalescer{}, nil
}

func (c *structCoalescer) unionFieldCoalescer(structType reflect.Type, field reflect.StructField) (Coalescer, error) {
	if field.Type.Kind() != reflect.Slice {
		return nil, fmt.Errorf("field %s.%s: union strategy is only supported for slices", structType.String(), field.Name)
	}
	return &sliceMergeCoalescer{
		fallback:     c.fallback,
		mergeKeyFunc: SliceUnion,
	}, nil
}

func (c *structCoalescer) indexFieldCoalescer(structType reflect.Type, field reflect.StructField) (Coalescer, error) {
	if field.Type.Kind() != reflect.Slice {
		return nil, fmt.Errorf("field %s.%s: index strategy is only supported for slices", structType.String(), field.Name)
	}
	return &sliceMergeCoalescer{
		fallback:     c.fallback,
		mergeKeyFunc: SliceIndex,
	}, nil
}

func (c *structCoalescer) mergeFieldCoalescer(structType reflect.Type, field reflect.StructField, strategy string) (Coalescer, error) {
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
	return &sliceMergeCoalescer{
		fallback:     c.fallback,
		mergeKeyFunc: newMergeByField(key),
	}, nil
}
