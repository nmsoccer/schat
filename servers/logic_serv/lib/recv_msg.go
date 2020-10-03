package lib

import (
	"schat/proto/ss"
	"time"
)

const (
	MESSAGE_LEN = ss.MAX_SS_MSG_SIZE //200k
)

type Msg struct {
	sender int
	msg    []byte
}

var pmsg *Msg

func init() {
	pmsg = new(Msg)
	pmsg.msg = make([]byte, MESSAGE_LEN)
}

func RecvMsg(pconfig *Config) int64 {
	var _func_ = "RecvMsg"
	var start_time int64

	//init
	pmsg.msg = pmsg.msg[0:cap(pmsg.msg)]
	var recv int
	var log = pconfig.Comm.Log
	var proc = pconfig.Comm.Proc
	var msg = pmsg.msg
	var handle_pkg = 0

	start_time = time.Now().UnixNano()
	//keep recving
	for {
		msg = msg[0:cap(msg)]
		recv = proc.Recv(msg, cap(msg), &(pmsg.sender))
		if recv < 0 { //no package
			break
		}

		handle_pkg++
		//unpack
		msg = msg[0:recv]
		var ss_msg = new(ss.SSMsg)
		err := ss.UnPack(msg, ss_msg)
		if err != nil {
			log.Err("unpack failed! err:%v", err)
			continue
		}
		//log.Debug("unpack success! v:%v and %v" , *ss_req , *(ss_req.GetHeartBeat()));

		//dispatch
		switch ss_msg.ProtoType {
		case ss.SS_PROTO_TYPE_HEART_BEAT_REQ:
			RecvHeartBeatReq(pconfig, ss_msg.GetHeartBeatReq(), pmsg.sender)
		case ss.SS_PROTO_TYPE_PING_REQ:
			RecvPingReq(pconfig, ss_msg.GetPingReq(), pmsg.sender)
		case ss.SS_PROTO_TYPE_LOGIN_REQ:
			RecvLoginReq(pconfig, ss_msg.GetLoginReq(), msg, pmsg.sender)
		case ss.SS_PROTO_TYPE_LOGIN_RSP:
			RecvLoginRsp(pconfig, ss_msg.GetLoginRsp(), msg)
		case ss.SS_PROTO_TYPE_LOGOUT_REQ:
			RecvLogoutReq(pconfig, ss_msg.GetLogoutReq())
		case ss.SS_PROTO_TYPE_REG_REQ:
			DirecToDb(pconfig , ss_msg.ProtoType , pmsg.sender , msg);
		case ss.SS_PROTO_TYPE_REG_RSP:
			DirecToConnect(pconfig , ss_msg.ProtoType , pmsg.sender , msg);
		case ss.SS_PROTO_TYPE_CREATE_GROUP_REQ:
			RecvCreateGroupReq(pconfig , ss_msg.GetCreateGroupReq() , msg)
		case ss.SS_PROTO_TYPE_CREATE_GROUP_RSP:
			RecvCreateGroupRsp(pconfig , ss_msg.GetCreateGroupRsp() , msg)
		case ss.SS_PROTO_TYPE_USE_DISP_PROTO:
			RecvDispMsg(pconfig , ss_msg.GetMsgDisp())
		case ss.SS_PROTO_TYPE_APPLY_GROUP_REQ:
			RecvApplyGroupReq(pconfig , ss_msg.GetApplyGroupReq())
		case ss.SS_PROTO_TYPE_FETCH_APPLY_GROUP_RSP:
			RecvFetchApplyGroupRsp(pconfig , ss_msg.GetFetchApplyRsp())
		case ss.SS_PROTO_TYPE_APPLY_GROUP_AUDIT:
			RecvApplyGroupAudit(pconfig , ss_msg.GetApplyGroupAudit() , msg)
		case ss.SS_PROTO_TYPE_COMMON_NOTIFY:
			RecvCommonNotify(pconfig , ss_msg.GetCommonNotify() , pmsg.sender)
		case ss.SS_PROTO_TYPE_FETCH_AUDIT_GROUP_RSP:
			RecvFetchAuditGroupRsp(pconfig , ss_msg.GetFetchAuditRsp())
		case ss.SS_PROTO_TYPE_SEND_CHAT_REQ:
			RecvSendChatReq(pconfig , ss_msg.GetSendChatReq())
		case ss.SS_PROTO_TYPE_FETCH_CHAT_RSP:
			RecvFetchChatRsp(pconfig , ss_msg.GetFetchChatRsp())
		default:
			log.Err("%s fail! unknown proto type:%v", _func_, ss_msg.ProtoType)
		}
	}

	//return
	if handle_pkg == 0 {
		return 0
	} else {
		return (time.Now().UnixNano() - start_time) / 1000000 //millisec
	}
}

func DirecToDb(pconfig *Config , proto_id ss.SS_PROTO_TYPE , from int , msg []byte) {
	var _func_ = "<DirecToDb>";
	log := pconfig.Comm.Log;

	log.Debug("%s proto:%v from:%d" , _func_ , proto_id , from);
	//to db
	ok := SendToDb(pconfig, msg)
	if !ok {
		log.Err("%s send failed! proto:%v from:%d", _func_ , proto_id , from);
		return
	}
}

func DirecToConnect(pconfig *Config , proto_id ss.SS_PROTO_TYPE , from int , msg []byte) {
	var _func_ = "<DirecToConnect>";
	log := pconfig.Comm.Log;

	log.Debug("%s proto:%v from:%d" , _func_ , proto_id , from);
	ok := SendToConnect(pconfig, msg)
	if !ok {
		log.Err("%s send to connect failed! proto:%v from:%d", _func_, proto_id , from);
		return
	}
}