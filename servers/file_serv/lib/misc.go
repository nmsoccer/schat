package lib

import (
	"schat/proto/ss"
	"schat/servers/comm"
)

const(
	//TOKEN LEN
	FILE_SERV_TOKEN_LEN = 12

	PERIOD_UPDATE_TOKEN = 3600000 //1h update token
	PERIOD_SYNC_TOKEN = 600000	//10min sync token to dir
)

func UpdateServToken(arg interface{}) {
	var _func_ = "<UpdateServToken>"
	pconfig , ok := arg.(*Config)
	if !ok {
		return
	}
	log := pconfig.Comm.Log

	//update token
	new_token , err := comm.GenRandNumStr(FILE_SERV_TOKEN_LEN)
	if err != nil {
		log.Err("%s fail! rand str err:%v" , _func_ , err)
		return
	}

	pconfig.NowToken = new_token
	log.Info("%s update new token:%s" , _func_ , new_token)

	//to file_server
	pmsg := new(FileMsg)
	pmsg.msg_type = FILE_MSG_UPDATE_TOKEN
	pmsg.str_v = new_token
	pconfig.FileServer.Send(pmsg)

	//to dir serv
	SyncServToken(pconfig)
}

func SyncServToken(arg interface{}) {
	var _func_ = "<SyncServToken>"
	pconfig , ok := arg.(*Config)
	if !ok {
		return
	}
	log := pconfig.Comm.Log

	//to dir
	pnotify := new(ss.MsgCommonNotify)
	pnotify.NotifyType = ss.COMMON_NOTIFY_TYPE_NOTIFY_FILE_TOKEN
	pnotify.IntV = int64(pconfig.FileConfig.ServIndex)
	pnotify.StrV = pconfig.NowToken

	pss_msg , err := comm.GenDispMsg(ss.DISP_MSG_TARGET_DIR_SERVER , ss.DISP_MSG_METHOD_ALL , ss.DISP_PROTO_TYPE_DISP_COMMON_NOTIFY ,
		0 , pconfig.ProcId , 0 , pnotify)
	if err != nil {
		log.Err("%s gen ss failed! err:%v" , _func_ , err)
		return
	}

	//send
	SendToDisp(pconfig , 0 , pss_msg)
}

//dir --> file query
func RecvFileTokenNotify(pconfig *Config , pnotify *ss.MsgCommonNotify , dir_serv int) {
	var _func_ = "<RecvFileTokenNotify>"
	log := pconfig.Comm.Log
	serv_index := int(pnotify.IntV)

	//check index
	if serv_index != pconfig.FileConfig.ServIndex {
		log.Info("%s not my index! %d:%d abandon!" , _func_ , serv_index , pconfig.FileConfig.ServIndex)
		return
	}

	//back to dir
	log.Info("%s get token query from dir_serv:%d will send back! serv_index:%d" , _func_ , dir_serv , serv_index)
	SyncServToken(pconfig)
}