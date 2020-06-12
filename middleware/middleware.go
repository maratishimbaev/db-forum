package middleware

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/labstack/echo"
	"time"
)

func TimeMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		fmt.Printf("%s %s - start\n",
			c.Request().Method,
			c.Request().URL)
		start := time.Now()
		err := next(c)
		result := time.Now().Sub(start)
		color.Yellow("%s %s - %.3fs\n",
			c.Request().Method,
			c.Request().URL,
			result.Seconds())
		return err
	}
}
