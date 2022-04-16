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
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

func TestNewPointerCoalescer(t *testing.T) {
	tests := []struct {
		name string
		opts []PointerCoalescerOption
		want Coalescer
	}{
		{
			name: "no opts",
			opts: nil,
			want: &pointerCoalescer{fallback: &defaultCoalescer{}},
		},
		{
			name: "with opts",
			opts: []PointerCoalescerOption{
				func(c *pointerCoalescer) {
					c.fallback = nil
				},
			},
			want: &pointerCoalescer{fallback: nil},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NewPointerCoalescer(tt.opts...))
		})
	}
}

func Test_pointerCoalescer_Coalesce(t *testing.T) {
	type foo struct {
		Int int
	}
	tests := []struct {
		name string
		v1   interface{}
		v2   interface{}
		want interface{}
	}{
		{
			"ints zeroes",
			(*int)(nil),
			(*int)(nil),
			(*int)(nil),
		},
		{
			"ints",
			intPtr(1),
			intPtr(2),
			intPtr(2),
		},
		{
			"structs zeroes",
			(*foo)(nil),
			(*foo)(nil),
			(*foo)(nil),
		},
		{
			"structs",
			&foo{Int: 1},
			&foo{Int: 2},
			&foo{Int: 2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coalescer := NewPointerCoalescer()
			got, err := coalescer.Coalesce(reflect.ValueOf(tt.v1), reflect.ValueOf(tt.v2))
			require.NoError(t, err)
			assert.Equal(t, tt.want, got.Interface())
		})
	}
	t.Run("type errors", func(t *testing.T) {
		_, err := NewPointerCoalescer().Coalesce(reflect.ValueOf(intPtr(1)), reflect.ValueOf(stringPtr("a")))
		assert.EqualError(t, err, "types do not match: *int != *string")
		_, err = NewPointerCoalescer().Coalesce(reflect.ValueOf(1), reflect.ValueOf(2))
		assert.EqualError(t, err, "values have wrong kind: expected ptr, got int")
	})
	t.Run("fallback error", func(t *testing.T) {
		m := new(mockCoalescer)
		m.On("Coalesce", mock.Anything, mock.Anything).Return(reflect.Value{}, errors.New("fake"))
		m.Test(t)
		coalescer := NewPointerCoalescer()
		coalescer.WithFallback(m)
		_, err := coalescer.Coalesce(reflect.ValueOf(intPtr(1)), reflect.ValueOf(intPtr(2)))
		assert.EqualError(t, err, "fake")
	})
}

func intPtr(i int) *int {
	return &i
}

func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}
