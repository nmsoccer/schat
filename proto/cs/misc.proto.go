package cs


//server --> client common notify
type CSCommonNotify struct {
	NotifyType int 	  `json:"type"` //refer COMMON_NOTIFY_T_XX
	GrpId      int64  `json:"grp_id"`
	IntV       int64  `json:"intv"`
	StrV       string `json:"strv"`
	StrS       []string `json:"strs"`
}

//client --> server common query
type CSCommonQuery struct {
	QueryType int `json:"type"`
	GrpId      int64  `json:"grp_id"`
	IntV       int64  `json:"int_v"`
	StrV       string `json:"str_v"`
}