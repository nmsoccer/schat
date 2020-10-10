package lib

import (
	"schat/proto/ss"
)

func RecvDelGroupNotify(pconfig *Config , preq *ss.MsgCommonNotify) {
	var _func_ = "<RecvDelGroupNotify>"
	log := pconfig.Comm.Log

    log.Debug("%s grp_id:%d master:%d grp_name:%s" , _func_ , preq.GrpId , preq.Uid , preq.StrV)
	//check member
	if preq.Members==nil || len(preq.Members)==0 {
		return
	}

	//to logic
	pnotify := new(ss.MsgCommonNotify)
	pnotify.GrpId = preq.GrpId
	pnotify.NotifyType = preq.NotifyType
	pnotify.StrV = preq.StrV
	var puser_info *UserInfo

	for uid , _ := range preq.Members {
		puser_info = GetUserInfo(pconfig , uid)
		if puser_info != nil {
			pnotify.Uid = uid
			SendSpecCommNotify(pconfig , ss.DISP_MSG_TARGET_LOGIC_SERVER , puser_info.login_serv , pnotify)
		}
	}

}

func RecvKickGroupNotify(pconfig *Config , pnotify *ss.MsgCommonNotify) {
	var _func_ = "<RecvDelGroupNotify>"
	log := pconfig.Comm.Log
	uid := pnotify.Uid

	log.Debug("%s grp_id:%d uid:%d grp_name:%s" , _func_ , pnotify.GrpId , uid , pnotify.StrV)
	//check logic
	puser_info := GetUserInfo(pconfig , uid)
	if puser_info == nil {
		return
	}

	//send
	SendSpecCommNotify(pconfig , ss.DISP_MSG_TARGET_LOGIC_SERVER , puser_info.login_serv , pnotify)
}

func RecvChgMemberNotify(pconfig *Config , preq *ss.MsgCommonNotify) {
	master_uid := preq.Uid
	var uid int64 = 0

	//notify
	pnotify := new(ss.MsgCommonNotify)
	pnotify.GrpId = preq.GrpId
	pnotify.NotifyType = preq.NotifyType
	pnotify.StrV = preq.StrV
	pnotify.IntV = preq.IntV
	var puser_info *UserInfo

	//to master
	puser_info = GetUserInfo(pconfig , master_uid)
	if puser_info != nil {
		pnotify.Uid = master_uid
		SendSpecCommNotify(pconfig , ss.DISP_MSG_TARGET_LOGIC_SERVER , puser_info.login_serv , pnotify)
	}

	//to members
	if preq.Members==nil || len(preq.Members)==0 {
		return
	}

	for uid , _ = range preq.Members {
		puser_info = GetUserInfo(pconfig , uid)
		if puser_info != nil {
			pnotify.Uid = uid
			SendSpecCommNotify(pconfig , ss.DISP_MSG_TARGET_LOGIC_SERVER , puser_info.login_serv , pnotify)
		}
	}

}

