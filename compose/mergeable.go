package compose

import (
	"fmt"
	"reflect"
)

const mergeableMethodName = "Merge"

type Mergeable[T any] interface {
	Merge(others ...T) T
}

func IsMergeable(v any) bool {
	return isMergeable(reflect.TypeOf(v))
}

func isMergeable(t reflect.Type) bool {
	method, ok := t.MethodByName(mergeableMethodName)
	if !ok {
		return false
	}
	if method.Type.NumIn() != 2 { // one receiver and one argument
		return false
	}
	if method.Type.NumOut() != 1 {
		return false
	}

	in := method.Type.In(1)
	if in.Kind() != reflect.Slice {
		return false
	}
	if in.Elem() != t {
		return false
	}

	if method.Type.Out(0) != t {
		return false
	}
	return true
}

func mergeMergeable(vs []any) (any, error) {
	t := reflect.TypeOf(vs[0])
	method, _ := t.MethodByName(mergeableMethodName)

	rvs := make([]reflect.Value, 0, len(vs))
	for _, v := range vs {
		if vt := reflect.TypeOf(v); vt != t {
			return nil, fmt.Errorf(
				"(mergeMergeable) type mismatch. expected: '%v', got: '%v'", t, vt)
		}
		rvs = append(rvs, reflect.ValueOf(v))
	}

	out := method.Func.Call(rvs)
	return out[0].Interface(), nil
}
