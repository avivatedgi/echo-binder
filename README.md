# Echo Binder

A custom binder for the [Echo](https://echo.labstack.com/) web framework that replaces echo's DefaultBinder.
This one supports the same syntax as [gongular](https://github.com/mustafaakin/gongular)'s binder and uses [validator](https://github.com/go-playground/validator) to validate the binded structs.

Most of the time, most of your echo request handlers will start/be filled with binding the body & query data into structures, parsing the headers/path parameters and create a lot of boiler-plate for no reason. This binder aims to reduce the repetitive work while making it easy and user friendly by binding the headers, path, query and body parameters into one struct automatically by using the same `context.Bind` function (without changing your code at all)!

## Features

* Binding URL Query Parameters
* Binding Path Parameters
* Binding Headers
* Binding Body
* Binding Forms
* Struct Validation

## Usage

### Installation

Download [echo-binder](https://github.com/avivatedgi/echo-binder) by using:

```bash
get get -u github.com/avivatedgi/echo-binder
```

And import following in your code:

```go
import "github.com/avivatedgi/echo-binder"
```

Wherever you initiate your `echo` engine, just insert the following line:

```go
e := echo.New()

// Set the echo's binder to use echo-binder instead the DefaultBinder
e.Binder = echo_binder.New()
```

### URL Query Parameters

Query parameters are optional key-value pairs that appear to the right of the `?` in a URL. For example, the following URL has two query params, `sort` and `page`, with respective values `ASC` and `2`:

`http://example.com/articles?sort=ASC&page=2`

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

`http://example.com/users?id=1234&values=1&values=2&values=3`

<details>
  <summary>The struct will be filled with the following values:</summary>

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

</details>

### Path Parameters

Path parameters are variable parts of a URL path. They are typically used to point to a specific resource within a collection, such as a user identified by ID. A URL can have several path parameters, each prefixed with colon `:`. For example the following URL has two path parameters, `userId` and `postId`:

`http://example.com/users/:userId/posts/:postId`

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

`http://example.com/users/1234/posts/5678`

<details>
  <summary>The struct will be filled with the following values:</summary>

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

</details>

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

```http
Accept-Language: en-US,en;q=0.5
User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10.9; rv:50.0) Gecko/20100101 Firefox/50.0
```

<details>
  <summary>The struct will be filled with the following values:</summary>

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

</details>

### Body

The type of the body of the request is indicated by the `Content-Type` header. This functionallity bind the data under the `Body` attribute under your struct, but the logic here is exactly as in [echo](https://echo.labstack.com/)'s body binder.

<details>
  <summary><b>Example</b></summary>

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

</details>

</br>The data will be binded according to the specific `Content-Type` header, if it's `application/json` it will use the json attributes, if it's `application/xml` it will use the xml attributes, etc...

### Forms

Actually, forms are supposed to be also part of the Body binding (in [echo](https://echo.labstack.com/) they actually are, under the `application/x-www-form-urlencoded` Content-Type). So binding forms can be used by two ways:

<details>
  <summary>By the <b>Body</b> attribute</summary>
  
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

</details>

<details>
  <summary>By the <b>Form</b> attribute</summary>
  
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

</details>

### Validation

The structs that are binded by this `Binder` are automatically validated by the `validate` attribute using the [validator](https://github.com/go-playground/validator) package. For more information about the validator check the [documentation](https://pkg.go.dev/github.com/go-playground/validator).

### Notes

* All of the sub-structures in the request (`Path`, `Query`, `Header`, `Body`, `Form`) can have embedded struct 