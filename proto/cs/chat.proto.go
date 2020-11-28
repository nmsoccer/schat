package cs

const (
  CHAT_MSG_TYPE_TEXT = 0
  CHAT_MSG_TYPE_IMG  = 1

  //sync group field
  SYNC_GROUP_FIELD_ALL = 1
  SYNC_GROUP_FIELD_SNAP = 2

  //group attr
  GROUP_ATTR_VISIBLE = 0
  GROUP_ATTR_INVISIBLE = 1
  GROUP_ATTR_DESC = 2
  GROUP_ATTR_GRP_NAME = 3
  GROUP_ATTR_GRP_HEAD = 4
)

//create group
type CSCreateGroupReq struct {
	Name string `json:"name"`
	Pass string `json:"pass"`
	Desc string `json:"desc"`
}

type CSCreateGroupRsp struct {
	Result int `json:"result"`
	GrpId int64 `json:"grp_id"`
	Name string `json:"name"`
	MemberCnt int `json:"member_count"`
	CreateTs int64 `json:"create_ts"`
	Desc string `json:"desc"`
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
	SyncType int8 `json:"sync_type"` //0:normal 1:history
	GrpId  int64 `json:"grp_id"`
	Count  int   `json:"count"`
	ChatList []*ChatMsg `json:"chat_list"`
}

type CSExitGroupReq  struct {
	GrpId  int64 `json:"grp_id"`
}

type CSExitGroupRsp  struct {
	Result int `json:"result"`
	GrpId  int64 `json:"grp_id"`
	GrpName string `json:"grp_name"`
	DelGroup int8 `json:"del_group"`
	ByKick   int8 `json:"by_kick"`
}

type CSChatHistoryReq struct {
	GrpId  int64 `json:"grp_id"`
	//fetch chat history before now_mst_id(not include now_msg_id) max 40. aka from [now_msg_id-40 , now_msg_id) if 0 fetch from latest_msg_id
	NowMsgId int64 `json:"now_msg_id"`
}

type CSKickGroupReq struct {
	GrpId  int64 `json:"grp_id"`
	KickUid int64 `json:"kick_uid"`
}

//QUERY_FLAG
type CSQueryGroupReq struct {
	GrpId  int64 `json:"grp_id"`
}


type CSSyncGroupInfo struct {
	Field  int32 `json:"field"` //refer SYNC_GROUP_FIELD_xx
	GrpId  int64 `json:"grp_id"`
	GrpInfo *ChatGroup `json:"grp_info"`
	GrpSnap *GroupGroundItem `json:"grp_snap"`
}

type CSFetchUserProfileReq struct {
	TargetList []int64  `json:"target_list"`
}

type CSFetchUserProfileRsp struct {
	Profiles map[int64]*UserProfile  `json:"profiles"`
}

type CSChgGroupAttrReq struct {
	Attr   int   `json:"attr"` //refer GROUP_ATTR_XX
	GrpId  int64 `json:"grp_id"`
	IntV   int64  `json:"int_v"`
	StrV   string `json:"str_v"`
}

type CSChgGroupAttrRsp struct {
	Result int   `json:"result"` //COMMON_RESULT_XX
	Attr   int   `json:"attr"` //refer GROUP_ATTR_XX
	GrpId  int64 `json:"grp_id"`
	IntV   int64  `json:"int_v"`
	StrV   string `json:"str_v"`
}

type CSGroupGroundReq struct {
	StartIndex int `json:"start"` //search start index
}

type CSGroupGroundRsp struct {
	Count int `json:"count"`
	ItemList []*GroupGroundItem `json:"item_list"`
}

type CSUpdateChatReq struct {
    UpdateType int `json:"upt_type"`
    Grpid int64 `json:"grp_id"`
    MsgId int64 `json:"msg_id"`
}

type CSUpdateChatRsp struct {
	Result int   `json:"result"` //COMMON_RESULT_XX
	UpdateType int `json:"upt_type"`
	GrpId int64 `json:"grp_id"`
	MsgId int64 `json:"msg_id"`
}
