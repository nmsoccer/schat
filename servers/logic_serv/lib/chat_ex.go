package lib

import (
	"fmt"
	"schat/proto/ss"
	"schat/servers/comm"
	"time"
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
	//user info
	puser_info := GetUserInfo(pconfig, uid)
	if puser_info == nil {
		log.Err("%s offline! uid:%d grp_id:%d", _func_, uid, grp_id)
		return
	}

	//some set
	switch prsp.Attr {
	case ss.GROUP_ATTR_TYPE_GRP_ATTR_GRP_NAME:
		if prsp.Result != ss.SS_COMMON_RESULT_SUCCESS {
			break
		}
		pgrp_info := GetUserGroup(pconfig , uid , grp_id)
		if pgrp_info == nil {
			log.Err("%s change group name but not in group! uid:%d grp_id:%d" , _func_ , uid , grp_id)
		} else {
			log.Info("%s change group name! uid:%d grp_id:%d grp_name %s-->%s" , _func_ , uid , grp_id , pgrp_info.GroupName , prsp.StrV)
			pgrp_info.GroupName = prsp.StrV
			send_content := fmt.Sprintf("群组改名为[%s]" , pgrp_info.GroupName)
			SendSysChat(pconfig , grp_id , send_content , ss.CHAT_MSG_FLAG_CHAT_FLAG_NORMAL)
		}
	default:
		break
	}


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

func SyncUserGroupSnap(pconfig *Config , uid int64) {
	var _func_ = "<SyncUserGroupSnap>"
	log := pconfig.Comm.Log

	//user info
	puser_info := GetUserInfo(pconfig , uid)
	if puser_info == nil {
		log.Err("%s offline! uid:%d" , _func_ , uid)
		return
	}

	curr_ts := time.Now().Unix()
	if (curr_ts - puser_info.last_sync_own_snap) < 60 {
		log.Err("%s too frequent! uid:%d" , _func_ , uid)
		return
	}


		//chat info
	pchat_info := puser_info.user_info.BlobInfo.ChatInfo
	if pchat_info.AllGroup<=0 || pchat_info.AllGroups==nil {
		log.Info("%s no group enter!" , _func_ , uid)
		return
	}


	//check and send
	pquery := new(ss.MsgBatchQueryGroupSnap)
	pquery.Uid = uid
	pquery.Count = 0
	pquery.TargetList = make([]int64 , 32)
	for grp_id , _ := range pchat_info.AllGroups {
		//per 30 as a package
		if pquery.Count < 30 {
			pquery.TargetList[pquery.Count] = grp_id
			pquery.Count++
			continue
		}

		//full and send
		SendBatchQueryGroupSnap(pconfig , uid , pquery)
		//reset
		pquery.Count = 0
	}
	//last check
	if pquery.Count>0 {
		SendBatchQueryGroupSnap(pconfig , uid , pquery)
	}

	puser_info.last_sync_own_snap = curr_ts
}


func SendBatchQueryGroupSnap(pconfig *Config , uid int64 , pquery *ss.MsgBatchQueryGroupSnap) {
	var _func_ = "<SendBatchQueryGroupSnap>"
	log := pconfig.Comm.Log

	//check
	if pquery.Count <= 0 {
		log.Err("%s empty list! uid:%d" , _func_ , uid)
		return
	}

	//ss_msg
	var ss_msg ss.SSMsg
	err := comm.FillSSPkg(&ss_msg , ss.SS_PROTO_TYPE_BATCH_QUERY_GROUP_SNAP , pquery)
	if err != nil {
		log.Err("%s gen ss failed! err:%v uid:%d" , _func_ , err , uid)
		return
	}

	SendToDb(pconfig , &ss_msg)
}

func RecvUpdateChatReq(pconfig *Config , preq *ss.MsgUpdateChatReq , msg []byte) {
	var _func_ = "<RecvUpdateChatReq>"
	log := pconfig.Comm.Log
	uid := preq.Uid
	grp_id := preq.GrpId

	//check type
	switch preq.UpdateType {
	case ss.UPDATE_CHAT_TYPE_UPT_CHAT_DEL , ss.UPDATE_CHAT_TYPE_UPT_CHAT_CANCEL:
		//nothing
	default:
		log.Err("%s illegal type:%d uid:%d" , _func_ , preq.UpdateType , uid)
		return
	}


	//chat info
	ok , err := UserInGroup(pconfig , uid , grp_id)
	if err != nil {
		log.Err("%s check in group fail! err:%v uid:%d grp_id:%d" , _func_ , err , uid , grp_id)
		return
	}
	if !ok {
		log.Err("%s not in group! uid:%d grp_id:%d" , _func_ , uid , grp_id)
		return
	}

	//to db
	SendToDb(pconfig , msg)
}

func RecvUpdateChatRsp(pconfig *Config , prsp *ss.MsgUpdateChatRsp , msg []byte) {
	var _func_ = "<RecvUpdateChatRsp>"
	log := pconfig.Comm.Log
	uid := prsp.Uid
	grp_id := prsp.GrpId

	log.Info("%s uid:%d type:%d grp_id:%d msg_id:%d result:%d" , _func_ , uid , prsp.UpdateType , grp_id , prsp.MsgId , prsp.Result)
	//online
	puser_info := GetUserInfo(pconfig , uid)
	if puser_info == nil {
		log.Err("%s offline! uid:%d" , _func_ , uid)
		return
	}

	//if success will send sys-canceler chat
	if prsp.Result == ss.SS_COMMON_RESULT_SUCCESS {
		switch prsp.UpdateType {
		case ss.UPDATE_CHAT_TYPE_UPT_CHAT_CANCEL:
			log.Info("%s will send canceler chat! uid:%d grp_id:%d" , _func_ , uid , grp_id)
			send_content := fmt.Sprintf("%d" , prsp.MsgId)
			SendSysChat(pconfig , grp_id , send_content , ss.CHAT_MSG_FLAG_CHAT_FLAG_CANCELLER)

			//del old file
			if prsp.SrcType == ss.CHAT_MSG_TYPE_CHAT_TYPE_IMG && len(prsp.SrcContent)>0 {
				log.Info("%s will del old chat img! uid:%d grp_id:%d del_url:%s" , _func_ , uid , grp_id , prsp.SrcContent)
				SendDelOldFile(pconfig , prsp.SrcContent , uid , grp_id)
			}
		default:
			//nothing
		}
	}

	//to connect
	SendToConnect(pconfig , msg)
}