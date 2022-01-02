package iocgo

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"unsafe"
)

var (
	errNotFound = errors.New("not found")
)

type binding struct {
	specifiedParameters map[int]interface{} //构造对象时参数指定的值
	dependsOn           map[int]string      //构造对象时依赖的其他对象的name
	constructor         interface{}         //构造函数指针，用于构造对应的实例
	instance            interface{}         //在默认单例情况下，存储对应的绑定的实例
	isTransient         bool                //是否是临时对象
	isDefault           bool                //是否是默认对象
	name                string              //对应的名字
	resolveTypes        []reflect.Type      //指定构造函数返回的参数列表对应的接口类型，如果不指定某个返回值，可以设置为nil
	optionalIndexes     map[int]bool        //哪些参数是可选的，如果可选，那么即使无法找到对应实例也不会报错
}

func (b *binding) Clone() *binding {
	clone := &binding{
		specifiedParameters: make(map[int]interface{}, len(b.specifiedParameters)),
		dependsOn:           make(map[int]string, len(b.dependsOn)),
		constructor:         b.constructor,
		instance:            b.instance,
		isTransient:         b.isTransient,
		isDefault:           b.isDefault,
		name:                b.name,
		resolveTypes:        b.resolveTypes,
		optionalIndexes:     make(map[int]bool, len(b.optionalIndexes)),
	}
	for k, v := range b.specifiedParameters {
		clone.specifiedParameters[k] = v
	}
	for k, v := range b.dependsOn {
		clone.dependsOn[k] = v
	}
	for k, v := range b.optionalIndexes {
		clone.optionalIndexes[k] = v
	}
	return clone
}

// resolve creates an appropriate implementation of the related abstraction
func (b *binding) resolve(c *Container) (interface{}, error) {
	if b.instance != nil {
		return b.instance, nil
	}

	instList, err := c.invoke(b.constructor, b.specifiedParameters, b.dependsOn, b.optionalIndexes)
	if err != nil {
		return nil, err
	}
	if len(instList) == 0 {
		return nil, errors.New("resolve function must return instance")
	}
	inst := instList[0]
	if !b.isTransient {
		b.instance = inst
	}
	return inst, err
}

type namedBinding struct {
	defaultBinding *binding
	namedBinding   map[string]*binding
}

func (b *namedBinding) Clone() *namedBinding {
	clone := &namedBinding{
		defaultBinding: b.defaultBinding.Clone(),
		namedBinding:   make(map[string]*binding, len(b.namedBinding)),
	}
	for k, v := range b.namedBinding {
		clone.namedBinding[k] = v.Clone()
	}
	return clone
}
func newNamedBinding(b *binding) *namedBinding {
	bindings := make(map[string]*binding)
	bindings[b.name] = b
	return &namedBinding{defaultBinding: b, namedBinding: bindings}
}
func (nb *namedBinding) addNewBinding(b *binding, isDefault bool) {
	if isDefault {
		nb.defaultBinding = b
	}
	nb.namedBinding[b.name] = b
}

// Container interface类型->map["name"]binding对象，如果没有命名实例，那么name就是""
type Container struct {
	bind  map[reflect.Type]*namedBinding
	alias map[reflect.Type]reflect.Type
}

// NewContainer creates a new instance of the Container
func NewContainer() *Container {
	return &Container{
		bind:  make(map[reflect.Type]*namedBinding),
		alias: make(map[reflect.Type]reflect.Type),
	}
}

//Register 注册一个对象的构造函数到容器中，该构造函数接收其他interface对象或者值对象作为参数，返回interface对象
//注意返回的应该是interface，而不应该是具体的struct类型的指针
func (c *Container) Register(constructor interface{}, options ...Option) error {
	//检查resolver必须是一个构造函数
	reflectedResolver := reflect.TypeOf(constructor)
	if reflectedResolver.Kind() != reflect.Func {
		return errors.New("container: the constructor must be a function")
	}
	//遍历构造函数的输出，找到具体构造的类型，并将这些类型放入到container中
	for i := 0; i < reflectedResolver.NumOut(); i++ {
		//构造新的binding对象
		b := &binding{constructor: constructor, specifiedParameters: make(map[int]interface{})}
		for _, op := range options {
			err := op(b)
			if err != nil {
				return err
			}
		}
		resolveType := reflectedResolver.Out(i)
		if len(b.resolveTypes) > i && b.resolveTypes[i] != nil { //如果指定了映射的interface，则使用指定的
			if !resolveType.Implements(b.resolveTypes[i]) {
				return errors.New("resolve type " + resolveType.String() + " not implement " + b.resolveTypes[i].String())
			}
			resolveType = b.resolveTypes[i]
		}
		if namedBinding, has := c.bind[resolveType]; has { //增加新binding
			namedBinding.addNewBinding(b, b.isDefault)
		} else { //没有注册过这个接口的任何绑定
			c.bind[resolveType] = newNamedBinding(b)
		}
	}

	return nil
}

//RegisterInstance 注册一个对象的实例到容器中
//参数interfacePtr 是一个接口的指针
//参数instance 是实例值
func (c *Container) RegisterInstance(interfacePtr interface{}, instance interface{}, options ...Option) error {
	b := &binding{instance: instance}
	for _, op := range options {
		err := op(b)
		if err != nil {
			return err
		}
	}
	t, err := getTypeFromInterface(interfacePtr)
	if err != nil {
		return err
	}
	if namedBinding, has := c.bind[t]; has { //增加新的绑定
		namedBinding.addNewBinding(b, b.isDefault)
	} else { //没有注册过这个接口的任何绑定
		c.bind[t] = newNamedBinding(b)
	}
	return nil
}

//RegisterSubInterface 注册一个子接口到某个接口
func (c *Container) RegisterSubInterface(subInterfacePtr interface{}, interfacePtr interface{}) error {
	stype, err := getTypeFromInterface(subInterfacePtr)
	if err != nil {
		return err
	}
	itype, err := getTypeFromInterface(interfacePtr)
	if err != nil {
		return err
	}
	c.alias[stype] = itype
	return nil
}

func (c *Container) SetDefaultBinding(interfacePtr interface{}, defaultName string) error {
	itype, err := getTypeFromInterface(interfacePtr)
	if err != nil {
		return err
	}
	if nameBinding, ok := c.bind[itype]; ok {
		if theBinding, found := nameBinding.namedBinding[defaultName]; found {
			nameBinding.defaultBinding = theBinding
			return nil
		}
	}
	return errNotFound
}
func getTypeFromInterface(interfacePtr interface{}) (reflect.Type, error) {
	ptr := reflect.TypeOf(interfacePtr)
	if ptr == nil || ptr.Kind() != reflect.Ptr {
		return nil, errors.New("interfacePtr must be a interface point, not a interface value")
	}
	return ptr.Elem(), nil
}

// arguments 通过容器获得一个函数的传入参数的值列表
func (c *Container) arguments(function interface{}, specifiedParameters map[int]interface{},
	dependsOn map[int]string, optionalIndexes map[int]bool) ([]reflect.Value, error) {
	reflectedFunction := reflect.TypeOf(function)
	argumentsCount := reflectedFunction.NumIn()
	arguments := make([]reflect.Value, argumentsCount)

	for i := 0; i < argumentsCount; i++ {
		abstraction := reflectedFunction.In(i)
		if specifiedValue, has := specifiedParameters[i]; has { //如果是指定了参数的，直接获得参数值
			if isNil(specifiedValue) { //如果在指定参数中设置了nil，那么表示强制将该值设为空,
				arguments[i] = reflect.Zero(abstraction)
				continue
			}
			//如果参数是struct类型，需要调用Fill填充这个struct中的字段
			fieldKind := reflect.TypeOf(specifiedValue).Kind()
			if fieldKind == reflect.Struct {
				err := c.Fill(&specifiedValue)
				if err != nil {
					return nil, err
				}
			}
			if fieldKind == reflect.Ptr { //如果是指针，那么获得对应的值类型，判断是否struct，是则Fill
				elem := reflect.TypeOf(specifiedValue).Elem()
				if elem.Kind() == reflect.Struct {
					err := c.Fill(specifiedValue)
					if err != nil {
						return nil, err
					}
				}
			}
			arguments[i] = reflect.ValueOf(specifiedValue)
			continue
		}
		name := dependsOn[i]
		b, err := c.getBinding(abstraction, name)
		if err != nil {
			//找不到该函数对应的参数类型的映射，如果是optional的，则设为空，否则报错
			if _, optional := optionalIndexes[i]; optional {
				arguments[i] = reflect.Zero(abstraction)
				continue
			}
			//必填字段找不到，报错
			resolveType := ""
			o := reflectedFunction.Out(0)
			if o != nil {
				resolveType = o.String()
			}
			return nil, errors.New("resolve type: " + resolveType + " no concrete found for: " + abstraction.String())
		}

		instance, err := b.resolve(c)
		if err != nil {
			return nil, err
		}
		arguments[i] = reflect.ValueOf(instance)
	}
	return arguments, nil
}

func (c *Container) getBinding(theType reflect.Type, name string) (*binding, error) {
	if namedBinding, exist := c.bind[theType]; exist {
		//从容器中找到了对应的binding
		//如果使用DependsOn指定了依赖的对象的name，那么通过指定的name获取binding
		if name != "" {
			if b, ok := namedBinding.namedBinding[name]; ok {
				return b, nil
			}
			return nil, errors.New("container: no concrete found for: " + theType.String() + " name: " + name)

		}
		//没有通过name指定，那么就取默认的binding
		return namedBinding.defaultBinding, nil
	}
	//找不到该函数对应的参数类型的映射，在alias中找
	if aType, ok := c.alias[theType]; ok {
		return c.getBinding(aType, name)
	}
	return nil, errNotFound
}

func (c *Container) invoke(function interface{}, specifiedParameters map[int]interface{},
	dependsOn map[int]string, optionalIndexes map[int]bool) (
	[]interface{}, error) {
	args, err := c.arguments(function, specifiedParameters, dependsOn, optionalIndexes)
	if err != nil {
		return nil, err
	}
	returns := reflect.ValueOf(function).Call(args)
	if len(returns) == 0 {
		return nil, nil
	}
	returnList := []interface{}{}
	var errPtr *error
	errType := reflect.TypeOf(errPtr).Elem()
	for _, rt := range returns {
		if rt.Type().Implements(errType) && !rt.IsNil() { //返回类型中有不为空的error
			return nil, rt.Interface().(error)
		}
		returnList = append(returnList, rt.Interface())
	}
	return returnList, nil
}

//Resolve input interface, resolve instance. 传入接口的指针，获得对应的实例
func (c *Container) Resolve(abstraction interface{}, options ...ResolveOption) error {
	receiverType := reflect.TypeOf(abstraction)
	if receiverType == nil {
		return errors.New("container: invalid abstraction")
	}
	option := &resolveOption{}
	for _, op := range options {
		err := op(option)
		if err != nil {
			return err
		}
	}
	if receiverType.Kind() == reflect.Ptr {
		elem := receiverType.Elem()
		b, err := c.getBinding(elem, option.name)
		if err != nil {
			return errors.New("resolve type: " + receiverType.String() + " no concrete found for: " + elem.String())
		}

		args := b.specifiedParameters
		if len(option.args) > 0 {
			for i, v := range option.args {
				args[i] = v
			}
		}
		oldArgs := b.specifiedParameters
		b.specifiedParameters = args
		defer func() {
			b.specifiedParameters = oldArgs
		}()
		instance, err := b.resolve(c)
		if err != nil {
			return err //errors.New("resolve type: " + receiverType.String() + " " + err.Error())
		}
		reflect.ValueOf(abstraction).Elem().Set(reflect.ValueOf(instance))
		return nil
	}
	return errors.New("container: invalid abstraction")
}

// Call takes a function (receiver) with one or more arguments of the abstractions (interfaces).
// It invokes the function (receiver) and passes the related implementations.
func (c *Container) Call(function interface{}, options ...CallOption) ([]interface{}, error) {
	receiverType := reflect.TypeOf(function)
	if receiverType == nil || receiverType.Kind() != reflect.Func {
		return nil, errors.New("container: invalid function")
	}
	callOption := &resolveOption{}
	for _, op := range options {
		err := op(callOption)
		if err != nil {
			return nil, err
		}
	}
	return c.invoke(function, callOption.args, callOption.dependsOn, nil) //TODO optional

}

// Fill takes a struct and resolves the fields with the tag `optional:"true"` or `name:"dependOnName1"`
func (c *Container) Fill(structure interface{}) error {
	// 获取入参类型
	receiverType := reflect.TypeOf(structure)
	if receiverType == nil {
		return errors.New("container: invalid structure")
	}

	if receiverType.Kind() == reflect.Ptr {
		elem := receiverType.Elem()
		if elem.Kind() == reflect.Struct {
			s := reflect.ValueOf(structure).Elem()
			for i := 0; i < s.NumField(); i++ {
				// 获取第i个字段
				f := s.Field(i)
				fType := f.Type()
				//如果是interface的数组，那么就填充所有实现
				sliceFill := false
				if f.Kind() == reflect.Slice && f.Type().Elem().Kind() == reflect.Interface {
					sliceFill = true
					fType = f.Type().Elem()
				} else if f.Kind() != reflect.Interface { //只对interface类型执行Fill逻辑
					continue
				}
				//是否是可选的
				optional := false
				if b, exist := s.Type().Field(i).Tag.Lookup("optional"); exist {
					optional = strings.ToLower(b) == "true"
				}

				namedBinding, ok := c.bind[fType]
				if !ok {
					if optional {
						continue
					}
					return errors.New("container: no concrete found for: " + f.Type().String())
				}
				if sliceFill {
					for _, b := range namedBinding.namedBinding {
						instance, _ := b.resolve(c)
						ptr := reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem()
						ptr.Set(reflect.Append(ptr, reflect.ValueOf(instance)))
					}
					continue
				}
				b := namedBinding.defaultBinding
				//指定了name字段说明该字段依赖的binding name
				if name, exist := s.Type().Field(i).Tag.Lookup("name"); exist {
					if concrete, exist := namedBinding.namedBinding[name]; exist {
						b = concrete
					} else {
						return fmt.Errorf("container: cannot resolve %v field", s.Type().Field(i).Name)
					}
				}
				//没有指定name，获得默认binding
				instance, _ := b.resolve(c)
				ptr := reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem()
				ptr.Set(reflect.ValueOf(instance))
			}
			return nil
		}
		return errors.New("container: invalid structure, input elem type:" + elem.Kind().String())
	}

	return errors.New("container: invalid structure, input type:" + receiverType.Kind().String())
}

// Reset deletes all the existing bindings and empties the container instance.
func (c *Container) Reset() {
	for k := range c.bind {
		delete(c.bind, k)
	}
	for k := range c.alias {
		delete(c.alias, k)
	}
}
func (c *Container) Clone() *Container {
	clone := &Container{
		bind:  make(map[reflect.Type]*namedBinding, len(c.bind)),
		alias: make(map[reflect.Type]reflect.Type, len(c.alias)),
	}
	for k, v := range c.bind {
		clone.bind[k] = v.Clone()
	}
	for k, v := range c.alias {
		clone.alias[k] = v
	}
	return clone
}

// container is the global repository of bindings
var container = NewContainer()

// Resolve takes an abstraction (interface reference) and fills it with the related implementation.
func Resolve(abstraction interface{}, options ...ResolveOption) error {
	return container.Resolve(abstraction, options...)
}

//Register register interface-> constructor
func Register(constructor interface{}, options ...Option) error {
	return container.Register(constructor, options...)
}

// Reset deletes all the existing bindings and empties the container instance.
func Reset() {
	container.Reset()
}

//Fill takes a struct and resolves the fields with the tag `optional:"true"` or `name:"dependOnName1"`
//argument must be a struct point
func Fill(structure interface{}) error {
	return container.Fill(structure)
}

//Call invoke function that use interface as parameters
func Call(function interface{}, options ...CallOption) ([]interface{}, error) {
	return container.Call(function, options...)
}

//RegisterInstance register interface->instance into container
func RegisterInstance(interfacePtr interface{}, instance interface{}, options ...Option) error {
	return container.RegisterInstance(interfacePtr, instance, options...)
}

//RegisterSubInterface set sub interface map to another interface
func RegisterSubInterface(subInterfacePtr interface{}, interfacePtr interface{}) error {
	return container.RegisterSubInterface(subInterfacePtr, interfacePtr)
}

//SetDefaultBinding change default binding to another name
func SetDefaultBinding(interfacePtr interface{}, defaultName string) error {
	return container.SetDefaultBinding(interfacePtr, defaultName)
}
func isNil(i interface{}) bool {
	vi := reflect.ValueOf(i)
	if vi.Kind() == reflect.Ptr {
		return vi.IsNil()
	}
	return false
}
func Clone() *Container {
	return container.Clone()
}
