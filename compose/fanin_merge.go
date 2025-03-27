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

package compose

import (
	"fmt"
	"reflect"

	"github.com/cloudwego/eino/internal"
)

// RegisterFanInMergeFunc registers a function to merge outputs from multiple nodes when fan-in.
// It's used to define how to merge for a specific type.
// For maps that can be merged naturally, you don't need to register a function unless you want to customize the merge logic.
// Please note that it's completely different from RegisterStreamChunkConcatFunc which is used to concat stream chunks.
func RegisterFanInMergeFunc[T any](fn func([]T) (T, error)) {
	internal.RegisterFanInMergeFunc(fn)
}

// the caller should ensure len(vs) > 1
func mergeValues(vs []any) (any, error) {
	v0 := reflect.ValueOf(vs[0])
	t0 := v0.Type()

	if fn := internal.GetMergeFunc(t0); fn != nil {
		return fn(vs)
	}

	if s, ok := vs[0].(streamReader); ok {
		t := s.getChunkType()
		if internal.GetMergeFunc(t) == nil {
			return nil, fmt.Errorf("(mergeValues | stream type)"+
				" unsupported chunk type: %v", t)
		}

		ss := make([]streamReader, len(vs)-1)
		for i := 0; i < len(ss); i++ {
			s_, ok_ := vs[i+1].(streamReader)
			if !ok_ {
				return nil, fmt.Errorf("(mergeStream) unexpected type. "+
					"expect: %v, got: %v", t0, reflect.TypeOf(vs[i]))
			}

			if st := s_.getChunkType(); st != t {
				return nil, fmt.Errorf("(mergeStream) chunk type mismatch. "+
					"expect: %v, got: %v", t, st)
			}

			ss[i] = s_
		}

		ms := s.merge(ss)

		return ms, nil
	}

	return nil, fmt.Errorf("(mergeValues) unsupported type: %v", t0)
}
