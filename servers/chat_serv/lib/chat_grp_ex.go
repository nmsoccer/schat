package lib

import (
	"schat/proto/ss"
	"schat/servers/comm"
)

const (
	INIT_GROUP_SCORE = 100
)

func RecvChgGroupAttrReq(pconfig *Config, preq *ss.MsgChgGroupAttrReq, logic_serv int) {
	var _func_ = "<RecvChgGroupAttrReq>"
	log := pconfig.Comm.Log
	uid := preq.Uid
	grp_id := preq.GrpId

	//group info
	pgrp_info := GetGroupInfo(pconfig, grp_id)
	if pgrp_info != nil {
		//online
		log.Debug("%s group online! attr:%d uid:%d grp_id:%d", _func_, preq.Attr, uid, grp_id)
		ChgGroupAttr(pconfig, preq, logic_serv)
		return
	}

	//load group first
	log.Debug("%s group offline! will load first! attr:%d uid:%d grp_id:%d", _func_, preq.Attr, uid, grp_id)
	LoadGroup(pconfig, uid, grp_id, ss.LOAD_GROUP_REASON_LOAD_GRP_CHG_GROUP_ATTR, int64(logic_serv), preq)
	return
}

func RecvChgGroupAttrRsp(pconfig *Config, prsp *ss.MsgChgGroupAttrRsp) {
	var _func_ = "<RecvChgGroupAttrRsp>"
	log := pconfig.Comm.Log
	uid := prsp.Uid
	grp_id := prsp.GrpId
	attr := prsp.Attr

	log.Info("%s result:%d grp_id:%d uid:%d attr:%d", _func_, prsp.Result, grp_id, uid, attr)
	for {
		//success
		if prsp.Result == ss.SS_COMMON_RESULT_SUCCESS {
			break
		}

		//failed may reverse data
		pgrp_info := GetGroupInfo(pconfig, grp_id)
		if pgrp_info == nil {
			log.Err("%s failed but grp offline! data may not consistent attr:%d grp_id:%d", _func_, attr, grp_id)
			break
		}

		//reverse
		switch attr {
		case ss.GROUP_ATTR_TYPE_GRP_ATTR_VISIBLE:
			pgrp_info.db_group_info.BlobInfo.Visible = 0
		case ss.GROUP_ATTR_TYPE_GRP_ATTR_INVISIBLE:
			pgrp_info.db_group_info.BlobInfo.Visible = 1
		default:
			//nothing
		}
		break
	}

	//ss
	pss_msg, err := comm.GenDispMsg(ss.DISP_MSG_TARGET_LOGIC_SERVER, ss.DISP_MSG_METHOD_SPEC, ss.DISP_PROTO_TYPE_DISP_CHG_GROUP_ATTR_RSP,
		int(prsp.Occupy), pconfig.ProcId, 0, prsp)
	if err != nil {
		log.Err("%s gen ss failed! err:%v uid:%d grp_id:%d attr:%d", _func_, err, uid, grp_id, attr)
		return
	}

	//send
	SendToDisp(pconfig, 0, pss_msg)
}

func ChgGroupAttr(pconfig *Config, preq *ss.MsgChgGroupAttrReq, logic_serv int) {
	var _func_ = "<ChgGroupAttr>"
	log := pconfig.Comm.Log
	uid := preq.Uid
	grp_id := preq.GrpId
	direct_back := false

	//grp info
	pgrp_info := GetGroupInfo(pconfig, grp_id)
	if pgrp_info == nil {
		log.Err("%s grp offline! grp_id:%d", _func_, grp_id)
		return
	}

	//handle
	switch preq.Attr {
	case ss.GROUP_ATTR_TYPE_GRP_ATTR_VISIBLE:
		if pgrp_info.db_group_info.BlobInfo.Visible > 0 {
			log.Info("%s already visible! uid:%d grp_id:%d", _func_, uid, grp_id)
			direct_back = true
			break
		}
		//set visible
		log.Info("%s will set group visible! uid:%d grp_id:%d", _func_, uid, grp_id)
		pgrp_info.db_group_info.BlobInfo.Visible = 1
		preq.IntV = int64(pgrp_info.db_group_info.BlobInfo.VisibleScore)
	case ss.GROUP_ATTR_TYPE_GRP_ATTR_INVISIBLE:
		if pgrp_info.db_group_info.BlobInfo.Visible <= 0 {
			log.Info("%s already invisible! uid:%d grp_id:%d", _func_, uid, grp_id)
			direct_back = true
			break
		}
		//set visible
		log.Info("%s will set group invisible! uid:%d grp_id:%d", _func_, uid, grp_id)
		pgrp_info.db_group_info.BlobInfo.Visible = 0
	default:
		log.Err("%s illegal attr:%d uid:%d grp_id:%d", _func_, preq.Attr, uid, grp_id)
		return
	}

	//direct back
	if direct_back {
		//resp
		prsp := new(ss.MsgChgGroupAttrRsp)
		prsp.Uid = uid
		prsp.GrpId = grp_id
		prsp.Attr = preq.Attr
		prsp.Result = ss.SS_COMMON_RESULT_SUCCESS
		pss_msg, err := comm.GenDispMsg(ss.DISP_MSG_TARGET_LOGIC_SERVER, ss.DISP_MSG_METHOD_SPEC, ss.DISP_PROTO_TYPE_DISP_CHG_GROUP_ATTR_RSP,
			logic_serv, pconfig.ProcId, 0, prsp)
		if err != nil {
			log.Err("%s logic ss failed! err:%v uid:%d grp_id:%d attr:%d", _func_, err, uid, grp_id, prsp.Attr)
			return
		}

		SendToDisp(pconfig, 0, pss_msg)
		return
	}

	//to db
	var ss_msg ss.SSMsg
	preq.StrV = pgrp_info.db_group_info.GroupName
	preq.Occupy = int64(logic_serv)
	err := comm.FillSSPkg(&ss_msg, ss.SS_PROTO_TYPE_CHG_GROUP_ATTR_REQ, preq)
	if err != nil {
		log.Err("%s gen ss failed! err:%v uid:%d grp_id:%d attr:%d", _func_, err, uid, grp_id, preq.Attr)
		return
	}

	SendToDb(pconfig, &ss_msg)
}
