package echo_binder

import (
	"net/http"
	"reflect"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type Binder struct {
	validator *validator.Validate
}

func New() *Binder {
	return &Binder{
		validator: validator.New(),
	}
}

func (binder Binder) Bind(i interface{}, c echo.Context) error {
	structType := reflect.TypeOf(i)

	// Make sure that we get a structure to bind
	if structType.Kind() != reflect.Ptr {
		return BadRequestError(ErrorInvalidType)
	}

	// Get the actual element instead of the pointer
	structType = structType.Elem()

	// Check that the data is actually a struct
	if structType.Kind() != reflect.Struct {
		return BadRequestError(ErrorInvalidType)
	}

	structValue := reflect.ValueOf(i).Elem()

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
		if kind != reflect.Struct {
			return BadRequestError(GetInvalidTypeAtLocationError(typeField.Name))
		}

		// Get the structField of the field
		structField := structValue.Field(i)
		if err := handler(c, &structField); err != nil {
			return err
		}
	}

	if binder.validator != nil {
		if err := binder.validator.Struct(i); err != nil {
			return BadRequestError(err)
		}
	}

	return nil
}

type structFieldData struct {
	FieldName string
	Value     *reflect.Value
}

var fieldHandlers = map[string]func(echo.Context, *reflect.Value) error{
	PathField:   bindPath,
	QueryField:  bindQuery,
	BodyField:   bindBody,
	FormField:   bindForm,
	HeaderField: bindHeader,
}

func bindPath(c echo.Context, structField *reflect.Value) error {
	fields := getStructFields(structField)
	names := c.ParamNames()
	values := c.ParamValues()

	for i := 0; i < len(names); i++ {
		name := names[i]

		field, ok := fields[name]
		if !ok {
			// Didn't found a field to bound to this path parameter, should return a bad request error.
			return BadRequestError(GetMissingParamAtLocationError(PathField, name))
		}

		if !field.Value.CanSet() {
			// The field is not settable, should return an error
			return BadRequestError(GetNotSettableParamAtLocationError(PathField, name))
		}

		if err := setWithProperType(field.Value.Kind(), values[i], field.Value); err != nil {
			return BadRequestError(err)
		}
	}

	return nil
}

func bindQuery(c echo.Context, structField *reflect.Value) error {
	// Check if the method is valid for the query binding
	method := c.Request().Method
	if method != http.MethodGet && method != http.MethodDelete && method != http.MethodHead {
		return BadRequestError(GetUnsupportedHttpMethodError(QueryField, method))
	}

	fields := getStructFields(structField)
	params := c.QueryParams()

	for name, values := range params {
		field, ok := fields[name]
		if !ok {
			// Didn't found a field to bound to this query parameter, continue
			continue
		}

		if !field.Value.CanSet() {
			// The field is not settable, should return an error
			return BadRequestError(GetNotSettableParamAtLocationError(QueryField, name))
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
					return BadRequestError(err)
				}
			}

			// Set the slice to the field
			field.Value.Set(slice)

		default:
			if err := setWithProperType(field.Value.Kind(), values[0], field.Value); err != nil {
				return BadRequestError(err)
			}
		}
	}

	return nil
}

func bindBody(c echo.Context, structField *reflect.Value) error {
	// Check if the method is valid for body binding
	if c.Request().Method == http.MethodGet {
		return BadRequestError(GetUnsupportedHttpMethodError(BodyField, c.Request().Method))
	}

	// We want the body to bind exactly like echo does
	return new(echo.DefaultBinder).BindBody(c, structField.Addr().Interface())
}

func bindForm(c echo.Context, structField *reflect.Value) error {
	// Check if the method is valid for body binding
	if c.Request().Method == http.MethodGet {
		return BadRequestError(GetUnsupportedHttpMethodError(BodyField, c.Request().Method))
	}

	fields := getStructFields(structField)

	values, err := c.FormParams()
	if err != nil {
		return BadRequestError(err)
	}

	for name, values := range values {
		field, ok := fields[name]
		if !ok {
			// Didn't found a field to bound to this form parameter, continue
			continue
		}

		if !field.Value.CanSet() {
			// The field is not settable, should return an error
			return BadRequestError(GetNotSettableParamAtLocationError(FormField, name))
		}

		switch field.Value.Type().Kind() {
		case reflect.Slice:
			sliceKind := field.Value.Type().Elem().Kind()
			slice := reflect.MakeSlice(field.Value.Type(), len(values), len(values))

			// Build the slice with the values
			for i := 0; i < len(values); i++ {
				value := slice.Index(i)
				if err := setWithProperType(sliceKind, values[i], &value); err != nil {
					return BadRequestError(err)
				}
			}

			// Set the slice to the field
			field.Value.Set(slice)

		default:
			if err := setWithProperType(field.Value.Kind(), values[0], field.Value); err != nil {
				return BadRequestError(err)
			}
		}
	}

	return nil
}

func bindHeader(c echo.Context, structField *reflect.Value) error {
	fields := getStructFields(structField)
	header := c.Request().Header

	for name, field := range fields {
		headerValue := header.Get(name)
		if headerValue == "" {
			continue
		}

		if !field.Value.CanSet() {
			// The field is not settable, should return an error
			return BadRequestError(GetNotSettableParamAtLocationError(HeaderField, field.FieldName))
		}

		if err := setWithProperType(field.Value.Kind(), headerValue, field.Value); err != nil {
			return BadRequestError(err)
		}
	}

	return nil
}

// Returns a map of string to reflect.StructField out of a reflect.Value
// This function assumes that the reflect.Value is a struct, and it will panic if it is not
func getStructFields(structField *reflect.Value) map[string]*structFieldData {
	fields := make(map[string]*structFieldData)

	for i := 0; i < structField.Type().NumField(); i++ {
		fieldType := structField.Type().Field(i)
		fieldStruct := structField.Field(i)

		// If the field is an anonymous field, we need to get the fields of the struct it points to
		if fieldType.Anonymous {
			// Check if the field is a pointer to a structure
			if fieldType.Type.Kind() == reflect.Ptr && fieldType.Type.Elem().Kind() == reflect.Struct {
				// If the field is nil, so we need to set it to a default value
				if fieldStruct.IsNil() {
					fieldStruct.Set(reflect.New(fieldType.Type.Elem()))
				}

				fieldStruct = fieldStruct.Elem()
			}

			tempFields := getStructFields(&fieldStruct)

			for name, field := range tempFields {
				fields[name] = field
			}

			continue
		}

		identifier := fieldType.Tag.Get(TagIdentifier)
		if identifier == "" {
			identifier = fieldType.Name
		}

		fields[identifier] = &structFieldData{FieldName: fieldType.Name, Value: &fieldStruct}
	}

	return fields
}
