package hall

type (
	// 公告信息
	Announcement struct {
		Data  []byte   // 具体消息
		IDs   []string // 需要通知的 ID 组
		Reply string   // 主动通知地址
	}
)
