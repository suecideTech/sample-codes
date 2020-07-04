package echomw

import (
	"net/http"

	"github.com/blendle/zapdriver"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// 流用元：https://gist.github.com/ndrewnee/6187a01427b9203b9f11ca5864b8a60d

// ZapLogger is an example of echo middleware that logs requests using logger "zap"
func ZapLogger(log *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if err != nil {
				c.Error(err)
			}

			req := c.Request()
			echoRes := c.Response()

			var res http.Response
			res.StatusCode = echoRes.Status

			n := res.StatusCode
			switch {
			case n >= 500:
				log.Error("Server error", zapdriver.HTTP(zapdriver.NewHTTP(req, &res)))
			case n >= 400:
				log.Warn("Client error", zapdriver.HTTP(zapdriver.NewHTTP(req, &res)))
			case n >= 300:
				log.Info("Redirection", zapdriver.HTTP(zapdriver.NewHTTP(req, &res)))
			default:
				log.Info("Success", zapdriver.HTTP(zapdriver.NewHTTP(req, &res)))
			}

			return nil
		}
	}
}
