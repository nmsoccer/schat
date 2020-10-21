package lib

import (
	"schat/proto/ss"
	"schat/servers/comm"
)

func RecvDispMsg(pconfig *Config, pdisp *ss.MsgDisp) {
	var _func_ = "<RecvDispMsg>"
	log := pconfig.Comm.Log

	//log.Debug("%s disp_proto:%d disp_from:%d target:%d spec:%d method:%d", _func_, pdisp.ProtoType, pdisp.FromServer, pdisp.Target,
	//	pdisp.SpecServer, pdisp.Method)

	switch pdisp.ProtoType {
	case ss.DISP_PROTO_TYPE_DISP_HELLO:
		pmsg := pdisp.GetHello()
		log.Debug("%s hello! content:%s", _func_, pmsg.Content)
	case ss.DISP_PROTO_TYPE_DISP_COMMON_NOTIFY:
		pmsg := pdisp.GetCommonNotify()
		RecvCommNotify(pconfig, pmsg, int(pdisp.FromServer))
	default:
		log.Err("%s  unkown disp_proto:%d", _func_, pdisp.ProtoType)
	}

	return
}

func RecvCommNotify(pconfig *Config, pnotify *ss.MsgCommonNotify, src_serv int) {
	var _func_ = "<RecvCommNotify>"
	log := pconfig.Comm.Log

	//log.Debug("%s notify:%d src_serv:%d" , _func_ , pnotify.NotifyType , src_serv)
	switch pnotify.NotifyType {
	case ss.COMMON_NOTIFY_TYPE_NOTIFY_USER_LOGIN:
		RecvLoginNotify(pconfig, pnotify.Uid, src_serv)
	case ss.COMMON_NOTIFY_TYPE_NOTIFY_USER_LOGOUT:
		RecvLogoutNotify(pconfig, pnotify.Uid, src_serv)
	case ss.COMMON_NOTIFY_TYPE_NOTIFY_NEW_AUDIT:
		RecvApplyGroupAuditNotify(pconfig , pnotify)
	case ss.COMMON_NOTIFY_TYPE_NOTIFY_NEW_MSG:
		RecvNewMsgNotify(pconfig, pnotify)
	case ss.COMMON_NOTIFY_TYPE_NOTIFY_UPLOAD_FILE:
		RecvUploadFileNotify(pconfig, pnotify, src_serv)
	case ss.COMMON_NOTIFY_TYPE_NOTIFY_DEL_GROUP:
		RecvDelGroupNotify(pconfig, pnotify)
	case ss.COMMON_NOTIFY_TYPE_NOTIFY_BATCH_USER_ONLINE:
		RecvBatchOnLineNotify(pconfig, pnotify, src_serv)
	case ss.COMMON_NOTIFY_TYPE_NOTIFY_KICK_GROUP:
		RecvKickGroupNotify(pconfig, pnotify)
	case ss.COMMON_NOTIFY_TYPE_NOTIFY_ADD_MEMBER, ss.COMMON_NOTIFY_TYPE_NOTIFY_DEL_MEMBER:
		RecvChgMemberNotify(pconfig, pnotify)
	default:
		log.Err("%s unhandled notify:%d src:%d", _func_, pnotify.NotifyType, src_serv)
	}
}

func SendSpecCommNotify(pconfig *Config, disp_target ss.DISP_MSG_TARGET, spec_serv int, pnotify *ss.MsgCommonNotify) {
	var _func_ = "<SendSpecCommNotify>"
	log := pconfig.Comm.Log

	//GEN SS
	pss_msg, err := comm.GenDispMsg(disp_target, ss.DISP_MSG_METHOD_SPEC, ss.DISP_PROTO_TYPE_DISP_COMMON_NOTIFY,
		spec_serv, pconfig.ProcId, 0, pnotify)
	if err != nil {
		log.Err("%s gen ss failed! err:%v uid:%d grp_id:%d disp_target:%d spec:%d", _func_, err, pnotify.Uid, pnotify.GrpId,
			disp_target, spec_serv)
		return
	}

	//Send
	SendToDisp(pconfig, 0, pss_msg)
}
