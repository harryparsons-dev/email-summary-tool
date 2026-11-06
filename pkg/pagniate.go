package pkg

import (
	"strconv"

	"github.com/labstack/echo/v5"
	"github.com/uptrace/bun"
)

func Paginate(c *echo.Context) (int, int, func(q *bun.SelectQuery) *bun.SelectQuery) {
	page, err := strconv.Atoi(c.QueryParam("page"))
	if err != nil {
		page = 1
	}
	if page == 0 {
		page = 1
	}

	pageSize, err := strconv.Atoi(c.QueryParam("page_size"))
	if err != nil {
		pageSize = 10
	}
	if pageSize == 0 {
		pageSize = 10
	}
	return page, pageSize, func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.Limit(pageSize).Offset((page - 1) * pageSize)
	}
}
