

package iocgo

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Logger interface {
	Debug(msg string)
}
type GoLogger struct {
	name string
}

func (l GoLogger) Debug(msg string) {
	fmt.Println(l.name, msg)
}

type Store interface {
	Save(height uint64)
}
type LevelDBStore struct {
	heightList []uint64
}

func (s *LevelDBStore) Save(h uint64) {
	s.heightList = append(s.heightList, h)
	fmt.Println("LevelDB save height", s.heightList)
}

type MySQLStore struct {
}

func (s *MySQLStore) Save(h uint64) {
	fmt.Println("MySQL save height", h)
}

type BlockStore interface {
	SaveBlock(h uint64)
}
type CMBlockStore struct {
	store   Store
	l       Logger
	chainId string
}

func (cmstore *CMBlockStore) SaveBlock(h uint64) {
	if cmstore.l != nil {
		cmstore.l.Debug(cmstore.chainId + " start ...")
	} else {
		fmt.Println(cmstore.chainId, "no logger")
	}
	cmstore.store.Save(h)
	if cmstore.l != nil {
		cmstore.l.Debug(cmstore.chainId + " end.")
	}
}
func NewBlockStore(store Store, l Logger, chainId string) BlockStore {
	fmt.Println("New BlockStore")
	return &CMBlockStore{
		store:   store,
		l:       l,
		chainId: chainId,
	}
}
func NewBlockStoreImpl(store Store, l Logger, chainId string) *CMBlockStore {
	fmt.Println("New BlockStore implement")
	return &CMBlockStore{
		store:   store,
		l:       l,
		chainId: chainId,
	}
}

type InitBlockStoreArg struct {
	store   Store  `name:"leveldb"`
	l       Logger `optional:"true"`
	chainId string
}

func InitBlockStore(input *InitBlockStoreArg) BlockStore {
	fmt.Println("Init BlockStore")
	return &CMBlockStore{
		store:   input.store,
		l:       input.l,
		chainId: input.chainId,
	}
}
func TestContainer_RegisterParameters(t *testing.T) {
	container := NewContainer()
	container.Register(NewBlockStore, Parameters(map[int]interface{}{2: "chain1"}))
	container.Register(func() Logger { return GoLogger{} })
	container.Register(func() Store { return &LevelDBStore{} })
	var cmStore BlockStore
	err := container.Resolve(&cmStore)
	assert.Nil(t, err)
	cmStore.SaveBlock(123)
	var store Store
	container.Resolve(&store)
	store.Save(456)
}
func TestContainer_RegisterDefault(t *testing.T) {
	container := NewContainer()
	container.Register(NewBlockStore, Parameters(map[int]interface{}{2: "chain1"}))
	container.Register(func() Logger { return GoLogger{} })
	container.Register(func() Store { return &LevelDBStore{} }, Name("leveldb"))
	container.Register(func() Store { return &MySQLStore{} }, Name("mysql"), Default())
	var cmStore BlockStore
	err := container.Resolve(&cmStore)
	assert.Nil(t, err)
	cmStore.SaveBlock(123)
}

func TestContainer_RegisterOptional(t *testing.T) {
	container := NewContainer()
	container.Register(NewBlockStore, Parameters(map[int]interface{}{2: "chain1"}), Optional(1))
	container.Register(func() Store { return &LevelDBStore{} }, Name("leveldb"))
	container.Register(func() Store { return &MySQLStore{} }, Name("mysql"), Default())
	var cmStore BlockStore
	err := container.Resolve(&cmStore)
	assert.Nil(t, err)
	cmStore.SaveBlock(123)
}
func TestContainer_RegisterInterface(t *testing.T) {
	container := NewContainer()
	var s *BlockStore
	container.Register(NewBlockStoreImpl, Parameters(map[int]interface{}{2: "chain1"}), Interface(s))
	container.Register(func() Logger { return GoLogger{} })
	container.Register(func() Store { return &LevelDBStore{} }, Name("leveldb"))
	container.Register(func() Store { return &MySQLStore{} }, Name("mysql"), Default())
	var cmStore BlockStore
	err := container.Resolve(&cmStore)
	assert.Nil(t, err)
	cmStore.SaveBlock(123)
	var l *Logger
	err = container.Register(NewBlockStoreImpl, Parameters(map[int]interface{}{2: "chain1"}), Interface(l), Default())
	assert.NotNil(t, err)
	t.Log(err)
}
func TestContainer_RegisterLifestyleTransient(t *testing.T) {
	container := NewContainer()
	container.Register(NewBlockStore, Parameters(map[int]interface{}{2: "chain1"}))
	container.Register(func() Logger { return GoLogger{} })
	container.Register(func() Store {
		fmt.Println("Init new LevelDBStore")
		return &LevelDBStore{heightList: make([]uint64, 0)}
	}, Lifestyle(true))
	var cmStore BlockStore
	err := container.Resolve(&cmStore)
	assert.Nil(t, err)
	cmStore.SaveBlock(123)
	cmStore.SaveBlock(456)
	var store Store
	container.Resolve(&store)
	store.Save(789)
}
func TestContainer_RegisterDependsOn(t *testing.T) {
	container := NewContainer()
	container.Register(NewBlockStore, Parameters(map[int]interface{}{2: "chain1"}), DependsOn(map[int]string{0: "leveldb"}))
	container.Register(func() Logger { return GoLogger{} })
	container.Register(func() Store { return &LevelDBStore{} }, Name("leveldb"))
	container.Register(func() Store { return &MySQLStore{} }, Name("mysql"), Default())
	var cmStore BlockStore
	err := container.Resolve(&cmStore)
	assert.Nil(t, err)
	cmStore.SaveBlock(123)
}

func TestContainer_RegisterInstance(t *testing.T) {
	container := NewContainer()
	container.Register(NewBlockStore, Parameters(map[int]interface{}{2: "chain1"}), DependsOn(map[int]string{0: "leveldb"}))
	l := &GoLogger{name: "mylogger"}
	var log Logger
	err := container.RegisterInstance(&log, l)
	assert.Nil(t, err)
	container.Register(func() Store { return &LevelDBStore{} }, Name("leveldb"))
	container.Register(func() Store { return &MySQLStore{} }, Name("mysql"), Default())
	var cmStore BlockStore
	err = container.Resolve(&cmStore)
	assert.Nil(t, err)
	cmStore.SaveBlock(123)
}
func TestContainer_Resolve(t *testing.T) {
	container := NewContainer()
	container.Register(NewBlockStore, Parameters(map[int]interface{}{2: "chain1"}))
	container.Register(func() Logger { return GoLogger{} })
	container.Register(func() Store { return &LevelDBStore{} }, Name("leveldb"))
	container.Register(func() Store { return &MySQLStore{} }, Name("mysql"), Default())
	var cmStore BlockStore
	err := container.Resolve(&cmStore, Arguments(map[int]interface{}{2: "chain2"}))
	assert.Nil(t, err)
	cmStore.SaveBlock(123)
	var store Store
	err = container.Resolve(&store, ResolveName("leveldb"))
	assert.Nil(t, err)
	store.Save(456)
	err = container.Resolve(&store, ResolveName("leveldb1"))
	assert.NotNil(t, err)
	t.Log(err)
}
func TestContainer_RegisterStructInitArg(t *testing.T) {
	container := NewContainer()
	arg := &InitBlockStoreArg{chainId: "chain3"}
	container.Register(InitBlockStore, Parameters(map[int]interface{}{0: arg}))
	container.Register(func() Logger { return GoLogger{} })
	container.Register(func() Store { return &LevelDBStore{} }, Name("leveldb"))
	container.Register(func() Store { return &MySQLStore{} }, Name("mysql"), Default())
	var cmStore BlockStore
	err := container.Resolve(&cmStore)
	assert.Nil(t, err)
	cmStore.SaveBlock(123)
}

func TestContainer_Fill(t *testing.T) {
	container := NewContainer()
	arg := &InitBlockStoreArg{chainId: "chain3"}
	container.Register(InitBlockStore, Parameters(map[int]interface{}{0: arg}))
	container.Register(func() Logger { return GoLogger{} })
	container.Register(func() Store { return &LevelDBStore{} }, Name("leveldb"))
	container.Register(func() Store { return &MySQLStore{} }, Name("mysql"), Default())
	err := container.Fill(arg)
	assert.Nil(t, err)
	t.Logf("%#v", arg)
}
func TestContainer_ResolveStructInitArg(t *testing.T) {
	container := NewContainer()
	container.Register(InitBlockStore)
	container.Register(func() Logger { return GoLogger{} })
	container.Register(func() Store { return &LevelDBStore{} }, Name("leveldb"))
	container.Register(func() Store { return &MySQLStore{} }, Name("mysql"), Default())
	var cmStore BlockStore
	arg := &InitBlockStoreArg{chainId: "chain3"}
	err := container.Resolve(&cmStore, Arguments(map[int]interface{}{0: arg}))
	assert.Nil(t, err)
	cmStore.SaveBlock(123)
}
func TestContainer_ResolveStructInitArgOptional(t *testing.T) {
	container := NewContainer()
	arg := &InitBlockStoreArg{chainId: "chain3"}
	container.Register(InitBlockStore, Parameters(map[int]interface{}{0: arg}))
	container.Register(func() Store { return &LevelDBStore{} }, Name("leveldb"))
	container.Register(func() Store { return &MySQLStore{} }, Name("mysql"), Default())
	err := container.Fill(arg)
	assert.Nil(t, err)
	t.Logf("%#v", arg)
}

type MultiStoreArg struct {
	store []Store
	l     Logger
}

func TestContainer_FillSlice(t *testing.T) {
	container := NewContainer()
	arg := &MultiStoreArg{}
	container.Register(InitBlockStore)
	container.Register(func() Logger { return GoLogger{} })
	container.Register(func() Store { return &LevelDBStore{} }, Name("leveldb"))
	container.Register(func() Store { return &MySQLStore{} }, Name("mysql"), Default())
	err := container.Fill(arg)
	assert.Nil(t, err)
	t.Logf("%#v", arg)
	for i, store := range arg.store {
		arg.l.Debug("call store.save ...")
		store.Save(uint64(i))
	}
}

func TestContainer_Call(t *testing.T) {
	container := NewContainer()
	container.Register(func() Logger { return GoLogger{} })
	container.Register(func() Store { return &LevelDBStore{} }, Name("leveldb"))
	container.Register(func() Store { return &MySQLStore{} }, Name("mysql"), Default())
	var save1 = func(l Logger, store Store, height uint64) {
		l.Debug("start...")
		store.Save(height)
		l.Debug("end.")
	}
	var save2 = func(l Logger, store Store) {
		save1(l, store, 1234)
	}
	var saveErr = func(l Logger, store Store) error {
		return errors.New("not implement")
	}
	_, err := container.Call(save2)
	assert.Nil(t, err)
	_, err = container.Call(save2, CallDependsOn(map[int]string{1: "leveldb"}))
	assert.Nil(t, err)
	_, err = container.Call(save1, CallDependsOn(map[int]string{1: "mysql"}), CallArguments(map[int]interface{}{2: uint64(1234)}))
	assert.Nil(t, err)
	_, err = container.Call(saveErr, CallDependsOn(map[int]string{1: "leveldb"}))
	assert.NotNil(t, err)
	t.Log(err)
}
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