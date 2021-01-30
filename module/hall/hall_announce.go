package hall

import (
	redisGo "github.com/gomodule/redigo/redis"
	hallModel "hallserver/model/hall"
	"jarvis/base/database"
	"log"
	"time"
)

const (
	// redis 全局公告键
	RedisAnnouncementKey = "GLOBAL:ANNOUNCEMENTS"
	// 主动推送到端的路径
	AnnounceRoute = "ANNOUNCE"
)

func (hm *hallModule) announce(t time.Time) bool {
	redisConn, err := database.GetRedisConn()
	if err != nil {
		log.Printf("get redis conn error : %s", err.Error())
		return true
	}
	defer redisConn.Close()

	v, err := redisGo.String(redisConn.Do("lpop", RedisAnnouncementKey))
	if err != nil {
		if err != redisGo.ErrNil {
			log.Printf("rpush announcement to redis error : %s", err.Error())
		}
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
