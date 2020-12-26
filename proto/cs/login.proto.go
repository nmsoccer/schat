package cs

type CSLoginReq struct {
	Name   string `json:"name"`
	Pass   string `json:"pass"`
	Device string `json:"device"`
	Version string `json:"version"`
	Flag    int64 `json:"flag"`
}

type CSLoginRsp struct {
	Result int        `json:"result"`
	Name   string     `json:"name"`
	Basic  UserBasic  `json:"basic"`
	Detail UserDetail `json:"user_detail"`
	Flag    int64 `json:"flag"`
	LastLogout int64 `json:"last_logout"`
}

type CSLogoutReq struct {
	Uid int64 `json:"uid"`
}

type CSLogoutRsp struct {
	Result int    `json:"result"`
	Uid    int64  `json:"uid"`
	Msg    string `json:"msg"`
}

type CSRegReq struct {
	Name string `json:"name"`
	Pass string `json:"pass"`
	RoleName string `json:"role_name"`
	Sex  uint8  `json:"sex"`
	Addr string `json:"addr"`
	Desc string `json:"desc"`
}

type CSRegRsp struct {
	Result int    `json:"result"`
	Name   string `json:"name"`
}

type CSUpdateUserReq struct {
	RoleName string `json:"role_name"`//if len>0 means update
	Addr string `json:"addr"`//if len>0 means update
	Desc string `json:"desc"`//if len>0 means update
	Passwd string `json:"pass"`//if len>0 will update
}

type CSUpdateUserRsp struct {
	Result int `json:"result"` //common result
	RoleName string `json:"role_name"`//if len>0 means update
	Addr string `json:"addr"`//if len>0 means update
	Desc string `json:"desc"`//if len>0 means update
	Passwd string `json:"pass"`//if len>0 will update
}