package iocgo

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainer_DependenceNotFound(t *testing.T) {
	defer Reset()
	err := Register(NewFoobar)
	assert.Nil(t, err)
	//Register(func() Fooer { return &Foo{} })
	err = Register(func() Barer { return &Bar{} })
	assert.Nil(t, err)
	var fb Foobarer
	err = Resolve(&fb)
	assert.NotNil(t, err)
	t.Log(err)
}
func TestContainer_RegisterNotAFunctionError(t *testing.T) {
	defer Reset()
	err := Register(&Foo{})
	assert.NotNil(t, err)
	t.Log(err)
}
func TestContainer_RegisterInterfaceError(t *testing.T) {
	defer Reset()
	err := Register(NewFoobar)
	assert.Nil(t, err)
	var f Fooer
	err = Register(NewFoo, Interface(f))
	assert.NotNil(t, err)
	t.Log(err)
	//var b Barer
	err = Register(NewBar, Interface(&f))
	assert.NotNil(t, err)
	t.Log(err)
	var fb Foobarer
	err = Resolve(&fb)
	assert.NotNil(t, err)
	t.Log(err)
}
func TestContainer_RegisterInstanceError(t *testing.T) {
	defer Reset()
	Register(NewFoobar)
	Register(func() Fooer { return &Foo{} })
	b := NewBar()
	var bar Barer
	err := RegisterInstance(bar, b)
	assert.NotNil(t, err)
	t.Log(err)
}

type Foobarer2 interface {
	Say2(int, string)
}

func TestContainer_ResolveError(t *testing.T) {
	defer Reset()
	Register(NewFoobarWithMsg, Parameters(map[int]interface{}{2: "studyzy"}))
	Register(func() Fooer { return &Foo{} })
	Register(func() Barer { return &Bar{} })
	var fb Foobarer
	err := Resolve(fb, Arguments(map[int]interface{}{2: "arg2"})) //resolve use new argument to replace register parameters
	assert.NotNil(t, err)
	t.Log(err)
	var fb2 Foobarer2
	err = Resolve(&fb2)
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
	defer Reset()
	Register(NewFoobarError, Optional(0))

	Register(func() Barer { return &Bar{} })
	var fb Foobarer
	err := Resolve(&fb)
	assert.NotNil(t, err)
	assert.Equal(t, "input nil", err.Error())
	t.Log(err)
	Register(func() Fooer { return &Foo{} })
	err = Resolve(&fb)
	assert.Nil(t, err)
}
