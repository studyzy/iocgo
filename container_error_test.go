package iocgo

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainer_DependenceNotFound(t *testing.T) {
	container := NewContainer()
	container.Register(NewFoobar)
	//container.Register(func() Fooer { return &Foo{} })
	container.Register(func() Barer { return &Bar{} })
	var fb Foobarer
	err := container.Resolve(&fb)
	assert.NotNil(t, err)
	t.Log(err)
}
func TestContainer_RegisterNotAFunctionError(t *testing.T) {
	container := NewContainer()
	err := container.Register(&Foo{})
	assert.NotNil(t, err)
	t.Log(err)
}
func TestContainer_RegisterInterfaceError(t *testing.T) {
	container := NewContainer()
	container.Register(NewFoobar)
	var f Fooer
	err := container.Register(NewFoo, Interface(f))
	assert.NotNil(t, err)
	t.Log(err)
	//var b Barer
	err = container.Register(NewBar, Interface(&f))
	assert.NotNil(t, err)
	t.Log(err)
	var fb Foobarer
	err = container.Resolve(&fb)
	assert.NotNil(t, err)
	t.Log(err)
}
func TestContainer_RegisterInstanceError(t *testing.T) {
	container := NewContainer()
	container.Register(NewFoobar)
	container.Register(func() Fooer { return &Foo{} })
	b := NewBar()
	var bar Barer
	err := container.RegisterInstance(bar, b)
	assert.NotNil(t, err)
	t.Log(err)
}

type Foobarer2 interface {
	Say2(int, string)
}

func TestContainer_ResolveError(t *testing.T) {
	container := NewContainer()
	container.Register(NewFoobarWithMsg, Parameters(map[int]interface{}{2: "studyzy"}))
	container.Register(func() Fooer { return &Foo{} })
	container.Register(func() Barer { return &Bar{} })
	var fb Foobarer
	err := container.Resolve(fb, Arguments(map[int]interface{}{2: "arg2"})) //resolve use new argument to replace register parameters
	assert.NotNil(t, err)
	t.Log(err)
	var fb2 Foobarer2
	err = container.Resolve(&fb2)
	assert.NotNil(t, err)
	t.Log(err)

}
func NewFoobarError(f Fooer, b Barer) (Foobarer, error) {
	if f == nil || b == nil {
		return nil, errors.New("input nil")
	}
	return &Foobar{
		foo: f,
		bar: b,
	}, nil
}
func TestContainer_ResolveReturnError(t *testing.T) {
	container := NewContainer()
	container.Register(NewFoobarError, Optional(0))

	container.Register(func() Barer { return &Bar{} })
	var fb Foobarer
	err := container.Resolve(&fb)
	assert.NotNil(t, err)
	assert.Equal(t, "input nil", err.Error())
	t.Log(err)
	container.Register(func() Fooer { return &Foo{} })
	err = container.Resolve(&fb)
	assert.Nil(t, err)
}
