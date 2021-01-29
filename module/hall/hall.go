package hall

import (
	"errors"
	"jarvis/base/network"
	"log"
	"sync"
)

type (
	// 模块定义
	hallModule struct {
		client     network.Client
		connManage *connManage
	}

	// 连接映射用户管理
	connManage struct {
		connMapLock sync.Mutex
		connMap     map[string]map[string]string // map[conn id]map[string]string{Token:token,Session:session}
	}
)

const (
	// 模块名定义
	ModuleName = "Hall"

	// 用户服务开放端口
	userServerAddress = ":8082"

	// 连接映射用户信息 Token 键
	connUserToken = "Token"

	// 连接映射用户信息 Session 键
	connUserSession = "Session"
)

var (
	// 默认模块
	defaultHall *hallModule
)

func init() {
	// 实例化客户端
	c := network.NewGRPCClient(userServerAddress, network.DefaultPackager())
	if err := c.Initialize(); err != nil {
		log.Fatalln(err.Error())
		return
	}

	// 默认模块实例化
	defaultHall = &hallModule{
		client: c,
		connManage: &connManage{
			connMapLock: sync.Mutex{},
			connMap:     make(map[string]map[string]string),
		},
	}
}

// 将默认模块声明为模块
func NewModule() network.Module {
	return defaultHall
}

// 将默认模块声明为观察者
func NewObserver() network.Observer {
	return defaultHall
}

// 模块要求实现函数: Name() string
func (hm *hallModule) Name() string {
	return ModuleName
}

// 模块要求实现函数: Route() map[string][]network.RouteHandleFunc
func (hm *hallModule) Route() map[string][]network.RouteHandleFunc {
	return map[string][]network.RouteHandleFunc{
		"login": {hm.login}, // 登录
	}
}

// 观察者要求实现函数: ObserveConnect(string)
func (hm *hallModule) ObserveConnect(id string) {
	if err := hm.connManage.AddConn(id); err != nil {
		log.Println(err.Error())
	}
}

// 观察者要求实现函数: ObserveDisconnect(string)
func (hm *hallModule) ObserveDisconnect(id string) {
	if err := hm.connManage.RemoveConn(id); err != nil {
		log.Println(err.Error())
	}
}

// 添加连接
func (cm *connManage) AddConn(id string) error {
	cm.connMapLock.Lock()
	defer cm.connMapLock.Unlock()

	if _, exist := cm.connMap[id]; exist {
		return errors.New(id + " exist")
	}

	cm.connMap[id] = map[string]string{}

	log.Printf("ConnManage : %+v", cm)

	return nil
}

// 删除连接
func (cm *connManage) RemoveConn(id string) error {
	cm.connMapLock.Lock()
	defer cm.connMapLock.Unlock()

	if _, exist := cm.connMap[id]; !exist {
		return errors.New(id + " doesn't Exist")
	}

	delete(cm.connMap, id)
	return nil
}

// 添加连接信息
func (cm *connManage) AddConnUserInfo(id, token, session string) error {
	if id == "" || token == "" || session == "" {
		return errors.New("id,token or session can't be empty")
	}

	cm.connMapLock.Lock()
	defer cm.connMapLock.Unlock()

	infoMap, exist := cm.connMap[id]
	if !exist {
		return errors.New(id + " doesn't exist")
	}

	infoMap[token] = session

	log.Printf("ConnManage [%s] user : %+v", id, infoMap)

	return nil
}

// 校验连接信息
func (cm *connManage) VerifyConnUserInfo(id, token, session string) (bool, error) {
	if id == "" || token == "" || session == "" {
		return false, errors.New("id,token or session can't be empty")
	}

	cm.connMapLock.Lock()
	defer cm.connMapLock.Unlock()

	infoMap, exist := cm.connMap[id]
	if !exist {
		return false, errors.New(id + " doesn't exist")
	}

	if session != infoMap[token] {
		return false, nil
	}

	return true, nil
}

// 打印回复错误
func printReplyError(err error) {
	if err == nil {
		return
	}

	log.Printf("Reply error : %s", err.Error())
}
