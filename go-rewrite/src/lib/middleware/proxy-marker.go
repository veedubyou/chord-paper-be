package middleware

import "github.com/labstack/echo/v4"

func ProxyMarkerOn(next echo.HandlerFunc) echo.HandlerFunc {
	return proxyMarker(next, "on")
}

func ProxyMarkerOff(next echo.HandlerFunc) echo.HandlerFunc {
	return proxyMarker(next, "off")
}

func proxyMarker(next echo.HandlerFunc, proxyStatus string) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set("X-Proxy-Marker", proxyStatus)
		return next(c)
	}
}
