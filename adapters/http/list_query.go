package http

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/banumusa/backend/core/ports"
)

type listQueryOptions struct {
	DefaultPageSize  int
	MaxPageSize      int
	DefaultSortBy    string
	DefaultSortOrder ports.SortOrder
	AllowedSortBy    map[string]struct{}
	AllowedFilters   map[string]struct{}
}

type listQuery struct {
	Params  ports.ListParams
	Filters map[string]string
}

func parseListQuery(r *http.Request, opts listQueryOptions) (listQuery, error) {
	query := listQuery{
		Params: ports.ListParams{
			Page:      1,
			PageSize:  opts.DefaultPageSize,
			SortBy:    opts.DefaultSortBy,
			SortOrder: opts.DefaultSortOrder,
		},
		Filters: make(map[string]string),
	}

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			return query, errors.New("invalid page parameter")
		}
		query.Params.Page = page
	}

	if pageSizeStr := r.URL.Query().Get("pageSize"); pageSizeStr != "" {
		pageSize, err := strconv.Atoi(pageSizeStr)
		if err != nil || pageSize < 1 {
			return query, errors.New("invalid pageSize parameter")
		}
		query.Params.PageSize = pageSize
	}

	if sortBy := r.URL.Query().Get("sortBy"); sortBy != "" {
		if len(opts.AllowedSortBy) > 0 {
			if _, ok := opts.AllowedSortBy[sortBy]; !ok {
				return query, errors.New("invalid sortBy parameter")
			}
		}
		query.Params.SortBy = sortBy
	}

	if sortOrder := r.URL.Query().Get("sortOrder"); sortOrder != "" {
		normalized := ports.SortOrder(strings.ToLower(sortOrder))
		if normalized != ports.SortOrderAsc && normalized != ports.SortOrderDesc {
			return query, errors.New("invalid sortOrder parameter")
		}
		query.Params.SortOrder = normalized
	}

	for filterName := range opts.AllowedFilters {
		if value := r.URL.Query().Get(filterName); value != "" {
			query.Filters[filterName] = value
		}
	}

	query.Params = query.Params.Normalize(
		opts.DefaultPageSize,
		opts.MaxPageSize,
		opts.DefaultSortBy,
		opts.DefaultSortOrder,
	)

	return query, nil
}
