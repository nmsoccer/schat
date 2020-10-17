package lib

import (
	"schat/proto/ss"
)

func RecvDispMsg(pconfig *Config, pdisp *ss.MsgDisp) {
	var _func_ = "<RecvDispMsg>"
	log := pconfig.Comm.Log

	log.Debug("%s disp_proto:%d disp_from:%d target:%d spec:%d method:%d", _func_, pdisp.ProtoType, pdisp.FromServer, pdisp.Target,
		pdisp.SpecServer, pdisp.Method)

	//dispatch
	switch pdisp.ProtoType {
	case ss.DISP_PROTO_TYPE_DISP_HELLO:
		pmsg := pdisp.GetHello()
		log.Info("%s hello! content:%s", _func_, pmsg.Content)
	case ss.DISP_PROTO_TYPE_DISP_KICK_DUPLICATE_USER:
		pmsg := pdisp.GetKickDupUser()
		RecvDupUserKick(pconfig, pmsg, int(pdisp.FromServer))
	case ss.DISP_PROTO_TYPE_DISP_APPLY_GROUP_RSP:
		pmsg := pdisp.GetApplyGroupRsp()
		RecvApplyGroupRsp(pconfig, pmsg)
	case ss.DISP_PROTO_TYPE_DISP_APPLY_GROUP_NOTIFY:
		pmsg := pdisp.GetApplyGroupNotify()
		RecvApplyGroupNotify(pconfig, pmsg)
	case ss.DISP_PROTO_TYPE_DISP_COMMON_NOTIFY:
		pmsg := pdisp.GetCommonNotify()
		RecvCommonNotify(pconfig, pmsg, int(pdisp.FromServer))
	case ss.DISP_PROTO_TYPE_DISP_ENTER_GROUP_RSP:
		pmsg := pdisp.GetEnterGroupRsp()
		RecvEnterGroupRsp(pconfig, pmsg)
	case ss.DISP_PROTO_TYPE_DISP_SEND_CHAT_RSP:
		pmsg := pdisp.GetSendChatRsp()
		RecvSendChatRsp(pconfig, pmsg)
	case ss.DISP_PROTO_TYPE_DISP_SYNC_GROUP_INFO:
		pmsg := pdisp.GetSyncGroupInfo()
		RecvSyncGroupInfo(pconfig, pmsg)
	case ss.DISP_PROTO_TYPE_DISP_CHG_GROUP_ATTR_RSP:
		pmsg := pdisp.GetChgGroupAttrRsp()
		RecvChgGroupAttrRsp(pconfig, pmsg)
	default:
		log.Err("%s convert disp-msg failed! unkown disp_proto:%d", _func_, pdisp.ProtoType)
		return
	}

	return
}

func RecvCommonNotify(pconfig *Config, pnotify *ss.MsgCommonNotify, src_serv int) {
	var _func_ = "<RecvCommonNotify>"
	log := pconfig.Comm.Log

	//log.Debug("%s uid:%d type:%d int_v:%d" , _func_ , pnotify.Uid , pnotify.NotifyType , pnotify.IntV)
	switch pnotify.NotifyType {
	case ss.COMMON_NOTIFY_TYPE_NOTIFY_NEW_AUDIT:
		RecvApplyGroupAuditNotify(pconfig, pnotify)
	case ss.COMMON_NOTIFY_TYPE_NOTIFY_NEW_MSG:
		RecvNewMsgNotify(pconfig, pnotify)
	case ss.COMMON_NOTIFY_TYPE_NOTIFY_UPLOAD_FILE:
		RecvUploadFileNotify(pconfig, pnotify, src_serv)
	case ss.COMMON_NOTIFY_TYPE_NOTIFY_SERV_LOAD:
		RecvLoadNotify(pconfig, pnotify)
	case ss.COMMON_NOTIFY_TYPE_NOTIFY_DEL_GROUP:
		RecvDelGroupNotify(pconfig, pnotify)
	case ss.COMMON_NOTIFY_TYPE_NOTIFY_FILE_ADDR:
		RecvFileAddrNotify(pconfig, pnotify)
	case ss.COMMON_NOTIFY_TYPE_NOTIFY_KICK_GROUP:
		RecvKickGroupNotify(pconfig, pnotify)
	case ss.COMMON_NOTIFY_TYPE_NOTIFY_ADD_MEMBER, ss.COMMON_NOTIFY_TYPE_NOTIFY_DEL_MEMBER:
		RecvChgMemNotify(pconfig, pnotify)
	default:
		log.Err("%s unhandled notify:%d uid:%d", _func_, pnotify.NotifyType, pnotify.Uid)
	}
}
