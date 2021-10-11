# iocgo
A lightweight IoC dependency injection container for Golang
[![Build status](https://github.com/studyzy/iocgo/workflows/Go/badge.svg)](https://github.com/studyzy/iocgo/actions)
[![Coverage Status](https://coveralls.io/repos/github/studyzy/iocgo/badge.svg)](https://coveralls.io/github/studyzy/iocgo)
# How to use
## Installation
it requires Go 1.15 or newer versions.
install package:

`go get github.com/studyzy/iocgo`

## Examples
### 1. Simple
```go
type Fooer interface {
	Foo(int)
}
type Foo struct {
}

func (Foo)Foo(i int)  {
	fmt.Println("foo:",i)
}
type Barer interface {
	Bar(string)
}
type Bar struct {

}
func (Bar) Bar(s string){
	fmt.Println("bar:",s)
}
type Foobarer interface {
	Say(int,string)
}
type Foobar struct {
	foo Fooer
	bar Barer
}
func NewFoobar(f Fooer,b Barer) Foobarer{
	return &Foobar{
		foo: f,
		bar: b,
	}
}
func (f Foobar)Say(i int ,s string)  {
	f.foo.Foo(i)
	f.bar.Bar(s)
}
func TestContainer_SimpleRegister(t *testing.T) {
	container := NewContainer()
	container.Register(NewFoobar)
	container.Register(func() Fooer { return &Foo{} })
	container.Register(func() Barer { return &Bar{} })
	var fb Foobarer
	container.Resolve(&fb)
	fb.Say(123,"Hello World")
}
```
### 2. Register options
iocgo support below options when register resolver:
* Name
* Optional
* Interface
* Lifestyle(isTransient)
* DependsOn
* Parameters
* Default

How to use these options? see test example:
[container_test.go](https://github.com/studyzy/iocgo/blob/main/container_test.go)

### 3. Register instance
If you already have an instance, you can use :

`RegisterInstance(interfacePtr interface{}, instance interface{}, options ...Option) error `
```go
b := &Bar{}
var bar Barer //interface
container.RegisterInstance(&bar, b) // register interface -> instance
```

### 4. Resolve instance
You can input an interface point to Resolve function, and this function will set correct instance to this interface.
if constructor function return an error, this Resolve function also return this error.
```go
var fb Foobarer
err:=container.Resolve(&fb)
```

**Notice: Resolve first parameter is an interface point, not an interface**

Resolve function also support options, belows are resolve options:
* Arguments
* ResolveName

### 5. Fill a struct fields
Define a struct include some interface fields, call Fill function can fill all interface fields to instance.
```go
type FoobarInput struct {
	foo Fooer
	bar Barer
	msg string
}
input := FoobarInput{
		msg: "studyzy",
	}
	container.Register(func() Fooer { return &Foo{} })
	container.Register(func() Barer { return &Bar{} })
	err := container.Fill(&input)
```
struct fields also support below tags:
* name //Resolve instance by resolver name
* optional //No instance found, keep nil, not throw error
For example:
```
type FoobarInputWithTag struct {
	foo Fooer `optional:"true"`
	bar Barer `name:"baz"`
	msg string
}
```

### 6. Call a function
If you have a function that use interface as parameter, you can direct use Call to invoke this function.
For example:
```
func SayHi1(f Fooer, b Barer) {
	f.Foo(1234)
	b.Bar("hi")
}
Register(func() Fooer { return &Foo{} })
Register(func() Barer { return &Bar{} })
Call(SayHi1)
```
The same as Resolve function, Call function also support options:
* CallArguments
* CallDependsOn

By the way, if invoked function return an error, Call function also return same error. If function return multi values, Call function also return same values as []interface{}

## References:
* https://github.com/golobby/container
* https://github.com/castleproject/Windsor