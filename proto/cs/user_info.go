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
	GroupId int64
	GroupName string
	LastMsgId int64 //last readed
	EnterTs   int64 //enter ts
}

type UserChatInfo struct {
	AllGroup int32
	AllGroups map[int64] *UserChatGroup
	MasterGroup int32
	MasterGroups map[int64] bool
}

type UserDetail struct {
	Exp int32 `json:"exp"`
	//Depot *UserDepot `json:"user_depot"`
	ChatInfo *UserChatInfo `json:"chat_info"`
}

type UserProfile struct {
	Uid   int64  `json:"uid"`
	Name  string `json:"name"`
	Addr  string `json:"addr"`
	Sex   uint8  `json:"sex"` //refer SEX_XX
	Level int32  `json:"level"`
	HeadUrl string `json:"head_url"`
}

