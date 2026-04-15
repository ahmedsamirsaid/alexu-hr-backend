package ports

type SortOrder string

const (
	SortOrderAsc  SortOrder = "asc"
	SortOrderDesc SortOrder = "desc"
)

type ListParams struct {
	Page      int
	PageSize  int
	SortBy    string
	SortOrder SortOrder
}

func (p ListParams) Normalize(defaultPageSize, maxPageSize int, defaultSortBy string, defaultSortOrder SortOrder) ListParams {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 {
		p.PageSize = defaultPageSize
	}
	if maxPageSize > 0 && p.PageSize > maxPageSize {
		p.PageSize = maxPageSize
	}
	if p.SortBy == "" {
		p.SortBy = defaultSortBy
	}
	if p.SortOrder != SortOrderAsc && p.SortOrder != SortOrderDesc {
		p.SortOrder = defaultSortOrder
	}
	return p
}

func (p ListParams) Offset() int {
	if p.Page < 1 || p.PageSize < 1 {
		return 0
	}
	return (p.Page - 1) * p.PageSize
}

func TotalPages(total, pageSize int) int {
	if total <= 0 || pageSize <= 0 {
		return 0
	}
	return (total + pageSize - 1) / pageSize
}
