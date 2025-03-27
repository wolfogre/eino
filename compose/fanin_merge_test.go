package compose

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cloudwego/eino/schema"
)

func Test_mergeValues(t *testing.T) {
	t.Run("merge maps", func(t *testing.T) {
		m1 := map[int]int{1: 1, 2: 2, 3: 3, 4: 4}
		m2 := map[int]int{5: 5, 6: 6, 7: 7, 8: 8}
		m3 := map[int]int{9: 9, 10: 10, 11: 11}

		t.Run("regular", func(t *testing.T) {
			mergedM, err := mergeValues([]any{m1, m2, m3})
			assert.NoError(t, err)

			m := mergedM.(map[int]int)

			// len(m) == len(m1) + len(m2) + len(m3)
			assert.Equal(t, len(m), len(m1)+len(m2)+len(m3))
		})

		t.Run("duplicated key", func(t *testing.T) {
			_, err := mergeValues([]any{m1, m2, m3, map[int]int{1: 1}})
			assert.ErrorContains(t, err, "duplicated key")
		})

		t.Run("type mismatch", func(t *testing.T) {
			_, err := mergeValues([]any{m1, m2, m3, map[int]string{1: "1"}})
			assert.ErrorContains(t, err, "type mismatch")
		})
	})

	t.Run("merge stream", func(t *testing.T) {
		ass := []any{
			packStreamReader(schema.StreamReaderFromArray[map[int]string]([]map[int]string{{1: "1"}})),
			packStreamReader(schema.StreamReaderFromArray[map[int]string]([]map[int]string{{2: "2"}})),
			packStreamReader(schema.StreamReaderFromArray[map[int]string]([]map[int]string{{3: "3", 4: "4"}})),
		}
		isr, err := mergeValues(ass)
		require.NoError(t, err)
		ret, ok := unpackStreamReader[map[int]string](isr.(streamReader))
		require.True(t, ok)
		defer ret.Close()

		got := make(map[int]string)
		for i := 0; i < 3; i++ {
			m, err := ret.Recv()
			require.NoError(t, err)
			for k, v := range m {
				got[k] = v
			}
		}
		_, err = ret.Recv()
		require.ErrorIs(t, err, io.EOF)

		assert.Equal(t, map[int]string{
			1: "1",
			2: "2",
			3: "3",
			4: "4",
		}, got)
	})
}
