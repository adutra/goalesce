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
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_zero(t *testing.T) {
	assert.Equal(t, 0, zero[int]())
	assert.Equal(t, "", zero[string]())
	assert.Equal(t, false, zero[bool]())
	type foo struct {
		A int
		B *int
		C *foo
	}
	assert.Equal(t, foo{}, zero[foo]())
	assert.Equal(t, (*int)(nil), zero[*int]())
	assert.Equal(t, []int(nil), zero[[]int]())
	assert.Equal(t, map[string]int(nil), zero[map[string]int]())
}

func Test_cast(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		got, err := cast[int](reflect.ValueOf(123))
		assert.Equal(t, 123, got)
		assert.NoError(t, err)
	})
	t.Run("error", func(t *testing.T) {
		got, err := cast[int](reflect.ValueOf("abc"))
		assert.Equal(t, 0, got)
		assert.EqualError(t, err, "cannot convert string to int")
	})
}

func Test_indirect(t *testing.T) {
	tests := []struct {
		name string
		t    reflect.Type
		want reflect.Type
	}{
		{
			name: "pointer",
			t:    reflect.TypeOf(intPtr(0)),
			want: reflect.TypeOf(0),
		},
		{
			name: "non pointer",
			t:    reflect.TypeOf(0),
			want: reflect.TypeOf(0),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, indirect(tt.t))
		})
	}
}

func Test_safeIndirect(t *testing.T) {
	tests := []struct {
		name string
		v    reflect.Value
		want reflect.Value
	}{
		{
			name: "pointer",
			v:    reflect.ValueOf(intPtr(0)),
			want: reflect.ValueOf(0),
		},
		{
			name: "nil pointer",
			v:    reflect.ValueOf((*int)(nil)),
			want: reflect.ValueOf(0),
		},
		{
			name: "non pointer",
			v:    reflect.ValueOf(0),
			want: reflect.ValueOf(0),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want.Interface(), safeIndirect(tt.v).Interface())
		})
	}
}

func Test_checkZero(t *testing.T) {
	tests := []struct {
		name      string
		v1        reflect.Value
		v2        reflect.Value
		wantValue reflect.Value
		wantDone  bool
	}{
		{"both zero", reflect.ValueOf(0), reflect.ValueOf(0), reflect.ValueOf(0), true},
		{"v1 zero", reflect.ValueOf(0), reflect.ValueOf(1), reflect.ValueOf(1), true},
		{"v2 zero", reflect.ValueOf(1), reflect.ValueOf(0), reflect.ValueOf(1), true},
		{"none zero", reflect.ValueOf(1), reflect.ValueOf(2), reflect.Value{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValue, gotDone := checkZero(tt.v1, tt.v2)
			assert.Equal(t, tt.wantValue, gotValue)
			assert.Equal(t, tt.wantDone, gotDone)
		})
	}
}

func Test_checkTypesMatch(t *testing.T) {
	tests := []struct {
		name    string
		v1      reflect.Type
		v2      reflect.Type
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "same type",
			v1:      reflect.TypeOf(0),
			v2:      reflect.TypeOf(1),
			wantErr: assert.NoError,
		},
		{
			name: "different type",
			v1:   reflect.TypeOf(0),
			v2:   reflect.TypeOf("abc"),
			wantErr: func(t assert.TestingT, err error, args ...interface{}) bool {
				return assert.EqualError(t, err, "types do not match: int != string")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkTypesMatch(tt.v1, tt.v2)
			tt.wantErr(t, err)
		})
	}
}
