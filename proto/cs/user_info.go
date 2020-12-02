package cs

const (
	SEX_MALE = 1
	SEX_FEMALE = 2
)

type UserBasic struct {
	Uid   int64  `json:"uid"`
	Name  string `json:"name"`
	Addr  string `json:"addr"`
	Sex   uint8  `json:"sex"` //refer SEX_XX
	Level int32  `json:"level"`
	HeadUrl string `json:"head_url"`
}

type UserChatGroup struct {
	GroupId int64  `json:"grp_id"`
	GroupName string  `json:"grp_name"`
	LastMsgId int64   `json:"last_msg"` //last readed
	EnterTs   int64  `json:"enter_ts"` //enter ts
}

type UserChatInfo struct {
	AllGroup int32  `json:"all_group"`
	AllGroups map[int64] *UserChatGroup  `json:"all_groups"`
	MasterGroup int32  `json:"master_group"`
	MasterGroups map[int64] bool  `json:"master_groups"`
}

type UserDetail struct {
	Exp int32 `json:"exp"`
	//Depot *UserDepot `json:"user_depot"`
	ChatInfo *UserChatInfo `json:"chat_info"`
	Desc string `json:"desc"`
	ClientDesKey string `json:"c_des_key"`
}

type UserProfile struct {
	Uid   int64  `json:"uid"`
	Name  string `json:"name"`
	Addr  string `json:"addr"`
	Sex   uint8  `json:"sex"` //refer SEX_XX
	Level int32  `json:"level"`
	HeadUrl string `json:"head_url"`
	Desc string `json:"desc"`
}

