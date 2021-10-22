# 1. iocgo简介
习惯于Java或者C#开发的人应该对控制反转与依赖注入应该再熟悉不过了。在Java平台有鼎鼎大名的Spring框架，在C#平台有Autofac,Unity,Windsor等，我当年C#开发时用的最多的就是Windsor。使用IoC容器是面向对象开发中非常方便的解耦模块之间的依赖的方法。各个模块之间不依赖于实现，而是依赖于接口，然后在构造函数或者属性或者方法中注入特定的实现，方便了各个模块的拆分以及模块的独立单元测试。
在[长安链]的设计中，各个模块可以灵活组装，模块之间的依赖基于protocol中定义的接口，每个接口有一个或者多个官方实现，当然第三方也可以提供该接口更多的实现。为了实现更灵活的组装各个模块，管理各个模块的依赖关系，于是我写了iocgo这个轻量级的golang版Ioc容器。

[English]((https://github.com/studyzy/iocgo/blob/main/README.md)) | 中文

# 2. iocgo如何使用
## 2.1 iocgo包的安装
现在go官方版本已经出到1.17了，当然我在代码中其实也没有用什么新版本的新特性，于是就用1.15版本或者之后的Go版本即可。要使用iocgo包，直接通过go get添加到项目中：

`go get github.com/studyzy/iocgo`

## 2.2 使用示例与说明
### 2.2.1 最简单的例子：
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
这里我使用NewContainer()创建了一个新的容器，然后在容器中调用Register方法注册了3个接口和对应的构造函数，分别是：
1. Foobarer接口对应NewFoobar(f Fooer,b Barer)构造函数
2. Fooer接口对应构造&Foo{}的匿名函数。
3. Barer接口对应构造&Bar{}的匿名函数。

接下来调用Resolve函数，并传入var fb Foobarer 这个接口变量的指针，iocgo就会自动去构建Foobarer对应的实例，并最终将实例赋值到fb这个变量上，于是最后我们就可以正常调用fb.Say实例方法了。

### 2.22. Register 的选项
iocgo的注册interface到对象的函数定义如下：

`func Register(constructor interface{}, options ...Option) error`

iocgo为Register函数提供了以下参数选项可根据实际情况选择性使用:
* Name 为某个interface->对象的映射命名
* Optional 表名这个构造函数中哪些注入的interface参数是可选的，如果是可选，那么就算找不到interface对应的实例也不会报错。
* Interface 显式声明这个构造函数返回的实例是映射到哪个interface。
* Lifestyle(isTransient) 声明这个构造函数在构造实例后是构造的临时实例还是单例实例，如果是临时实例，那么下次再获取该interface对应的实例时需要再次调用构造函数，如果是单例，那么就缓存实例到容器中，下次再想获得interface对应的实例时直接使用缓存中的，不需要再次构造。
* DependsOn 这个主要是指定构造函数中的某个参数在通过容器获得对应的实例时，应该通过哪个Name去获得对应的实例。
* Parameters 这个主要用于指定构造函数中的某些非容器托管的参数，比如某构造函数中有int，string等参数，而这些参数的实例是不需要通过ioc容器进行映射托管的，那么就在这里直接指定。
* Default 这个主要用于设置一个interface对应的默认的实例，也就是如果没有指定Name的情况下，应该找哪个实例。
  关于每一个参数该如何使用，我都写了UT样例，具体参考：
  [container_test.go](https://github.com/studyzy/iocgo/blob/main/container_test.go)

### 2.2.3. 注册实例
如果我们已经有了某个对象的实例，那么可以将该实例和其想映射的interface直接注册到ioc容器中，方便其他依赖的对象获取，RegisterInstance函数定义如下:

`RegisterInstance(interfacePtr interface{}, instance interface{}, options ...Option) error `
使用上也很简单，直接将实例对应的interface的指针作为参数1，实例本身作为参数2，传入RegisterInstance即可：
```go
b := &Bar{}
var bar Barer //interface
container.RegisterInstance(&bar, b) // register interface -> instance
```

### 2.2.4. 获得实例
相关映射我们通过Register函数和RegisterInstance函数已经注册到容器中，接下来就需要从容器获得指定的实例了。获得实例需要调用函数：

`func Resolve(abstraction interface{}, options ...ResolveOption) error`
这里第一个参数abstraction是我们想要获取的某个interface的指针，第二个参数是可选参数，目前提供的选项有：
* ResolveName 指定使用哪个name的interface和实例的映射，如果不指定，那么就是默认映射。
* Arguments 指定在调用对应的构造函数获得实例时，传递的参数，比如int，string等类型的不在ioc容器中托管的参数，可以在这里指定。如果构造函数本身需要这些参数，而且在前面Register的时候已经通过Parameters选项进行了指定，那么这里新的指定会覆盖原有Register的指定。
```go
var fb Foobarer
err:=container.Resolve(&fb)
```
另外如果我们的构造函数return的值中支持error，而且实际构造的时候确实返回了error，那么Resolve函数也会返回对应的这个err。
**特别注意：Resolve的第一个参数是申明的某个interface的指针，一定要是指针，不能直接传interface**

### 2.2.5. 结构体参数和字段填充
有些时候构造函数的入参非常多，于是我们可以申明一个结构体，把所有入参都放入这个结构体中，这样构造函数就只需要一个参数了。iocgo也支持自动填充这个结构体中interface对应的实例，从而构造新的对象。另外iocgo也提供了Fill方法，可以直接填充某个结构体，比如：
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
结构体中的字段还支持tag，目前提供的tag有两种：

* name //指定这个字段在获得对应的实例时使用的name
* optional //指定这个字段是否是可选的，如果是，那么就算获得不到对应的实例，也不会报错。
  示例example:
```
type FoobarInputWithTag struct {
	foo Fooer `optional:"true"`
	bar Barer `name:"baz"`
	msg string
}
```

### 2.2.6. 函数调用
除了构造函数注入之外，iocgo也支持函数注入，我们申明一个函数，这个函数的参数中有些参数是interface，那么通过调用iocgo中的Call方法，可以为这个函数注入对应的实例作为参数，并最终完成函数的调用。
示例 example:
```
func SayHi1(f Fooer, b Barer) {
	f.Foo(1234)
	b.Bar("hi")
}
Register(func() Fooer { return &Foo{} })
Register(func() Barer { return &Bar{} })
Call(SayHi1)
```
Call函数也是支持选项的，目前提供了2个选项:
* CallArguments 指定函数中某个参数的值
* CallDependsOn 指定函数中某个参数在通过ioc容器获得实例时使用哪个name来获得实例。
  最后函数调用完成，如果函数本身有多个返回值，有error返回，那么Call函数也会返回对应的结果。


## 2.3 参考:
在写这个iocgo的代码时，主要参考了以下两个Ioc相关的项目：
* https://github.com/golobby/container
* https://github.com/castleproject/Windsor

# 3. 总结
iocgo是一个纯Golang语言开发的用于管理依赖注入的IoC容器，使用这个容器可以很好的实现go语言下的面向对象开发，模块解耦。现已经开源，欢迎大家使用，开源地址：https://github.com/studyzy/iocgo