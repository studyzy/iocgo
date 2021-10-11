package iocgo

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type FoobarInput struct {
	foo Fooer
	bar Barer
	msg string
}

func NewFoobarWithInput(input *FoobarInput) Foobarer {
	return &Foobar{
		foo: input.foo,
		bar: input.bar,
		msg: input.msg,
	}
}
func TestContainer_RegisterStructInitArg(t *testing.T) {
	log = ""
	container := NewContainer()
	input := &FoobarInput{
		msg: "studyzy",
	}
	container.Register(NewFoobarWithInput, Parameters(map[int]interface{}{0: input}))
	container.Register(func() Fooer { return &Foo{} })
	container.Register(func() Barer { return &Bar{} })
	var fb Foobarer
	err := container.Resolve(&fb)
	assert.Nil(t, err)
	fb.Say(123, "Hello World")
	t.Log(log)
	assert.True(t, strings.Contains(log, "foo:"))
	assert.True(t, strings.Contains(log, "bar:"))
	assert.True(t, strings.Contains(log, "studyzy"))
}

func TestContainer_Fill(t *testing.T) {
	log = ""
	container := NewContainer()
	input := FoobarInput{
		msg: "studyzy",
	}
	container.Register(func() Fooer { return &Foo{} })
	container.Register(func() Barer { return &Bar{} })
	err := container.Fill(&input)
	assert.Nil(t, err)
	t.Logf("%#v", input)
	assert.NotNil(t, input.foo)
	assert.NotNil(t, input.bar)
}
func TestContainer_ResolveStructInitArg(t *testing.T) {
	log = ""
	container := NewContainer()
	container.Register(NewFoobarWithInput)
	container.Register(func() Fooer { return &Foo{} })
	container.Register(func() Barer { return &Bar{} })
	var fb Foobarer
	input := &FoobarInput{
		msg: "studyzy",
	}
	err := container.Resolve(&fb, Arguments(map[int]interface{}{0: input}))
	assert.Nil(t, err)
	fb.Say(123, "Hello World")
	t.Log(log)
	assert.True(t, strings.Contains(log, "foo:"))
	assert.True(t, strings.Contains(log, "bar:"))
	assert.True(t, strings.Contains(log, "studyzy"))
}

type FoobarInputWithTag struct {
	foo Fooer `optional:"true"`
	bar Barer `name:"baz"`
	msg string
}

func NewFoobarWithInputTag(input *FoobarInputWithTag) Foobarer {
	return &Foobar{
		foo: input.foo,
		bar: input.bar,
		msg: input.msg,
	}
}
func TestContainer_ResolveStructInitArgOptional(t *testing.T) {
	log = ""
	container := NewContainer()
	input := &FoobarInputWithTag{
		msg: "studyzy",
	}
	container.Register(NewFoobarWithInputTag, Parameters(map[int]interface{}{0: input}))
	container.Register(func() Barer { return &Bar{} }, Name("bar"))
	container.Register(func() Barer { return &Baz{} }, Name("baz"))
	var fb Foobarer
	err := container.Resolve(&fb)
	assert.Nil(t, err)
	fb.Say(123, "Hello World")
	t.Log(log)
	assert.False(t, strings.Contains(log, "foo:"))
	assert.True(t, strings.Contains(log, "baz:"))
	assert.True(t, strings.Contains(log, "studyzy"))
}

type FoobarInputMultiBar struct {
	foo Fooer
	bar []Barer
	msg string
}

func TestContainer_FillSlice(t *testing.T) {
	log = ""
	container := NewContainer()
	input := FoobarInputMultiBar{
		msg: "studyzy",
	}
	container.Register(func() Fooer { return &Foo{} })
	container.Register(func() Barer { return &Bar{} }, Name("bar"))
	container.Register(func() Barer { return &Baz{} }, Name("baz"))
	err := container.Fill(&input)
	assert.Nil(t, err)
	t.Logf("%#v", input)
	assert.NotNil(t, input.foo)
	assert.NotNil(t, input.bar)
	for _, bar := range input.bar {
		bar.Bar("Hi")
	}
	assert.True(t, strings.Contains(log, "bar:"))
	assert.True(t, strings.Contains(log, "baz:"))
}
