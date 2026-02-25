package response

import "math"

type Collection[T any] struct {
	Items      []T   `json:"items"`
	TotalCount int64 `json:"totalCount"`
	PageSize   int   `json:"pageSize"`
	PageNumber int   `json:"pageNumber"`
	PageCount  int64 `json:"pageCount"`
}

func NewCollection[T any](items []T, total int64, page int, size int) Collection[T] {
	pageCount := int64(math.Ceil(float64(total) / float64(size)))
	return Collection[T]{
		Items:      items,
		TotalCount: total,
		PageSize:   size,
		PageNumber: page,
		PageCount:  pageCount,
	}
}
