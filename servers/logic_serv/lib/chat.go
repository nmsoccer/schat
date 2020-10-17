package lib

import (
	"errors"
	"fmt"
	"schat/proto/ss"
	"schat/servers/comm"
	"time"
)

const (
	FETCH_APPLY_GROUP_COUNT       = 20
	FETCH_AUDIT_GROUP_COUNT       = 20
	FETCH_CHAT_COUNT              = 40
	FETCH_OFFLINE_INFO_COUNT      = 40
	FETCH_GROUND_GRP_COUNT        = 40
	PERIOD_FETCH_APPLY_GROUP_INTV = 10000 //10s
)

func InitUserChatInfo(pconfig *Config, pinfo *ss.UserChatInfo, uid int64) {
	var _func_ = "<InitUserChatInfo>"
	log := pconfig.Comm.Log

	//init map
	if pinfo.AllGroup <= 0 {
		pinfo.AllGroups = make(map[int64]*ss.UserChatGroup) //no effect
		pinfo.AllGroup = 0
	}

	if pinfo.MasterGroup <= 0 {
		pinfo.MasterGroups = make(map[int64]bool) //no effect
		pinfo.MasterGroup = 0
	}

	log.Info("%s finish! uid:%d", _func_, uid)
}

//check user in group
//@return (bool , error)
func UserInGroup(pconfig *Config, uid int64, grp_id int64) (bool, error) {
	var err_msg string

	//user info
	puser_info := GetUserInfo(pconfig, uid)
	if puser_info == nil {
		err_msg = fmt.Sprintf("user offline! uid:%d grp_id:%d", uid, grp_id)
		return false, errors.New(err_msg)
	}

	//chat info
	pchat_info := puser_info.user_info.BlobInfo.ChatInfo
	if pchat_info.AllGroup <= 0 || pchat_info.AllGroups == nil {
		return false, nil
	}

	//check
	_, ok := pchat_info.AllGroups[grp_id]
	if !ok {
		return false, nil
	}

	return true, nil
}

func DelUserGroup(pconfig *Config, uid int64, grp_id int64, kick_ts int64) error {
	var err_msg string
	//user info
	puser_info := GetUserInfo(pconfig, uid)
	if puser_info == nil {
		err_msg = fmt.Sprintf("uid:%d offline", uid)
		return errors.New(err_msg)
	}

	//chat info
	pchat_info := puser_info.user_info.BlobInfo.ChatInfo
	if pchat_info.AllGroup == 0 || pchat_info.AllGroups == nil {
		err_msg = fmt.Sprintf("uid:%d enter no group", uid)
		return errors.New(err_msg)
	}

	//all group
	pgrp, ok := pchat_info.AllGroups[grp_id]
	if !ok {
		err_msg = fmt.Sprintf("uid:%d not in group! grp_id:%d", uid, grp_id)
		return errors.New(err_msg)
	}

	//check time stamp
	if kick_ts > 0 {
		if pgrp.EnterTs >= kick_ts { //kick is older than enter will miss it
			err_msg = fmt.Sprintf("kick_ts:%d is older than enter_ts:%d! grp_id:%d", kick_ts, pgrp.EnterTs, grp_id)
			return errors.New(err_msg)
		}
	}

	//delete
	delete(pchat_info.AllGroups, grp_id)
	pchat_info.AllGroup--
	if pchat_info.AllGroup < 0 {
		pchat_info.AllGroup = 0
	}

	//master no error
	for {
		if pchat_info.MasterGroup == 0 || pchat_info.MasterGroups == nil {
			break
		}

		_, ok = pchat_info.MasterGroups[grp_id]
		if !ok {
			break
		}

		delete(pchat_info.MasterGroups, grp_id)
		pchat_info.MasterGroup--
		if pchat_info.MasterGroup < 0 {
			pchat_info.MasterGroup = 0
		}
		break
	}

	return nil
}

func RecvCreateGroupReq(pconfig *Config, preq *ss.MsgCreateGrpReq, msg []byte) {
	var _func_ = "<RecvCreateGroupReq>"
	log := pconfig.Comm.Log
	uid := preq.Uid

	log.Info("%s uid:%d grp_name:%s", _func_, uid, preq.GrpName)
	//get user
	puser := GetUserInfo(pconfig, uid)
	if puser == nil {
		log.Err("%s fail! uid:%d offline!", _func_, uid)
		return
	}

	//check chat
	pchat_info := puser.user_info.BlobInfo.GetChatInfo()
	if pchat_info == nil {
		puser.user_info.BlobInfo.ChatInfo = new(ss.UserChatInfo)
		InitUserChatInfo(pconfig, puser.user_info.BlobInfo.ChatInfo, uid)
		pchat_info = puser.user_info.BlobInfo.GetChatInfo()
	}

	//check max
	if pchat_info.MasterGroup >= 100 {
		log.Err("%s fail! master group to max! uid:%d", _func_, uid)
		SendCreateGroupErrRsp(pconfig, uid, ss.CREATE_GROUP_RESULT_CREATE_RET_MAX_NUM)
		return
	}

	//check duplicate
	for grp_id, _ := range pchat_info.MasterGroups {
		if pgrp, ok := pchat_info.AllGroups[grp_id]; ok {
			if pgrp.GroupName == preq.GrpName {
				log.Err("%s fail! created group %s already in use! uid:%d", _func_, preq.GrpName, uid)
				SendCreateGroupErrRsp(pconfig, uid, ss.CREATE_GROUP_RESULT_CREATE_RET_DUPLICATE)
				return
			}
		}
	}

	//To DB
	SendToDb(pconfig, msg)
}

func RecvCreateGroupRsp(pconfig *Config, prsp *ss.MsgCreateGrpRsp, msg []byte) {
	var _func_ = "<RecvCreateGroupRsp>"
	log := pconfig.Comm.Log
	uid := prsp.Uid

	log.Info("%s uid:%d ret:%d grp_name:%s", _func_, uid, prsp.Ret, prsp.GrpName)
	//get user
	puser := GetUserInfo(pconfig, uid)
	if puser == nil {
		log.Err("%s fail! uid:%d offline! ret:%d group:%s", _func_, uid, prsp.Ret, prsp.GrpName)
		return
	}

	//if failed direct to connect
	if prsp.Ret != ss.CREATE_GROUP_RESULT_CREATE_RET_SUCCESS {
		SendToConnect(pconfig, msg)
		return
	}

	//success
	pchat_info := puser.user_info.BlobInfo.GetChatInfo()
	if pchat_info == nil {
		puser.user_info.BlobInfo.ChatInfo = new(ss.UserChatInfo)
		InitUserChatInfo(pconfig, puser.user_info.BlobInfo.ChatInfo, uid)
		pchat_info = puser.user_info.BlobInfo.GetChatInfo()
	}

	//set chat info
	pchat_info.MasterGroup += 1
	if pchat_info.MasterGroups == nil {
		pchat_info.MasterGroups = make(map[int64]bool)
	}
	pchat_info.MasterGroups[prsp.GrpId] = true

	pchat_info.AllGroup += 1
	if pchat_info.AllGroups == nil {
		pchat_info.AllGroups = make(map[int64]*ss.UserChatGroup)
	}
	pchat_info.AllGroups[prsp.GrpId] = new(ss.UserChatGroup)
	pgrp := pchat_info.AllGroups[prsp.GrpId]
	pgrp.GroupId = prsp.GrpId
	pgrp.GroupName = prsp.GrpName
	pgrp.LastReadId = 0
	pgrp.EnterTs = time.Now().Unix()
	log.Info("%s create group success! uid:%d grp:%s grp_id:%d master_group:%d all_group:%d", _func_,
		uid, prsp.GrpName, prsp.GrpId, pchat_info.MasterGroup, pchat_info.AllGroup)
	//to Connect
	SendToConnect(pconfig, msg)

}

func SendCreateGroupErrRsp(pconfig *Config, uid int64, ret ss.CREATE_GROUP_RESULT) {
	var _func_ = "<SendCreateGroupRsp>"
	log := pconfig.Comm.Log

	//ss_msg
	var ss_msg ss.SSMsg
	pCreateGroupRsp := new(ss.MsgCreateGrpRsp)
	pCreateGroupRsp.Ret = ret
	pCreateGroupRsp.Uid = uid

	//gen
	err := comm.FillSSPkg(&ss_msg, ss.SS_PROTO_TYPE_CREATE_GROUP_RSP, pCreateGroupRsp)
	if err != nil {
		log.Err("%s gen ss! uid:%d result:%d", _func_, uid, ret)
		return
	}

	//send to connect
	SendToConnect(pconfig, &ss_msg)
}

//Apply
func RecvApplyGroupReq(pconfig *Config, preq *ss.MsgApplyGroupReq) {
	var _func_ = "<RecvApplyGroupReq>"
	log := pconfig.Comm.Log
	uid := preq.ApplyUid

	log.Info("%s uid:%d grp_id:%d", _func_, uid, preq.GroupId)
	//get user
	puser := GetUserInfo(pconfig, uid)
	if puser == nil {
		log.Err("%s fail! uid:%d offline!", _func_, uid)
		return
	}

	//check exist
	chat_info := puser.user_info.BlobInfo.GetChatInfo()
	if chat_info.AllGroup > 0 && chat_info.AllGroups != nil {
		if grp_info, exist := chat_info.AllGroups[preq.GroupId]; exist {
			log.Err("%s fail! already in group! uid:%d grp_id:%d", _func_, uid, preq.GroupId)
			prsp := new(ss.MsgApplyGroupRsp)
			prsp.ApplyUid = preq.ApplyUid
			prsp.Result = ss.APPLY_GROUP_RESULT_APPLY_GRP_EXIST
			prsp.GroupName = grp_info.GroupName
			prsp.GroupId = preq.GroupId
			SendApplyGroupRsp(pconfig, prsp)
			return
		}
	}

	//fill info
	preq.ApplyName = puser.user_info.BasicInfo.Name

	//Pack Disp Pkg
	//apply disp msg == MsgApplyGroupReq
	pss_msg, err := comm.GenDispMsg(ss.DISP_MSG_TARGET_CHAT_SERVER, ss.DISP_MSG_METHOD_HASH, ss.DISP_PROTO_TYPE_DISP_APPLY_GROUP_REQ,
		0, pconfig.ProcId, preq.GroupId, preq)
	if err != nil {
		log.Err("%s GenDispMsg Failed! uid:%d err:%v", _func_, uid, err)
		return
	}

	//pack
	enc_data, err := ss.Pack(pss_msg)
	if err != nil {
		log.Err("%s pack failed! uid:%d err:%v", _func_, uid, err)
		return
	}

	//To Disp
	SendToDisp(pconfig, 0, enc_data)
}

func RecvApplyGroupRsp(pconfig *Config, prsp *ss.MsgApplyGroupRsp) {
	var _func_ = "<RecvApplyGroupRsp>"
	log := pconfig.Comm.Log
	uid := prsp.ApplyUid

	//check uid
	puser_info := GetUserInfo(pconfig, uid)
	if puser_info == nil {
		log.Err("%s uid:%d offline!", _func_, uid)
		return
	}

	log.Debug("%s uid:%d result:%d grp_id:%d grp_name:%s", _func_, uid, prsp.Result, prsp.GroupId, prsp.GroupName)
	//back to connect
	SendApplyGroupRsp(pconfig, prsp)
}

func SendApplyGroupRsp(pconfig *Config, prsp *ss.MsgApplyGroupRsp) {
	var _func_ = "<SendApplyGroupRsp>"
	log := pconfig.Comm.Log

	//ss_msg
	var ss_msg ss.SSMsg

	//gen ss
	err := comm.FillSSPkg(&ss_msg, ss.SS_PROTO_TYPE_APPLY_GROUP_RSP, prsp)
	if err != nil {
		log.Err("%s gen ssfail! uid:%d result:%d", _func_, prsp.ApplyUid, prsp.Result)
		return
	}

	//send to connect
	SendToConnect(pconfig, &ss_msg)
}

func RecvApplyGroupNotify(pconfig *Config, pnotify *ss.MsgApplyGroupNotify) {
	var _func = "<RecvApplyGroupNotify>"
	log := pconfig.Comm.Log

	log.Debug("%s uid:%d grp_id:%d grp_name:%s", _func, pnotify.MasterUid, pnotify.GroupId, pnotify.GroupName)
	//check online
	puser := GetUserInfo(pconfig, pnotify.MasterUid)
	if puser == nil {
		log.Info("%s uid:%d offline!", _func, pnotify.MasterUid)
		return
	}

	//to client
	var ss_msg ss.SSMsg
	err := comm.FillSSPkg(&ss_msg, ss.SS_PROTO_TYPE_APPLY_GROUP_NOTIFY, pnotify)
	if err != nil {
		log.Err("%s fill ss failed! uid:%d", _func, pnotify.MasterUid)
		return
	}

	SendToConnect(pconfig, &ss_msg)
}

//fetch apply group
func SendFetchApplyGroupReq(pconfig *Config, uid int64) {
	var _func_ = "<SendFetchApplyGroupReq>"
	log := pconfig.Comm.Log

	//check online
	puser := GetUserInfo(pconfig, uid)
	if puser == nil {
		log.Err("%s fail! uid:%d offline!", _func_, uid)
		return
	}

	//ss_msg
	var ss_msg ss.SSMsg
	pfetch := new(ss.MsgFetchApplyGrpReq)
	pfetch.Uid = uid
	pfetch.FetchCount = FETCH_APPLY_GROUP_COUNT

	err := comm.FillSSPkg(&ss_msg, ss.SS_PROTO_TYPE_FETCH_APPLY_GROUP_REQ, pfetch)
	if err != nil {
		log.Err("%s gen ss fail! err:%v", _func_, err)
		return
	}

	//to db
	SendToDb(pconfig, &ss_msg)
}

//recv fetch apply
func RecvFetchApplyGroupRsp(pconfig *Config, pfetch *ss.MsgFetchApplyGrpRsp) {
	var _func_ = "<RecvFetchApplyGroupRsp>"
	log := pconfig.Comm.Log
	uid := pfetch.Uid

	log.Debug("%s uid:%d fetch_count:%d complete:%d", _func_, uid, pfetch.FetchCount, pfetch.Complete)
	//get user info
	puser := GetUserInfo(pconfig, pfetch.Uid)
	if puser == nil {
		log.Info("%s user offline! uid:%d", _func_, pfetch.Uid)
		return
	}

	//condition
	if pfetch.Complete == 1 {
		log.Debug("%s fetch complete! uid:%d", _func_, uid)
		puser.fetch_apply_complete = true
		return
	}
	if pfetch.FetchCount == 0 || len(pfetch.NotifyList) == 0 {
		log.Debug("%s fetch empty! uid:%d", _func_, uid)
		return
	}

	//pack and send
	var ss_msg ss.SSMsg
	var err error
	for i := 0; i < int(pfetch.FetchCount); i++ {
		err = comm.FillSSPkg(&ss_msg, ss.SS_PROTO_TYPE_APPLY_GROUP_NOTIFY, pfetch.NotifyList[i])
		if err != nil {
			log.Err("%s gen ss failed! err:%v uid:%d grp_id:%d grp_name:%s", _func_, err, uid, pfetch.NotifyList[i].GroupId,
				pfetch.NotifyList[i].GroupName)
			continue
		}
		SendToConnect(pconfig, &ss_msg)
	}

}

func RecvApplyGroupAudit(pconfig *Config, preq *ss.MsgApplyGroupAudit, msg []byte) {
	var _func_ = "<RecvApplyGroupAudit>"
	log := pconfig.Comm.Log
	uid := preq.Uid

	//check online
	puser := GetUserInfo(pconfig, uid)
	if puser == nil {
		log.Err("%s uid:%d offline!", _func_, uid)
		return
	}
	log.Debug("%s uid:%d grp_id:%d apply_uid:%d result:%d", _func_, uid, preq.GroupId, preq.ApplyUid, preq.Result)

	//check master
	pchat_info := puser.user_info.BlobInfo.ChatInfo
	if pchat_info.MasterGroup == 0 || pchat_info.MasterGroups == nil {
		log.Err("%s uid:%d owns none group!", _func_, uid)
		return
	}
	_, exist := pchat_info.MasterGroups[preq.GroupId]
	if !exist {
		log.Err("%s do not own this group! uid:%d grp_id:%d", _func_, uid, preq.GroupId)
		return
	}

	SendToDb(pconfig, msg)
}

func RecvApplyGroupAuditNotify(pconfig *Config, pnotify *ss.MsgCommonNotify) {
	var _func_ = "<RecvApplyGroupAuditNotify>"
	log := pconfig.Comm.Log
	uid := pnotify.Uid

	//diff logic
	if int(pnotify.IntV) != pconfig.ProcId {
		log.Debug("%s will trans to logic:%d uid:%d", _func_, pnotify.IntV, uid)
		pss_msg, err := comm.GenDispMsg(ss.DISP_MSG_TARGET_LOGIC_SERVER, ss.DISP_MSG_METHOD_SPEC, ss.DISP_PROTO_TYPE_DISP_COMMON_NOTIFY,
			int(pnotify.IntV), pconfig.ProcId, 0, pnotify)
		if err != nil {
			log.Err("%s gen dispmsg failed! err:%v uid:%d", _func_, err, uid)
			return
		}
		SendToDisp(pconfig, 0, pss_msg)
		return
	}

	//same logic
	puser := GetUserInfo(pconfig, uid)
	if puser == nil {
		log.Err("%s offline! uid:%d", _func_, uid)
		return
	}
	log.Debug("%s get auidit notify! uid:%d", _func_, uid)

	//fetch
	SendFetchAuditGroupReq(pconfig, uid)
}

func SendFetchAuditGroupReq(pconfig *Config, uid int64) {
	var _func_ = "<SendFetchAuditGroupReq>"
	log := pconfig.Comm.Log

	//get user info
	puser := GetUserInfo(pconfig, uid)
	if puser == nil {
		log.Err("%s user offline! uid:%d", _func_, uid)
		return
	}

	//ss
	var ss_msg ss.SSMsg
	preq := new(ss.MsgFetchAuditGrpReq)
	preq.Uid = uid
	preq.FetchCount = FETCH_AUDIT_GROUP_COUNT

	err := comm.FillSSPkg(&ss_msg, ss.SS_PROTO_TYPE_FETCH_AUDIT_GROUP_REQ, preq)
	if err != nil {
		log.Err("%s gen ss failed! err:%v uid:%d", _func_, err, uid)
		return
	}

	//send
	SendToDb(pconfig, &ss_msg)
}

func RecvFetchAuditGroupRsp(pconfig *Config, prsp *ss.MsgFetchAuditGrpRsp) {
	var _func_ = "<RecvFetchAuditGroupRsp>"
	log := pconfig.Comm.Log
	uid := prsp.Uid

	//get user info
	puser := GetUserInfo(pconfig, uid)
	if puser == nil {
		log.Err("%s user offline! uid:%d", _func_, uid)
		return
	}

	//complete
	if prsp.Complete == 1 {
		log.Debug("%s complete! uid:%d", _func_, uid)
	}

	if prsp.FetchCount == 0 {
		log.Debug("%s fetch nothing! uid:%d", _func_, uid)
		return
	}

	//entering group
	pchat_info := puser.user_info.BlobInfo.ChatInfo
	//init
	if pchat_info.EnteringGroup == nil {
		pchat_info.EnteringGroup = make(map[int64]bool)
	}
	if pchat_info.AllGroups == nil {
		pchat_info.AllGroups = make(map[int64]*ss.UserChatGroup)
	}

	//iter
	var ss_msg ss.SSMsg
	var exist bool
	var paudit *ss.MsgApplyGroupAudit
	for i := 0; i < int(prsp.FetchCount); i++ {
		//get pointer
		paudit = prsp.AuditList[i]
		if paudit == nil {
			log.Err("%s nil audit! i:%d fetch:%d", _func_, i, prsp.FetchCount)
			continue
		}

		//allow
		if paudit.Result == ss.APPLY_GROUP_RESULT_APPLY_GRP_ALLOW {
			_, exist = pchat_info.AllGroups[paudit.GroupId]
			if exist {
				log.Debug("%s already in group! uid:%d grp_id:%d", _func_, prsp.Uid, paudit.GroupId)
				continue
			}

			if pchat_info.EnteringGroup[paudit.GroupId] {
				log.Info("%s send reentering grp req! uid:%d grp_id:%d", _func_, prsp.Uid, paudit.GroupId)
				SendEnterGroupReq(pconfig, uid, paudit.GroupId)
				continue
			}

			//enter group
			pchat_info.EnteringGroup[paudit.GroupId] = true //entering group
			SendEnterGroupReq(pconfig, uid, paudit.GroupId)
		}

		//send to client apply result
		papply := new(ss.MsgApplyGroupRsp)
		papply.Result = prsp.AuditList[i].Result
		papply.GroupName = prsp.AuditList[i].GroupName
		papply.GroupId = prsp.AuditList[i].GroupId
		papply.ApplyUid = prsp.Uid

		err := comm.FillSSPkg(&ss_msg, ss.SS_PROTO_TYPE_APPLY_GROUP_RSP, papply)
		if err != nil {
			log.Err("%s gen apply group rsp failed! err:%v uid:%d group_id:%d", _func_, err, prsp.Uid, paudit.GroupId)
			continue
		}

		//back
		SendToConnect(pconfig, &ss_msg)
	}

}

func SendEnterGroupReq(pconfig *Config, uid int64, grp_id int64) {
	var _func_ = "<SendEnterGroupReq>"
	log := pconfig.Comm.Log

	//dis ss msg
	preq := new(ss.MsgEnterGroupReq)
	preq.Uid = uid
	preq.GrpId = grp_id
	pss_msg, err := comm.GenDispMsg(ss.DISP_MSG_TARGET_CHAT_SERVER, ss.DISP_MSG_METHOD_HASH, ss.DISP_PROTO_TYPE_DISP_ENTER_GROUP_REQ, 0,
		pconfig.ProcId, grp_id, preq)
	if err != nil {
		log.Err("%s failed! err:%v uid:%d grp_id:%d", _func_, err, uid, grp_id)
		return
	}

	//send
	SendToDisp(pconfig, grp_id, pss_msg)
}

func CheckEnteringGroup(pconfig *Config, uid int64) {
	var _func_ = "<CheckEnteringGroup>"
	log := pconfig.Comm.Log

	//user info
	puser_info := GetUserInfo(pconfig, uid)
	if puser_info == nil {
		log.Err("%s offline! uid:%d", _func_, uid)
		return
	}

	//chat info
	pchat_info := puser_info.user_info.BlobInfo.ChatInfo
	if pchat_info.EnteringGroup != nil {
		for grp_id, _ := range pchat_info.EnteringGroup {
			log.Debug("%s will enter %d uid:%d", _func_, grp_id, uid)
			SendEnterGroupReq(pconfig, uid, grp_id)
		}
	}

	log.Debug("%s finish! uid:%d", _func_, uid)
}

func RecvEnterGroupRsp(pconfig *Config, prsp *ss.MsgEnterGroupRsp) {
	var _func_ = "<RecvEnterGroupRsp>"
	log := pconfig.Comm.Log
	uid := prsp.Uid
	grp_id := prsp.GrpId

	log.Debug("%s uid:%d grp_id:%d grp_name:%s ret:%d", _func_, uid, grp_id, prsp.GrpName, prsp.Result)
	//get user
	puser_info := GetUserInfo(pconfig, uid)
	if puser_info == nil {
		log.Err("%s offline! uid:%d grp_id:%d", _func_, uid, grp_id)
		return
	}

	pchat_info := puser_info.user_info.BlobInfo.ChatInfo
	if prsp.Result == 0 { //enter success
		//Add
		if pchat_info.AllGroups == nil {
			pchat_info.AllGroups = make(map[int64]*ss.UserChatGroup)
		}

		//check exist
		_, ok := pchat_info.AllGroups[grp_id]
		if ok {
			log.Info("%s already in group:%d uid:%d", _func_, grp_id, uid)
		} else {
			pchat_info.AllGroups[grp_id] = new(ss.UserChatGroup)
			pchat_info.AllGroups[grp_id].GroupId = grp_id
			pchat_info.AllGroups[grp_id].GroupName = prsp.GrpName
			pchat_info.AllGroups[grp_id].LastReadId = prsp.MsgCount //set read id
			pchat_info.AllGroups[grp_id].EnterTs = time.Now().Unix()
			pchat_info.AllGroup++
			log.Info("%s enter group success! uid:%d grp_id:%d grp_name:%s", _func_, uid, grp_id, prsp.GrpName)

			//Fetch Group Chat
			//SendFetchChatReq(pconfig , uid , grp_id)
		}
	} else { //no group
		log.Info("%s group not exist anymore! uid:%d grp_id:%d", _func_, uid, grp_id)
	}

	//del enter flag
	if pchat_info.EnteringGroup != nil {
		delete(pchat_info.EnteringGroup, grp_id)
	}

}

func TickFetchApplyGroup(arg interface{}) {
	pconfig, ok := arg.(*Config)
	if !ok {
		return
	}

	//fetch
	if pconfig.Users.curr_online > 0 {
		for uid, info := range pconfig.Users.user_map {
			if info != nil && info.fetch_apply_complete == false {
				SendFetchApplyGroupReq(pconfig, uid)
			}
		}
	}

}

func RecvSendChatReq(pconfig *Config, preq *ss.MsgSendChatReq) {
	var _func_ = "<RecvSendChatReq>"
	log := pconfig.Comm.Log

	log.Debug("%s uid:%d temp_id:%d type:%d content:%s grp_id:%d", _func_, preq.Uid, preq.TempId, preq.ChatMsg.ChatType,
		preq.ChatMsg.Content, preq.ChatMsg.GroupId)
	//get user
	puser_info := GetUserInfo(pconfig, preq.Uid)
	if puser_info == nil {
		log.Err("%s offline! uid:%d", _func_, preq.Uid)
		return
	}

	//check in group
	pchat_info := puser_info.user_info.BlobInfo.ChatInfo
	if pchat_info.AllGroup <= 0 || pchat_info.AllGroups == nil {
		log.Err("%s in none group! uid:%d", _func_, preq.Uid)
		return
	}
	_, exist := pchat_info.AllGroups[preq.ChatMsg.GroupId]
	if !exist {
		log.Err("%s not in group! uid:%d group:%d", _func_, preq.Uid, preq.ChatMsg.GroupId)
		return
	}

	//fill info
	preq.ChatMsg.SenderUid = preq.Uid
	preq.ChatMsg.Sender = puser_info.user_info.BasicInfo.Name

	//disp ss
	pss_msg, err := comm.GenDispMsg(ss.DISP_MSG_TARGET_CHAT_SERVER, ss.DISP_MSG_METHOD_HASH, ss.DISP_PROTO_TYPE_DISP_SEND_CHAT_REQ,
		0, pconfig.ProcId, preq.ChatMsg.GroupId, preq)
	if err != nil {
		log.Err("%s gen disp ss fail! uid:%d err:%v", _func_, preq.Uid, err)
		return
	}

	//send
	SendToDisp(pconfig, 0, pss_msg)
}

func RecvSendChatRsp(pconfig *Config, prsp *ss.MsgSendChatRsp) {
	var _func_ = "<RecvSendChatRsp>"
	log := pconfig.Comm.Log

	log.Debug("%s uid:%d temp_id:%d type:%d content:%s grp_id:%d result:%d", _func_, prsp.Uid, prsp.TempId, prsp.ChatMsg.ChatType,
		prsp.ChatMsg.Content, prsp.ChatMsg.GroupId, prsp.Result)
	//get user
	puser_info := GetUserInfo(pconfig, prsp.Uid)
	if puser_info == nil {
		log.Err("%s offline! uid:%d", _func_, prsp.Uid)
		return
	}

	//gen ss
	var ss_msg ss.SSMsg
	err := comm.FillSSPkg(&ss_msg, ss.SS_PROTO_TYPE_SEND_CHAT_RSP, prsp)
	if err != nil {
		log.Err("%s gen ss fail! err:%v uid:%d", _func_, err, prsp.Uid)
		return
	}

	//to connect
	SendToConnect(pconfig, &ss_msg)
}

func RecvFetchChatReq(pconfig *Config, preq *ss.MsgFetchChatReq) {
	var _func_ = "<RecvFetchChatReq>"
	log := pconfig.Comm.Log

	switch preq.FetchType {
	case ss.SS_COMMON_TYPE_COMM_TYPE_HISTORY:
		DoFetchChatHistroy(pconfig, preq)
	default:
		log.Err("%s type not support! type:%d uid:%d grp_id:%d", _func_, preq.FetchType, preq.Uid, preq.GrpId)
		return
	}

}

func DoFetchChatHistroy(pconfig *Config, preq *ss.MsgFetchChatReq) {
	var _func_ = "<DoFetchChatHistroy>"
	log := pconfig.Comm.Log
	uid := preq.Uid
	grp_id := preq.GrpId

	//user_info
	puser_info := GetUserInfo(pconfig, uid)
	if puser_info == nil {
		log.Err("%s user offline! uid:%d grp_id:%d", _func_, uid, grp_id)
		return
	}

	//chat_info
	pchat_info := puser_info.user_info.BlobInfo.ChatInfo
	if pchat_info.AllGroup <= 0 || pchat_info.AllGroups == nil {
		log.Err("%s user enter no group! uid:%d grp_id:%d", _func_, uid, grp_id)
		return
	}

	//grp_info
	pgrp, ok := pchat_info.AllGroups[grp_id]
	if !ok {
		log.Err("%s user not enter group! uid:%d grp_id:%d", _func_, uid, grp_id)
		return
	}

	//fill info
	if preq.LatestMsgId == 0 {
		preq.LatestMsgId = pgrp.LastReadId
	}
	preq.LatestMsgId -= FETCH_CHAT_COUNT //fetch before 40items
	preq.LatestMsgId -= 1
	if preq.LatestMsgId < 0 {
		preq.LatestMsgId = 0
	}
	preq.FetchCount = FETCH_CHAT_COUNT

	//ss
	var ss_msg ss.SSMsg
	err := comm.FillSSPkg(&ss_msg, ss.SS_PROTO_TYPE_FETCH_CHAT_REQ, preq)
	if err != nil {
		log.Err("%s gen ss failed! err:%v uid:%d grp_id:%d", _func_, err, uid, grp_id)
		return
	}

	log.Debug("%s finish! uid:%d grp_id:%d now_id:%d", _func_, uid, grp_id, preq.LatestMsgId)
	SendToDb(pconfig, &ss_msg)
}

//grp_id==0 fetch all group
//only for normal fetch
func SendFetchChatReq(pconfig *Config, uid int64, spec_grp int64) {
	var _func_ = "<SendFetchChatReq>"
	log := pconfig.Comm.Log

	//get user info
	puser := GetUserInfo(pconfig, uid)
	if puser == nil {
		log.Err("%s user offline! uid:%d", _func_, uid)
		return
	}

	//get chat_info
	pchat_info := puser.user_info.BlobInfo.ChatInfo
	if pchat_info.AllGroup <= 0 || pchat_info.AllGroups == nil {
		return
	}

	//ss
	var ss_msg ss.SSMsg
	var grp_id int64
	var grp_info *ss.UserChatGroup
	var ok bool

	preq := new(ss.MsgFetchChatReq)
	preq.Uid = uid
	preq.FetchCount = FETCH_CHAT_COUNT
	preq.FetchType = ss.SS_COMMON_TYPE_COMM_TYPE_NORMAL
	if spec_grp == 0 { //all group
		for grp_id, grp_info = range pchat_info.AllGroups {
			preq.GrpId = grp_id
			preq.LatestMsgId = grp_info.LastReadId

			err := comm.FillSSPkg(&ss_msg, ss.SS_PROTO_TYPE_FETCH_CHAT_REQ, preq)
			if err != nil {
				log.Err("%s gen ss failed! err:%v uid:%d grp_id:%d", _func_, err, uid, grp_id)
				continue
			}

			//send
			SendToDb(pconfig, &ss_msg)
		}
	} else { //spec group
		grp_info, ok = pchat_info.AllGroups[spec_grp]
		if !ok {
			log.Err("%s failed! spec_grp:%d not in! uid:%d", _func_, spec_grp, uid)
			return
		}
		preq.GrpId = spec_grp
		preq.LatestMsgId = grp_info.LastReadId

		err := comm.FillSSPkg(&ss_msg, ss.SS_PROTO_TYPE_FETCH_CHAT_REQ, preq)
		if err != nil {
			log.Err("%s gen ss2 failed! err:%v uid:%d spec_grp:%d", _func_, err, uid, spec_grp)
			return
		}

		//send
		SendToDb(pconfig, &ss_msg)
	}
}

func RecvFetchChatRsp(pconfig *Config, prsp *ss.MsgFetchChatRsp) {
	var _func_ = "<RecvFetchChatRsp>"
	log := pconfig.Comm.Log
	uid := prsp.Uid
	grp_id := prsp.GrpId

	log.Debug("%s uid:%d grp_id:%d result:%d", _func_, uid, grp_id, prsp.Result)
	//check result
	if prsp.Result == ss.SS_COMMON_RESULT_FAILED {
		return
	}

	//get user info
	puser := GetUserInfo(pconfig, uid)
	if puser == nil {
		log.Err("%s user offline! uid:%d", _func_, uid)
		return
	}

	//get chat_info
	pchat_info := puser.user_info.BlobInfo.ChatInfo
	if pchat_info.AllGroup <= 0 || pchat_info.AllGroups == nil {
		log.Err("%s has nothing group! uid:%d", _func_, uid)
		return
	}

	//check group
	grp_info, ok := pchat_info.AllGroups[grp_id]
	if !ok {
		log.Err("%s not enter group! uid:%d grp_id:%d", _func_, uid, grp_id)
		return
	}

	//Not Exist
	if prsp.Result == ss.SS_COMMON_RESULT_NOEXIST {
		log.Info("%s group is deleted! uid:%d grp_id:%d", _func_, uid, grp_id)
		//master group may not happen here
		if pchat_info.MasterGroups != nil && pchat_info.MasterGroups[grp_id] == true {
			delete(pchat_info.MasterGroups, grp_id)
			pchat_info.MasterGroup--
		}

		delete(pchat_info.AllGroups, grp_id)
		pchat_info.AllGroup--

		if pchat_info.MasterGroup < 0 {
			pchat_info.MasterGroup = 0
		}
		if pchat_info.AllGroup < 0 {
			pchat_info.AllGroup = 0
		}
		return
	}

	//step.Sync Chat
	//no data
	if prsp.FetchCount <= 0 || len(prsp.ChatList) <= 0 {
		return
	}

	//sync chat
	var old_readed = grp_info.LastReadId
	var ss_msg ss.SSMsg
	var idx int

	psync := new(ss.MsgSyncChatList)
	psync.Uid = prsp.Uid
	psync.GrpId = prsp.GrpId
	psync.SyncType = prsp.FetchType
	psync.ChatList = make([]*ss.ChatMsg, prsp.FetchCount)
	for i := 0; i < int(prsp.FetchCount); i++ {
		if prsp.ChatList[i] == nil {
			log.Err("%s chat msg nil! uid:%d grp_id:%d", _func_, uid, grp_id)
			continue
		}

		//fill
		psync.ChatList[idx] = prsp.ChatList[i]
		idx++
		log.Debug("%s msg_id:%d content:%s", _func_, prsp.ChatList[i].MsgId, prsp.ChatList[i].Content)

		//update
		if prsp.ChatList[i].MsgId > old_readed {
			old_readed = prsp.ChatList[i].MsgId
		}
	}
	psync.Count = int32(idx)

	//pack
	err := comm.FillSSPkg(&ss_msg, ss.SS_PROTO_TYPE_SYNC_CHAT_LIST, psync)
	if err != nil {
		log.Err("%s gen sync ss failed! uid:%d grp_id:%d err:%v", _func_, uid, grp_id, err)
		return
	}

	//send
	ok = SendToConnect(pconfig, &ss_msg)
	if !ok {
		return
	}

	//check type
	//history not active fetch-chain
	if prsp.FetchType == ss.SS_COMMON_TYPE_COMM_TYPE_HISTORY {
		return
	}

	grp_info.LastReadId = old_readed
	//step. ReFetch
	if prsp.Complete == 1 {
		return
	}

	SendFetchChatReq(pconfig, uid, grp_id)
}

func RecvNewMsgNotify(pconfig *Config, preq *ss.MsgCommonNotify) {
	var _func_ = "<RecvNewMsgNotify>"
	log := pconfig.Comm.Log
	uid := preq.Uid
	grp_id := preq.GrpId

	log.Debug("%s grp_id:%d uid:%d", _func_, grp_id, uid)
	//Get User
	puser_info := GetUserInfo(pconfig, uid)
	if puser_info == nil {
		log.Err("%s offline! uid:%d", _func_, uid)
		return
	}

	//Check chat
	pchat_info := puser_info.user_info.BlobInfo.ChatInfo
	if pchat_info.AllGroups == nil {
		log.Err("%s all group nil! uid:%d", _func_, uid)
		return
	}

	pgrp_info, ok := pchat_info.AllGroups[grp_id]
	if !ok {
		log.Err("%s not in group! uid:%d grp_id:%d", _func_, uid, grp_id)
		return
	}

	//Check Msg
	//proper next msg direct to client
	if preq.ChatMsg != nil && preq.ChatMsg.MsgId == pgrp_info.LastReadId+1 {
		log.Debug("%s just the next msg! to client! uid:%d latest_read:%d msg_id:%d grp_id:%d", _func_, uid, pgrp_info.LastReadId,
			preq.ChatMsg.MsgId, grp_id)

		psync := new(ss.MsgSyncChatList)
		psync.Uid = uid
		psync.GrpId = grp_id
		psync.ChatList = make([]*ss.ChatMsg, 1)
		psync.ChatList[0] = preq.ChatMsg
		psync.Count = 1

		//gen ss
		var ss_msg ss.SSMsg
		err := comm.FillSSPkg(&ss_msg, ss.SS_PROTO_TYPE_SYNC_CHAT_LIST, psync)
		if err != nil {
			log.Err("%s gen sync ss failed! err:%v uid:%d grp_id:%d", _func_, err, uid, grp_id)
		} else {
			SendToConnect(pconfig, &ss_msg)
			pgrp_info.LastReadId++
			return
		}

	}

	//Fetch From Db
	if !puser_info.fetch_apply_complete { //in fetching chain
		return
	}

	SendFetchChatReq(pconfig, uid, grp_id)
}

func RecvExitGroupReq(pconfig *Config, preq *ss.MsgExitGroupReq) {
	var _func_ = "<RecvExitGroupReq>"
	log := pconfig.Comm.Log
	uid := preq.Uid
	grp_id := preq.GrpId

	//check user
	puser_info := GetUserInfo(pconfig, uid)
	if puser_info == nil {
		log.Err("%s offline! uid:%d", _func_, uid)
		return
	}

	//check chat
	pchat_info := puser_info.user_info.BlobInfo.ChatInfo
	if pchat_info.AllGroup <= 0 || pchat_info.AllGroups == nil {
		log.Err("%s enter no group! uid:%d", _func_, uid)
		return
	}

	//if master group will del group
	info, ok := pchat_info.AllGroups[grp_id]
	if !ok {
		log.Err("%s not enter group! uid:%d grp_id:%d", _func_, uid, grp_id)
		return
	}

	preq.GrpName = info.GroupName
	ok = pchat_info.MasterGroups[grp_id]
	if ok {
		log.Info("%s master group! will del group! grp_id:%d uid:%d", _func_, grp_id, uid)
		preq.DelGroup = 1
	}

	//gen ss
	var ss_msg ss.SSMsg
	err := comm.FillSSPkg(&ss_msg, ss.SS_PROTO_TYPE_EXIT_GROUP_REQ, preq)
	if err != nil {
		log.Err("%s gen ss failed! err:%v uid:%d grp_id:%d", _func_, err, uid, grp_id)
	}

	//to db
	SendToDb(pconfig, &ss_msg)
}

func RecvExitGroupRsp(pconfig *Config, prsp *ss.MsgExitGroupRsp) {
	var _func_ = "<RecvExitGroupRsp>"
	log := pconfig.Comm.Log
	uid := prsp.Uid
	grp_id := prsp.GrpId

	log.Debug("%s grp_id:%d uid:%d result:%d del_group:%d", _func_, grp_id, uid, prsp.Result, prsp.DelGroup)
	//check user
	puser_info := GetUserInfo(pconfig, uid)
	if puser_info == nil {
		log.Err("%s user offline! uid:%d grp_id:%d result:%d", _func_, uid, grp_id, prsp.Result)
		return
	}

	//handle
	switch prsp.Result {
	case ss.SS_COMMON_RESULT_FAILED:
		log.Err("%s fail! grp_id:%d uid:%d result:%d del_group:%d", _func_, grp_id, uid, prsp.Result, prsp.DelGroup)
	case ss.SS_COMMON_RESULT_SUCCESS, ss.SS_COMMON_RESULT_NOEXIST:
		err := DelUserGroup(pconfig, uid, grp_id, 0)
		if err != nil {
			log.Err("%s del group fail! err:%v uid:%d grp_id:%d del_group:%d", _func_, err, uid, grp_id, prsp.DelGroup)
			return
		}
	default:
		log.Err("%s illegal result:%d uid:%d grp_id:%d", _func_, prsp.Result, uid, grp_id)
		return
	}

	//to client
	var ss_msg ss.SSMsg
	err := comm.FillSSPkg(&ss_msg, ss.SS_PROTO_TYPE_EXIT_GROUP_RSP, prsp)
	if err != nil {
		log.Err("%s gen resp failed! err:%v uid:%d grp_id:%d", _func_, err, uid, grp_id)
	} else {
		SendToConnect(pconfig, &ss_msg)
	}

	//noitfy
	if prsp.Result != ss.SS_COMMON_RESULT_SUCCESS {
		return
	}

	// gen notify
	pnotify := new(ss.MsgCommonNotify)
	if prsp.DelGroup == 0 {
		pnotify.NotifyType = ss.COMMON_NOTIFY_TYPE_NOTIFY_EXIT_GROUP
	} else {
		pnotify.NotifyType = ss.COMMON_NOTIFY_TYPE_NOTIFY_DEL_GROUP
	}
	pnotify.GrpId = grp_id
	pnotify.Uid = uid
	pnotify.StrV = prsp.GrpName

	pss_msg, err := comm.GenDispMsg(ss.DISP_MSG_TARGET_CHAT_SERVER, ss.DISP_MSG_METHOD_HASH, ss.DISP_PROTO_TYPE_DISP_COMMON_NOTIFY,
		0, pconfig.ProcId, grp_id, pnotify)
	if err != nil {
		log.Err("%s gen notify ss failed! err:%v grp_id:%d type:%d", _func_, err, grp_id, pnotify.NotifyType)
		return
	}

	//to chat
	SendToDisp(pconfig, 0, pss_msg)
}

func RecvDelGroupNotify(pconfig *Config, pnotify *ss.MsgCommonNotify) {
	var _func_ = "<RecvDelGroupNotify>"
	log := pconfig.Comm.Log
	uid := pnotify.Uid
	grp_id := pnotify.GrpId

	//puser
	puser_info := GetUserInfo(pconfig, uid)
	if puser_info == nil {
		log.Err("%s user offline! uid:%d grp_id:%d", _func_, uid, grp_id)
		return
	}

	//chat_info
	pchat_info := puser_info.user_info.BlobInfo.ChatInfo
	_, ok := pchat_info.AllGroups[grp_id]
	if !ok {
		log.Err("%s not enter group:%d uid:%d", _func_, grp_id, uid)
		return
	}

	//del
	delete(pchat_info.AllGroups, grp_id)
	pchat_info.AllGroup--
	if pchat_info.AllGroup < 0 {
		pchat_info.AllGroup = 0
	}
	log.Info("%s del group:%d success! uid:%d grp_name:%s", _func_, grp_id, uid, pnotify.StrV)

	//to client
	prsp := new(ss.MsgExitGroupRsp)
	prsp.GrpId = grp_id
	prsp.Uid = uid
	prsp.GrpName = pnotify.StrV
	prsp.DelGroup = 1
	prsp.Result = ss.SS_COMMON_RESULT_SUCCESS

	var ss_msg ss.SSMsg
	err := comm.FillSSPkg(&ss_msg, ss.SS_PROTO_TYPE_EXIT_GROUP_RSP, prsp)
	if err != nil {
		log.Err("%s gen ss failed! err:%v uid:%d grp_id:%d", _func_, err, uid, grp_id)
		return
	}

	SendToConnect(pconfig, &ss_msg)
}

func RecvKickGroupReq(pconfig *Config, preq *ss.MsgKickGroupReq) {
	var _func_ = "<RecvKickGroupReq>"
	log := pconfig.Comm.Log
	uid := preq.Uid
	grp_id := preq.GrpId

	//user info
	puser_info := GetUserInfo(pconfig, uid)
	if puser_info == nil {
		log.Err("%s user offline! uid:%d", _func_, uid)
		return
	}

	//chat info
	pchat_info := puser_info.user_info.BlobInfo.ChatInfo
	if pchat_info.MasterGroup == 0 || pchat_info.MasterGroups == nil {
		log.Err("%s owns nothing group! uid:%d", _func_, uid)
		return
	}

	_, ok := pchat_info.MasterGroups[grp_id]
	if !ok {
		log.Err("%s not master! uid:%d grp_id:%d", _func_, uid, grp_id)
		return
	}

	pgrp, ok := pchat_info.AllGroups[grp_id]
	if !ok {
		log.Err("%s has nothing group info! uid:%d grp_id:%d", _func_, uid, grp_id)
		return
	}

	//fill info
	preq.GrpName = pgrp.GroupName

	//ss
	var ss_msg ss.SSMsg
	err := comm.FillSSPkg(&ss_msg, ss.SS_PROTO_TYPE_KICK_GROUP_REQ, preq)
	if err != nil {
		log.Err("%s gen ss failed! err:%v uid:%d grp_id:%d kick:%d", _func_, err, uid, grp_id, preq.KickUid)
		return
	}

	//to db
	SendToDb(pconfig, &ss_msg)
}

//master who kick member
func RecvKickGroupRsp(pconfig *Config, prsp *ss.MsgKickGroupRsp) {
	var _func_ = "<RecvKickGroupRsp>"
	log := pconfig.Comm.Log
	uid := prsp.Uid
	grp_id := prsp.GrpId

	log.Info("%s uid:%d grp_id:%d result:%d kick:%d", _func_, uid, grp_id, prsp.Result, prsp.KickUid)
	if prsp.Result != ss.SS_COMMON_RESULT_SUCCESS {
		return
	}

	//notify to group
	pnotify := new(ss.MsgCommonNotify)
	pnotify.NotifyType = ss.COMMON_NOTIFY_TYPE_NOTIFY_KICK_GROUP
	pnotify.GrpId = prsp.GrpId
	pnotify.Uid = prsp.KickUid
	pnotify.StrV = prsp.GrpName

	pss_msg, err := comm.GenDispMsg(ss.DISP_MSG_TARGET_CHAT_SERVER, ss.DISP_MSG_METHOD_HASH, ss.DISP_PROTO_TYPE_DISP_COMMON_NOTIFY,
		0, pconfig.ProcId, grp_id, pnotify)
	if err != nil {
		log.Err("%s gen notify ss failed! err:%v uid:%d grp_id:%d kick:%d", _func_, err, uid, grp_id, prsp.KickUid)
		return
	}

	//to chat serv
	SendToDisp(pconfig, 0, pss_msg)
}

//kick by master
func RecvKickGroupNotify(pconfig *Config, pnotify *ss.MsgCommonNotify) {
	var _func_ = "<RecvKickGroupNotify>"
	log := pconfig.Comm.Log
	uid := pnotify.Uid
	grp_id := pnotify.GrpId

	//user info
	puser_info := GetUserInfo(pconfig, uid)
	if puser_info == nil {
		log.Err("%s user offline! uid:%d grp_id:%d", _func_, uid, grp_id)
		return
	}

	//exit group
	err := DelUserGroup(pconfig, uid, grp_id, pnotify.IntV)
	if err != nil {
		log.Err("%s del group failed! uid:%d grp_id:%d err:%v", _func_, uid, grp_id, err)
		return
	}

	//notify client
	prsp := new(ss.MsgExitGroupRsp)
	prsp.Uid = uid
	prsp.GrpId = grp_id
	prsp.GrpName = pnotify.StrV
	prsp.ByKick = 1
	prsp.Result = ss.SS_COMMON_RESULT_SUCCESS

	//ss
	var ss_msg ss.SSMsg
	err = comm.FillSSPkg(&ss_msg, ss.SS_PROTO_TYPE_EXIT_GROUP_RSP, prsp)
	if err != nil {
		log.Err("%s gen ss failed! err:%v uid:%d grp_id:%d", _func_, err, uid, grp_id)
		return
	}

	//to connect
	SendToConnect(pconfig, &ss_msg)
}

func RecvChgMemNotify(pconfig *Config, pnotify *ss.MsgCommonNotify) {
	var _func_ = "<RecvChgMemNotify>"
	log := pconfig.Comm.Log
	uid := pnotify.Uid
	grp_id := pnotify.GrpId

	log.Debug("%s uid:%d grp_id:%d grp_name:%s type:%d chg_uid:%d", _func_, uid, grp_id, pnotify.StrV, pnotify.NotifyType,
		pnotify.IntV)
	//ss
	var ss_msg ss.SSMsg
	err := comm.FillSSPkg(&ss_msg, ss.SS_PROTO_TYPE_COMMON_NOTIFY, pnotify)
	if err != nil {
		log.Err("%s gen ss fail! uid:%d grp_id:%d grp_name:%s type:%d chg_uid:%d err:%v", _func_, uid, grp_id, pnotify.StrV, pnotify.NotifyType,
			pnotify.IntV, err)
		return
	}

	//to connect
	SendToConnect(pconfig, &ss_msg)
}

func SendFetchOfflineInfoReq(pconfig *Config, uid int64) {
	var _func_ = "<SendFetchOfflineInfoReq>"
	log := pconfig.Comm.Log

	//get user info
	puser := GetUserInfo(pconfig, uid)
	if puser == nil {
		log.Err("%s user offline! uid:%d", _func_, uid)
		return
	}

	//ss
	var ss_msg ss.SSMsg
	preq := new(ss.MsgFetchOfflineInfoReq)
	preq.Uid = uid
	preq.FetchCount = FETCH_OFFLINE_INFO_COUNT

	err := comm.FillSSPkg(&ss_msg, ss.SS_PROTO_TYPE_FETCH_OFFLINE_INFO_REQ, preq)
	if err != nil {
		log.Err("%s gen ss failed! err:%v uid:%d", _func_, err, uid)
		return
	}

	//send
	SendToDb(pconfig, &ss_msg)
}

func RecvFetchOfflineInfoRsp(pconfig *Config, prsp *ss.MsgFetchOfflineInfoRsp) {
	var _func_ = "<RecvFetchOfflineInfoRsp>"
	log := pconfig.Comm.Log
	uid := prsp.Uid

	//user info
	puser_info := GetUserInfo(pconfig, uid)
	if puser_info == nil {
		log.Err("%s offline! uid:%d", _func_, uid)
		return
	}

	//parse offline info
	log.Debug("%s will parse offline_info! count:%d uid:%d complete:%d result:%d", _func_, prsp.FetchCount, uid, prsp.Complete,
		prsp.Result)

	if prsp.Result != ss.SS_COMMON_RESULT_SUCCESS {
		return
	}

	//iter
	for i := 0; i < int(prsp.FetchCount); i++ {
		ParseOfflineInfo(pconfig, uid, prsp.InfoList[i])
	}

	//continue?
	if prsp.Complete == 1 {
		return
	}

	SendFetchApplyGroupReq(pconfig, uid)
}

func RecvQueryGroupReq(pconfig *Config, preq *ss.MsgQueryGroupReq) {
	var _func_ = "<RecvQueryGroupReq>"
	log := pconfig.Comm.Log
	uid := preq.Uid
	grp_id := preq.GrpId

	//check in group
	ok, _ := UserInGroup(pconfig, uid, grp_id)
	if !ok {
		log.Err("%s not in group! uid:%d grp_id:%d", _func_, uid, grp_id)
		return
	}

	//to chat serv
	pss_msg, err := comm.GenDispMsg(ss.DISP_MSG_TARGET_CHAT_SERVER, ss.DISP_MSG_METHOD_HASH, ss.DISP_PROTO_TYPE_DISP_QUERY_GROUP_REQ,
		0, pconfig.ProcId, grp_id, preq)
	if err != nil {
		log.Err("%s gen ss failed! uid:%d grp_id:%d err:%v", _func_, uid, grp_id, err)
		return
	}

	SendToDisp(pconfig, 0, pss_msg)
}

func RecvSyncGroupInfo(pconfig *Config, pinfo *ss.MsgSyncGroupInfo) {
	var _func_ = "<RecvQueryGroupReq>"
	log := pconfig.Comm.Log
	uid := pinfo.Uid
	grp_id := pinfo.GrpId

	//check online
	puser_info := GetUserInfo(pconfig, uid)
	if puser_info == nil {
		log.Err("%s user offline! uid:%d grp_id:%d", _func_, uid, grp_id)
		return
	}

	//ss
	var ss_msg ss.SSMsg
	err := comm.FillSSPkg(&ss_msg, ss.SS_PROTO_TYPE_SYNC_GROUP_INFO, pinfo)
	if err != nil {
		log.Err("%s gen ss failed! err:%v uid:%d grp_id:%d", _func_, err, uid, grp_id)
		return
	}

	//to conn
	SendToConnect(pconfig, &ss_msg)
}
