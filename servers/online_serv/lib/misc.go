package lib

import (
	"schat/proto/ss"
	"schat/servers/comm"
)

func RecvUploadFileNotify(pconfig *Config, pnotify *ss.MsgCommonNotify, file_server int) {
	var _func_ = "<RecvUploadFileNotify>"
	log := pconfig.Comm.Log
	uid := pnotify.Uid
	grp_id := pnotify.GrpId
	file_type := pnotify.Occupy

	log.Debug("%s. uid:%d grp_id:%d url:%s file_type:%d", _func_, uid, grp_id, pnotify.StrV , file_type)
	//check online
	puser_info := GetUserInfo(pconfig, uid)
	if puser_info == nil {
		log.Err("%s not online! will send back! uid:%d", _func_, uid)

		pnotify.IntV = comm.FILE_UPT_CHECK_ONLINE
		//gen ss
		pss_msg, err := comm.GenDispMsg(ss.DISP_MSG_TARGET_FILE_SERVER, ss.DISP_MSG_METHOD_SPEC, ss.DISP_PROTO_TYPE_DISP_COMMON_NOTIFY,
			file_server, pconfig.ProcId, 0, pnotify)
		if err != nil {
			log.Err("%s gen -->file ss msg failed! err:%v uid:%d grp_id:%d url:%s", _func_, err, uid, grp_id, pnotify.StrV)
			return
		}

		//send
		SendToDisp(pconfig, 0, pss_msg)
		return
	}

	//to online-logic
	log.Debug("%s will transport to logic:%d uid:%d", _func_, puser_info.login_serv, uid)
	//gen ss
	pss_msg, err := comm.GenDispMsg(ss.DISP_MSG_TARGET_LOGIC_SERVER, ss.DISP_MSG_METHOD_SPEC, ss.DISP_PROTO_TYPE_DISP_COMMON_NOTIFY,
		puser_info.login_serv, file_server, 0, pnotify)
	if err != nil {
		log.Err("%s gen -->logic ss msg failed! err:%v uid:%d grp_id:%d url:%s", _func_, err, uid, grp_id, pnotify.StrV)
		return
	}

	//send
	SendToDisp(pconfig, 0, pss_msg)
}

func RecvApplyGroupAuditNotify(pconfig *Config, pnotify *ss.MsgCommonNotify) {
	var _func_ = "<RecvApplyGroupAuditNotify>"
	log := pconfig.Comm.Log
	uid := pnotify.Uid
	grp_id := pnotify.GrpId

	//user
	puser_info := GetUserInfo(pconfig , uid)
	if puser_info == nil {
		log.Debug("%s offline! uid:%d grp_id:%d" , _func_ , uid , grp_id)
		return
	}

	//to logic
	SendSpecCommNotify(pconfig , ss.DISP_MSG_TARGET_LOGIC_SERVER , puser_info.login_serv , pnotify)
}
