package lib

import (
	"schat/proto/ss"
	"schat/servers/comm"
)

func RecvChgGroupAttrReq(pconfig *Config, preq *ss.MsgChgGroupAttrReq) {
	var _func_ = "<RecvChgGroupAttrReq>"
	log := pconfig.Comm.Log
	uid := preq.Uid
	grp_id := preq.GrpId

	//user info
	puser_info := GetUserInfo(pconfig, uid)
	if puser_info == nil {
		log.Err("%s offline! uid:%d grp_id:%d", _func_, uid, grp_id)
		return
	}

	//chat info
	pchat_info := puser_info.user_info.BlobInfo.ChatInfo
	if pchat_info.MasterGroup == 0 || len(pchat_info.MasterGroups) == 0 {
		log.Err("%s owns nothing group! uid:%d", _func_, uid)
		return
	}

	_, ok := pchat_info.MasterGroups[grp_id]
	if !ok {
		log.Err("%s not own group! uid:%d grp_id:%d", _func_, uid, grp_id)
		return
	}

	//to chat serv
	log.Debug("%s try to update attr:%d uid:%d grp_id:%d", _func_, preq.Attr, uid, grp_id)
	pss_msg, err := comm.GenDispMsg(ss.DISP_MSG_TARGET_CHAT_SERVER, ss.DISP_MSG_METHOD_HASH, ss.DISP_PROTO_TYPE_DISP_CHG_GROUP_ATTR_REQ,
		0, pconfig.ProcId, grp_id, preq)
	if err != nil {
		log.Err("%s gen ss failed! err:%v attr:%d uid:%d grp_id:%d", _func_, err, preq.Attr, uid, grp_id)
		return
	}

	//send
	SendToDisp(pconfig, 0, pss_msg)
}

func RecvChgGroupAttrRsp(pconfig *Config, prsp *ss.MsgChgGroupAttrRsp) {
	var _func_ = "<RecvChgGroupAttrRsp>"
	log := pconfig.Comm.Log
	uid := prsp.Uid
	grp_id := prsp.GrpId

	log.Debug("%s result:%d attr:%d uid:%d grp_id:%d", _func_, prsp.Result, prsp.Attr, uid, grp_id)
	//ss
	var ss_msg ss.SSMsg
	err := comm.FillSSPkg(&ss_msg, ss.SS_PROTO_TYPE_CHG_GROUP_ATTR_RSP, prsp)
	if err != nil {
		log.Err("%s gen ss failed! err:%v uid:%d grp_id:%d", _func_, err, uid, grp_id)
		return
	}

	//to connect
	SendToConnect(pconfig, &ss_msg)
}

func RecvGroupGroundReq(pconfig *Config, preq *ss.MsgGroupGroudReq) {
	var _func_ = "<RecvGroupGroundRsp>"
	log := pconfig.Comm.Log
	uid := preq.Uid

	//user info
	puser_info := GetUserInfo(pconfig, uid)
	if puser_info == nil {
		log.Err("%s offline! uid:%d", _func_, uid)
		return
	}

	//ss
	if preq.StartIndex < 0 {
		puser_info.grp_ground_start = 0 //reset
	}
	preq.StartIndex = puser_info.grp_ground_start
	preq.Count = FETCH_GROUND_GRP_COUNT
	var ss_msg ss.SSMsg

	err := comm.FillSSPkg(&ss_msg, ss.SS_PROTO_TYPE_GROUP_GROUND_REQ, preq)
	if err != nil {
		log.Err("%s gen ss failed! err:%v uid:%d", _func_, err, uid)
		return
	}

	//to db
	SendToDb(pconfig, &ss_msg)
}

func RecvGroupGroundRsp(pconfig *Config, prsp *ss.MsgGroupGroudRsp, msg []byte) {
	var _func_ = "<RecvGroupGroundRsp>"
	log := pconfig.Comm.Log
	uid := prsp.Uid

	log.Debug("%s uid:%d count:%d", _func_, uid, prsp.Count)
	//user info
	puser_info := GetUserInfo(pconfig, uid)
	if puser_info == nil {
		log.Err("%s offline! uid:%d", _func_, uid)
		return
	}
	puser_info.grp_ground_start += prsp.Count
	//puser_info.grp_ground_start++ //next started

	//to connect
	SendToConnect(pconfig, msg)
}
