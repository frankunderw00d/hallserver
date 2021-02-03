package hall

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	hallModel "hallserver/model/hall"
	"jarvis/base/database"
	"log"
	"time"
)

const (
	// 主动推送到端的路径
	AnnounceRoute = "ANNOUNCE"
)

// 每10秒推送一条公告记录中最新的公告
func (hm *hallModule) announce(t time.Time) bool {
	c, err := database.GetMongoConn("hall_announce_record")
	if err != nil {
		log.Printf("get mongo conn error : %s", err.Error())
		return true
	}

	var result map[string]interface{}
	err = c.FindOne(context.Background(), bson.D{}, options.FindOne().SetSort(bson.D{{"time", -1}})).Decode(&result)
	if err != nil {
		log.Printf("mongo conn find one decode error : %s", err.Error())
		return true
	}

	a, ok := result["announcement"]
	if !ok {
		log.Println("announcement key doesn't exists")
		return true
	}
	v, ok := a.(string)
	if !ok {
		log.Println("announcement value aren't string")
		return true
	}

	// 发送本服务器公告
	inform := hallModel.Announcement{
		Data:  []byte(v),
		IDs:   hm.connManage.AllConnIDList(),
		Reply: AnnounceRoute,
	}
	hm.announceChannel <- inform

	return true
}
