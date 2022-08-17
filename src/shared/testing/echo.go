package testing

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

func PrepareEchoContext(request *http.Request, response http.ResponseWriter) echo.Context {
	e := echo.New()
	return e.NewContext(request, response)
}
