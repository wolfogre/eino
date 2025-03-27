package compose

import (
	"fmt"
	"reflect"

	"github.com/cloudwego/eino/internal"
)

func RegisterFanInMergeFunc[T any](fn func([]T) (T, error)) {
	internal.RegisterFanInMergeFunc(fn)
}

func mergeMap(vs []any) (any, error) {
	typ := reflect.TypeOf(vs[0])
	merged := reflect.MakeMap(typ)
	for _, v := range vs {
		if reflect.TypeOf(v) != typ {
			return nil, fmt.Errorf(
				"(mergeMap) field type mismatch. expected: '%v', got: '%v'", typ, reflect.TypeOf(v))
		}

		iter := reflect.ValueOf(v).MapRange()
		for iter.Next() {
			key, val := iter.Key(), iter.Value()
			if merged.MapIndex(key).IsValid() {
				return nil, fmt.Errorf("(mergeMap) duplicated key ('%v') found", key.Interface())
			}
			merged.SetMapIndex(key, val)
		}
	}

	return merged.Interface(), nil
}

// the caller should ensure len(vs) > 1
func mergeValues(vs []any) (any, error) {
	v0 := reflect.ValueOf(vs[0])
	t0 := v0.Type()
	k0 := t0.Kind()

	if fn := internal.GetMergeFunc(t0); fn != nil {
		return fn(vs)
	}

	if k0 == reflect.Map {
		return mergeMap(vs)
	}

	if s, ok := vs[0].(streamReader); ok {
		if t := s.getChunkType(); t.Kind() != reflect.Map && internal.GetMergeFunc(t) == nil {
			return nil, fmt.Errorf("(mergeValues | stream type)"+
				" unsupported chunk type: %v", s.getChunkType())
		}

		ss := make([]streamReader, len(vs)-1)
		for i := 0; i < len(ss); i++ {
			s_, ok_ := vs[i+1].(streamReader)
			if !ok_ {
				return nil, fmt.Errorf("(mergeStream) unexpected type. "+
					"expect: %v, got: %v", t0, reflect.TypeOf(vs[i]))
			}

			if s_.getChunkType() != s.getChunkType() {
				return nil, fmt.Errorf("(mergeStream) chunk type mismatch. "+
					"expect: %v, got: %v", s.getChunkType(), s_.getChunkType())
			}

			ss[i] = s_
		}

		ms := s.merge(ss)

		return ms, nil
	}

	return nil, fmt.Errorf("(mergeValues) unsupported type: %v", t0)
}
