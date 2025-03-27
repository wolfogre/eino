package compose

import "github.com/cloudwego/eino/internal"

func RegisterFanInMergeFunc[T any](fn func([]T) (T, error)) {
	internal.RegisterFanInMergeFunc(fn)
}
