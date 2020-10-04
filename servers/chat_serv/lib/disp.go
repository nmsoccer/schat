package lib

import (
	"schat/proto/ss"
)

func RecvDispMsg(pconfig *Config, pdisp *ss.MsgDisp) {
	var _func_ = "<RecvDispMsg>"
	log := pconfig.Comm.Log

	log.Debug("%s disp_proto:%d disp_from:%d target:%d spec:%d method:%d", _func_, pdisp.ProtoType, pdisp.FromServer, pdisp.Target,
		pdisp.SpecServer, pdisp.Method)

	switch pdisp.ProtoType {
	case ss.DISP_PROTO_TYPE_DISP_HELLO:
		pmsg := pdisp.GetHello()
		log.Debug("%s hello! content:%s", _func_, pmsg.Content)
	case ss.DISP_PROTO_TYPE_DISP_APPLY_GROUP_REQ:
        pmsg := pdisp.GetApplyGroupReq()
        RecvApplyGroupReq(pconfig , pmsg , pdisp)
	case ss.DISP_PROTO_TYPE_DISP_ENTER_GROUP_REQ:
		pmsg := pdisp.GetEnterGroupReq()
		RecvEnterGroupReq(pconfig , pmsg , pdisp.FromServer)
	case ss.DISP_PROTO_TYPE_DISP_SEND_CHAT_REQ:
		pmsg := pdisp.GetSendChatReq()
		RecvSendChatReq(pconfig , pmsg , int(pdisp.FromServer))
	case ss.DISP_PROTO_TYPE_DISP_COMMON_NOTIFY:
		pmsg := pdisp.GetCommonNotify()
		RecvCommNotify(pconfig , pmsg , int(pdisp.FromServer))
	default:
		log.Err("%s convert disp-msg fail! unkown disp_proto:%d", _func_, pdisp.ProtoType)
	}

	return
}


func RecvCommNotify(pconfig *Config , pnotify *ss.MsgCommonNotify , src_serv int) {
	var _func_ = "<RecvCommNotify>"
	log := pconfig.Comm.Log

	//log.Debug("%s notify:%d src_serv:%d" , _func_ , pnotify.NotifyType , src_serv)
	switch pnotify.NotifyType {
	case ss.COMMON_NOTIFY_TYPE_NOTIFY_EXIT_GROUP:
		RecvExitGroupNotify(pconfig , pnotify)
	case ss.COMMON_NOTIFY_TYPE_NOTIFY_DEL_GROUP:
		RecvDelGroupNotify(pconfig , pnotify)
	default:
		log.Err("%s unhandled notify:%d src:%d" , _func_ , pnotify.NotifyType , src_serv)
	}
}