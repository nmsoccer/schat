package cs


type UserBasic struct {
	Uid   int64  `json:"uid"`
	Name  string `json:"name"`
	Addr  string `json:"addr"`
	Sex   uint8  `json:"sex"` //1:male 2:female
	Level int32  `json:"level"`
}

type UserChatGroup struct {
	GroupId int64
	GroupName string
	LastMsgId int64 //last readed
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
