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
)

func (c *mainCoalescer) coalesceInterface(v1, v2 reflect.Value) (reflect.Value, error) {
	if v1.Elem().Type() != v2.Elem().Type() {
		return v2, nil
	}
	coalesced := reflect.New(v1.Type())
	coalescedTarget, err := c.coalesce(v1.Elem(), v2.Elem())
	if err != nil {
		return reflect.Value{}, err
	}
	coalesced.Elem().Set(coalescedTarget)
	return coalesced.Elem(), nil
}
