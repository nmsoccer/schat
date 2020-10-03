package cs

const (
  CHAT_MSG_TYPE_TEXT = 0
  CHAT_MSG_TYPE_IMG  = 1
)

//create group
type CSCreateGroupReq struct {
	Name string `json:"name"`
	Pass string `json:"pass"`
}

type CSCreateGroupRsp struct {
	Result int `json:"result"`
	GrpId int64 `json:"grp_id"`
	Name string `json:"name"`
	MemberCnt int `json:"member_count"`
	CreateTs int64 `json:"create_ts"`
}

//apply group
type CSApplyGroupReq struct {
	GrpId   int64  `json:"grp_id"`
	GrpName string `json:"grp_name"`
	Pass    string `json:"pass"`
	Msg     string `json:"msg"`
}

type CSApplyGroupRsp struct {
	Result  int `json:"result"`
	GrpId   int64  `json:"grp_id"`
	GrpName string `json:"grp_name"`
}

type CSApplyGroupNotify struct {
	ApplyUid  int64  `json:"apply_uid"`
	ApplyName string  `json:"apply_name"`
	ApplyMsg  string `json:"apply_msg"`
	GrpId     int64  `json:"grp_id"`
	GrpName   string `json:"grp_name"`
}

type CSApplyGroupAudit struct {
	Audit     int  `json:"audit"` //0:deny 1:pass
	ApplyUid  int64  `json:"apply_uid"`
	GrpId     int64  `json:"grp_id"`
	GrpName   string `json:"grp_name"`
}

//Send Chat
type ChatMsg struct {
	ChatType  int   `json:"chat_type"` //CHAT_MSG_TYPE_XX
	MsgId     int64 `json:"msg_id"`
	GrpId     int64 `json:"grp_id"`
	SenderUid int64 `json:"sender_uid"`
	SenderName string `json:"sender"`
	SendTs    int64 `json:"send_ts"`
	Content   string `json:"content"`
}

type CSSendChatReq struct {
    TempMsgId int64   `json:"temp_id"`
	ChatType  int   `json:"chat_type"` //CHAT_MSG_TYPE_XX
	GrpId     int64 `json:"grp_id"`
	Content   string `json:"content"`
}

type CSSendChatRsp struct {
	TempMsgId int64   `json:"temp_id"`
	Result    int     `json:"result"`
	ChatMsg   *ChatMsg `json:"chat_msg"`
}

type CSSyncChatList struct {
	GrpId  int64 `json:"grp_id"`
	Count  int   `json:"count"`
	ChatList []*ChatMsg `json:"chat_list"`
}