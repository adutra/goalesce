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
)

// safeIndirect is a variant of reflect.Indirect that returns a zero value if the value is a nil
// pointer. Because of that, this function never returns an invalid value.
func safeIndirect(v reflect.Value) reflect.Value {
	indirect := reflect.Indirect(v)
	if !indirect.IsValid() {
		// nil pointer: return zero value
		indirect = reflect.Zero(v.Type().Elem())
	}
	return indirect
}

func indirect(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

func checkZero(v1, v2 reflect.Value) (reflect.Value, bool) {
	if v1.IsZero() {
		return v2, true
	} else if v2.IsZero() {
		return v1, true
	}
	return reflect.Value{}, false
}

func checkTypesMatch(v1, v2 reflect.Value) error {
	if v1.Type() != v2.Type() {
		return fmt.Errorf("types do not match: %s != %s", v1.Type().String(), v2.Type().String())
	}
	return nil
}

func checkTypesMatchWithKind(v1, v2 reflect.Value, kind reflect.Kind) error {
	if err := checkTypesMatch(v1, v2); err != nil {
		return err
	} else if v1.Kind() != kind {
		return fmt.Errorf("values have wrong kind: expected %s, got %s", kind, v1.Kind())
	}
	return nil
}
