package cs

const (
	COMMON_NOTIFY_T_FILE_ADDR = 1
)

//server --> client common notify
type CSCommonNotify struct {
	NotifyType int 	  `json:"type"` //refer COMMON_NOTIFY_T_XX
	GrpId      int64  `json:"grp_id"`
	IntV       int64  `json:"intv"`
	StrV       string `json:"strv"`
	StrS       []string `json:"strs"`
}
