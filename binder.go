// A custom binder for the echo web framework that replaces echo's DefaultBinder.
// This one supports the same syntax as gongular's binder and uses go-playground/validator to validate the binded structs.
package echo_binder

import (
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// A replacement for the echo.DefaultBinder that binds the Path, Query, Header, Body and Form params
// into nested structures that passed into the binder, and finally valiate the structure with the go-playground/validator
// package. For more information about the validator check: https://pkg.go.dev/github.com/go-playground/validator
//
// To use this binder, just add it to the echo.Echo instance:
//		e := echo.New()
// 		e.Binder = echo_binder.New()
//
// For example, for this struct defined:
// 		type RequestExample struct {
// 			Body struct {
// 				Name string `json:"name" validate:"required"`
// 			}
//
// 			Query struct {
// 				PostId int `binder:"postId" validate:"required"`
// 			}
//
// 			Path struct {
// 				UserId int `binder:"id" validate:"required"`
// 			}
//
//			Header struct {
// 				AcceptLanguage string `binder:"Accept-Language"`
// 				UserAgent string `binder:"User-Agent"`
// 			}
// 		}
// And this code execution:
// 		func requestHandler(c echo.Context) error {
// 			user := &RequestExample{}
// 			if err := binder.Bind(user, c); err != nil {
// 				return err
// 			}
//
// 			// Do something with the request
// 		}
// The binder will bind the following params:
// From the body, the name field will be bound to the Name field of the struct.
// From the query, the postId field will be bound to the PostId field of the struct.
// From the path, the id field will be bound to the UserId field of the struct.
// From the header, the Accept-Language field will be bound to the AcceptLanguage field of the struct.
// From the header, the User-Agent field will be bound to the UserAgent field of the struct.
type Binder struct {
	validator                    *validator.Validate
	callEchoDefaultBinderOnError bool
	defaultBinder                *echo.DefaultBinder
}

func New() *Binder {
	return &Binder{
		validator:                    validator.New(),
		callEchoDefaultBinderOnError: false,
		defaultBinder:                new(echo.DefaultBinder),
	}
}

func (binder *Binder) CallEchoDefaultBinderOnError(value bool) {
	binder.callEchoDefaultBinderOnError = value
}

func (binder Binder) Bind(i interface{}, c echo.Context) error {
	structType := reflect.TypeOf(i)

	// Make sure that we get a structure to bind
	if structType.Kind() != reflect.Ptr {
		return badRequestError(errorInvalidType)
	}

	// Get the actual element instead of the pointer
	structType = structType.Elem()

	// Check that the data is actually a struct
	if structType.Kind() != reflect.Struct {
		if binder.callEchoDefaultBinderOnError {
			return binder.defaultBinder.Bind(i, c)
		}

		return badRequestError(errorInvalidType)
	}

	structValue := reflect.ValueOf(i).Elem()

	calledHandler := false

	// Iterate over all the fields of the structure and check for the path, query and body members
	for i := 0; i < structType.NumField(); i++ {
		typeField := structType.Field(i)

		// Find the handler for the field by its name
		handler, ok := fieldHandlers[typeField.Name]
		if !ok {
			// Didn't found a handler for this field name, skip it
			continue
		}

		kind := typeField.Type.Kind()

		// If the kind is a pointer get the actual kind
		if kind == reflect.Ptr {
			kind = typeField.Type.Elem().Kind()
		}

		// If the field is not a structure, return an error for that field
		// Only if the field is not a body
		if kind != reflect.Struct && typeField.Name != bodyField {
			if binder.callEchoDefaultBinderOnError {
				return binder.defaultBinder.Bind(i, c)
			}

			return badRequestError(getInvalidTypeAtLocationError(typeField.Name, structTypeString))
		}

		// Get the structField of the field
		structField := structValue.Field(i)
		calledHandler = true
		if err := handler(c, structType, &structValue, &structField); err != nil {
			return badRequestError(err)
		}
	}

	if !calledHandler && binder.callEchoDefaultBinderOnError {
		return binder.defaultBinder.Bind(i, c)
	}

	if binder.validator != nil {
		if err := binder.validator.Struct(i); err != nil {
			return badRequestError(err)
		}
	}

	return nil
}

type structFieldData struct {
	FieldName string
	Value     *reflect.Value
}

var fieldHandlers = map[string]func(echo.Context, reflect.Type, *reflect.Value, *reflect.Value) error{
	pathField:   bindPath,
	queryField:  bindQuery,
	bodyField:   bindBody,
	formField:   bindForm,
	headerField: bindHeader,
}

func bindPath(c echo.Context, structType reflect.Type, structValue *reflect.Value, structField *reflect.Value) error {
	fields, err := getStructFields(structField)
	if err != nil {
		return badRequestError(err)
	}

	names := c.ParamNames()
	values := c.ParamValues()

	for i := 0; i < len(names); i++ {
		name := names[i]

		field, ok := fields[name]
		if !ok {
			// Didn't found a field to bound to this path parameter, should return a bad request error.
			return badRequestError(getMissingParamAtLocationError(pathField, name))
		}

		if !field.Value.CanSet() {
			// The field is not settable, should return an error
			return badRequestError(getNotSettableParamAtLocationError(pathField, name))
		}

		if err := setWithProperType(field.Value.Kind(), values[i], field.Value); err != nil {
			return badRequestError(err)
		}
	}

	return nil
}

func bindQuery(c echo.Context, structType reflect.Type, structValue *reflect.Value, structField *reflect.Value) error {
	// Check if the method is valid for the query binding
	method := c.Request().Method
	if method != http.MethodGet && method != http.MethodDelete && method != http.MethodHead {
		return badRequestError(getUnsupportedHttpMethodError(queryField, method))
	}

	fields, err := getStructFields(structField)
	if err != nil {
		return badRequestError(getInvalidAnonymousFieldError(pathField))
	}

	params := c.QueryParams()

	for name, values := range params {
		field, ok := fields[name]
		if !ok {
			// Didn't found a field to bound to this query parameter, continue
			continue
		}

		if !field.Value.CanSet() {
			// The field is not settable, should return an error
			return badRequestError(getNotSettableParamAtLocationError(queryField, name))
		}

		switch field.Value.Type().Kind() {
		case reflect.Slice:
			// sliceKind := field.StructField.Type.Elem().Kind()
			sliceKind := field.Value.Type().Elem().Kind()
			slice := reflect.MakeSlice(field.Value.Type(), len(values), len(values))

			// Build the slice with the values
			for i := 0; i < len(values); i++ {
				value := slice.Index(i)
				if err := setWithProperType(sliceKind, values[i], &value); err != nil {
					return badRequestError(err)
				}
			}

			// Set the slice to the field
			field.Value.Set(slice)

		default:
			if err := setWithProperType(field.Value.Kind(), values[0], field.Value); err != nil {
				return badRequestError(err)
			}
		}
	}

	return nil
}

func bindBody(c echo.Context, structType reflect.Type, structValue *reflect.Value, structField *reflect.Value) (err error) {
	request := c.Request()

	// Check if the method is valid for body binding and if there is content in the body
	if request.Method == http.MethodGet {
		return badRequestError(getUnsupportedHttpMethodError(bodyField, request.Method))
	} else if request.ContentLength == 0 {
		return nil
	}

	// Check if the content type is valid for body binding
	contentType := request.Header.Get(echo.HeaderContentType)

	body, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return internalServerError(err)
	}

	switch {
	case strings.HasPrefix(contentType, echo.MIMEApplicationJSON):
		if err := json.Unmarshal(body, structField.Addr().Interface()); err != nil {
			return badRequestError(err)
		}

	case strings.HasPrefix(contentType, echo.MIMEApplicationXML), strings.HasPrefix(contentType, echo.MIMETextXML):
		if err := xml.Unmarshal(body, structField.Addr().Interface()); err != nil {
			return badRequestError(err)
		}
	}

	if structField.Type().Kind() != reflect.Struct {
		// If the body is not a struct, no need to fill the BodySentFields field.
		return nil
	}

	field, found := structType.FieldByName(bodySentFields)
	if !found {
		// Didn't found the body sent field, so we just don't bind it.
		return nil
	} else if field.Type.Kind() != reflect.TypeOf(RecursiveLookupTable{}).Kind() {
		return badRequestError(getInvalidTypeAtLocationError(bodySentFields, lookupTypeString))
	}

	fieldValue := structValue.FieldByName(bodySentFields)
	if !fieldValue.CanSet() {
		return badRequestError(getNotSettableParamAtLocationError(structValue.Type().Name(), bodySentFields))
	}

	data := lookupTable{}
	if err := json.Unmarshal(body, &data); err != nil {
		return badRequestError(err)
	}

	fieldValue.Set(reflect.ValueOf(data.IntoRecursiveLookupTable()))
	return nil
}

func bindForm(c echo.Context, structType reflect.Type, structValue *reflect.Value, structField *reflect.Value) error {
	request := c.Request()

	// Check if the method is valid for body binding and if there is content in the body
	if request.Method == http.MethodGet {
		return badRequestError(getUnsupportedHttpMethodError(bodyField, request.Method))
	} else if request.ContentLength == 0 {
		return nil
	}

	// Check if the content type is valid for form binding
	contentType := request.Header.Get(echo.HeaderContentType)
	if !strings.HasPrefix(contentType, echo.MIMEApplicationForm) && !strings.HasPrefix(contentType, echo.MIMEMultipartForm) {
		return nil
	}

	fields, err := getStructFields(structField)
	if err != nil {
		return badRequestError(getInvalidAnonymousFieldError(formField))
	}

	values, err := c.FormParams()
	if err != nil {
		return badRequestError(err)
	}

	for name, values := range values {
		field, ok := fields[name]
		if !ok {
			// Didn't found a field to bound to this form parameter, continue
			continue
		}

		if !field.Value.CanSet() {
			// The field is not settable, should return an error
			return badRequestError(getNotSettableParamAtLocationError(formField, name))
		}

		switch field.Value.Type().Kind() {
		case reflect.Slice:
			sliceKind := field.Value.Type().Elem().Kind()
			slice := reflect.MakeSlice(field.Value.Type(), len(values), len(values))

			// Build the slice with the values
			for i := 0; i < len(values); i++ {
				value := slice.Index(i)
				if err := setWithProperType(sliceKind, values[i], &value); err != nil {
					return badRequestError(err)
				}
			}

			// Set the slice to the field
			field.Value.Set(slice)

		default:
			if err := setWithProperType(field.Value.Kind(), values[0], field.Value); err != nil {
				return badRequestError(err)
			}
		}
	}

	return nil
}

func bindHeader(c echo.Context, structType reflect.Type, structValue *reflect.Value, structField *reflect.Value) error {
	fields, err := getStructFields(structField)
	if err != nil {
		return badRequestError(getInvalidAnonymousFieldError(headerField))
	}

	header := c.Request().Header

	for name, field := range fields {
		headerValue := header.Get(name)
		if headerValue == "" {
			continue
		}

		if !field.Value.CanSet() {
			// The field is not settable, should return an error
			return badRequestError(getNotSettableParamAtLocationError(headerField, field.FieldName))
		}

		if err := setWithProperType(field.Value.Kind(), headerValue, field.Value); err != nil {
			return badRequestError(err)
		}
	}

	return nil
}

// Returns a map of string to reflect.StructField out of a reflect.Value
// This function assumes that the reflect.Value is a struct, and it will panic if it is not
func getStructFields(structField *reflect.Value) (map[string]*structFieldData, error) {
	fields := make(map[string]*structFieldData)

	for i := 0; i < structField.Type().NumField(); i++ {
		fieldType := structField.Type().Field(i)
		fieldStruct := structField.Field(i)

		// If the field is an anonymous field, we need to get the fields of the struct it points to
		if fieldType.Anonymous {
			kind := fieldType.Type.Kind()

			// If the kind is a pointer let's get the real kind
			if kind == reflect.Ptr {
				kind = fieldType.Type.Elem().Kind()
			}

			// If its not a struct, we can't get the fields of it
			if kind != reflect.Struct {
				return nil, errorInvalidAnonymousField
			}
		}

		kind := fieldType.Type.Kind()
		isPointer := false

		// If the kind is a pointer let's get the real kind
		if kind == reflect.Ptr {
			kind = fieldType.Type.Elem().Kind()
			isPointer = true
		}

		// If the kind is a struct, let's get the fields of it.
		if kind == reflect.Struct {
			if isPointer && fieldStruct.IsNil() {
				fieldStruct.Set(reflect.New(fieldType.Type.Elem()))
				fieldStruct = fieldStruct.Elem()
			}

			tempFields, err := getStructFields(&fieldStruct)
			if err != nil {
				return nil, err
			}

			for name, field := range tempFields {
				fields[name] = field
			}

			continue
		}

		identifier := fieldType.Tag.Get(TagIdentifier)
		if identifier == "" {
			identifier = fieldType.Name
		} else if identifier == "-" {
			// Make sure we don't add this field
			continue
		}

		fields[identifier] = &structFieldData{FieldName: fieldType.Name, Value: &fieldStruct}
	}

	return fields, nil
}
