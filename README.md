# iocgo
A lightweight IoC dependency injection container for Golang

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











## References:
* https://github.com/golobby/container
* https://github.com/castleproject/Windsor