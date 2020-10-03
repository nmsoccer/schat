package lib

import (
	"schat/proto/cs"
	"schat/proto/ss"
	"schat/servers/comm"
)

func SendCreateGroupReq(pconfig *Config , uid int64 , preq *cs.CSCreateGroupReq) bool {
	var _func_= "<SendCreateGroupReq>"
	log := pconfig.Comm.Log

	//ss msg
	var ss_msg ss.SSMsg
	pCreateGroupReq := new(ss.MsgCreateGrpReq)
	pCreateGroupReq.Uid = uid
	pCreateGroupReq.GrpName = preq.Name
	pCreateGroupReq.GrpPass = preq.Pass


	//fill
	err := comm.FillSSPkg(&ss_msg , ss.SS_PROTO_TYPE_CREATE_GROUP_REQ , pCreateGroupReq)
	if err != nil {
		log.Err("%s gen ss fail! err:%v uid:%d" , _func_ , err , uid)
		return false
	}

	//send
	return SendToLogic(pconfig , &ss_msg)
}

func RecvCreateGroupRsp(pconfig *Config , prsp *ss.MsgCreateGrpRsp) {
	var _func_ = "<RecvCreateGroupRsp>"
	log := pconfig.Comm.Log

	log.Info("%s uid:%d result:%v" , _func_ , prsp.Uid , *prsp)
	//get c_key
	c_key  , ok := pconfig.Uid2Ckey[prsp.Uid]
	if !ok {
		log.Err("%s uid offline! uid:%d" , _func_ , prsp.Uid)
		return
	}

	//to client
    var pmsg *cs.CSCreateGroupRsp
	pv , err := cs.Proto2Msg(cs.CS_PROTO_CREATE_GRP_RSP)
	if err != nil {
		log.Err("%s proto2msg failed! proto:%d err:%v" , _func_ , cs.CS_PROTO_CREATE_GRP_RSP , err)
		return
	}
	pmsg , ok = pv.(*cs.CSCreateGroupRsp);
	if !ok {
		log.Err("%s proto2msg type illegal!  proto:%d" , _func_ , cs.CS_PROTO_CREATE_GRP_RSP)
		return
	}

	//fill
    if prsp.Ret == ss.CREATE_GROUP_RESULT_CREATE_RET_SUCCESS {
    	pmsg.Name = prsp.GrpName
    	pmsg.Result = 0
    	pmsg.CreateTs = prsp.CreateTs
    	pmsg.GrpId = prsp.GrpId
    	pmsg.MemberCnt = int(prsp.MemCount)
	} else {
		pmsg.Result = int(prsp.Ret)
	}

	//to client
	SendToClient(pconfig , c_key , cs.CS_PROTO_CREATE_GRP_RSP , pmsg)
}

//Apply Group
func SendApplyGroupReq(pconfig *Config , uid int64 , preq *cs.CSApplyGroupReq) bool {
	var _func_= "<SendApplyGroupReq>"
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
	err := comm.FillSSPkg(&ss_msg , ss.SS_PROTO_TYPE_APPLY_GROUP_REQ , pApplyGroupReq)
	if err != nil {
		log.Err("%s pack fail! err:%v uid:%d" , _func_ , err , uid)
		return false
	}

	//send
	return SendToLogic(pconfig , &ss_msg)
}

func RecvApplyGroupRsp(pconfig *Config , prsp *ss.MsgApplyGroupRsp) {
	var _func_ = "<RecvCreateGroupRsp>"
	log := pconfig.Comm.Log

	log.Info("%s uid:%d result:%v" , _func_ , prsp.ApplyUid , *prsp)
	//get c_key
	c_key  , ok := pconfig.Uid2Ckey[prsp.ApplyUid]
	if !ok {
		log.Err("%s uid offline! uid:%d" , _func_ , prsp.ApplyUid)
		return
	}

	//to client
	var pmsg *cs.CSApplyGroupRsp
	pv , err := cs.Proto2Msg(cs.CS_PROTO_APPLY_GRP_RSP)
	if err != nil {
		log.Err("%s proto2msg failed! proto:%d err:%v" , _func_ , cs.CS_PROTO_APPLY_GRP_RSP , err)
		return
	}
	pmsg , ok = pv.(*cs.CSApplyGroupRsp);
	if !ok {
		log.Err("%s proto2msg type illegal!  proto:%d" , _func_ , cs.CS_PROTO_APPLY_GRP_RSP)
		return
	}

	//fill
	pmsg.GrpId = prsp.GroupId
	pmsg.Result = int(prsp.Result)
	pmsg.GrpName = prsp.GroupName


	//to client
	SendToClient(pconfig , c_key , cs.CS_PROTO_APPLY_GRP_RSP , pmsg)
}

func RecvApplyGroupNotify(pconfig *Config , pnotify *ss.MsgApplyGroupNotify) {
	var _func_ = "<RecvApplyGroupNotify>"
	log := pconfig.Comm.Log

	log.Debug("%s uid:%d grp_id:%d grp_name:%s msg:%s" , _func_ , pnotify.MasterUid , pnotify.GroupId , pnotify.GroupName ,
		pnotify.ApplyMsg)
	//get c_key
	c_key := GetClientKey(pconfig , pnotify.MasterUid)
	if c_key <= 0 {
		log.Info("%s uid:%d not online!" , _func_ , pnotify.MasterUid)
		return
	}

	//cs
	pv , err := cs.Proto2Msg(cs.CS_PROTO_APPLY_GRP_NOTIFY)
	if err != nil {
		log.Err("%s get msg fail! uid:%d err:%v" , _func_ , pnotify.MasterUid , err)
		return
	}
	pmsg , ok := pv.(*cs.CSApplyGroupNotify)
	if !ok {
		log.Err("%s not CSApplyGroupNotify! uid:%d" , _func_ , pnotify.MasterUid)
		return
	}
	pmsg.ApplyName = pnotify.ApplyName
	pmsg.ApplyUid = pnotify.ApplyUid
	pmsg.GrpId = pnotify.GroupId
	pmsg.GrpName = pnotify.GroupName
	pmsg.ApplyMsg = pnotify.ApplyMsg

	//to client
	SendToClient(pconfig , c_key , cs.CS_PROTO_APPLY_GRP_NOTIFY , pmsg)
}

func SendApplyGroupAudit(pconfig *Config , uid int64 , paudit *cs.CSApplyGroupAudit) {
	var _func_ = "<SendApplyGroupAudit>"
	log := pconfig.Comm.Log

	log.Debug("%s audit:%d apply_uid:%d grp_id:%d grp_name:%s uid:%d" , _func_ , paudit.Audit , paudit.ApplyUid , paudit.GrpId ,
		paudit.GrpName , uid)
	//ss
	var ss_msg ss.SSMsg
	preq := new(ss.MsgApplyGroupAudit)
	preq.ApplyUid = paudit.ApplyUid
	preq.GroupId = paudit.GrpId
	preq.GroupName = paudit.GrpName
	preq.Uid = uid
	if paudit.Audit == 1 {
		preq.Result = ss.APPLY_GROUP_RESULT_APPLY_GRP_ALLOW
	} else {
		preq.Result = ss.APPLY_GROUP_RESULT_APPLY_GRP_DENY
	}

	err := comm.FillSSPkg(&ss_msg , ss.SS_PROTO_TYPE_APPLY_GROUP_AUDIT , preq)
	if err != nil {
		log.Err("%s gen ss failed! err:%v uid:%d" , _func_ , err , uid)
		return
	}

    //send
    SendToLogic(pconfig , &ss_msg)
}

func SendSendChatReq(pconfig *Config , uid int64 , psend *cs.CSSendChatReq) {
	var _func_ = "<SendSendChatReq>"
	log := pconfig.Comm.Log

	log.Debug("%s uid:%d grp_id:%d type:%d temp_id:%d content:%s" , _func_ , uid , psend.GrpId , psend.ChatType ,
		psend.TempMsgId , psend.Content)

	//ss
	preq := new(ss.MsgSendChatReq)
	preq.ChatMsg = new(ss.ChatMsg)
	preq.Uid = uid
	preq.TempId = psend.TempMsgId
	preq.ChatMsg.Content = psend.Content
	preq.ChatMsg.ChatType = ss.CHAT_MSG_TYPE(psend.ChatType)
	preq.ChatMsg.GroupId = psend.GrpId

	var ss_msg ss.SSMsg
    err := comm.FillSSPkg(&ss_msg , ss.SS_PROTO_TYPE_SEND_CHAT_REQ , preq)
    if err != nil {
    	log.Err("%s gen ss fail! err:%v uid:%d content:%s" , _func_ , err , uid , psend.Content)
    	return
	}

	//to logic
	SendToLogic(pconfig , &ss_msg)
}

func RecvSendChatRsp(pconfig *Config , prsp *ss.MsgSendChatRsp) {
	var _func_ = "<RecvSendChatRsp>"
	log := pconfig.Comm.Log

	//get c_key
	c_key := GetClientKey(pconfig , prsp.Uid)
	if c_key <= 0 {
		log.Info("%s uid:%d not online!" , _func_ , prsp.Uid)
		return
	}

	//cs
	pv , err := cs.Proto2Msg(cs.CS_PROTO_SEND_CHAT_RSP)
	if err != nil {
		log.Err("%s get msg fail! uid:%d err:%v" , _func_ , prsp.Uid , err)
		return
	}
	pmsg , ok := pv.(*cs.CSSendChatRsp)
	if !ok {
		log.Err("%s not CSSendChatRsp! uid:%d" , _func_ , prsp.Uid)
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
	SendToClient(pconfig , c_key , cs.CS_PROTO_SEND_CHAT_RSP , pmsg)
}

func RecvSyncChatList(pconfig *Config , prsp *ss.MsgSyncChatList) {
	var _func_ = "<RecvSyncChatList>"
	log := pconfig.Comm.Log

	//get c_key
	c_key := GetClientKey(pconfig , prsp.Uid)
	if c_key <= 0 {
		log.Info("%s uid:%d not online!" , _func_ , prsp.Uid)
		return
	}

	//cs
	pv , err := cs.Proto2Msg(cs.CS_PROTO_SYNC_CHAT_LIST)
	if err != nil {
		log.Err("%s get msg fail! uid:%d err:%v" , _func_ , prsp.Uid , err)
		return
	}
	pmsg , ok := pv.(*cs.CSSyncChatList)
	if !ok {
		log.Err("%s not CSSyncChatList! uid:%d" , _func_ , prsp.Uid)
		return
	}

	log.Debug("%s uid:%d grp_id:%d count:%d" , _func_ , prsp.Uid , prsp.GrpId , prsp.Count)
	//Fill Info
	if prsp.Count <= 0 {
		return
	}

	pmsg.GrpId = prsp.GrpId
	pmsg.Count = int(prsp.Count)
	pmsg.ChatList = make([]*cs.ChatMsg , pmsg.Count)
	for i:=0; i<int(prsp.Count); i++ {
		pchat := new(cs.ChatMsg)
		pchat.GrpId = prsp.GrpId
		pchat.MsgId = prsp.ChatList[i].MsgId
		pchat.SenderName = prsp.ChatList[i].Sender
		pchat.SenderUid = prsp.ChatList[i].SenderUid
		pchat.SendTs = prsp.ChatList[i].SendTs
		pchat.Content = prsp.ChatList[i].Content
		pchat.ChatType = int(prsp.ChatList[i].ChatType)
		pmsg.ChatList[i] = pchat
	}

	//To Client
	SendToClient(pconfig , c_key , cs.CS_PROTO_SYNC_CHAT_LIST , pmsg)
}