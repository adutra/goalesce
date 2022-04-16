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

// Package goalesce is a library for coalescing (a.k.a. merging) objects in Go. It can coalesce any type of object,
// including structs, maps, and slices, even nested ones.
//
// Introduction
//
// The main entry point is the Coalesce function:
//
//   func Coalesce(o1, o2 interface{}, opts ...MainCoalescerOption) (coalesced interface{}, err error)
//
// It merges the 2 values into a single value. When called with no options, the function uses the following default
// algorithm:
//
//   - If both values are nil, return nil.
//   - If one value is nil, return the other value.
//   - If both values are zero values for the type, return the type's zero value.
//   - If one value is a zero value for the type, return the other value.
//   - If both values are non-zero values, the values are coalesced using the following rules:
//     - If both values are pointers, coalesce the values pointed to.
//     - If both values are structs, coalesce the structs recursively, field by field.
//     - If both values are maps, coalesce the maps recursively, key by key.
//     - Otherwise, return the second value.
//
// The Coalesce function can be called with a list of options to modify its default coalescing behavior. See the
// documentation of each option for details.
//
// Advanced usage
//
// The Coalescer interface allows for custom coalescing algorithms to be implemented. By passing custom coalescers to
// the Coalesce function, its behavior can be modified in any way.
package goalesce
