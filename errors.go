package echo_binder

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

var (
	ErrorInvalidType           = errors.New("binding element must be a pointer to a struct")
	ErrorInvalidAnonymousField = errors.New("binding element cannot have embedded fields that arent struct")
)

func GetInvalidTypeAtLocationError(location string) error {
	return fmt.Errorf("binding element at `%s` must be a struct", location)
}

func GetMissingParamAtLocationError(location, param string) error {
	return fmt.Errorf("missing param `%s` at `%s`", param, location)
}

func GetNotSettableParamAtLocationError(location, param string) error {
	return fmt.Errorf("param `%s` at `%s` is not settable", param, location)
}

func GetUnsupportedHttpMethodError(location, method string) error {
	return fmt.Errorf("unsupported http method `%s` at `%s`", method, location)
}

func GetInvalidAnonymousFieldError(location string) error {
	return fmt.Errorf("binding element at `%s` cannot have embedded fields that arent struct", location)
}

func BadRequestError(err error) *echo.HTTPError {
	return echo.NewHTTPError(http.StatusBadRequest, err.Error()).SetInternal(err)
}
