package hall

import (
	"errors"
	hallModel "hallserver/model/hall"
	"jarvis/base/network"
	uTime "jarvis/util/time"
	"log"
	"sync"
	"time"
)

type (
	// 模块定义
	hallModule struct {
		usClient        network.Client // userServer client
		connManage      *connManage
		announceChannel chan hallModel.Announcement // 公告通道
	}

	// 连接映射用户管理
	connManage struct {
		connMapLock sync.Mutex
		connMap     map[string]string // map[conn id]token
	}
)

const (
	// 模块名定义
	ModuleName = "Hall"

	// 用户服务开放端口
	userServerAddress = ":8082"

	// 公告通道大小
	announceChannelSize = 10
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
		usClient: c,
		connManage: &connManage{
			connMapLock: sync.Mutex{},
			connMap:     make(map[string]string),
		},
		announceChannel: make(chan hallModel.Announcement, announceChannelSize),
	}
}

// 将默认模块声明为模块
func NewModule() network.Module {
	ticker := uTime.NewTicker(time.Second*time.Duration(1), defaultHall.announce)
	ticker.Run()

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
// todo : 1.公告
// todo : 2.排行榜
// todo : 3.Banner
// todo : 4.游戏
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

// 观察者要求实现函数: InitiativeSend(network.Context)
func (hm *hallModule) InitiativeSend(ctx network.Context) {
	for {
		announcement, ok := <-hm.announceChannel
		if !ok {
			break
		}

		for _, id := range announcement.IDs {
			go printReplyError(ctx.FindAndSendReply(id, announcement.Reply, announcement.Data))
		}
	}
}

// 添加连接
func (cm *connManage) AddConn(id string) error {
	cm.connMapLock.Lock()
	defer cm.connMapLock.Unlock()

	if _, exist := cm.connMap[id]; exist {
		return errors.New(id + " exist")
	}

	cm.connMap[id] = ""

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
func (cm *connManage) AddConnUserInfo(id, token string) error {
	if id == "" || token == "" {
		return errors.New("id,token or session can't be empty")
	}

	cm.connMapLock.Lock()
	defer cm.connMapLock.Unlock()

	_, exist := cm.connMap[id]
	if !exist {
		return errors.New(id + " doesn't exist")
	}

	cm.connMap[id] = token

	log.Printf("ConnManage [%s] user : %s", id, token)

	return nil
}

// 获取当前所有连接的 ID
func (cm *connManage) AllConnIDList() []string {
	ids := make([]string, 0)

	cm.connMapLock.Lock()
	defer cm.connMapLock.Unlock()

	for id := range cm.connMap {
		ids = append(ids, id)
	}

	return ids
}

// 打印回复错误
func printReplyError(err error) {
	if err == nil {
		return
	}

	log.Printf("Reply error : %s", err.Error())
}
