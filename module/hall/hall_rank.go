package hall

import (
	"baseservice/middleware/authenticate"
	"baseservice/model/user"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	mongoGo "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	hallModel "hallserver/model/hall"
	"jarvis/base/database"
	"jarvis/base/database/redis"
	"jarvis/base/network"
)

const (
	// 默认每页10个项
	defaultNumberPerPage = 10
)

func (hm *hallModule) rank(ctx network.Context) {
	// 反序列化数据
	request := hallModel.RankRequest{}
	if err := json.Unmarshal(ctx.Request().Data, &request); err != nil {
		printReplyError(ctx.ServerError(err))
		return
	}

	// 实例化响应
	response := &hallModel.RankResponse{}
	// 调用函数
	err := hm.getRankList(request, response)
	if err != nil {
		fmt.Printf("rank error : %s", err.Error())
		printReplyError(ctx.ServerError(err))
		return
	}

	newSession := ctx.Extra(authenticate.ContextExtraSessionKey, "")
	response.Session = newSession.(string)

	// 序列化响应
	data, err := json.Marshal(response)
	if err != nil {
		fmt.Printf("marshal response error : %s", err.Error())
		printReplyError(ctx.ServerError(err))
		return
	}

	// 返回响应
	printReplyError(ctx.Success(data))
}

func (hm *hallModule) getRankList(request hallModel.RankRequest, response *hallModel.RankResponse) error {
	if request.NumberPerPage == 0 {
		request.NumberPerPage = defaultNumberPerPage
	}
	// 默认第一页
	if request.CurrentPage <= 0 {
		request.CurrentPage = 1
	}

	switch request.RankType {
	case 0: // 0-全部
		{
			if err := hm.onlineRank(request, response); err != nil {
				return err
			}
			if err := hm.ownMoneyRank(request, response); err != nil {
				return err
			}
			return hm.earnMoneyRank(request, response)
		}
	case 1: // 1-在线时长排行
		return hm.onlineRank(request, response)
	case 2: // 2-拥有金币排行
		return hm.ownMoneyRank(request, response)
	case 3: // 3-累计赚取金币排行
		return hm.earnMoneyRank(request, response)
	default:
		return errors.New("wrong rank type")
	}
}

// 1-在线时长排行
func (hm *hallModule) onlineRank(request hallModel.RankRequest, response *hallModel.RankResponse) error {
	return nil
}

// 2-拥有金币排行
func (hm *hallModule) ownMoneyRank(request hallModel.RankRequest, response *hallModel.RankResponse) error {
	// 获取 MySQL 连接
	conn, err := database.GetMySQLConn()
	if err != nil {
		return err
	}
	defer conn.Close()

	order := "select name,account_balance from `jarvis`.`dynamic_userInfo` order by account_balance desc ,account_token limit ? , ?"
	rows, err := conn.QueryContext(context.Background(), order, (request.CurrentPage-1)*request.NumberPerPage, request.NumberPerPage)
	if err != nil {
		return err
	}
	defer rows.Close()

	ownMoneyList := make([]hallModel.OwnMoneyItem, 0)
	for rows.Next() {
		item := hallModel.OwnMoneyItem{}
		if err := rows.Scan(&item.Name, &item.Money); err != nil {
			return err
		}
		ownMoneyList = append(ownMoneyList, item)
	}

	// 赋值
	response.OwnMoneyList = ownMoneyList

	// 如果是特定 rank 排行榜，todo : 查询总页码数
	if request.RankType == 2 {
		response.CurrentPage = request.CurrentPage
		response.NumberPerPage = request.NumberPerPage
		//response.TotalPage = 10
	}

	return nil
}

// 3-累计赚取金币排行
// 查询所有非充值帐变记录总和
func (hm *hallModule) earnMoneyRank(request hallModel.RankRequest, response *hallModel.RankResponse) error {
	conn, err := database.GetMongoConn("dynamic_user_account_balance_update_record")
	if err != nil {
		return err
	}

	// 构建 mongo 过滤筛选 {"$match":{"type":2}}
	matchState := bson.D{
		{"$match", bson.M{"type": 2}},
	}

	groupState := bson.D{
		{
			"$group", bson.D{
				{"_id", "$user"},
				{"total", bson.M{
					"$sum": "$amount",
				}},
			},
		},
	}

	sortState := bson.D{
		{"$sort", bson.M{"total": -1, "_id": 1}},
	}

	limitState := bson.D{
		{"$limit", request.NumberPerPage},
	}

	skipState := bson.D{
		{"$skip", (request.CurrentPage - 1) * request.NumberPerPage},
	}

	curcor, err := conn.Aggregate(context.Background(), mongoGo.Pipeline{matchState, groupState, sortState, limitState, skipState}, &options.AggregateOptions{})
	if err != nil {
		return err
	}

	var results []map[string]interface{}
	if err := curcor.All(context.Background(), &results); err != nil {
		return err
	}

	earnMoneyList := make([]hallModel.EarnMoneyItem, 0)
	for _, result := range results {
		// _id 取出校验
		_id, exists := result["_id"]
		if !exists {
			return errors.New("_id key not exists")
		}
		token, ok := _id.(string)
		if !ok {
			return errors.New("_id id not string type")
		}
		// 从 redis 取得用户信息
		u, err := GetUserInfoFromRedis(token)
		if err != nil {
			return err
		}
		// total 取出校验
		total, exists := result["total"]
		if !exists {
			return errors.New("total key not exists")
		}
		earn, ok := total.(int64)
		if !ok {
			return errors.New("total id not int type")
		}

		// 构建 item
		item := hallModel.EarnMoneyItem{
			Name:  u.Info.Name,
			Money: earn,
		}

		earnMoneyList = append(earnMoneyList, item)
	}

	response.EarnMoneyList = earnMoneyList

	return nil
}

// 根据 token 从 redis 中获取用户数据
func GetUserInfoFromRedis(token string) (user.User, error) {
	infoStr, err := redis.HGet("UsersInfo", "User:"+token)
	if err != nil {
		return user.User{}, err
	}

	u := user.User{}

	return u, json.Unmarshal([]byte(infoStr), &u)
}
