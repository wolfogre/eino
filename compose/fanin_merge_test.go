package compose

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cloudwego/eino/schema"
)

func Test_mergeValues(t *testing.T) {
	t.Run("merge maps", func(t *testing.T) {
		m1 := map[int]int{1: 1, 2: 2, 3: 3, 4: 4}
		m2 := map[int]int{5: 5, 6: 6, 7: 7, 8: 8}
		m3 := map[int]int{9: 9, 10: 10, 11: 11}
		mergedM, err := mergeValues([]any{m1, m2, m3})
		assert.Nil(t, err)

		m := mergedM.(map[int]int)

		// len(m) == len(m1) + len(m2) + len(m3)
		assert.Equal(t, len(m), len(m1)+len(m2)+len(m3))

		_, err = mergeValues([]any{m1, m2, m3, map[int]int{1: 1}})
		assert.NotNil(t, err)

		_, err = mergeValues([]any{m1, m2, m3, map[int]string{1: "1"}})
		assert.NotNil(t, err)
	})

	t.Run("merge stream", func(t *testing.T) {
		ass := []any{
			packStreamReader(schema.StreamReaderFromArray[map[int]bool]([]map[int]bool{{1: true}})),
			packStreamReader(schema.StreamReaderFromArray[map[int]bool]([]map[int]bool{{2: true}})),
			packStreamReader(schema.StreamReaderFromArray[map[int]bool]([]map[int]bool{{3: true}})),
		}
		isr, err := mergeValues(ass)
		assert.Nil(t, err)
		ret, ok := unpackStreamReader[map[int]bool](isr.(streamReader))
		defer ret.Close()

		// check if merge ret is StreamReader
		assert.True(t, ok)

		for i := 1; i <= 3; i++ {
			num, err := ret.Recv()
			assert.Nil(t, err)

			if num[i] != true {
				t.Fatalf("stream read num:%d is out of expect", i)
			}
		}
		_, err = ret.Recv()
		if err != io.EOF {
			t.Fatalf("stream reader isn't return EOF as expect: %v", err)
		}
	})
}
