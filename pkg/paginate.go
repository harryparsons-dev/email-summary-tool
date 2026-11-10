package pkg

import (
	"strconv"

	"github.com/labstack/echo/v5"
	"github.com/uptrace/bun"
)

const (
	defaultPageSize = 10
	maxPageSize     = 100
)

// Paginate reads the page and page_size query parameters and returns them
// with a function that applies them to a query. Missing or invalid values
// fall back to the first page and the default page size, and page_size is
// capped at maxPageSize.
func Paginate(c *echo.Context) (int, int, func(q *bun.SelectQuery) *bun.SelectQuery) {
	page, err := strconv.Atoi(c.QueryParam("page"))
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(c.QueryParam("page_size"))
	if err != nil || pageSize < 1 {
		pageSize = defaultPageSize
	}
	pageSize = min(pageSize, maxPageSize)

	return page, pageSize, func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.Limit(pageSize).Offset((page - 1) * pageSize)
	}
}
