package lib

import (
	"schat/proto/ss"
	"schat/servers/comm"
)

func RecvSendChatReq(pconfig *Config , preq *ss.MsgSendChatReq , from_logic int) {
	var _func_ = "<RecvSendChatReq>"
	log := pconfig.Comm.Log

	log.Debug("%s uid:%d grp_id:%d content:%s" , _func_ , preq.Uid , preq.ChatMsg.GroupId , preq.ChatMsg.Content)
	if preq.ChatMsg.GroupId <=0 {
		log.Err("%s group_id illegal! uid:%d group_id:%d" , _func_ , preq.Uid , preq.ChatMsg.GroupId)
		return
	}

	var ss_msg ss.SSMsg
    //Check Group
    pgroup := GetGroupInfo(pconfig , preq.ChatMsg.GroupId)
    if pgroup != nil {
		//Online Send to Db Direct
		preq.Occupy = int64(from_logic)
		preq.ChatMsg.MsgId = pgroup.db_group_info.LatestMsgId + 1 //this may not be precise
		err := comm.FillSSPkg(&ss_msg , ss.SS_PROTO_TYPE_SEND_CHAT_REQ , preq)
		if err != nil {
			log.Err("%s gen send ss failed! err:%v uid:%d" , _func_ , err , preq.Uid)
			return
		}
		SendToDb(pconfig , &ss_msg)
    	return
	}

	//Load Group First
	log.Debug("%s will load group first! uid:%d group_id:%d" , _func_ , preq.Uid , preq.ChatMsg.GroupId)
    pload := new(ss.MsgLoadGroupReq)
    pload.Uid = preq.Uid
    pload.TempId = preq.TempId
    pload.Reason = ss.LOAD_GROUP_REASON_LOAD_GRP_SEND_CHAT
    pload.GrpId = preq.ChatMsg.GroupId
    pload.ChatMsg = preq.ChatMsg
    pload.Occoupy = int64(from_logic)

    err := comm.FillSSPkg(&ss_msg , ss.SS_PROTO_TYPE_LOAD_GROUP_REQ , pload)
    if err != nil {
    	log.Err("%s gen load ss failed! err:%v uid:%d" , _func_ , err , preq.Uid)
    	return
	}
	SendToDb(pconfig , &ss_msg)
}

func RecvSendChatRsp(pconfig *Config , prsp *ss.MsgSendChatRsp) {
	var _func_ = "RecvSendChatRsp"
	log := pconfig.Comm.Log
	uid := prsp.Uid
	grp_id := prsp.ChatMsg.GroupId

	log.Debug("%s uid:%d grp_id:%d result:%d" , _func_ , uid , grp_id , prsp.Result)
	//Back To Logic
	pss_msg , err := comm.GenDispMsg(ss.DISP_MSG_TARGET_LOGIC_SERVER , ss.DISP_MSG_METHOD_SPEC , ss.DISP_PROTO_TYPE_DISP_SEND_CHAT_RSP ,
		int(prsp.Occupy) , pconfig.ProcId , 0 , prsp)
	if err!= nil {
		log.Err("%s gen send_chat_rsp ss failed! err:%v uid:%d" , _func_ , err , prsp.Uid)
	} else {
		SendToDisp(pconfig, 0, pss_msg)
	}

	if prsp.Result != ss.SEND_CHAT_RESULT_SEND_CHAT_SUCCESS {
		return
	}
	//Send Success
	//Update Group Info
	pgrp_info := GetGroupInfo(pconfig , grp_id)
	if pgrp_info == nil {	//group should be online
		log.Err("%s grp_info not online! grp_id:%d" , _func_ , grp_id)
		return
	}
		//update real msg id
	if pgrp_info.db_group_info.LatestMsgId < prsp.ChatMsg.MsgId {
		pgrp_info.db_group_info.LatestMsgId = prsp.ChatMsg.MsgId
	}

	//Broadcast
    pnotify := new(ss.MsgCommonNotify)
    pnotify.NotifyType = ss.COMMON_NOTIFY_TYPE_NOTIFY_NEW_MSG
    pnotify.GrpId = grp_id
    pnotify.Uid = pgrp_info.db_group_info.MasterUid //master
    if pgrp_info.db_group_info.MemCount > 0 && pgrp_info.db_group_info.Members!=nil {
		pnotify.Members = pgrp_info.db_group_info.Members
	}
    pnotify.ChatMsg = prsp.ChatMsg //carry latest
    pss_msg , err = comm.GenDispMsg(ss.DISP_MSG_TARGET_ONLINE_SERVER , ss.DISP_MSG_METHOD_RAND , ss.DISP_PROTO_TYPE_DISP_COMMON_NOTIFY ,
    	0 , pconfig.ProcId , 0 , pnotify)
    if err!= nil {
    	log.Err("%s broadcast failed! gen ss fail! err:%v grp_id:%d" , _func_ , err , grp_id)
    	return
	}

	SendToDisp(pconfig , 0 , pss_msg)
}
