package lib

import (
	"schat/proto/ss"
	"schat/servers/comm"
	"time"
)

const (
	NOTIFY_ONLINE_FLAG_LOGIN = 1
	NOTIFY_ONLINE_FLAG_LOGOUT = 2
)

//Notify To OnlineServ
//flag:refer NOTIFY_ONLINE_FLAG_xx
func NotifyOnline(pconfig *Config , uid int64 , flag int) {
	var _func_ = "<NotifyOnline>"
	log := pconfig.Comm.Log

	//notify
	log.Debug("%s uid:%d flag:%d" , _func_ , uid , flag)
	curr_ts := time.Now().Unix()
	pnotify := new(ss.MsgCommonNotify)
	switch  flag {
	case NOTIFY_ONLINE_FLAG_LOGIN:
		pnotify.NotifyType = ss.COMMON_NOTIFY_TYPE_NOTIFY_USER_LOGIN
	case NOTIFY_ONLINE_FLAG_LOGOUT:
		pnotify.NotifyType = ss.COMMON_NOTIFY_TYPE_NOTIFY_USER_LOGOUT
	default:
		log.Err("%s illegal flag:%d uid:%d" , _func_ , flag , uid)
		return
	}
	pnotify.Uid = uid
	pnotify.IntV = curr_ts
	pss_msg , err := comm.GenDispMsg(ss.DISP_MSG_TARGET_ONLINE_SERVER , ss.DISP_MSG_METHOD_RAND , ss.DISP_PROTO_TYPE_DISP_COMMON_NOTIFY ,
		0 , pconfig.ProcId , 0 , pnotify)
	if err != nil {
		log.Err("%s gen disp ss failed! err:%v uid:%d" , _func_ , err , uid)
		return
	}

	//Send
	SendToDisp(pconfig , 0 , pss_msg)
}

