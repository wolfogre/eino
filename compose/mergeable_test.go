package compose

import (
	"reflect"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

type intMergeable int

var _ Mergeable[intMergeable] = intMergeable(0)

func (m intMergeable) Merge(others ...intMergeable) intMergeable {
	ret := m
	for _, v := range others {
		ret += v
	}
	return ret
}

type structMergeable struct {
	A int
	B []string
}

var _ Mergeable[structMergeable] = structMergeable{}

func (m structMergeable) Merge(others ...structMergeable) structMergeable {
	ret := structMergeable{
		A: m.A,
		B: append([]string{}, m.B...),
	}
	for _, v := range others {
		ret.A += v.A
		ret.B = append(ret.B, v.B...)
	}
	sort.Strings(ret.B)
	return ret
}

type pointerMergeable struct {
	A int
	B []string
}

var _ Mergeable[*pointerMergeable] = (*pointerMergeable)(nil)

func (m *pointerMergeable) Merge(others ...*pointerMergeable) *pointerMergeable {
	ret := pointerMergeable{}
	if m != nil {
		ret.A = m.A
		ret.B = append([]string{}, m.B...)
	}
	for _, v := range others {
		if v == nil {
			continue
		}
		ret.A += v.A
		ret.B = append(ret.B, v.B...)
	}
	sort.Strings(ret.B)
	return &ret
}

type fakeMergeable int

// This is a negative example, it should be:
// var _ Mergeable[fakeMergeable] = fakeMergeable(0)
var _ Mergeable[string] = fakeMergeable(0)

func (m fakeMergeable) Merge(others ...string) string {
	return ""
}

type wrongMethodName int

func (m wrongMethodName) MergeOthers(others ...wrongMethodName) wrongMethodName {
	ret := m
	for _, v := range others {
		ret += v
	}
	return ret
}

func TestIsMergeable(t *testing.T) {
	t.Run(reflect.TypeFor[intMergeable]().Name(), func(t *testing.T) {
		assert.True(t, IsMergeable(intMergeable(0)))
	})

	t.Run(reflect.TypeFor[structMergeable]().Name(), func(t *testing.T) {
		assert.True(t, IsMergeable(structMergeable{}))
	})

	t.Run(reflect.TypeFor[pointerMergeable]().Name(), func(t *testing.T) {
		assert.True(t, IsMergeable((*pointerMergeable)(nil)))
	})

	t.Run(reflect.TypeFor[fakeMergeable]().Name(), func(t *testing.T) {
		assert.False(t, IsMergeable(fakeMergeable(0)))
	})

	t.Run(reflect.TypeFor[wrongMethodName]().Name(), func(t *testing.T) {
		assert.False(t, IsMergeable((*wrongMethodName)(nil)))
	})

	// TODO: more test cases
}

func Test_mergeMergeable(t *testing.T) {
	t.Run(reflect.TypeFor[intMergeable]().Name(), func(t *testing.T) {
		v := intMergeable(1)
		others := []intMergeable{2, 3, 4}
		ret, err := mergeMergeable([]any{v, others[0], others[1], others[2]})
		assert.NoError(t, err)
		assert.Equal(t, intMergeable(10), ret)
	})

	// TODO: more test cases
}
