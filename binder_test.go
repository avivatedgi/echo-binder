package echo_binder

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

type validEmbedded struct {
	Key1   string `binder:"key1"`
	Key2   int    `binder:"key2"`
	Unbind string `binder:"-"`
}

type allTypes struct {
	Bool      bool     `binder:"bool"`
	Int       int      `binder:"int"`
	Int8      int8     `binder:"int8"`
	Int16     int16    `binder:"int16"`
	Int32     int32    `binder:"int32"`
	Int64     int64    `binder:"int64"`
	Uint      uint     `binder:"uint"`
	Uint8     uint8    `binder:"uint8"`
	Uint16    uint16   `binder:"uint16"`
	Uint32    uint32   `binder:"uint32"`
	Uint64    uint64   `binder:"uint64"`
	Float32   float32  `binder:"float32"`
	Float64   float64  `binder:"float64"`
	String    string   `binder:"string"`
	IntPtr    *int     `binder:"intptr"`
	UintPtr   *uint    `binder:"uintptr"`
	FloatPtr  *float32 `binder:"floatptr"`
	StringPtr *string  `binder:"stringptr"`
	BoolPtr   *bool    `binder:"boolptr"`
}

type invalidEmbedded struct {
	string
}

type pathNormalTester struct {
	Path struct {
		Name   string
		Id     int
		Unused *string
	}
}

type pathTagTester struct {
	Path struct {
		Name   string `binder:"custom"`
		Id     float32
		Unused *string
	}
}

type pathEmbedddedTester struct {
	Path struct {
		string
	}
}

type pathValidEmbeddedTester struct {
	Path struct {
		validEmbedded
	}
}

func TestBinderErrors(t *testing.T) {
	assert := assert.New(t)

	e := echo.New()
	e.Binder = New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/users/:Name/:Id")
	c.SetParamNames("Name", "Id")
	c.SetParamValues("Omri Siniver", "3")

	normal := new(pathNormalTester)
	// Can bind only pointers
	err := c.Bind(*normal)
	assert.Error(err)

	// Can bind only structs
	err = c.Bind(new(int))
	assert.Error(err)

	// Cannot bind Path, Query, Form, Header, and Body if they are not structures.
	type invalidPath struct {
		Path   string
		Query  string
		Form   int
		Header int
		Body   float32
	}

	invalid := invalidPath{}
	err = c.Bind(&invalid)
	assert.Error(err)

	// Can not bind embedded struct if it has invalid embedded fields
	type invalidEmbedded2 struct {
		Path struct {
			invalidEmbedded
		}
	}

	invalid2 := invalidEmbedded2{}
	invalid2.Path.invalidEmbedded.string = ""
	err = c.Bind(&invalid2)
	assert.Error(err)
}

type unhandledStructsTester struct {
	ShouldNotBeHandled struct {
		Id int `json:"Id"`
	}

	ShouldNotBeHandled2 struct {
		Name string `json:"Name"`
	}
}

func TestBinderUnhandledStructs(t *testing.T) {
	assert := assert.New(t)

	e := echo.New()
	e.Binder = New()

	req := httptest.NewRequest(http.MethodGet, "/users?Name=5&Id=5", strings.NewReader(`{"Name":"Omri Siniver", "Id": 7}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/users/:Name/:Id")
	c.SetParamNames("Name", "Id")
	c.SetParamValues("Omri Siniver", "3")

	normal := unhandledStructsTester{
		ShouldNotBeHandled: struct {
			Id int `json:"Id"`
		}{Id: 10},
		ShouldNotBeHandled2: struct {
			Name string `json:"Name"`
		}{Name: "Roy"},
	}
	err := c.Bind(&normal)
	if assert.NoError(err) {
		assert.Equal(10, normal.ShouldNotBeHandled.Id)
		assert.Equal("Roy", normal.ShouldNotBeHandled2.Name)
	}
}

func TestPathBinder(t *testing.T) {
	assert := assert.New(t)
	e := echo.New()
	e.Binder = New()

	// Normal request tests
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/users/:Name/:Id")
	c.SetParamNames("Name", "Id")
	c.SetParamValues("Omri Siniver", "3")

	normal := new(pathNormalTester)
	err := c.Bind(normal)
	if assert.NoError(err) {
		assert.Equal(3, normal.Path.Id)
		assert.Equal("Omri Siniver", normal.Path.Name)
		assert.Equal((*string)(nil), normal.Path.Unused)
	}

	// Test the custom tags
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath("/users/:custom/:Id")
	c.SetParamNames("custom", "Id")
	c.SetParamValues("Omer Chen", "3.53")

	tagged := new(pathTagTester)
	err = c.Bind(tagged)
	if assert.NoError(err) {
		assert.Equal(float32(3.53), tagged.Path.Id)
		assert.Equal("Omer Chen", tagged.Path.Name)
		assert.Equal((*string)(nil), tagged.Path.Unused)
	}

	// Test the custom tags
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath("/users/:custom/:Id")
	c.SetParamNames("custom", "Id")
	c.SetParamValues("Omer Chen", "3.53")

	embedded := new(pathEmbedddedTester)
	err = c.Bind(embedded)
	assert.Error(err)

	// Test the valid embedded struct
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath("/users/:key1/:key2")
	c.SetParamNames("key1", "key2")
	c.SetParamValues("value1", "2")

	validEmbedded := new(pathValidEmbeddedTester)
	err = c.Bind(validEmbedded)
	if assert.NoError(err) {
		assert.Equal("value1", validEmbedded.Path.Key1)
		assert.Equal(2, validEmbedded.Path.Key2)
	}

	// Check the unbind
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath("/users/:key1/:key2/:-")
	c.SetParamNames("key1", "key2", "-")
	c.SetParamValues("value1", "2", "BLIBLI")

	validEmbedded2 := new(pathValidEmbeddedTester)
	validEmbedded.Path.validEmbedded.Unbind = "BLABLA"
	err = c.Bind(validEmbedded2)
	assert.Error(err)
}

type bodyNormalTester struct {
	Body struct {
		Name string `json:"name"`
	}

	Header struct{}
}

type bodyDifferentType struct {
	Body []string
}

type bodyDifferentType2 struct {
	Body int
}

func TestBodyBinder(t *testing.T) {
	// There is no to much to check in here, the logic is mostly echo's,
	// The only logic here is to pass the `struct.Body` into the `echo.DefaultBinder.BindBody`
	assert := assert.New(t)

	e := echo.New()
	e.Binder = New()

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"Omri Siniver"}`))
	rec := httptest.NewRecorder()
	req.Header.Set("Content-Type", "application/json")
	c := e.NewContext(req, rec)

	u := new(bodyNormalTester)
	err := c.Bind(u)
	if assert.NoError(err) {
		assert.Equal("Omri Siniver", u.Body.Name)
	}

	req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`["Daniel", "Israeli"]`))
	rec = httptest.NewRecorder()
	req.Header.Set("Content-Type", "application/json")
	c = e.NewContext(req, rec)

	u2 := new(bodyDifferentType)
	err = c.Bind(u2)
	if assert.NoError(err) {
		assert.Equal(u2.Body, []string{"Daniel", "Israeli"})
	}

	req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`151`))
	rec = httptest.NewRecorder()
	req.Header.Set("Content-Type", "application/json")
	c = e.NewContext(req, rec)

	u3 := new(bodyDifferentType2)
	err = c.Bind(u3)
	if assert.NoError(err) {
		assert.Equal(u3.Body, 151)
	}
}

type queryTester struct {
	Query struct {
		Name      string
		Age       float64
		Data      []string
		OtherData []int
	}
}

type queryTagTester struct {
	Query struct {
		Name string   `binder:"name"`
		Age  float64  `binder:"custom"`
		Data []string `binder:"data"`
	}
}

type queryEmbedddedTester struct {
	Query struct {
		string
	}
}

type queryValidEmbeddedTester struct {
	Query struct {
		validEmbedded
	}
}

func TestQueryBinder(t *testing.T) {
	assert := assert.New(t)

	e := echo.New()
	e.Binder = New()

	// Normal query tester
	req := httptest.NewRequest(http.MethodGet, "/users?Name=Omri&Age=3.14157&Data=1&Data=2&Data=3&OtherData=1&OtherData=2&OtherData=3&F=5", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	normal := new(queryTester)
	err := c.Bind(normal)
	if assert.NoError(err) {
		assert.Equal("Omri", normal.Query.Name)
		assert.Equal(float64(3.14157), normal.Query.Age)
		assert.Equal([]string{"1", "2", "3"}, normal.Query.Data)
		assert.Equal([]int{1, 2, 3}, normal.Query.OtherData)
	}

	// Tester with custom tags
	req = httptest.NewRequest(http.MethodGet, "/users?name=Omri&custom=3.14157&data=1&data=2&data=3&Data=5&OtherData=1&OtherData=2&OtherData=3&F=5", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	custom := new(queryTagTester)
	err = c.Bind(custom)
	if assert.NoError(err) {
		assert.Equal("Omri", custom.Query.Name)
		assert.Equal(float64(3.14157), custom.Query.Age)
		assert.Equal([]string{"1", "2", "3"}, custom.Query.Data)
	}

	// Tester with embeedded field
	req = httptest.NewRequest(http.MethodGet, "/users?name=Omri&custom=3.14157&data=1&data=2&data=3&Data=5&OtherData=1&OtherData=2&OtherData=3&F=5", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	embedded := new(queryEmbedddedTester)
	err = c.Bind(embedded)
	assert.Error(err)

	// Tester with valid embedded field
	req = httptest.NewRequest(http.MethodGet, "/users?name=Omri&custom=3.14157&data=1&data=2&data=3&Data=5&OtherData=1&OtherData=2&OtherData=3&F=5&key1=value1&key2=2", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	validEmbedded := new(queryValidEmbeddedTester)
	err = c.Bind(validEmbedded)
	if assert.NoError(err) {
		assert.Equal("value1", validEmbedded.Query.Key1)
		assert.Equal(2, validEmbedded.Query.Key2)
	}

	// Tester with unbind
	req = httptest.NewRequest(http.MethodGet, "/users?name=Omri&custom=3.14157&data=1&data=2&data=3&Data=5&OtherData=1&OtherData=2&OtherData=3&F=5&key1=value1&key2=2&Unbind=BLIBLI&-=blibli", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	unbind := new(queryValidEmbeddedTester)
	unbind.Query.Unbind = "BLABLA"
	err = c.Bind(unbind)
	if assert.NoError(err) {
		assert.Equal("value1", unbind.Query.Key1)
		assert.Equal(2, unbind.Query.Key2)
		assert.Equal("BLABLA", unbind.Query.Unbind)
	}

	// Test the all types
	req = httptest.NewRequest(http.MethodGet, "/users?bool=true&int=150&int8=127&int16=150&int32=150&int64=150&uint=150&uint8=150&uint16=150&uint32=150&uint64=150&float32=150&float64=150&string=150&intptr=150&uintptr=150&floatptr=150&stringptr=150&boolptr=true", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	type AllTypes struct {
		Query struct {
			allTypes
		}
	}

	all := new(AllTypes)
	err = c.Bind(all)
	if assert.NoError(err) {
		assert.Equal(all.Query.Bool, true)
		assert.Equal(all.Query.Int, int(150))
		assert.Equal(all.Query.Int8, int8(127))
		assert.Equal(all.Query.Int16, int16(150))
		assert.Equal(all.Query.Int32, int32(150))
		assert.Equal(all.Query.Int64, int64(150))
		assert.Equal(all.Query.Uint, uint(150))
		assert.Equal(all.Query.Uint8, uint8(150))
		assert.Equal(all.Query.Uint16, uint16(150))
		assert.Equal(all.Query.Uint32, uint32(150))
		assert.Equal(all.Query.Uint64, uint64(150))
		assert.Equal(all.Query.Float32, float32(150))
		assert.Equal(all.Query.Float64, float64(150))
		assert.Equal(all.Query.String, "150")
		assert.Equal(all.Query.IntPtr, getReference(150))
		assert.Equal(all.Query.UintPtr, getReference[uint](150))
		assert.Equal(all.Query.FloatPtr, getReference[float32](150))
		assert.Equal(all.Query.StringPtr, getReference("150"))
		assert.Equal(all.Query.BoolPtr, getReference(true))
	}
}

func getReference[T any](data T) *T {
	return &data
}

type embeddedHeader struct {
	Omer string `binder:"harari"`
}

type AnotherEmbeddedHeader struct {
	Yaeli string `binder:"yerushalmi"`
}

type headerTester struct {
	Header struct {
		embeddedHeader
		*AnotherEmbeddedHeader
		Name  string
		Build int `binder:"custom"`
	}
}

type headerEmbbeddedFieldTester struct {
	Header struct {
		string
	}
}

func TestHeaderBinder(t *testing.T) {
	assert := assert.New(t)

	e := echo.New()
	e.Binder = New()

	// Test normal header binding
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Request().Header.Set("Name", "Omri")
	c.Request().Header.Set("Custom", "132")
	c.Request().Header.Set("harari", "1234")
	c.Request().Header.Set("yerushalmi", "0525381648")

	normal := new(headerTester)
	err := c.Bind(normal)
	if assert.NoError(err) {
		assert.Equal("Omri", normal.Header.Name)
		assert.Equal(132, normal.Header.Build)
		assert.Equal("1234", normal.Header.embeddedHeader.Omer)
		assert.Equal("0525381648", normal.Header.AnotherEmbeddedHeader.Yaeli)
	}

	// Test invalid embedded type binding
	req = httptest.NewRequest(http.MethodGet, "/users", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	embedded := new(headerEmbbeddedFieldTester)
	err = c.Bind(embedded)
	assert.Error(err)
}

type formTester struct {
	Form struct {
		Name string
		Age  int       `binder:"custom"`
		Data []float64 `binder:"data"`
	}
}

type formEmbeddedTester struct {
	Form struct {
		string
	}
}

type formValidEmbeddedTester struct {
	Form struct {
		validEmbedded
	}
}

func TestFormBinder(t *testing.T) {
	assert := assert.New(t)

	e := echo.New()
	e.Binder = New()

	// Validate normal form binding
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader("Name=Koren&custom=15&data=3.14157&data=152.32&Data=0"))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	normal := new(formTester)
	err := c.Bind(normal)
	if assert.NoError(err) {
		assert.Equal("Koren", normal.Form.Name)
		assert.Equal(15, normal.Form.Age)
		assert.Equal([]float64{3.14157, 152.32}, normal.Form.Data)
	}

	// Validate embedded form binding
	req = httptest.NewRequest(http.MethodPost, "/users", strings.NewReader("Name=Koren&custom=15&data=3.14157&data=152.32&Data=0"))
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	embedded := new(formEmbeddedTester)
	err = c.Bind(embedded)
	assert.Error(err)

	// Validate valid embedded form binding
	req = httptest.NewRequest(http.MethodPost, "/users", strings.NewReader("Name=Koren&custom=15&data=3.14157&data=152.32&Data=0&key1=value1&key2=2"))
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	validEmbedded := new(formValidEmbeddedTester)
	err = c.Bind(validEmbedded)
	if assert.NoError(err) {
		assert.Equal("value1", validEmbedded.Form.Key1)
		assert.Equal(2, validEmbedded.Form.Key2)
	}
}

type validateTester struct {
	Header struct {
		Name    string `validate:"required"`
		Version string
	}
}

func TestValidator(t *testing.T) {
	assert := assert.New(t)

	e := echo.New()
	e.Binder = New()

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Request().Header.Set("Name", "Omri")
	c.Request().Header.Set("Version", "132")

	normal := new(validateTester)
	err := c.Bind(normal)
	if assert.NoError(err) {
		assert.Equal("Omri", normal.Header.Name)
		assert.Equal("132", normal.Header.Version)
	}

	req = httptest.NewRequest(http.MethodGet, "/users", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.Request().Header.Set("Version", "132")

	normal = new(validateTester)
	err = c.Bind(normal)
	assert.Error(err)
}

type bodySentFieldsTester struct {
	Body struct {
		Name   string `json:"name"`
		Age    int    `json:"age"`
		Nested struct {
			Field         bool `json:"field"`
			AnotherNested struct {
				Field bool `json:"field"`
			} `json:"nested"`
		} `json:"nested"`
	}

	BodySentFields RecursiveLookupTable
}

func TestBodySentFieldsBinder(t *testing.T) {
	// There is no to much to check in here, the logic is mostly echo's,
	// The only logic here is to pass the `struct.Body` into the `echo.DefaultBinder.BindBody`
	assert := assert.New(t)

	e := echo.New()
	e.Binder = New()

	data := `{"name":"Omri","age":15,"nested":{"field":true,"nested":{"field":false}}}`

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(data))
	rec := httptest.NewRecorder()
	req.Header.Set("Content-Type", "application/json")
	c := e.NewContext(req, rec)

	u := new(bodySentFieldsTester)
	err := c.Bind(u)
	if assert.NoError(err) {
		assert.Equal("Omri", u.Body.Name)
		assert.Equal(15, u.Body.Age)
		assert.Equal(true, u.Body.Nested.Field)
		assert.Equal(false, u.Body.Nested.AnotherNested.Field)

		assert.True(u.BodySentFields.FieldExists("name"))
		assert.True(u.BodySentFields.FieldExists("age"))
		assert.True(u.BodySentFields.FieldExists("nested"))
		assert.False(u.BodySentFields.FieldExists("nested2"))
		assert.True(u.BodySentFields.FieldExists("nested.nested"))
		assert.True(u.BodySentFields.FieldExists("nested.field"))
		assert.False(u.BodySentFields.FieldExists("nested.field2"))
		assert.True(u.BodySentFields.FieldExists("nested.nested.field"))
	}
}
