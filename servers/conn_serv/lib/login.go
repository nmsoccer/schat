package lib

import (
	"schat/proto/cs"
	"schat/proto/ss"
	"schat/servers/comm"
)

func SendLoginReq(pconfig *Config, client_key int64, plogin_req *cs.CSLoginReq) {
	var _func_ = "<SendLoginReq>"
	log := pconfig.Comm.Log

	log.Debug("%s send login pkg to logic! user:%s device:%s", _func_, plogin_req.Name, plogin_req.Device)
	//create pkg
	var ss_msg ss.SSMsg
	pLoginReq := new(ss.MsgLoginReq)
	pLoginReq.CKey = client_key
	pLoginReq.Name = plogin_req.Name
	pLoginReq.Pass = plogin_req.Pass
	pLoginReq.Device = plogin_req.Device
	pLoginReq.Version = plogin_req.Version

	//ss_msg
	err := comm.FillSSPkg(&ss_msg, ss.SS_PROTO_TYPE_LOGIN_REQ, pLoginReq)
	if err != nil {
		log.Err("%s gen ss_pkg failed! err:%v ckey:%v", _func_, err, client_key)
		return
	}

	//send
	ok := SendToLogic(pconfig, &ss_msg)
	if !ok {
		log.Err("%s send failed! client_key:%v", _func_, client_key)
		return
	}
}

func RecvLoginRsp(pconfig *Config, prsp *ss.MsgLoginRsp) {
	var _func_ = "<RecvLoginRsp>"
	log := pconfig.Comm.Log

	//log
	log.Debug("%s result:%d user:%s c_key:%v", _func_, prsp.Result, prsp.Name, prsp.CKey)

	//response
	var pmsg *cs.CSLoginRsp
	pv, err := cs.Proto2Msg(cs.CS_PROTO_LOGIN_RSP)
	if err != nil {
		log.Err("%s proto2msg failed! proto:%d err:%v", _func_, cs.CS_PROTO_LOGIN_RSP, err)
		return
	}
	pmsg, ok := pv.(*cs.CSLoginRsp)
	if !ok {
		log.Err("%s proto2msg type illegal!  proto:%d", _func_, cs.CS_PROTO_LOGIN_RSP)
		return
	}

	//msg
	pmsg.Result = int(prsp.Result)
	pmsg.Name = prsp.Name
	//success
	if prsp.Result == ss.USER_LOGIN_RET_LOGIN_SUCCESS {
		for {
			//check if multi-clients connect to same connect_serv
			old_key, ok := pconfig.Uid2Ckey[prsp.Uid]
			if ok && old_key != prsp.CKey {
				log.Err("%s user:%d already logon at %d now is:%d abandon!", _func_, prsp.Uid, old_key, prsp.CKey)
				pmsg.Result = int(ss.USER_LOGIN_RET_LOGIN_MULTI_ON)
				break
			}

			//basic
			pmsg.Basic.Uid = prsp.GetUserInfo().BasicInfo.Uid
			pmsg.Basic.Name = prsp.GetUserInfo().BasicInfo.Name
			pmsg.Basic.Addr = prsp.UserInfo.BasicInfo.Addr
			pmsg.Basic.Level = prsp.UserInfo.BasicInfo.Level
			pmsg.Basic.HeadUrl = prsp.UserInfo.BasicInfo.HeadUrl
			if prsp.UserInfo.BasicInfo.Sex {
				pmsg.Basic.Sex = 1
			} else {
				pmsg.Basic.Sex = 0
			}

			//blob
			pblob := prsp.UserInfo.GetBlobInfo()

			//detail
			pmsg.Detail.Exp = prsp.UserInfo.BlobInfo.Exp
			pmsg.Detail.ChatInfo = new(cs.UserChatInfo)
			pmsg.Detail.Desc = prsp.UserInfo.BlobInfo.UserDesc
			pmsg.Detail.ClientDesKey = prsp.UserInfo.BlobInfo.ClientEncDesKey
			if pblob.ChatInfo != nil && pblob.ChatInfo.AllGroup > 0 {
				blob_chat_info := pblob.GetChatInfo()
				cs_chat_info := pmsg.Detail.ChatInfo

				//FILL ALL GROUP
				cs_chat_info.AllGroups = make(map[int64]*cs.UserChatGroup)
				for grp_id, grp_info := range blob_chat_info.GetAllGroups() {
					cs_chat_info.AllGroups[grp_id] = new(cs.UserChatGroup)
					cs_chat_info.AllGroups[grp_id].GroupId = grp_info.GroupId
					cs_chat_info.AllGroups[grp_id].GroupName = grp_info.GroupName
					cs_chat_info.AllGroups[grp_id].LastMsgId = grp_info.LastReadId
					cs_chat_info.AllGroups[grp_id].EnterTs = grp_info.EnterTs
				}
				cs_chat_info.AllGroup = blob_chat_info.AllGroup

				//FILL MASTER GROUP
				if blob_chat_info.MasterGroup > 0 {
					cs_chat_info.MasterGroups = make(map[int64]bool)
					for grp_id, v := range blob_chat_info.MasterGroups {
						cs_chat_info.MasterGroups[grp_id] = v
					}
				}
				cs_chat_info.MasterGroup = blob_chat_info.MasterGroup
			}

			//create map refer
			pconfig.Ckey2Uid[prsp.CKey] = pmsg.Basic.Uid
			pconfig.Uid2Ckey[pmsg.Basic.Uid] = prsp.CKey
			break
		}
	}

	//to client
	SendToClient(pconfig, prsp.CKey, cs.CS_PROTO_LOGIN_RSP, pmsg)
}

func SendLogoutReq(pconfig *Config, uid int64, reason ss.USER_LOGOUT_REASON) {
	var _func_ = "<SendLogoutReq>"
	log := pconfig.Comm.Log

	//construct
	var ss_msg ss.SSMsg
	pLogoutReq := new(ss.MsgLogoutReq)
	pLogoutReq.Uid = uid
	pLogoutReq.Reason = reason

	//pack
	err := comm.FillSSPkg(&ss_msg, ss.SS_PROTO_TYPE_LOGOUT_REQ, pLogoutReq)
	if err != nil {
		log.Err("%s gen ss_pkg failed! err:%v uid:%v reason:%v", _func_, err, uid, reason)
		return
	}

	//send
	ok := SendToLogic(pconfig, &ss_msg)
	if !ok {
		log.Err("%s send failed! uid:%v reason:%v", _func_, uid, reason)
		return
	}
}

func RecvLogoutRsp(pconfig *Config, prsp *ss.MsgLogoutRsp) {
	var _func_ = "<RecvLogoutRsp>"
	log := pconfig.Comm.Log
	log.Info("%s uid:%v reason:%v msg:%s", _func_, prsp.Uid, prsp.Reason, prsp.Msg)

	//get client key
	c_key, ok := pconfig.Uid2Ckey[prsp.Uid]
	if !ok {
		log.Err("%s no c_key found! uid:%v reason:%v msg:%s", _func_, prsp.Uid, prsp.Reason, prsp.Msg)
		return
	}

	//response
	var pmsg *cs.CSLogoutRsp
	pv, err := cs.Proto2Msg(cs.CS_PROTO_LOGOUT_RSP)
	if err != nil {
		log.Err("%s proto2msg failed! proto:%d err:%v", _func_, cs.CS_PROTO_LOGOUT_RSP, err)
		return
	}
	pmsg, ok = pv.(*cs.CSLogoutRsp)
	if !ok {
		log.Err("%s proto2msg type illegal!  proto:%d", _func_, cs.CS_PROTO_LOGOUT_RSP)
		return
	}

	//fill
	pmsg.Msg = prsp.Msg
	pmsg.Result = int(prsp.Reason)
	pmsg.Uid = prsp.Uid

	//to client
	SendToClient(pconfig, c_key, cs.CS_PROTO_LOGOUT_RSP, pmsg)

	//should check close connection positively
	switch prsp.Reason {
	case ss.USER_LOGOUT_REASON_LOGOUT_CLIENT_TIMEOUT, ss.USER_LOGOUT_REASON_LOGOUT_SERVER_KICK_BAN,
		ss.USER_LOGOUT_REASON_LOGOUT_SERVER_KICK_RECONN, ss.USER_LOGOUT_REASON_LOGOUT_SERVER_SHUT:
		CloseClient(pconfig, c_key)
	default:
		//nothing to do
	}

	//clear map
	delete(pconfig.Uid2Ckey, prsp.Uid)
	delete(pconfig.Ckey2Uid, c_key)
}

func SendRegReq(pconfig *Config, client_key int64, preq *cs.CSRegReq) {
	var _func_ = "<SendRegReq>"
	log := pconfig.Comm.Log

	log.Debug("%s send reg pkg to logic! user:%s addr:%s sex:%v role_name:%s", _func_, preq.Name, preq.Addr, preq.Sex, preq.RoleName)
	//create pkg
	var ss_msg ss.SSMsg
	pRegReq := new(ss.MsgRegReq)
	pRegReq.Name = preq.Name
	pRegReq.Pass = preq.Pass
	pRegReq.Addr = preq.Addr
	pRegReq.CKey = client_key
	pRegReq.RoleName = preq.RoleName
	pRegReq.Desc = preq.Desc
	if preq.Sex == 1 {
		pRegReq.Sex = true
	} else {
		pRegReq.Sex = false
	}

	//gen
	err := comm.FillSSPkg(&ss_msg, ss.SS_PROTO_TYPE_REG_REQ, pRegReq)
	if err != nil {
		log.Err("%s gen ss failed! err:%v ckey:%v", _func_, err, client_key)
		return
	}

	//send
	ok := SendToLogic(pconfig, &ss_msg)
	if !ok {
		log.Err("%s send failed! client_key:%v", _func_, client_key)
		return
	}
}

func RecvRegRsp(pconfig *Config, prsp *ss.MsgRegRsp) {
	var _func_ = "<RecvRegRsp>"
	log := pconfig.Comm.Log
	log.Info("%s name:%s result:%v c_key:%d", _func_, prsp.Name, prsp.Result, prsp.CKey)

	//response
	/*
		var gmsg cs.GeneralMsg
		gmsg.ProtoId = cs.CS_PROTO_REG_RSP;
		psub := new(cs.CSRegRsp)
		gmsg.SubMsg = psub
	*/
	//response
	var pmsg *cs.CSRegRsp
	pv, err := cs.Proto2Msg(cs.CS_PROTO_REG_RSP)
	if err != nil {
		log.Err("%s proto2msg failed! proto:%d err:%v", _func_, cs.CS_PROTO_REG_RSP, err)
		return
	}
	pmsg, ok := pv.(*cs.CSRegRsp)
	if !ok {
		log.Err("%s proto2msg type illegal!  proto:%d", _func_, cs.CS_PROTO_REG_RSP)
		return
	}

	//fill
	pmsg.Result = int(prsp.Result)
	pmsg.Name = prsp.Name

	//to client
	SendToClient(pconfig, prsp.CKey, cs.CS_PROTO_REG_RSP, pmsg)
	return
}
