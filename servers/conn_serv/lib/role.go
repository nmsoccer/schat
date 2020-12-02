package lib

import (
	"schat/proto/cs"
	"schat/proto/ss"
	"schat/servers/comm"
)

func SendFetchUserProfileReq(pconfig *Config, uid int64, pfetch *cs.CSFetchUserProfileReq) {
	var _func_ = "<SendFetchUserProfileReq>"
	log := pconfig.Comm.Log

	//check
	if len(pfetch.TargetList) <= 0 {
		log.Err("%s fetch nothing! uid:%d", _func_, uid)
		return
	}

	//ss
	var ss_msg ss.SSMsg
	preq := new(ss.MsgFetchUserProfileReq)
	preq.Uid = uid
	preq.TargetList = pfetch.TargetList

	err := comm.FillSSPkg(&ss_msg, ss.SS_PROTO_TYPE_FETCH_USER_PROFILE_REQ, preq)
	if err != nil {
		log.Err("%s gen ss failed! err:%v uid:%d", _func_, err, uid)
		return
	}

	//to logic
	SendToLogic(pconfig, &ss_msg)
}

func RecvFetchUserProfileRsp(pconfig *Config, prsp *ss.MsgFetchUserProfileRsp) {
	var _func_ = "<RecvFetchUserProfileRsp>"
	log := pconfig.Comm.Log
	uid := prsp.Uid

	//c_key
	c_key := GetClientKey(pconfig, uid)
	if c_key <= 0 {
		log.Err("%s user offline! uid:%d", _func_, uid)
		return
	}

	//check result
	if prsp.Result != ss.SS_COMMON_RESULT_SUCCESS {
		log.Err("%s result:%d uid:%d", _func_, prsp.Result, uid)
		return
	}

	//length
	if len(prsp.Profiles) <= 0 {
		log.Err("%s profile list 0! uid:%d", _func_, uid)
		return
	}

	//cs
	pv, err := cs.Proto2Msg(cs.CS_PROTO_FETCH_USER_PROFILE_RSP)
	if err != nil {
		log.Err("%s get msg fail! uid:%d err:%v", _func_, uid, err)
		return
	}
	pmsg, ok := pv.(*cs.CSFetchUserProfileRsp)
	if !ok {
		log.Err("%s not CSFetchUserProfileRsp! uid:%d", _func_, uid)
		return
	}

	//fill msg
	var tuid int64
	var ss_info *ss.UserProfile
	var cs_info *cs.UserProfile
	pmsg.Profiles = make(map[int64]*cs.UserProfile)
	for tuid, ss_info = range prsp.Profiles {
		if ss_info == nil {
			pmsg.Profiles[tuid] = nil
			continue
		}
		cs_info = new(cs.UserProfile)
		cs_info.Uid = ss_info.Uid
		cs_info.Addr = ss_info.Addr
		cs_info.Sex = uint8(ss_info.Sex)
		cs_info.Level = ss_info.Level
		cs_info.Name = ss_info.Name
		cs_info.HeadUrl = ss_info.HeadUrl
		cs_info.Desc = ss_info.UserDesc
		pmsg.Profiles[tuid] = cs_info
	}

	//to client
	SendToClient(pconfig, c_key, cs.CS_PROTO_FETCH_USER_PROFILE_RSP, pmsg)
}

func SendUpdateUserReq(pconfig *Config, uid int64, pupdate *cs.CSUpdateUserReq) {
	var _func_ = "<SendUpdateUserReq>"
	log := pconfig.Comm.Log


	//ss
	var ss_msg ss.SSMsg
	preq := new(ss.MsgUpdateUserReq)
	preq.Uid = uid
	preq.RoleName = pupdate.RoleName
	preq.Addr = pupdate.Addr
	preq.Desc = pupdate.Desc
	preq.Passwd = pupdate.Passwd

	err := comm.FillSSPkg(&ss_msg, ss.SS_PROTO_TYPE_UPDATE_USER_REQ, preq)
	if err != nil {
		log.Err("%s gen ss failed! err:%v uid:%d", _func_, err, uid)
		return
	}

	//to logic
	SendToLogic(pconfig, &ss_msg)
}

func RecvUpdateUserRsp(pconfig *Config , prsp *ss.MsgUpdateUserRsp) {
	var _func_ = "<RecvUpdateUserRsp>"
	log := pconfig.Comm.Log
	uid := prsp.Uid

	//c_key
	c_key := GetClientKey(pconfig, uid)
	if c_key <= 0 {
		log.Err("%s user offline! uid:%d", _func_, uid)
		return
	}

	//cs
	pv, err := cs.Proto2Msg(cs.CS_PROTO_UPDATE_USER_RSP)
	if err != nil {
		log.Err("%s get msg fail! uid:%d err:%v", _func_, uid, err)
		return
	}
	pmsg, ok := pv.(*cs.CSUpdateUserRsp)
	if !ok {
		log.Err("%s not CSFetchUserProfileRsp! uid:%d", _func_, uid)
		return
	}

	log.Info("%s result:%d uid:%d" , _func_ , prsp.Result , uid)
	//fill info
	pmsg.Result = int(prsp.Result)
	pmsg.RoleName = prsp.RoleName
	pmsg.Addr = prsp.Addr
	pmsg.Desc = prsp.Desc
	pmsg.Passwd = prsp.Passwd

	//to client
	SendToClient(pconfig , c_key , cs.CS_PROTO_UPDATE_USER_RSP , pmsg)
}