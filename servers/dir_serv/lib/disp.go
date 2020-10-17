package lib

import (
	"schat/proto/ss"
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
	case ss.COMMON_NOTIFY_TYPE_NOTIFY_SERV_LOAD:
		RecvLoadNotify(pconfig, pnotify)
	case ss.COMMON_NOTIFY_TYPE_NOTIFY_FILE_ADDR:
		RecvFileAddrNotify(pconfig, pnotify, src_serv)
	case ss.COMMON_NOTIFY_TYPE_NOTIFY_FILE_TOKEN:
		RecvFileTokenNotify(pconfig, pnotify, src_serv)
	case ss.COMMON_NOTIFY_TYPE_NOTIFY_UPLOAD_FILE:
		RecvUploadFileNotify(pconfig, pnotify)
	default:
		log.Err("%s unhandled notify:%d src:%d", _func_, pnotify.NotifyType, src_serv)
	}
}
