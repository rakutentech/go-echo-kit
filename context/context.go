package context

import (
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

// CustomContext ...
type CustomContext struct {
	echo.Context
}

// NewCustomContext ...
func NewCustomContext(c echo.Context) CustomContext {
	return CustomContext{c}
}

// ParamUint64 ...
func (cc *CustomContext) ParamUint64(key string) uint64 {
	value, _ := strconv.Atoi(cc.Param(key))
	return uint64(value)
}

// QueryUint64 ...
func (cc *CustomContext) QueryUint64(key string) uint64 {
	value, _ := strconv.Atoi(cc.QueryParam(key))
	return uint64(value)
}

// ParamTime ...
func (cc *CustomContext) ParamTime(key string) (time.Time, error) {
	date := cc.Param(key)
	return stringToTime(date)
}

func stringToTime(date string) (time.Time, error) {
	return time.Parse(time.RFC3339, date)
}
