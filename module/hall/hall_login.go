package hall

import (
	"encoding/json"
	"jarvis/base/network"
	uRand "jarvis/util/rand"
	"log"
)

// 转接到用户服务中
func (hm *hallModule) login(ctx network.Context) {
	log.Println(string(ctx.Request().Data))

	message, err := hm.client.RequestSync(network.Message{
		Module: "User",
		Route:  "login",
		Data:   ctx.Request().Data,
		Reply:  uRand.RandomString(8),
	})
	if err != nil {
		printReplyError(ctx.ServerError(err))
		return
	}

	log.Println(string(message.Data))

	printReplyError(ctx.BinaryReply(message.Data))

	// 将用户登录信息保存到连接管理中
	type B struct {
		Token   string `json:"token"`
		Session string `json:"session"`
	}

	reply := network.Reply{}
	if err := json.Unmarshal(message.Data, &reply); err != nil {
		log.Printf("unmarshal Message.Data [%s] to Reply error : %s", string(message.Data), err.Error())
		return
	}

	if reply.Code != 200 {
		return
	}

	b := B{}
	if err := json.Unmarshal(reply.Data, &b); err != nil {
		log.Printf("unmarshal reply.Data to B error : %s", err.Error())
		return
	}

	if err := hm.connManage.AddConnUserInfo(ctx.Request().ID, b.Token, b.Session); err != nil {
		log.Printf("hm.connManage.AddConnUserInfo: %s", err.Error())
	}
}
