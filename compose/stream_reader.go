/*
 * Copyright 2024 CloudWeGo Authors
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
	"io"
	"reflect"

	"github.com/cloudwego/eino/internal/generic"
	"github.com/cloudwego/eino/schema"
)

type streamReader interface {
	copy(n int) []streamReader
	getType() reflect.Type
	getChunkType() reflect.Type
	merge([]streamReader) streamReader
	concatAndMerge([]streamReader, func([]any) (any, error)) streamReader
	withKey(string) streamReader
	close()
	toAnyStreamReader() *schema.StreamReader[any]
}

type streamReaderPacker[T any] struct {
	sr *schema.StreamReader[T]
}

func (srp streamReaderPacker[T]) close() {
	srp.sr.Close()
}

func (srp streamReaderPacker[T]) copy(n int) []streamReader {
	ret := make([]streamReader, n)
	srs := srp.sr.Copy(n)

	for i := 0; i < n; i++ {
		ret[i] = streamReaderPacker[T]{srs[i]}
	}

	return ret
}

func (srp streamReaderPacker[T]) getType() reflect.Type {
	return reflect.TypeOf(srp.sr)
}

func (srp streamReaderPacker[T]) getChunkType() reflect.Type {
	return generic.TypeOf[T]()
}

func (srp streamReaderPacker[T]) merge(isrs []streamReader) streamReader {
	srs := make([]*schema.StreamReader[T], len(isrs)+1)
	srs[0] = srp.sr
	for i := 1; i < len(srs); i++ {
		sr, ok := unpackStreamReader[T](isrs[i-1])
		if !ok {
			return nil
		}

		srs[i] = sr
	}

	sr := schema.MergeStreamReaders(srs)

	return packStreamReader(sr)
}

func (srp streamReaderPacker[T]) concatAndMerge(isrs []streamReader, fn func([]any) (any, error)) streamReader {
	srs := make([]*schema.StreamReader[T], len(isrs)+1)
	srs[0] = srp.sr
	for i := 1; i < len(srs); i++ {
		sr, ok := unpackStreamReader[T](isrs[i-1])
		if !ok {
			return nil
		}

		srs[i] = sr
	}

	sent := false
	ret := schema.StreamReaderFromGenerator(func() (T, error) {
		var zero T
		if sent {
			return zero, io.EOF
		}
		var values []any
		for _, sr := range srs {
			v, err := concatStreamReader(sr)
			if err != nil {
				return zero, fmt.Errorf("concat stream reader: %w", err)
			}
			values = append(values, v)
		}
		merged, err := fn(values)
		if err != nil {
			return zero, fmt.Errorf("merge values: %w", err)
		}
		sent = true
		return merged.(T), nil
	})

	return packStreamReader(ret)
}

func (srp streamReaderPacker[T]) withKey(key string) streamReader {
	convert := func(v T) (map[string]any, error) {
		return map[string]any{key: v}, nil
	}

	ret := schema.StreamReaderWithConvert[T, map[string]any](srp.sr, convert)

	return packStreamReader(ret)
}

func (srp streamReaderPacker[T]) toAnyStreamReader() *schema.StreamReader[any] {
	return schema.StreamReaderWithConvert(srp.sr, func(t T) (any, error) {
		return t, nil
	})
}

func packStreamReader[T any](sr *schema.StreamReader[T]) streamReader {
	return streamReaderPacker[T]{sr}
}

func unpackStreamReader[T any](isr streamReader) (*schema.StreamReader[T], bool) {
	c, ok := isr.(streamReaderPacker[T])
	if ok {
		return c.sr, true
	}

	typ := generic.TypeOf[T]()
	if typ.Kind() == reflect.Interface {
		return schema.StreamReaderWithConvert(isr.toAnyStreamReader(), func(t any) (T, error) {
			return t.(T), nil
		}), true
	}

	return nil, false
}
