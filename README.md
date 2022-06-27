# Echo Binder

A custom binder for the [Echo](https://echo.labstack.com/) web framework that replaces echo's DefaultBinder.
This one supports the same syntax as [gongular](https://github.com/mustafaakin/gongular)'s binder and uses [validator](https://github.com/go-playground/validator) to validate the binded structs.

## Features

* Binding URL Query Parameters
* Binding Path Parameters
* Binding Headers
* Binding Body
* Binding Forms
* Struct Validation

## Usage

### Installation

Download [Echo Binder](https://github.com/avivatedgi/echo-binder) by using:

```bash
$ get get -u github.com/avivatedgi/echo-binder
```

And import following in your code:

```go
import "github.com/avivatedgi/echo-binder"
```

Whenever you initiate your `echo` engine, just insert the following line:

```go
e := echo.New()
e.Binder = echo_binder.New()
```

### URL Query Parameters

Query parameters are optional key-value pairs that appear to the right of the `?` in a URL. For example, the following URL has two query params, `sort` and `page`, with respective values `ASC` and `2`:

```http
http://example.com/articles?sort=ASC&page=2
```

Query parameters are case sensitive, and can only hold primitives and slices of primitives, for example for the following structure:

```go
type QueryExample struct {
    Query struct {
        Id      string   `binder:"id"`
        Values  []int    `binder:"values"`
    }
}
```

And the following URL:

```http
http://example.com/users?id=1234&values=1&values=2&values=3
```

The struct will be filled with the following values:

```go
func handler(c echo.Context) error {
    var example QueryExample
    if err := c.Bind(&example); err != nil {
        return err
    }

    fmt.Println(example.Query.Id)         // "1234"
    fmt.Println(example.Query.Values)     // ["1", "2", "3"]
}
```

### Path Parameters

Path parameters are variable parts of a URL path. They are typically used to point to a specific resource within a collection, such as a user identified by ID. A URL can have several path parameters, each prefixed with colon `:`. For example the following URL has two path parameters, `userId` and `postId`:

```http
http://example.com/users/:userId/posts/:postId
```

Path parameters are case sensitive, and can only hold primitives, for example the following structure:

```go
type PathExample struct {
    Path struct {
        UserId      string   `binder:"userId"`
        PostId      string   `binder:"postId"`
    }
}
```

And the following URL:

```http
http://example.com/users/1234/posts/5678
```

The struct will be filled with the following values:

```go
func handler(c echo.Context) error {
    var example PathExample
    if err := c.Bind(&example); err != nil {
        return err
    }

    fmt.Println(example.Path.UserId)      // "1234"
    fmt.Println(example.Path.PostId)      // "5678"
}
```

### Headers

HTTP headers let the client and the server pass additional information with an HTTP request or response. HTTP headers names are case insensitive followed by a colon (`:`), then by its value.

Header values can only hold primitives, for example for the following structure:

```go
type HeaderExample struct {
    Header struct {
        AcceptLanguage  string  `binder:"Accept-Language"`
        UserAgent       string  `binder:"User-Agent"`
    }
}
```

And the following values:

```
Accept-Language: en-US,en;q=0.5
User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10.9; rv:50.0) Gecko/20100101 Firefox/50.0
```

The struct will be filled with the following values:

```go
func handler(c echo.Context) error {
    var example HeaderExample
    if err := c.Bind(&example); err != nil {
        return err
    }

    fmt.Println(example.Header.AcceptLanguage)    // "en-US,en;q=0.5"
    fmt.Println(example.Header.UserAgent)         // "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.9; rv:50.0) Gecko/20100101 Firefox/50.0"
}
```

### Body

The type of the body of the request is indicated by the `Content-Type` header. This functionallity bind the data under the `Body` attribute under your struct, but the logic here is exactly as in [echo](https://echo.labstack.com/)'s body binder.

```go
type BodyExample struct {
    Body struct {
        Username    string      `json:"username" xml:"username"`
        Password    string      `json:"password" xml:"password"`
    }
}

func handler(c echo.Context) error {
    var example BodyExample
    if err := c.Bind(&example); err != nil {
        return err
    }

    fmt.Println(example.Body.Username)    // avivatedgi
    fmt.Println(example.Body.Password)    // *********
}
```

The data will be binded according to the specific `Content-Type` header, if it's `application/json` it will use the json attributes, if it's `application/xml` it will use the xml attributes, etc...

### Forms

Actually, forms are supposed to be also part of the Body binding (in [echo](https://echo.labstack.com/) they actually are, under the `application/x-www-form-urlencoded` Content-Type). So binding forms can be used by two ways:

* By the `Body` attribute:

```go
type FormBodyExample struct {
    Body struct {
        Username    string  `form:"username"`
        Password    string  `form:"password"`
    }
}

func handler(c echo.Context) error {
    var example FormBodyExample
    if err := c.Bind(&example); err != nil {
        return err
    }

    fmt.Println(example.Body.Username)    // avivatedgi
    fmt.Println(example.Body.Password)    // *********
}
```

* By the `Form` attribute:

```go
type FormExample struct {
    Form struct {
        Username    string  `binder:"username"`
        Password    string  `binder:"password"`
    }
}

func handler(c echo.Context) error {
    var example FormExample
    if err := c.Bind(&example); err != nil {
        return err
    }

    fmt.Println(example.Form.Username)    // avivatedgi
    fmt.Println(example.Form.Password)    // *********
}
```

### Validation

The structs that are binded by this `Binder` are automatically validated by the `validate` attribute using the [validator](https://github.com/go-playground/validator) package.
