package lib

import (
	"schat/proto/cs"
	"schat/proto/ss"
	"schat/servers/comm"
)

func SendCreateGroupReq(pconfig *Config, uid int64, preq *cs.CSCreateGroupReq) bool {
	var _func_ = "<SendCreateGroupReq>"
	log := pconfig.Comm.Log

	//ss msg
	var ss_msg ss.SSMsg
	pCreateGroupReq := new(ss.MsgCreateGrpReq)
	pCreateGroupReq.Uid = uid
	pCreateGroupReq.GrpName = preq.Name
	pCreateGroupReq.GrpPass = preq.Pass
	pCreateGroupReq.Desc = preq.Desc
	if len(pCreateGroupReq.Desc) <= 0 {
		pCreateGroupReq.Desc = "..."
	}

	//fill
	err := comm.FillSSPkg(&ss_msg, ss.SS_PROTO_TYPE_CREATE_GROUP_REQ, pCreateGroupReq)
	if err != nil {
		log.Err("%s gen ss fail! err:%v uid:%d", _func_, err, uid)
		return false
	}

	//send
	return SendToLogic(pconfig, &ss_msg)
}

func RecvCreateGroupRsp(pconfig *Config, prsp *ss.MsgCreateGrpRsp) {
	var _func_ = "<RecvCreateGroupRsp>"
	log := pconfig.Comm.Log

	log.Info("%s uid:%d result:%v", _func_, prsp.Uid, *prsp)
	//get c_key
	c_key, ok := pconfig.Uid2Ckey[prsp.Uid]
	if !ok {
		log.Err("%s uid offline! uid:%d", _func_, prsp.Uid)
		return
	}

	//to client
	var pmsg *cs.CSCreateGroupRsp
	pv, err := cs.Proto2Msg(cs.CS_PROTO_CREATE_GRP_RSP)
	if err != nil {
		log.Err("%s proto2msg failed! proto:%d err:%v", _func_, cs.CS_PROTO_CREATE_GRP_RSP, err)
		return
	}
	pmsg, ok = pv.(*cs.CSCreateGroupRsp)
	if !ok {
		log.Err("%s proto2msg type illegal!  proto:%d", _func_, cs.CS_PROTO_CREATE_GRP_RSP)
		return
	}

	//fill
	if prsp.Ret == ss.CREATE_GROUP_RESULT_CREATE_RET_SUCCESS {
		pmsg.Name = prsp.GrpName
		pmsg.Result = int(ss.SS_COMMON_RESULT_SUCCESS)
		pmsg.CreateTs = prsp.CreateTs
		pmsg.GrpId = prsp.GrpId
		pmsg.MemberCnt = int(prsp.MemCount)
		pmsg.Desc = prsp.Desc
	} else {
		pmsg.Result = int(prsp.Ret)
	}

	//to client
	SendToClient(pconfig, c_key, cs.CS_PROTO_CREATE_GRP_RSP, pmsg)
}

//Apply Group
func SendApplyGroupReq(pconfig *Config, uid int64, preq *cs.CSApplyGroupReq) bool {
	var _func_ = "<SendApplyGroupReq>"
	log := pconfig.Comm.Log

	//ss msg
	var ss_msg ss.SSMsg
	pApplyGroupReq := new(ss.MsgApplyGroupReq)
	pApplyGroupReq.GroupId = preq.GrpId
	pApplyGroupReq.GroupName = preq.GrpName
	pApplyGroupReq.Pass = preq.Pass
	pApplyGroupReq.ApplyUid = uid
	pApplyGroupReq.ApplyMsg = preq.Msg

	//fill
	err := comm.FillSSPkg(&ss_msg, ss.SS_PROTO_TYPE_APPLY_GROUP_REQ, pApplyGroupReq)
	if err != nil {
		log.Err("%s pack fail! err:%v uid:%d", _func_, err, uid)
		return false
	}

	//send
	return SendToLogic(pconfig, &ss_msg)
}

func RecvApplyGroupRsp(pconfig *Config, prsp *ss.MsgApplyGroupRsp) {
	var _func_ = "<RecvCreateGroupRsp>"
	log := pconfig.Comm.Log

	log.Info("%s uid:%d result:%v", _func_, prsp.ApplyUid, *prsp)
	//get c_key
	c_key, ok := pconfig.Uid2Ckey[prsp.ApplyUid]
	if !ok {
		log.Err("%s uid offline! uid:%d", _func_, prsp.ApplyUid)
		return
	}

	//to client
	var pmsg *cs.CSApplyGroupRsp
	pv, err := cs.Proto2Msg(cs.CS_PROTO_APPLY_GRP_RSP)
	if err != nil {
		log.Err("%s proto2msg failed! proto:%d err:%v", _func_, cs.CS_PROTO_APPLY_GRP_RSP, err)
		return
	}
	pmsg, ok = pv.(*cs.CSApplyGroupRsp)
	if !ok {
		log.Err("%s proto2msg type illegal!  proto:%d", _func_, cs.CS_PROTO_APPLY_GRP_RSP)
		return
	}

	//fill
	pmsg.GrpId = prsp.GroupId
	pmsg.Result = int(prsp.Result)
	pmsg.GrpName = prsp.GroupName
	pmsg.Flag = int(prsp.Flag)

	//to client
	SendToClient(pconfig, c_key, cs.CS_PROTO_APPLY_GRP_RSP, pmsg)
}

func RecvApplyGroupNotify(pconfig *Config, pnotify *ss.MsgApplyGroupNotify) {
	var _func_ = "<RecvApplyGroupNotify>"
	log := pconfig.Comm.Log

	log.Debug("%s uid:%d grp_id:%d grp_name:%s msg:%s", _func_, pnotify.MasterUid, pnotify.GroupId, pnotify.GroupName,
		pnotify.ApplyMsg)
	//get c_key
	c_key := GetClientKey(pconfig, pnotify.MasterUid)
	if c_key <= 0 {
		log.Info("%s uid:%d not online!", _func_, pnotify.MasterUid)
		return
	}

	//cs
	pv, err := cs.Proto2Msg(cs.CS_PROTO_APPLY_GRP_NOTIFY)
	if err != nil {
		log.Err("%s get msg fail! uid:%d err:%v", _func_, pnotify.MasterUid, err)
		return
	}
	pmsg, ok := pv.(*cs.CSApplyGroupNotify)
	if !ok {
		log.Err("%s not CSApplyGroupNotify! uid:%d", _func_, pnotify.MasterUid)
		return
	}
	pmsg.ApplyName = pnotify.ApplyName
	pmsg.ApplyUid = pnotify.ApplyUid
	pmsg.GrpId = pnotify.GroupId
	pmsg.GrpName = pnotify.GroupName
	pmsg.ApplyMsg = pnotify.ApplyMsg

	//to client
	SendToClient(pconfig, c_key, cs.CS_PROTO_APPLY_GRP_NOTIFY, pmsg)
}

func SendApplyGroupAudit(pconfig *Config, uid int64, paudit *cs.CSApplyGroupAudit) {
	var _func_ = "<SendApplyGroupAudit>"
	log := pconfig.Comm.Log

	log.Debug("%s audit:%d apply_uid:%d grp_id:%d grp_name:%s uid:%d", _func_, paudit.Audit, paudit.ApplyUid, paudit.GrpId,
		paudit.GrpName, uid)
	//ss
	var ss_msg ss.SSMsg
	preq := new(ss.MsgApplyGroupAudit)
	preq.ApplyUid = paudit.ApplyUid
	preq.GroupId = paudit.GrpId
	preq.GroupName = paudit.GrpName
	preq.Flag = int32(paudit.Flag)
	preq.Uid = uid
	if paudit.Audit == 1 {
		preq.Result = ss.APPLY_GROUP_RESULT_APPLY_GRP_ALLOW
	} else {
		preq.Result = ss.APPLY_GROUP_RESULT_APPLY_GRP_DENY
	}

	err := comm.FillSSPkg(&ss_msg, ss.SS_PROTO_TYPE_APPLY_GROUP_AUDIT, preq)
	if err != nil {
		log.Err("%s gen ss failed! err:%v uid:%d", _func_, err, uid)
		return
	}

	//send
	SendToLogic(pconfig, &ss_msg)
}

func SendSendChatReq(pconfig *Config, uid int64, psend *cs.CSSendChatReq) {
	var _func_ = "<SendSendChatReq>"
	log := pconfig.Comm.Log

	log.Debug("%s uid:%d grp_id:%d type:%d temp_id:%d content:%s", _func_, uid, psend.GrpId, psend.ChatType,
		psend.TempMsgId, psend.Content)

	//ss
	preq := new(ss.MsgSendChatReq)
	preq.ChatMsg = new(ss.ChatMsg)
	preq.Uid = uid
	preq.TempId = psend.TempMsgId
	preq.ChatMsg.Content = psend.Content
	preq.ChatMsg.ChatType = ss.CHAT_MSG_TYPE(psend.ChatType)
	preq.ChatMsg.GroupId = psend.GrpId

	var ss_msg ss.SSMsg
	err := comm.FillSSPkg(&ss_msg, ss.SS_PROTO_TYPE_SEND_CHAT_REQ, preq)
	if err != nil {
		log.Err("%s gen ss fail! err:%v uid:%d content:%s", _func_, err, uid, psend.Content)
		return
	}

	//to logic
	SendToLogic(pconfig, &ss_msg)
}

func RecvSendChatRsp(pconfig *Config, prsp *ss.MsgSendChatRsp) {
	var _func_ = "<RecvSendChatRsp>"
	log := pconfig.Comm.Log

	//get c_key
	c_key := GetClientKey(pconfig, prsp.Uid)
	if c_key <= 0 {
		log.Info("%s uid:%d not online!", _func_, prsp.Uid)
		return
	}

	//cs
	pv, err := cs.Proto2Msg(cs.CS_PROTO_SEND_CHAT_RSP)
	if err != nil {
		log.Err("%s get msg fail! uid:%d err:%v", _func_, prsp.Uid, err)
		return
	}
	pmsg, ok := pv.(*cs.CSSendChatRsp)
	if !ok {
		log.Err("%s not CSSendChatRsp! uid:%d", _func_, prsp.Uid)
		return
	}

	//generate
	pmsg.Result = int(prsp.Result)
	pmsg.TempMsgId = prsp.TempId

	if prsp.ChatMsg != nil {
		pmsg.ChatMsg = new(cs.ChatMsg)
		pmsg.ChatMsg.Content = prsp.ChatMsg.Content
		pmsg.ChatMsg.GrpId = prsp.ChatMsg.GroupId
		pmsg.ChatMsg.ChatType = int(prsp.ChatMsg.ChatType)
		pmsg.ChatMsg.MsgId = prsp.ChatMsg.MsgId
		pmsg.ChatMsg.SenderName = prsp.ChatMsg.Sender
		pmsg.ChatMsg.SenderUid = prsp.ChatMsg.SenderUid
		pmsg.ChatMsg.SendTs = prsp.ChatMsg.SendTs
	}

	//to client
	SendToClient(pconfig, c_key, cs.CS_PROTO_SEND_CHAT_RSP, pmsg)
}

func RecvSyncChatList(pconfig *Config, prsp *ss.MsgSyncChatList) {
	var _func_ = "<RecvSyncChatList>"
	log := pconfig.Comm.Log

	//get c_key
	c_key := GetClientKey(pconfig, prsp.Uid)
	if c_key <= 0 {
		log.Info("%s uid:%d not online!", _func_, prsp.Uid)
		return
	}

	//cs
	pv, err := cs.Proto2Msg(cs.CS_PROTO_SYNC_CHAT_LIST)
	if err != nil {
		log.Err("%s get msg fail! uid:%d err:%v", _func_, prsp.Uid, err)
		return
	}
	pmsg, ok := pv.(*cs.CSSyncChatList)
	if !ok {
		log.Err("%s not CSSyncChatList! uid:%d", _func_, prsp.Uid)
		return
	}

	log.Debug("%s uid:%d grp_id:%d count:%d", _func_, prsp.Uid, prsp.GrpId, prsp.Count)
	//Fill Info
	if prsp.Count <= 0 {
		return
	}

	pmsg.GrpId = prsp.GrpId
	pmsg.SyncType = int8(prsp.SyncType)
	pmsg.Count = int(prsp.Count)
	pmsg.ChatList = make([]*cs.ChatMsg, pmsg.Count)
	for i := 0; i < int(prsp.Count); i++ {
		pchat := new(cs.ChatMsg)
		pchat.GrpId = prsp.GrpId
		pchat.MsgId = prsp.ChatList[i].MsgId
		pchat.SenderName = prsp.ChatList[i].Sender
		pchat.SenderUid = prsp.ChatList[i].SenderUid
		pchat.SendTs = prsp.ChatList[i].SendTs
		pchat.Content = prsp.ChatList[i].Content
		pchat.ChatType = int(prsp.ChatList[i].ChatType)
		pchat.Flag = int64(prsp.ChatList[i].ChatFlag)
		pmsg.ChatList[i] = pchat
	}

	//To Client
	SendToClient(pconfig, c_key, cs.CS_PROTO_SYNC_CHAT_LIST, pmsg)
}

func SendExitGroupReq(pconfig *Config, uid int64, pexit *cs.CSExitGroupReq) {
	var _func_ = "<SendExitGroupReq>"
	log := pconfig.Comm.Log

	//ss
	var ss_msg ss.SSMsg
	preq := new(ss.MsgExitGroupReq)
	preq.Uid = uid
	preq.GrpId = pexit.GrpId

	err := comm.FillSSPkg(&ss_msg, ss.SS_PROTO_TYPE_EXIT_GROUP_REQ, preq)
	if err != nil {
		log.Err("%s gen ss failed! err:%v uid:%d grp_id:%d", _func_, err, uid, pexit.GrpId)
		return
	}

	//send
	SendToLogic(pconfig, &ss_msg)
}

func RecvExitGroupRsp(pconfig *Config, prsp *ss.MsgExitGroupRsp) {
	var _func_ = "<SendExitGroupRsp>"
	log := pconfig.Comm.Log

	//c_key
	c_key := GetClientKey(pconfig, prsp.Uid)
	if c_key <= 0 {
		log.Err("%s user offline! uid:%d", _func_, prsp.Uid)
		return
	}

	//cs
	pv, err := cs.Proto2Msg(cs.CS_PROTO_EXIT_GROUP_RSP)
	if err != nil {
		log.Err("%s get msg fail! uid:%d err:%v", _func_, prsp.Uid, err)
		return
	}
	pmsg, ok := pv.(*cs.CSExitGroupRsp)
	if !ok {
		log.Err("%s not CSExitGroupRsp! uid:%d", _func_, prsp.Uid)
		return
	}

	log.Debug("%s uid:%d grp_id:%d result:%d", _func_, prsp.Uid, prsp.GrpId, prsp.Result)
	//fill
	pmsg.Result = int(prsp.Result)
	pmsg.GrpId = prsp.GrpId
	pmsg.GrpName = prsp.GrpName
	pmsg.DelGroup = int8(prsp.DelGroup)
	pmsg.ByKick = int8(prsp.ByKick)

	//to client
	SendToClient(pconfig, c_key, cs.CS_PROTO_EXIT_GROUP_RSP, pmsg)
}

func SendFetchChatHistroyReq(pconfig *Config, uid int64, pfetch *cs.CSChatHistoryReq) {
	var _func_ = "<SendFetchChatHistroyReq>"
	log := pconfig.Comm.Log

	//ss
	preq := new(ss.MsgFetchChatReq)
	preq.Uid = uid
	preq.GrpId = pfetch.GrpId
	preq.FetchType = ss.SS_COMMON_TYPE_COMM_TYPE_HISTORY
	preq.LatestMsgId = pfetch.NowMsgId
	if preq.LatestMsgId < 0 {
		preq.LatestMsgId = 0
	}
	preq.FetchCount = pfetch.Count

	var ss_msg ss.SSMsg
	err := comm.FillSSPkg(&ss_msg, ss.SS_PROTO_TYPE_FETCH_CHAT_REQ, preq)
	if err != nil {
		log.Err("%s gen ss failed! uid:%d err:%v", _func_, err, uid)
		return
	}

	//to logic
	SendToLogic(pconfig, &ss_msg)
}

func SendKickGroupReq(pconfig *Config, uid int64, pkick *cs.CSKickGroupReq) {
	var _func_ = "<SendKickGroupReq>"
	log := pconfig.Comm.Log

	log.Debug("%s try to kick:%d grp_id:%d uid:%d", _func_, pkick.KickUid, pkick.GrpId, uid)
	//ss
	preq := new(ss.MsgKickGroupReq)
	preq.GrpId = pkick.GrpId
	preq.Uid = uid
	preq.KickUid = pkick.KickUid

	var ss_msg ss.SSMsg
	err := comm.FillSSPkg(&ss_msg, ss.SS_PROTO_TYPE_KICK_GROUP_REQ, preq)
	if err != nil {
		log.Err("%s gen ss failed! uid:%d kick:%d", _func_, uid, preq.KickUid)
		return
	}

	//to logic
	SendToLogic(pconfig, &ss_msg)
}

func SendQueryGroupReq(pconfig *Config, uid int64, pquery *cs.CSQueryGroupReq) {
	var _func_ = "<SendQueryGroupReq>"
	log := pconfig.Comm.Log

	log.Debug("%s uid:%d grp_id:%d", _func_, uid, pquery.GrpId)
	//ss
	preq := new(ss.MsgQueryGroupReq)
	preq.GrpId = pquery.GrpId
	preq.Uid = uid

	var ss_msg ss.SSMsg
	err := comm.FillSSPkg(&ss_msg, ss.SS_PROTO_TYPE_QUERY_GROUP_REQ, preq)
	if err != nil {
		log.Err("%s gen ss failed! uid:%d grp_id:%d err:%v", _func_, uid, preq.GrpId, err)
		return
	}

	//to logic
	SendToLogic(pconfig, &ss_msg)
}

func RecvSyncGroupInfo(pconfig *Config, pinfo *ss.MsgSyncGroupInfo) {
	var _func_ = "<RecvSyncGroupInfo>"
	log := pconfig.Comm.Log
	uid := pinfo.Uid
	grp_id := pinfo.GrpId

	//c_key
	c_key := GetClientKey(pconfig, uid)
	if c_key <= 0 {
		log.Err("%s user offline! uid:%d", _func_, uid)
		return
	}

	//cs
	pv, err := cs.Proto2Msg(cs.CS_PROTO_SYNC_GROUP_INFO)
	if err != nil {
		log.Err("%s get msg fail! uid:%d err:%v", _func_, uid, err)
		return
	}
	pmsg, ok := pv.(*cs.CSSyncGroupInfo)
	if !ok {
		log.Err("%s not CSSyncGroupInfo! uid:%d", _func_, uid)
		return
	}

	log.Debug("%s uid:%d grp_id:%d field:%d", _func_, uid, grp_id, pinfo.Field)
	//fill msg
	switch pinfo.Field {
	case ss.SS_GROUP_INFO_FIELD_GRP_FIELD_ALL:
		pmsg.Field = cs.SYNC_GROUP_FIELD_ALL
		pmsg.GrpId = grp_id
		if pinfo.GrpInfo == nil {
			log.Err("%s field all but group_info nil! uid:%d grp_id:%d", _func_, uid, grp_id)
			return
		}

		//fill group
		pmsg.GrpInfo = new(cs.ChatGroup)
		pmsg.GrpInfo.GrpId = grp_id
		pmsg.GrpInfo.GrpName = pinfo.GrpInfo.GroupName
		pmsg.GrpInfo.MasterUid = pinfo.GrpInfo.MasterUid
		pmsg.GrpInfo.MsgCount = pinfo.GrpInfo.LatestMsgId
		pmsg.GrpInfo.CreateTs = pinfo.GrpInfo.CreatedTs
		pmsg.GrpInfo.MemCount = pinfo.GrpInfo.MemCount
		pmsg.GrpInfo.Visible = pinfo.GrpInfo.BlobInfo.Visible
		pmsg.GrpInfo.Desc = pinfo.GrpInfo.BlobInfo.GroupDesc
		pmsg.GrpInfo.HeadUrl = pinfo.GrpInfo.BlobInfo.HeadUrl
		if pinfo.GrpInfo.MemCount > 0 || pinfo.GrpInfo.Members != nil {
			pmsg.GrpInfo.Members = make(map[int64]int32)
			pmsg.GrpInfo.Members = pinfo.GrpInfo.Members
		}
	case ss.SS_GROUP_INFO_FIELD_GRP_FIELD_SNAP:
		pmsg.Field = cs.SYNC_GROUP_FIELD_SNAP
		pmsg.GrpId = grp_id
		if pinfo.GrpSnap == nil {
			log.Err("%s field all but group_snap nil! uid:%d grp_id:%d", _func_, uid, grp_id)
			return
		}

		pmsg.GrpSnap = new(cs.GroupGroundItem)
		pmsg.GrpSnap.GrpId = grp_id
		pmsg.GrpSnap.GrpName = pinfo.GrpSnap.GrpName
		pmsg.GrpSnap.MemCount = pinfo.GrpSnap.MemCount
		pmsg.GrpSnap.Desc = pinfo.GrpSnap.Desc
		pmsg.GrpSnap.HeadUrl = pinfo.GrpSnap.HeadUrl
	default:
		log.Err("%s unknwon field:%d uid:%d grp_id:%d", _func_, pinfo.Field, uid, grp_id)
		return
	}

	//to client
	SendToClient(pconfig, c_key, cs.CS_PROTO_SYNC_GROUP_INFO, pmsg)
}

func SendChgGroupAttrReq(pconfig *Config, uid int64, pchg *cs.CSChgGroupAttrReq) {
	var _func_ = "<SendChgGroupAttrReq>"
	log := pconfig.Comm.Log
	grp_id := pchg.GrpId

	//req
	preq := new(ss.MsgChgGroupAttrReq)
	preq.Uid = uid
	preq.GrpId = grp_id

	//switch
	switch pchg.Attr {
	case cs.GROUP_ATTR_VISIBLE:
		preq.Attr = ss.GROUP_ATTR_TYPE_GRP_ATTR_VISIBLE
	case cs.GROUP_ATTR_INVISIBLE:
		preq.Attr = ss.GROUP_ATTR_TYPE_GRP_ATTR_INVISIBLE
	case cs.GROUP_ATTR_DESC:
		preq.Attr = ss.GROUP_ATTR_TYPE_GRP_ATTR_DESC
		preq.StrV = pchg.StrV
	case cs.GROUP_ATTR_GRP_NAME:
		preq.Attr = ss.GROUP_ATTR_TYPE_GRP_ATTR_GRP_NAME
		preq.StrV = pchg.StrV
	case cs.GROUP_ATTR_GRP_HEAD:
		preq.Attr = ss.GROUP_ATTR_TYPE_GRP_ATTR_HEAD_URL
		preq.StrV = pchg.StrV
	default:
		log.Err("%s illegal attr:%d uid:%d", _func_, pchg.Attr, uid)
		return
	}

	//send
	var ss_msg ss.SSMsg
	err := comm.FillSSPkg(&ss_msg, ss.SS_PROTO_TYPE_CHG_GROUP_ATTR_REQ, preq)
	if err != nil {
		log.Err("%s gen ss failed! err:%v uid:%d grp_id:%d attr:%d", _func_, err, uid, grp_id, pchg.Attr)
		return
	}

	//to logic
	SendToLogic(pconfig, &ss_msg)
}

func RecvChgGroupAttrRsp(pconfig *Config, prsp *ss.MsgChgGroupAttrRsp) {
	var _func_ = "<RecvChgGroupAttrRsp>"
	log := pconfig.Comm.Log
	uid := prsp.Uid
	grp_id := prsp.GrpId

	//c_key
	c_key := GetClientKey(pconfig, uid)
	if c_key <= 0 {
		log.Err("%s user offline! uid:%d", _func_, uid)
		return
	}

	//cs
	pv, err := cs.Proto2Msg(cs.CS_PROTO_CHG_GROUP_ATTR_RSP)
	if err != nil {
		log.Err("%s get msg fail! uid:%d err:%v", _func_, uid, err)
		return
	}
	pmsg, ok := pv.(*cs.CSChgGroupAttrRsp)
	if !ok {
		log.Err("%s not CSChgGroupAttrRsp! uid:%d", _func_, uid)
		return
	}

	log.Debug("%s uid:%d grp_id:%d attr:%d result:%d", _func_, uid, grp_id, prsp.Attr, prsp.Result)

	//fill info
	pmsg.Result = int(prsp.Result)
	pmsg.GrpId = grp_id
	switch prsp.Attr {
	case ss.GROUP_ATTR_TYPE_GRP_ATTR_VISIBLE:
		pmsg.Attr = cs.GROUP_ATTR_VISIBLE
	case ss.GROUP_ATTR_TYPE_GRP_ATTR_INVISIBLE:
		pmsg.Attr = cs.GROUP_ATTR_INVISIBLE
	case ss.GROUP_ATTR_TYPE_GRP_ATTR_DESC:
		pmsg.Attr = cs.GROUP_ATTR_DESC
		pmsg.StrV = prsp.StrV
	case ss.GROUP_ATTR_TYPE_GRP_ATTR_GRP_NAME:
		pmsg.Attr = cs.GROUP_ATTR_GRP_NAME
		pmsg.StrV = prsp.StrV
	case ss.GROUP_ATTR_TYPE_GRP_ATTR_HEAD_URL:
		pmsg.Attr = cs.GROUP_ATTR_GRP_HEAD
		pmsg.StrV = prsp.StrV
	default:
		log.Err("%s illegal attr:%d uid:%d grp_id:%d", _func_, prsp.Attr, uid, grp_id)
		return
	}

	//to client
	SendToClient(pconfig, c_key, cs.CS_PROTO_CHG_GROUP_ATTR_RSP, pmsg)
}

func SendGroupGroundReq(pconfig *Config, uid int64, pcs *cs.CSGroupGroundReq) {
	var _func_ = "<SendGroupGroundReq>"
	log := pconfig.Comm.Log

	//req
	preq := new(ss.MsgGroupGroudReq)
	preq.Uid = uid
	preq.StartIndex = int32(pcs.StartIndex)

	//ss
	var ss_msg ss.SSMsg
	err := comm.FillSSPkg(&ss_msg, ss.SS_PROTO_TYPE_GROUP_GROUND_REQ, preq)
	if err != nil {
		log.Err("%s gen ss failed! err:%v uid:%d", _func_, err, uid)
		return
	}

	//to logic
	SendToLogic(pconfig, &ss_msg)
}

func RecvGroupGroundRsp(pconfig *Config, prsp *ss.MsgGroupGroudRsp) {
	var _func_ = "<RecvGroupGroundRsp>"
	log := pconfig.Comm.Log
	uid := prsp.Uid

	//c_key
	c_key := GetClientKey(pconfig, uid)
	if c_key <= 0 {
		log.Err("%s user offline! uid:%d", _func_, uid)
		return
	}

	//cs
	pv, err := cs.Proto2Msg(cs.CS_PROTO_GROUP_GROUND_RSP)
	if err != nil {
		log.Err("%s get msg fail! uid:%d err:%v", _func_, uid, err)
		return
	}
	pmsg, ok := pv.(*cs.CSGroupGroundRsp)
	if !ok {
		log.Err("%s not CSChgGroupAttrRsp! uid:%d", _func_, uid)
		return
	}

	log.Debug("%s uid:%d count:%d", _func_, uid, prsp.Count)
	//fill info
	pmsg.Count = int(prsp.Count)
	if prsp.Count > 0 {
		pmsg.ItemList = make([]*cs.GroupGroundItem, pmsg.Count)
	}
	for i := 0; i < int(prsp.Count); i++ {
		pitem := new(cs.GroupGroundItem)
		pitem.GrpId = prsp.ItemList[i].GrpId
		pitem.GrpName = prsp.ItemList[i].GrpName
		pitem.MemCount = prsp.ItemList[i].MemCount
		pitem.Desc = prsp.ItemList[i].Desc
		pitem.HeadUrl = prsp.ItemList[i].HeadUrl
		pmsg.ItemList[i] = pitem
	}

	//to client
	SendToClient(pconfig, c_key, cs.CS_PROTO_GROUP_GROUND_RSP, pmsg)
}

func SendUpdateChatReq(pconfig *Config, uid int64, pcs *cs.CSUpdateChatReq) {
	var _func_ = "<SendGroupGroundReq>"
	log := pconfig.Comm.Log

	//req
	preq := new(ss.MsgUpdateChatReq)
	preq.Uid = uid
	preq.UpdateType = ss.UPDATE_CHAT_TYPE(pcs.UpdateType)
	preq.GrpId = pcs.Grpid
	preq.MsgId = pcs.MsgId

	//ss
	var ss_msg ss.SSMsg
	err := comm.FillSSPkg(&ss_msg, ss.SS_PROTO_TYPE_UPDATE_CHAT_REQ, preq)
	if err != nil {
		log.Err("%s gen ss failed! err:%v uid:%d", _func_, err, uid)
		return
	}

	//to logic
	SendToLogic(pconfig, &ss_msg)
}

func RecvUpdateChatRsp(pconfig *Config, prsp *ss.MsgUpdateChatRsp) {
	var _func_ = "<RecvGroupGroundRsp>"
	log := pconfig.Comm.Log
	uid := prsp.Uid

	//c_key
	c_key := GetClientKey(pconfig, uid)
	if c_key <= 0 {
		log.Err("%s user offline! uid:%d", _func_, uid)
		return
	}

	//cs
	pv, err := cs.Proto2Msg(cs.CS_PROTO_UPDATE_CHAT_RSP)
	if err != nil {
		log.Err("%s get msg fail! uid:%d err:%v", _func_, uid, err)
		return
	}
	pmsg, ok := pv.(*cs.CSUpdateChatRsp)
	if !ok {
		log.Err("%s not CSUpdateChatRsp! uid:%d", _func_, uid)
		return
	}

	log.Debug("%s uid:%d grp_id:%d result:%d msg_id:%d type:%d", _func_, uid, prsp.GrpId , prsp.Result , prsp.MsgId , prsp.UpdateType)
	//fill info
	pmsg.UpdateType = int(prsp.UpdateType)
	pmsg.GrpId = prsp.GrpId
	pmsg.Result = int(prsp.Result)
	pmsg.MsgId = prsp.MsgId

	//to client
	SendToClient(pconfig , c_key , cs.CS_PROTO_UPDATE_CHAT_RSP , pmsg)
}