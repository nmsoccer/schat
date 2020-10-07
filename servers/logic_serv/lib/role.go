package lib

import (
	"schat/proto/ss"
	"schat/servers/comm"
	"time"
)

const (
	NOTIFY_ONLINE_FLAG_LOGIN = 1
	NOTIFY_ONLINE_FLAG_LOGOUT = 2

	PERIOD_NOTIFY_USERS_ONLINE = 7000 // 7 sec per tick
	NOTIFY_ONLINE_TIME_SPAN = 22 //22 second
)

//Notify To OnlineServ(ALL)
//flag:refer NOTIFY_ONLINE_FLAG_xx
func NotifyOnline(pconfig *Config , uid int64 , flag int) {
	var _func_ = "<NotifyOnline>"
	log := pconfig.Comm.Log

	//notify
	//log.Debug("%s uid:%d flag:%d" , _func_ , uid , flag)
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
	pss_msg , err := comm.GenDispMsg(ss.DISP_MSG_TARGET_ONLINE_SERVER , ss.DISP_MSG_METHOD_ALL , ss.DISP_PROTO_TYPE_DISP_COMMON_NOTIFY ,
		0 , pconfig.ProcId , 0 , pnotify)
	if err != nil {
		log.Err("%s gen disp ss failed! err:%v uid:%d" , _func_ , err , uid)
		return
	}

	//Send
	SendToDisp(pconfig , 0 , pss_msg)
}

func QueryFileServAddr(pconfig *Config , uid int64) {
	var _func_ = "<QueryFileServAddr>"
	log := pconfig.Comm.Log

	//ss
	pnotify := new(ss.MsgCommonNotify)
	pnotify.Uid = uid
	pnotify.NotifyType = ss.COMMON_NOTIFY_TYPE_NOTIFY_FILE_ADDR

	pss_msg , err := comm.GenDispMsg(ss.DISP_MSG_TARGET_DIR_SERVER , ss.DISP_MSG_METHOD_RAND , ss.DISP_PROTO_TYPE_DISP_COMMON_NOTIFY ,
		0 , pconfig.ProcId , 0 , pnotify)
	if err != nil {
		log.Err("%s gen ss failed! err:%v uid:%d" , _func_ , err , uid)
		return
	}

	//to dir
	SendToDisp(pconfig , 0 , pss_msg)
}

func RecvFileAddrNotify(pconfig *Config , pnotify *ss.MsgCommonNotify) {
	var _func_ = "<RecvFileAddrNotify>"
	log := pconfig.Comm.Log
	uid := pnotify.Uid

	//check user
	puser_info := GetUserInfo(pconfig , uid)
	if puser_info == nil {
		log.Err("%s user offline! uid:%d" , _func_ , uid)
		return
	}

	//check file
	if pnotify.IntV==0 || pnotify.Strs==nil || len(pnotify.Strs)==0 {
		log.Err("%s file addr empty! uid:%d" , _func_ , uid)
		return
	}

	//to connect
	var ss_msg ss.SSMsg
	err := comm.FillSSPkg(&ss_msg , ss.SS_PROTO_TYPE_COMMON_NOTIFY , pnotify)
	if err != nil {
		log.Err("%s gen ss failed! uid:%d err:%v" , _func_ , uid , err)
		return
	}

	SendToConnect(pconfig , &ss_msg)
}


func TickNotifyOnline(arg interface{}) {
	pconfig , ok := arg.(*Config)
	if !ok {
		return
	}
	var _func_ = "<TickNotifyOnline>"
	log := pconfig.Comm.Log


	//check online
	if pconfig.Users == nil || pconfig.Users.curr_online<=0 {
		return
	}

	//pmsg
	pnotify := new(ss.MsgCommonNotify)
	pnotify.NotifyType = ss.COMMON_NOTIFY_TYPE_NOTIFY_BATCH_USER_ONLINE
	pnotify.Members = make(map[int64]int32)
	count := 0

	//check stat
	curr_ts := time.Now().Unix()
	for uid , info := range pconfig.Users.user_map {
		if info.last_notify_online + int64(NOTIFY_ONLINE_TIME_SPAN) < curr_ts {
			pnotify.Members[uid] = 1
			count++
			info.last_notify_online = curr_ts
		}
	}

	if count <= 0 {
		return
	}

	//ss
	pss_msg , err := comm.GenDispMsg(ss.DISP_MSG_TARGET_ONLINE_SERVER , ss.DISP_MSG_METHOD_ALL , ss.DISP_PROTO_TYPE_DISP_COMMON_NOTIFY ,
		0 , pconfig.ProcId , 0 , pnotify)
	if err != nil {
		log.Err("%s gen ss failed! err:%v count:%d" , _func_ , err , count)
		return
	}

	//to online
	SendToDisp(pconfig , 0 , pss_msg)
}


func SaveRolesOnExit(pconfig *Config) {
	var _func_ = "<SaveRolesOnExit>"
	log := pconfig.Comm.Log

	//check online
	if pconfig.Users.curr_online <= 0 || pconfig.Users.user_map == nil {
		return
	}

	//each role
	log.Info("%s will kickout online user! online:%d" , _func_ , pconfig.Users.curr_online)
	for uid , _ := range(pconfig.Users.user_map) {
		UserLogout(pconfig , uid , ss.USER_LOGOUT_REASON_LOGOUT_SERVER_SHUT);
		//send to client
		SendLogoutRsp(pconfig , uid , ss.USER_LOGOUT_REASON_LOGOUT_SERVER_SHUT , "server down");
	}
}
