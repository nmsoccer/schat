package cs

type ChatMsg struct {
	ChatType  int   `json:"chat_type"` //CHAT_MSG_TYPE_XX
	MsgId     int64 `json:"msg_id"`
	GrpId     int64 `json:"grp_id"`
	SenderUid int64 `json:"sender_uid"`
	SenderName string `json:"sender"`
	SendTs    int64 `json:"send_ts"`
	Content   string `json:"content"`
}

type ChatGroup struct {
	GrpId 	int64 `json:"grp_id"`
	GrpName string `json:"grp_name"`
	MasterUid int64 `json:"master"`
	MsgCount int64 `json:"msg_count"`
	CreateTs int64 `json:"create"`
	MemCount int32 `json:"mem_count"`
	Members  map[int64]int32 `json:"members"`
}
