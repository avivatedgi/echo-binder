package echo_binder

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

var (
	errorInvalidType           = errors.New("binding element must be a pointer to a struct")
	errorInvalidAnonymousField = errors.New("binding element cannot have embedded fields that arent struct")
)

func getInvalidTypeAtLocationError(location, requiredType string) error {
	return fmt.Errorf("binding element at `%s` must be a `%s`", location, requiredType)
}

func getMissingParamAtLocationError(location, param string) error {
	return fmt.Errorf("missing param `%s` at `%s`", param, location)
}

func getNotSettableParamAtLocationError(location, param string) error {
	return fmt.Errorf("param `%s` at `%s` is not settable", param, location)
}

func getUnsupportedHttpMethodError(location, method string) error {
	return fmt.Errorf("unsupported http method `%s` at `%s`", method, location)
}

func getInvalidAnonymousFieldError(location string) error {
	return fmt.Errorf("binding element at `%s` cannot have embedded fields that arent struct", location)
}

func badRequestError(err error) *echo.HTTPError {
	return echo.NewHTTPError(http.StatusBadRequest, err.Error()).SetInternal(err)
}

func internalServerError(err error) *echo.HTTPError {
	return echo.NewHTTPError(http.StatusInternalServerError, err.Error()).SetInternal(err)
}
