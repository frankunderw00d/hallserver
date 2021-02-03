package hall

import (
	"baseservice/base/basic"
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson"
	"jarvis/base/database"
	"jarvis/base/network"
	uRand "jarvis/util/rand"
	"log"
	"time"
)

const (
	// 用户模块名
	UserModule = "User"
	// 用户模块登录接口名
	UserLogin = "login"
	// 用户模块获取用户信息接口名
	UserGetUserInfo = "getUserInfo"
)

// 转接到用户服务中
func (hm *hallModule) login(ctx network.Context) {
	message, err := hm.usClient.RequestSync(network.Message{
		Module: UserModule,
		Route:  UserLogin,
		Data:   ctx.Request().Data,
		Reply:  uRand.RandomString(8),
	})
	if err != nil {
		printReplyError(ctx.ServerError(err))
		return
	}

	printReplyError(ctx.BinaryReply(message.Data))

	// 将用户登录信息保存到连接管理中
	reply := network.Reply{}
	if err := json.Unmarshal(message.Data, &reply); err != nil {
		log.Printf("unmarshal Message.Data [%s] to Reply error : %s", string(message.Data), err.Error())
		return
	}

	// 登录不成功
	if reply.Code != 200 {
		return
	}

	type LoginResponse struct {
		Token   string `json:"token"`
		Session string `json:"session"`
	}

	b := LoginResponse{}
	if err := json.Unmarshal(reply.Data, &b); err != nil {
		log.Printf("unmarshal reply.Data to B error : %s", err.Error())
		return
	}

	if err := hm.connManage.AddConnUserInfo(ctx.Request().ID, b.Token); err != nil {
		log.Printf("hm.connManage.AddConnUserInfo: %s", err.Error())
	}

	// 获取用户信息发送全服公告
	hm.getUserInfo(b.Token, b.Session)
}

// 获取用户信息
func (hm *hallModule) getUserInfo(token, session string) {
	type A struct {
		Token     string `json:"token"`      // 账号唯一标识
		Session   string `json:"session"`    // 会话标识
		SecretKey string `json:"secret_key"` // 加密 key
	}

	request := A{
		Token:     token,
		Session:   session,
		SecretKey: basic.EncryptSecretKey(token, session),
	}

	registerRequest, err := json.Marshal(&request)
	if err != nil {
		log.Printf("marshal register error : %s", err.Error())
		return
	}

	message, err := hm.usClient.RequestSync(network.Message{
		Module: UserModule,
		Route:  UserGetUserInfo,
		Data:   registerRequest,
		Reply:  uRand.RandomString(8),
	})
	if err != nil {
		log.Printf("request sync get user info error : %s", err.Error())
		return
	}

	reply := network.Reply{}
	if err := json.Unmarshal(message.Data, &reply); err != nil {
		log.Printf("unmarshal Message.Data [%s] to Reply error : %s", string(message.Data), err.Error())
		return
	}

	// 获取信息不成功
	if reply.Code != 200 {
		log.Printf("%s : %s", reply.Message, string(reply.Data))
		return
	}

	type LoginResponse struct {
		Session           string `json:"session"`              // 更新会话标识
		AccountType       int    `json:"type"`                 // 账号类型 0-游客 1-绑定用户
		Platform          int    `json:"platform"`             // 所属平台
		Name              string `json:"name"`                 // 用户名
		Age               int    `json:"age"`                  // 用户年龄
		Sex               bool   `json:"sex"`                  // 用户性别
		HeadImage         int    `json:"head_image"`           // 用户头像序号
		Vip               int    `json:"vip"`                  // 用户 vip 等级
		GameBgMusicVolume int    `json:"game_bg_music_volume"` // 背景音乐音量
		GameEffectVolume  int    `json:"game_effect_volume"`   // 音效音量
		AccountBalance    int64  `json:"account_balance"`      // 账户余额(单位:分)
	}

	b := LoginResponse{}
	if err := json.Unmarshal(reply.Data, &b); err != nil {
		log.Printf("unmarshal reply.Data to B error : %s", err.Error())
		return
	}

	log.Printf("Hall Login Get User Information : %+v", b)

	// 将公告记录保存到 mongo
	c, err := database.GetMongoConn("hall_announce_record")
	if err != nil {
		log.Printf("get mongo conn error : %s", err.Error())
		return
	}

	_, err = c.InsertOne(context.Background(), bson.M{
		"announcement": "热烈欢迎用户 " + b.Name + " 登录游戏",
		"from":         "service",
		"time":         time.Now().Format("20060102150405.000000000"),
	})
	if err != nil {
		log.Printf("Distributed sync publish error : %s", err.Error())
		return
	}
}
