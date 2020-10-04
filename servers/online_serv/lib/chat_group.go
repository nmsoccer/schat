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


	//to online
	pnotify := new(ss.MsgCommonNotify)
	pnotify.GrpId = preq.GrpId
	pnotify.NotifyType = preq.NotifyType
	pnotify.StrV = preq.StrV
	var puser_info *UserInfo

	for uid , _ := range(preq.Members) {
		puser_info = GetUserInfo(pconfig , uid)
		if puser_info != nil {
			pnotify.Uid = uid
			SendSpecCommNotify(pconfig , ss.DISP_MSG_TARGET_LOGIC_SERVER , puser_info.login_serv , pnotify)
		}
	}

}


