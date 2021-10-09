

package iocgo

import (
	"errors"
	"reflect"
)

type Option func(*binding) error

//Name 指定binding的name
func Name(name string) Option {
	return func(b *binding) error {
		b.name = name
		return nil
	}
}

//Optional 指定构造函数中哪些参数是可选的，即使没有Resolve出来，设置为nil即可，也不报错
func Optional(index ...int) Option {
	return func(b *binding) error {
		b.optionalIndexes = make(map[int]bool)
		for _, i := range index {
			b.optionalIndexes[i] = true
		}
		return nil
	}
}

//Interface 指定注册时构造函数返回对象对应的接口指针
//比如将Logger这个interface传入，代码为：
// var l *Logger
// Interface(l)
func Interface(it ...interface{}) Option {
	return func(b *binding) error {
		for _, i := range it {
			if i == nil {
				b.resolveTypes = append(b.resolveTypes, nil)
			}
			ptr := reflect.TypeOf(i)
			if ptr == nil || ptr.Kind() != reflect.Ptr {
				return errors.New("interface input must be a interface point")
			}
			t := ptr.Elem()
			b.resolveTypes = append(b.resolveTypes, t)
		}
		return nil
	}
}

//Lifestyle 指定在获得接口对应的实例时，是单例还是临时的
func Lifestyle(isTransient bool) Option {
	return func(b *binding) error {
		b.isTransient = isTransient
		return nil
	}
}

//DependsOn 指定这个构造函数依赖的接口对应的name
func DependsOn(dependsOn map[int]string) Option {
	return func(b *binding) error {
		b.dependsOn = dependsOn
		return nil
	}
}

//Parameters 指定这个构造函数的参数值
func Parameters(p map[int]interface{}) Option {
	return func(b *binding) error {
		b.specifiedParameters = p
		return nil
	}
}

//Default 指定这个接口对应构造函数是不是默认的映射
func Default() Option {
	return func(b *binding) error {
		b.isDefault = true
		return nil
	}
}

type ResolveOption func(*resolveOption) error
type resolveOption struct {
	name      string
	args      map[int]interface{}
	dependsOn map[int]string
}

//Arguments 指定在获得某接口的实例时，该实例构造函数的值
func Arguments(p map[int]interface{}) ResolveOption {
	return func(option *resolveOption) error {
		option.args = p
		return nil
	}
}

//ResolveName 指定在获得接口实例时，使用哪个name对应的构造函数
func ResolveName(name string) ResolveOption {
	return func(option *resolveOption) error {
		option.name = name
		return nil
	}
}

type CallOption func(*resolveOption) error

func CallArguments(p map[int]interface{}) CallOption {
	return func(option *resolveOption) error {
		option.args = p
		return nil
	}
}

//CallDependsOn 指定这个构造函数依赖的接口对应的name
func CallDependsOn(dependsOn map[int]string) CallOption {
	return func(option *resolveOption) error {
		option.dependsOn = dependsOn
		return nil
	}
}
