package hall

import "baseservice/middleware/authenticate"

type (
	// 在线时长项
	OnlineTimeItem struct {
		Name       string `json:"name"`        // 用户名
		OnlineTime int64  `json:"online_time"` // 在线时长统计，单位分钟
	}

	// 拥有金币项
	OwnMoneyItem struct {
		Name  string `json:"name"`  // 用户名
		Money int64  `json:"money"` // 单位:分
	}

	// 赚取金币项
	EarnMoneyItem struct {
		Name  string `json:"name"`  // 用户名
		Money int64  `json:"money"` // 单位:分
	}

	// 排行榜请求
	RankRequest struct {
		authenticate.Request
		RankType      int `json:"rank_type"`       // 排行榜类别 0-全部 1-在线时长排行 2-拥有金币排行 3-累计赚取金币排行
		NumberPerPage int `json:"number_per_page"` // 每页数量
		CurrentPage   int `json:"current_page"`    // 当前页码
	}

	// 排行榜响应
	RankResponse struct {
		authenticate.Response
		OnlineTimeList []OnlineTimeItem `json:"online_time_list"` // 在线时长列表
		OwnMoneyList   []OwnMoneyItem   `json:"own_money_list"`   // 拥有金币排行
		EarnMoneyList  []EarnMoneyItem  `json:"earn_money_list"`  // 累计赚取金币排行
		NumberPerPage  int              `json:"number_per_page"`  // 每页数量
		CurrentPage    int              `json:"current_page"`     // 当前页码
		TotalPage      int              `json:"total_page"`       // 总共页码数
	}
)
