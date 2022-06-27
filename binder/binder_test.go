package echo_binder

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

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

func TestPathBinder(t *testing.T) {
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
	err := c.Bind(normal)
	if assert.NoError(err) {
		assert.Equal(3, normal.Path.Id)
		assert.Equal("Omri Siniver", normal.Path.Name)
		assert.Equal((*string)(nil), normal.Path.Unused)
	}

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
}

type bodyNormalTester struct {
	Body struct {
		Name string `json:"name"`
	}

	Header struct{}
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

func TestQueryBinder(t *testing.T) {
	assert := assert.New(t)

	e := echo.New()
	e.Binder = New()

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

func TestHeaderBinder(t *testing.T) {
	assert := assert.New(t)

	e := echo.New()
	e.Binder = New()

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
}

type formTester struct {
	Form struct {
		Name string
		Age  int       `binder:"custom"`
		Data []float64 `binder:"data"`
	}
}

func TestFormBinder(t *testing.T) {
	assert := assert.New(t)

	e := echo.New()
	e.Binder = New()

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
