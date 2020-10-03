package lib

import (
    "schat/proto/ss"
	"schat/servers/comm"
)

func RecvPingReq(pconfig *Config , preq *ss.MsgPingReq , from int) {
	var _func_ = "<RecvPingReq>";
	log := pconfig.Comm.Log;

	log.Debug("%s get ping req! client_key:%v ts:%v from:%d" , _func_ , preq.ClientKey , preq.Ts , from);
	//Back
	var  ss_msg ss.SSMsg;
	pPingRsp := new(ss.MsgPingRsp);
	pPingRsp.ClientKey = preq.ClientKey;
	pPingRsp.Ts = preq.Ts;


	//encode
	err := comm.FillSSPkg(&ss_msg , ss.SS_PROTO_TYPE_PING_RSP , pPingRsp);
	if err != nil {
		log.Err("%s gen ss failed! err:%v key:%v" , _func_ , err , preq.ClientKey);
		return;
	}

	//sendback
	ok := SendToConnect(pconfig, &ss_msg);
	if !ok {
		log.Err("%s send back failed! key:%v" , _func_ , preq.ClientKey);
		return;
	}
	log.Debug("%s send back success! key:%v" , _func_ , preq.ClientKey);
	return;
}

//from conn
func RecvLoadNotify(pconfig *Config , pnotify *ss.MsgCommonNotify) {
	var _func_ = "<RecvLoadNotify>"
	log := pconfig.Comm.Log

	//gen ss
	pss_msg , err := comm.GenDispMsg(ss.DISP_MSG_TARGET_DIR_SERVER , ss.DISP_MSG_METHOD_ALL , ss.DISP_PROTO_TYPE_DISP_COMMON_NOTIFY , 0 ,
		pconfig.ProcId , 0 , pnotify)
	if err != nil {
		log.Err("%s gen ss failed! err:%v" , _func_ , err)
		return
	}

	//to dir
	SendToDisp(pconfig , 0 , pss_msg)
}


