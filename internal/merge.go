/*
 * Copyright 2025 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package internal

import (
	"fmt"
	"reflect"

	"github.com/cloudwego/eino/internal/generic"
)

var (
	mergeFuncs = map[reflect.Type]any{}
)

func RegisterFanInMergeFunc[T any](fn func([]T) (T, error)) {
	mergeFuncs[generic.TypeOf[T]()] = fn
}

func GetMergeFunc(tpe reflect.Type) func([]any) (any, error) {
	if fn, ok := mergeFuncs[tpe]; ok {
		return func(vs []any) (any, error) {
			rvs := reflect.MakeSlice(reflect.SliceOf(tpe), 0, len(vs))
			for _, v := range vs {
				if reflect.TypeOf(v) != tpe {
					return nil, fmt.Errorf(
						"(fan-in merge) field type mismatch. expected: '%v', got: '%v'", tpe, reflect.TypeOf(v))
				}
				rvs = reflect.Append(rvs, reflect.ValueOf(v))
			}

			rets := reflect.ValueOf(fn).Call([]reflect.Value{rvs})
			var err error
			if !rets[1].IsNil() {
				err = rets[1].Interface().(error)
			}
			return rets[0].Interface(), err
		}
	}

	return nil
}
