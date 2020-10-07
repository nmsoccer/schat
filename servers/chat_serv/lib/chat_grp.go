package lib

import (
	"schat/proto/ss"
	"schat/servers/comm"
	"time"
)

const (
	PERIOD_CHECK_SAVE_CHAT_GROUP = 10000 //10 sec
	SAVE_GROUP_TIME_SPAN = 40 //40sec expire
)

type OnLineGroup struct {
	load_ts int64
	last_save int64 //last save ts
	db_group_info *ss.GroupInfo
}

type OnLineGroupList struct {
	online_count int
	group_map map[int64]*OnLineGroup
}

func GetGroupInfo(pconfig *Config , gr_id int64) *OnLineGroup {
	if pconfig.GroupList.online_count == 0 {
		return nil
	}
	if pconfig.GroupList.group_map == nil {
		return nil
	}

	pinfo , ok := pconfig.GroupList.group_map[gr_id]
	if !ok {
		return nil
	}
	return pinfo
}


func RecvApplyGroupReq(pconfig *Config , preq *ss.MsgApplyGroupReq , pdisp *ss.MsgDisp) {
	var _func_ = "<RecvApplyGroupReq>"
	log := pconfig.Comm.Log

	log.Debug("%s grp_id:%d apply_uid:%d from:%d" , _func_ , preq.GroupId , preq.ApplyUid , pdisp.FromServer)
	//check cache


	//to db
	preq.Occupy = int64(pdisp.FromServer)
    var ss_msg ss.SSMsg

	//gen ss
	err := comm.FillSSPkg(&ss_msg , ss.SS_PROTO_TYPE_APPLY_GROUP_REQ , preq)
	if err != nil {
		log.Err("%s gen ss fail! apply_uid:%d err:%v" , _func_ , preq.ApplyUid , err)
		return
	}

	//to db
	SendToDb(pconfig , &ss_msg)
}

func RecvApplyGroupRsp(pconfig *Config , prsp *ss.MsgApplyGroupRsp) {
	var _func_ = "<RecvApplyGroupRsp>"
	log := pconfig.Comm.Log

	log.Debug("%s result:%d apply_uid:%d grp_id:%d from_logic:%d" , _func_ , prsp.Result , prsp.ApplyUid , prsp.GroupId ,
		prsp.Occupy)
	//disp to logic
	pss_msg , err := comm.GenDispMsg(ss.DISP_MSG_TARGET_NON_SERVER , ss.DISP_MSG_METHOD_SPEC , ss.DISP_PROTO_TYPE_DISP_APPLY_GROUP_RSP ,
		int(prsp.Occupy) , pconfig.ProcId , 0 , prsp)
	if err != nil {
		log.Err("%s gen disp mag fail! err:%v" , _func_ , err)
		return
	}

	//pack
	enc_data , err := ss.Pack(pss_msg)
	if err != nil {
		log.Err("%s pack failed! err:%v" , _func_ , err)
		return
	}

    //to Disp by rand
    SendToDisp(pconfig , 0 , enc_data)
}


func RecvApplyGroupNotify(pconfig *Config , pnotify *ss.MsgApplyGroupNotify) {
	var _func_ = "<RecvApplyGroupNotify>"
	log := pconfig.Comm.Log

	log.Debug("%s try to notify:%d group_uid:%d logic:%d" , _func_ , pnotify.MasterUid , pnotify.GroupId , pnotify.Occupy[0])
    //pack
	pss_msg , err := comm.GenDispMsg(ss.DISP_MSG_TARGET_NON_SERVER , ss.DISP_MSG_METHOD_SPEC , ss.DISP_PROTO_TYPE_DISP_APPLY_GROUP_NOTIFY ,
		int(pnotify.Occupy[0]) , pconfig.ProcId , 0 , pnotify)
	if err != nil {
		log.Err("%s gen ss failed! uid:%d grp_id:%d err:%v" , _func_ , pnotify.MasterUid , pnotify.GroupId , err)
		return
	}

	//disp
	SendToDisp(pconfig , pnotify.MasterUid, pss_msg)
}

func RecvEnterGroupReq(pconfig *Config , preq *ss.MsgEnterGroupReq , from_logic int32) {
	var _func_ = "<RecvEnterGroupReq>"
	log := pconfig.Comm.Log

	log.Debug("%s uid:%d grp_id:%d" , _func_ , preq.GrpId , preq.Uid)
	/*
	//group online
	pgrp_info := GetGroupInfo(pconfig , preq.GrpId)
	if pgrp_info != nil {
		log.Debug("%s group online! uid:%d grp_id:%d" , _func_ , preq.Uid , preq.GrpId)
		if pgrp_info.db_group_info.Members == nil {
			pgrp_info.db_group_info.Members = make(map[int64]int32)
		}
		pgrp_info.db_group_info.Members[preq.Uid] = from_logic
		SendEnterGroupRsp(pconfig , preq.Uid , preq.GrpId , pgrp_info.db_group_info.GroupName , int(from_logic) , 0)
		return
	}*/

	//to db
	var ss_msg ss.SSMsg
	preq.Occupy = int64(from_logic)
	err := comm.FillSSPkg(&ss_msg , ss.SS_PROTO_TYPE_ENTER_GROUP_REQ , preq)
	if err != nil {
		log.Err("%s gen ss fail! uid:%d grp_id:%d err:%v" , _func_ , preq.Uid , preq.GrpId, err)
		return
	}

	SendToDb(pconfig , &ss_msg)
}

func RecvEnterGroupRsp(pconfig *Config  , prsp *ss.MsgEnterGroupRsp) {
	var _func_ = "<RecvEnterGroupRsp>"
	log := pconfig.Comm.Log
	uid := prsp.Uid
	grp_id := prsp.GrpId

    log.Debug("%s uid:%d grp_id:%d ret:%d" , _func_ , uid , grp_id , prsp.Result)
	SendEnterGroupRsp(pconfig , uid , grp_id , prsp.GrpName , prsp.MsgCount , int(prsp.Occupy) , prsp.Result)

	//Update Online Group
	//fail
	if prsp.Result != 0 {
		return
	}
	//success
	pgroup := GetGroupInfo(pconfig , grp_id)
	if pgroup == nil {
		return
	}
	_ , ok := pgroup.db_group_info.Members[uid]
	if ok {
		return
	}
	pgroup.db_group_info.Members[uid] = 1
    pgroup.db_group_info.MemCount++
	log.Debug("%s append uid:%d grp_id:%d mem_count:%d" , _func_ , uid , grp_id , pgroup.db_group_info.MemCount)
}


func SendEnterGroupRsp(pconfig *Config , uid int64 , grp_id int64 , grp_name string , msg_count int64 , target_serv int , result int32) {
	var _func_ = "<SendEnterGroupRsp>"
	log := pconfig.Comm.Log

	//ss
	prsp := new(ss.MsgEnterGroupRsp)
	prsp.Uid = uid
	prsp.GrpId = grp_id
	prsp.GrpName = grp_name
	prsp.MsgCount = msg_count
	prsp.Result = result
	pss_msg , err := comm.GenDispMsg(ss.DISP_MSG_TARGET_LOGIC_SERVER , ss.DISP_MSG_METHOD_SPEC , ss.DISP_PROTO_TYPE_DISP_ENTER_GROUP_RSP ,
		target_serv , pconfig.ProcId , 0 , prsp)
	if err != nil {
		log.Err("%s gen disp ss fail! err:%v uid:%d" , _func_ , err , uid)
		return
	}

	//to disp
	SendToDisp(pconfig , 0 , pss_msg)
}

func RecvLoadGroupRsp(pconfig *Config , prsp *ss.MsgLoadGroupRsp) {
	var _func_ = "<RecvLoadGroupRsp>"
	log := pconfig.Comm.Log

	if prsp.LoadResult == ss.SS_COMMON_RESULT_FAILED {
		log.Err("%s load group failed! uid:%d group_id:%d" , _func_ , prsp.Uid , prsp.GrpId)
		return
	}
	if prsp.LoadResult == ss.SS_COMMON_RESULT_NOEXIST {
		LoadGroupNoExist(pconfig , prsp)
		return
	}

	//LoadResult == SUCCESS
	if prsp.GrpInfo == nil {
		log.Err("%s success but group_info nil! group_id:%d" , _func_ , prsp.GrpId)
		return
	}

	//Check Exist
	pgroup_info := GetGroupInfo(pconfig , prsp.GrpId)
	if pgroup_info == nil { //Now Load
		pgroup_list := pconfig.GroupList
		if pgroup_list.group_map == nil {
			pgroup_list.group_map = make(map[int64]*OnLineGroup)
			pgroup_list.online_count = 0
		}

		pgroup_info = new(OnLineGroup)
		pgroup_info.load_ts = time.Now().Unix()
		pgroup_info.db_group_info = prsp.GrpInfo
		pgroup_list.group_map[prsp.GrpId] = pgroup_info
		pgroup_list.online_count++
		log.Info("%s success! grp_id:%d online_group:%d group_info:%v", _func_, prsp.GrpId, pgroup_list.online_count,
			*pgroup_info.db_group_info)
	} else { //Already Loaded
		log.Info("%s group is loaded! group_id:%d" , _func_ , prsp.GrpId)
	}


	//AfterLoad Will Improve
	AfterLoadGroupSuccess(pconfig , prsp , pgroup_info.db_group_info)
	return
}


func AfterLoadGroupSuccess(pconfig *Config , prsp *ss.MsgLoadGroupRsp , pgrp_info *ss.GroupInfo) {
	var _func_ = "<RecvLoadGroupRsp>"
	log := pconfig.Comm.Log

	//check reason
	switch prsp.Reason {
	case ss.LOAD_GROUP_REASON_LOAD_GRP_SEND_CHAT:
		log.Debug("%s will send chat! uid:%d grp_id:%d" , _func_ , prsp.Uid , prsp.GrpId)
		if prsp.ChatMsg == nil {
			log.Err("%s fail! Reason:%d but no chat_msg found! uid:%d" , _func_ , prsp.Reason , prsp.Uid)
			break
		}
		var ss_msg ss.SSMsg
		preq := new(ss.MsgSendChatReq)
		preq.Uid = prsp.Uid
		preq.TempId = prsp.TempId
		preq.ChatMsg = prsp.ChatMsg
		preq.Occupy = prsp.Occoupy
		preq.ChatMsg.MsgId = pgrp_info.LatestMsgId + 1 //this may not be precise

		err := comm.FillSSPkg(&ss_msg , ss.SS_PROTO_TYPE_SEND_CHAT_REQ , preq)
		if err != nil {
			log.Err("%s gen send_chat_ss fail! uid:%d grp_id:%d err:%v" , _func_ , preq.Uid , prsp.GrpId , err)
			return
		}
		SendToDb(pconfig , &ss_msg)
	default:
		//nothing to do
	}

}

func LoadGroupNoExist(pconfig *Config , prsp *ss.MsgLoadGroupRsp) {
	var _func_ = "<LoadGroupNoExist>"
	log := pconfig.Comm.Log

	log.Info("%s uid:%d reason:%d grp_id:%d" , _func_ , prsp.Uid , prsp.GrpId , prsp.Reason)
	switch prsp.Reason {
	case ss.LOAD_GROUP_REASON_LOAD_GRP_SEND_CHAT:
		if prsp.Occoupy <= 0 {
			log.Err("%s from_serv empty! will not send back! uid:%d" , _func_ , prsp.Uid)
			break
		}
		pchat_rsp := new(ss.MsgSendChatRsp)
		pchat_rsp.Uid = prsp.Uid
		pchat_rsp.Result = ss.SEND_CHAT_RESULT_SEND_CHAT_NONE_GROUP
		pchat_rsp.TempId = prsp.TempId
		pchat_rsp.ChatMsg = prsp.ChatMsg

		pss_msg , err := comm.GenDispMsg(ss.DISP_MSG_TARGET_LOGIC_SERVER , ss.DISP_MSG_METHOD_SPEC , ss.DISP_PROTO_TYPE_DISP_SEND_CHAT_RSP ,
			int(prsp.Occoupy) , pconfig.ProcId , 0 , pchat_rsp)
		if err!= nil {
			log.Err("%s gen send_chat_rsp ss failed! err:%v uid:%d" , _func_ , err , prsp.Uid)
			break
		}
		SendToDisp(pconfig , 0 , pss_msg)
	default:
		//nothing to do
	}
}

func SaveChatGroup(pconfig *Config , grp_id int64 , reason ss.SS_COMMON_REASON) {
	var _func_ = "<SaveChatGroup>"
	log := pconfig.Comm.Log

	if reason != ss.SS_COMMON_REASON_REASON_TICK {
		log.Debug("%s grp_id:%d reason:%d", _func_, grp_id, reason)
	}
	//group_info
	pgroup_Info := GetGroupInfo(pconfig , grp_id)
	if pgroup_Info == nil {
		log.Err("%s fail! group offline! grp_id:%d reason:%d" , _func_ , grp_id , reason)
		return
	}

	//ss
	var ss_msg ss.SSMsg
	preq := new(ss.MsgSaveGroupReq)
	preq.GrpId = grp_id
	preq.MsgCount = pgroup_Info.db_group_info.LatestMsgId
	preq.Reason = reason
	preq.MemCount = pgroup_Info.db_group_info.MemCount
	if reason == ss.SS_COMMON_REASON_REASON_EXIT {
		preq.LoadServ = -1
	} else {
		preq.LoadServ = int32(pconfig.ProcId)
	}


	err := comm.FillSSPkg(&ss_msg , ss.SS_PROTO_TYPE_SAVE_GROUP_REQ ,preq)
	if err != nil {
		log.Err("%s gen ss failed! err:%v grp_id:%d" , _func_ , err ,grp_id)
		return
	}


	//to DB
	SendToDb(pconfig , &ss_msg)
}

func RecvSaveChatRsp(pconfig *Config , prsp *ss.MsgSaveGroupRsp) {
	var _func_ = "<RecvSaveChatRsp>"
	log := pconfig.Comm.Log
	grp_id := prsp.GrpId


	//log.Debug("%s grp_id:%d result:%d" , _func_ , grp_id , prsp.Result)
	if prsp.Result == ss.SS_COMMON_RESULT_FAILED {
		log.Err("%s failed! grp_id:%d" , _func_ , grp_id)
		return
	}
	if prsp.Result == ss.SS_COMMON_RESULT_NOEXIST {
		log.Info("%s not exist anymore! grp_id:%d" , _func_ , grp_id)
		//will del group
		return
	}

	//update ts
	pgroup_info := GetGroupInfo(pconfig , grp_id)
	if pgroup_info == nil {
		log.Err("%s offline! grp_id:%d" , _func_ , grp_id)
		return
	}

	//check member changed
	if prsp.MemberChged == 0 {
		return
	}

	//update members
	log.Info("%s will update member map! grp_id:%d" ,_func_ , grp_id)
	if prsp.Members == nil {	//no member any more
		log.Info("%s no member any more! will clear map! grp_id:%d" , _func_ , grp_id)
		pgroup_info.db_group_info.MemCount = 0
		pgroup_info.db_group_info.Members = make(map[int64]int32) //new map
		return
	}

	log.Info("%s will override member info!member count %d-->%d" , _func_ , pgroup_info.db_group_info.MemCount ,
		len(prsp.Members))
	pgroup_info.db_group_info.MemCount = int32(len(prsp.Members))
	pgroup_info.db_group_info.Members = prsp.Members
}


func SaveGroupOnTick(arg interface{}) {
	pconfig  , ok := arg.(*Config)
	if !ok {
		return
	}
	if pconfig.GroupList.online_count<=0 || pconfig.GroupList.group_map==nil {
		return
	}

	curr_ts := time.Now().Unix()
	//save each
	for grp_id , info := range pconfig.GroupList.group_map {
		if info.last_save + SAVE_GROUP_TIME_SPAN > curr_ts {
			continue
		}
		SaveChatGroup(pconfig , grp_id , ss.SS_COMMON_REASON_REASON_TICK)
		info.last_save = curr_ts
	}
}


func SaveGroupOnExit(pconfig *Config) {
    if pconfig.GroupList.online_count<=0 || pconfig.GroupList.group_map==nil {
    	return
	}

	//save each
	for grp_id , _ := range pconfig.GroupList.group_map {
		SaveChatGroup(pconfig , grp_id , ss.SS_COMMON_REASON_REASON_EXIT)
	}
}

func RecvExitGroupNotify(pconfig *Config , pnotify *ss.MsgCommonNotify) {
	var _func_ = "<RecvExitGroupNotify>"
	log := pconfig.Comm.Log
	uid := pnotify.Uid
	grp_id := pnotify.GrpId

	//grp_info
	pgrp_info := GetGroupInfo(pconfig , grp_id)
	if pgrp_info == nil {
		log.Info("%s grp offline! grp_id:%d" , _func_ , grp_id)
		return
	}

	//check member
	_ , ok := pgrp_info.db_group_info.Members[uid]
	if !ok {
		log.Info("%s grp:%d has no member:%d" , _func_ , grp_id , uid)
		return
	}

	//del it
	delete(pgrp_info.db_group_info.Members , uid)
	pgrp_info.db_group_info.MemCount--
	if pgrp_info.db_group_info.MemCount < 0 {
		pgrp_info.db_group_info.MemCount = 0
	}
    log.Info("%s success! mem_count:%d grp_id:%d uid:%d" , _func_ , pgrp_info.db_group_info.MemCount , grp_id , uid)
}


func RecvDelGroupNotify(pconfig *Config , pnotify *ss.MsgCommonNotify) {
	var _func_ = "<RecvExitGroupNotify>"
	log := pconfig.Comm.Log
	uid := pnotify.Uid
	grp_id := pnotify.GrpId

	//grp_info
	pgrp_info := GetGroupInfo(pconfig, grp_id)
	if pgrp_info == nil {
		log.Info("%s grp offline! grp_id:%d", _func_, grp_id)
		return
	}

	//check master
	if pgrp_info.db_group_info.MasterUid != uid {
		log.Err("%s fail! not master! grp_id:%d uid:%d master:%d" , _func_ , grp_id , uid , pgrp_info.db_group_info.MasterUid)
		return
	}

	//save members
	pnotify.Members = pgrp_info.db_group_info.Members

	//del online
	delete(pconfig.GroupList.group_map , grp_id)
	pconfig.GroupList.online_count--
	if pconfig.GroupList.online_count < 0 {
		pconfig.GroupList.online_count = 0
	}
	log.Info("%s done! grp_id:%d master:%d" , _func_ , grp_id , uid)


	//notify online member
	pss_msg , err := comm.GenDispMsg(ss.DISP_MSG_TARGET_ONLINE_SERVER , ss.DISP_MSG_METHOD_RAND , ss.DISP_PROTO_TYPE_DISP_COMMON_NOTIFY , 0 ,
		pconfig.ProcId , 0 , pnotify)
	if err != nil {
		log.Err("%s gen ss failed! err:%v grp_id:%d uid:%d" , _func_ , err , grp_id , uid)
		return
	}
	//to online
	SendToDisp(pconfig , 0 , pss_msg)
}