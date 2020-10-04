package lib

import (
	"schat/proto/ss"
	"schat/servers/comm"
)

func RecvNewMsgNotify(pconfig *Config , preq *ss.MsgCommonNotify) {
	var _func_ = "<RecvNewMsgNotify>"
	log := pconfig.Comm.Log

	log.Debug("%s grp_id:%d master:%d" , _func_ , preq.GrpId , preq.Uid)
	//check online
	if pconfig.world_online.world_online <= 0 || pconfig.world_online.user_map==nil {
		log.Debug("%s no online!" , _func_)
		return
	}


	//notify
	pnotify := new(ss.MsgCommonNotify)
	pnotify.GrpId = preq.GrpId
	pnotify.NotifyType = ss.COMMON_NOTIFY_TYPE_NOTIFY_NEW_MSG
	pnotify.ChatMsg = preq.ChatMsg

	//to master
    pusr_info := GetUserInfo(pconfig , preq.Uid)
    if pusr_info != nil {
    	pnotify.Uid = preq.Uid
		SendSpecCommNotify(pconfig , ss.DISP_MSG_TARGET_LOGIC_SERVER , pusr_info.login_serv , pnotify)
	}

	//to members
	if preq.Members == nil || len(preq.Members)==0 {
		log.Info("%s req no member contain!" , _func_)
		return
	}
	for uid , _ := range(preq.Members) {
		pusr_info = GetUserInfo(pconfig , uid)
		if pusr_info != nil {
			pnotify.Uid = uid
			//SendNewMsgNotify(pconfig , pnotify , pusr_info.login_serv)
			SendSpecCommNotify(pconfig , ss.DISP_MSG_TARGET_LOGIC_SERVER , pusr_info.login_serv , pnotify)
		}
	}

}

func SendNewMsgNotify(pconfig *Config , pnotify *ss.MsgCommonNotify , target_serv int) {
	var _func_ = "<SendNewMsgNotify>"
	log := pconfig.Comm.Log

	//GEN SS
	pss_msg , err := comm.GenDispMsg(ss.DISP_MSG_TARGET_LOGIC_SERVER , ss.DISP_MSG_METHOD_SPEC , ss.DISP_PROTO_TYPE_DISP_COMMON_NOTIFY ,
		target_serv , pconfig.ProcId , 0 , pnotify)
	if err != nil {
		log.Err("%s gen ss failed! err:%v uid:%d grp_id:%d" , _func_ , err , pnotify.Uid , pnotify.GrpId)
		return
	}

	//Send
	SendToDisp(pconfig , 0 , pss_msg)
}
