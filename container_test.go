package iocgo

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var log string

func Println(a ...interface{}) {
	msg := fmt.Sprintln(a...)
	log += msg
	fmt.Println(msg)
}

type Fooer interface {
	Foo(int)
	Foo2(string)
}
type Foo struct {
}

func NewFoo() *Foo {
	return &Foo{}
}
func (Foo) Foo(i int) {
	Println("foo:", i)
}
func (Foo) Foo2(s string) {
	Println("foo2:", s)
}

type SubFooer interface {
	Foo2(string)
}
type Barer interface {
	Bar(string)
}
type Bar struct {
}

func NewBar() *Bar {
	return &Bar{}
}
func (Bar) Bar(s string) {
	Println("bar:", s)
}

type Foobarer interface {
	Say(int, string)
}
type Foobar struct {
	foo Fooer
	bar Barer
	msg string
}

func NewFoobar(f Fooer, b Barer) Foobarer {
	return &Foobar{
		foo: f,
		bar: b,
	}
}
func NewFoobarWithMsg(f Fooer, b Barer, msg string) Foobarer {
	//Println("NewFoobarWithMsg~~~~~~~")
	return &Foobar{
		foo: f,
		bar: b,
		msg: msg,
	}
}
func (f *Foobar) Say(i int, s string) {
	if f.foo != nil {
		f.foo.Foo(i)
		f.foo.Foo2(s)
	}
	if f.bar != nil {
		f.bar.Bar(s)
	}
	f.msg += fmt.Sprintf("[%d,%s]", i, s)
	Println("Foobar msg:", f.msg)
}
func TestContainer_SimpleRegister(t *testing.T) {
	log = ""
	defer Reset()
	Register(NewFoobar)
	Register(func() Fooer { return &Foo{} })
	Register(func() Barer { return &Bar{} })
	var fb Foobarer
	Resolve(&fb)
	fb.Say(123, "Hello World")
	t.Log(log)
	assert.True(t, strings.Contains(log, "foo:"))
	assert.True(t, strings.Contains(log, "bar:"))

}
func TestContainer_RegisterParameters(t *testing.T) {
	log = ""
	defer Reset()
	Register(NewFoobarWithMsg, Parameters(map[int]interface{}{2: "studyzy"}))
	Register(func() Fooer { return &Foo{} })
	Register(func() Barer { return &Bar{} })
	var fb Foobarer
	Resolve(&fb)
	fb.Say(123, "Hello World")
	t.Log(log)
	assert.True(t, strings.Contains(log, "foo:"))
	assert.True(t, strings.Contains(log, "bar:"))
	assert.True(t, strings.Contains(log, "studyzy"))
}

type Baz struct{}

func (Baz) Bar(s string) {
	Println("baz:", s)
}
func TestContainer_RegisterDefault(t *testing.T) {
	log = ""
	defer Reset()
	Register(NewFoobar)
	Register(func() Fooer { return &Foo{} })
	Register(func() Barer { return &Bar{} })
	Register(func() Barer { return &Baz{} }, Default())
	var fb Foobarer
	Resolve(&fb)
	fb.Say(123, "Hello World")
	t.Log(log)
	assert.True(t, strings.Contains(log, "foo:"))
	assert.True(t, strings.Contains(log, "baz:"))
}

func TestContainer_RegisterOptional(t *testing.T) {
	log = ""
	defer Reset()
	Register(NewFoobar, Optional(0))
	Register(func() Barer { return &Bar{} })
	var fb Foobarer
	Resolve(&fb)
	fb.Say(123, "Hello World")
	t.Log(log)
	assert.True(t, !strings.Contains(log, "foo:"))
	assert.True(t, strings.Contains(log, "bar:"))
}
func TestContainer_RegisterInterface(t *testing.T) {
	log = ""
	defer Reset()
	Register(NewFoobar)
	var f Fooer
	err := Register(NewFoo, Interface(&f))
	assert.Nil(t, err)
	var b Barer
	err = Register(NewBar, Interface(&b))
	assert.Nil(t, err)
	var fb Foobarer
	err = Resolve(&fb)
	assert.Nil(t, err)
	fb.Say(123, "Hello World")
	t.Log(log)
	assert.True(t, strings.Contains(log, "foo:"))
	assert.True(t, strings.Contains(log, "bar:"))
}
func TestContainer_RegisterLifestyleTransient(t *testing.T) {
	log = ""
	defer Reset()
	Register(NewFoobarWithMsg, Parameters(map[int]interface{}{2: "studyzy"}), Lifestyle(true))
	Register(func() Fooer { return &Foo{} })
	Register(func() Barer { return &Bar{} })
	var fb Foobarer
	Resolve(&fb)
	fb.Say(123, "Hello World")
	t.Log(log)
	assert.True(t, strings.Contains(log, "foo:"))
	assert.True(t, strings.Contains(log, "bar:"))
	assert.True(t, strings.Contains(log, "123"))
	log = ""
	var fb2 Foobarer
	Resolve(&fb2) //resolve a new instance since Lifestyle(transient=true)
	fb2.Say(456, "Hi")
	t.Log(log)
	assert.True(t, strings.Contains(log, "Hi"))
	assert.True(t, !strings.Contains(log, "123"))

}
func TestContainer_RegisterDependsOn(t *testing.T) {
	log = ""
	defer Reset()
	Register(NewFoobar, DependsOn(map[int]string{1: "bar"})) //depend on "bar" name Barer
	Register(func() Fooer { return &Foo{} })
	Register(func() Barer { return &Bar{} }, Name("bar"))
	Register(func() Barer { return &Baz{} }, Name("baz"))
	var fb Foobarer
	Resolve(&fb)
	fb.Say(123, "Hello World")
	t.Log(log)
	assert.True(t, strings.Contains(log, "foo:"))
	assert.True(t, strings.Contains(log, "bar:"))
}
func TestContainer_RegisterInstance(t *testing.T) {
	log = ""
	defer Reset()
	Register(NewFoobar)
	Register(func() Fooer { return &Foo{} })
	b := &Bar{}
	var bar Barer
	RegisterInstance(&bar, b, Default()) // register interface -> instance
	var fb Foobarer
	Resolve(&fb)
	fb.Say(123, "Hello World")
	t.Log(log)
	assert.True(t, strings.Contains(log, "foo:"))
	assert.True(t, strings.Contains(log, "bar:"))
}
func TestContainer_Resolve(t *testing.T) {
	log = ""
	defer Reset()
	Register(NewFoobarWithMsg, Parameters(map[int]interface{}{2: "studyzy"}))                  //default Foobar register
	Register(NewFoobarWithMsg, Parameters(map[int]interface{}{2: "Devin"}), Name("instance2")) //named Foobar register
	Register(func() Fooer { return &Foo{} })
	Register(func() Barer { return &Bar{} })
	var fb Foobarer
	Resolve(&fb, Arguments(map[int]interface{}{2: "arg2"})) //resolve use new argument to replace register parameters
	fb.Say(123, "Hello World")
	t.Log(log)
	assert.True(t, strings.Contains(log, "foo:"))
	assert.True(t, strings.Contains(log, "bar:"))
	assert.True(t, strings.Contains(log, "arg2"))
	assert.False(t, strings.Contains(log, "studyzy"))
	log = ""
	var fb2 Foobarer
	err := Resolve(&fb2, ResolveName("instance2")) //resolve by name
	assert.Nil(t, err)
	fb2.Say(456, "New World")
	t.Log(log)
	assert.True(t, strings.Contains(log, "Devin"))

}

func TestContainer_Call(t *testing.T) {
	log = ""
	defer Reset()
	Register(NewFoobar)
	Register(func() Fooer { return &Foo{} })
	Register(func() Barer { return &Bar{} })

	Call(SayHi1)
	t.Log(log)
	assert.True(t, strings.Contains(log, "foo:"))
	assert.True(t, strings.Contains(log, "bar:"))
	log = ""
	Call(SayHi2, CallArguments(map[int]interface{}{2: "Devin"}))
	assert.True(t, strings.Contains(log, "Devin"))
	log = ""
	Register(func() Barer { return &Baz{} }, Name("baz"))
	Call(SayHi1, CallDependsOn(map[int]string{1: "baz"}))
	assert.True(t, strings.Contains(log, "foo:"))
	assert.True(t, strings.Contains(log, "baz:"))

}
func SayHi1(f Fooer, b Barer) {
	f.Foo(1234)
	b.Bar("hi")
}

func SayHi2(f Fooer, b Barer, hi string) {
	f.Foo(len(hi))
	b.Bar(hi)
	Println("SayHi")
}

func TestContainer_Reset(t *testing.T) {
	log = ""
	Register(NewFoobar)
	Register(func() Fooer { return &Foo{} })
	Register(func() Barer { return &Bar{} })
	var fb Foobarer
	err := Resolve(&fb)
	assert.Nil(t, err)
	fb.Say(123, "Hello World")
	t.Log(log)
	assert.True(t, strings.Contains(log, "foo:"))
	assert.True(t, strings.Contains(log, "bar:"))
	Reset()
	err = Resolve(&fb)
	assert.NotNil(t, err)
}

type SubFoobar struct {
	foo SubFooer
	bar Barer
}

func NewSubFoobar(f SubFooer, b Barer) Foobarer {
	return &SubFoobar{
		foo: f,
		bar: b,
	}
}

func (f *SubFoobar) Say(i int, s string) {
	if f.foo != nil {
		f.foo.Foo2(s)
	}
	if f.bar != nil {
		f.bar.Bar(s)
	}
	Println("SubFoobar")
}

func TestContainer_RegisterSubInterface(t *testing.T) {
	log = ""
	defer Reset()
	Register(NewSubFoobar)
	Register(func() Fooer { return &Foo{} })
	Register(func() Barer { return &Bar{} })
	var fb Foobarer
	err := Resolve(&fb)
	assert.NotNil(t, err)
	var sub SubFooer
	var foo Fooer
	RegisterSubInterface(&sub, &foo)
	err = Resolve(&fb)
	assert.Nil(t, err)
	fb.Say(123, "Hello World")
	t.Log(log)
	assert.True(t, strings.Contains(log, "foo2:"))
	assert.True(t, strings.Contains(log, "bar:"))
	err = Resolve(&sub)
	assert.Nil(t, err)
	sub.Foo2("Hello")
}

type Bar2 struct {
}

func (Bar2) Bar(s string) {
	Println("bar2:", s)
}

func TestContainer_SetDefaultBinding(t *testing.T) {
	log = ""
	defer Reset()
	Register(NewFoobar)
	Register(func() Fooer { return &Foo{} }, Name("foo1"))
	Register(func() Barer { return &Bar{} }, Name("bar1"))
	Register(func() Barer { return &Bar2{} }, Name("bar2"))
	var b Barer
	SetDefaultBinding(&b, "bar2")
	var fb Foobarer
	Resolve(&fb)
	fb.Say(123, "Hello World")
	t.Log(log)
	assert.True(t, strings.Contains(log, "foo:"))
	assert.True(t, strings.Contains(log, "bar2:"))

}
func TestContainer_Clone(t *testing.T) {
	log = ""
	defer Reset()
	Register(NewFoobar)
	Register(func() Fooer { return &Foo{} }, Name("foo1"))
	Register(func() Barer { return &Bar{} }, Name("bar1"))
	Register(func() Barer { return &Bar2{} }, Name("bar2"))
	var b Barer
	newContainer := Clone()
	SetDefaultBinding(&b, "bar2")
	var fb Foobarer
	newContainer.Resolve(&fb)
	fb.Say(123, "Hello World")
	t.Log(log)
	assert.True(t, strings.Contains(log, "foo:"))
	assert.True(t, strings.Contains(log, "bar:"))

}
