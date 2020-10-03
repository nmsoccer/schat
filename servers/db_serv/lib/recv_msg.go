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
		//log.Debug("recved:%d sender:%d v:%v" , recv , pmsg.sender , msg);
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
		case ss.SS_PROTO_TYPE_LOGIN_REQ:
			RecvUserLoginReq(pconfig, ss_msg.GetLoginReq(), pmsg.sender)
		case ss.SS_PROTO_TYPE_LOGOUT_REQ:
			RecvUserLogoutReq(pconfig, ss_msg.GetLogoutReq(), pmsg.sender)
		case ss.SS_PROTO_TYPE_REG_REQ:
			RecvRegReq(pconfig , ss_msg.GetRegReq() , pmsg.sender)
		case ss.SS_PROTO_TYPE_CREATE_GROUP_REQ:
			RecvCreateGroupReq(pconfig , ss_msg.GetCreateGroupReq() , pmsg.sender)
		case ss.SS_PROTO_TYPE_APPLY_GROUP_REQ:
			RecvApplyGroupReq(pconfig , ss_msg.GetApplyGroupReq() , pmsg.sender)
		case ss.SS_PROTO_TYPE_FETCH_APPLY_GROUP_REQ:
			RecvFetchApplyGroupReq(pconfig , ss_msg.GetFetchApplyReq() , pmsg.sender)
		case ss.SS_PROTO_TYPE_APPLY_GROUP_AUDIT:
			RecvApplyGroupAudit(pconfig , ss_msg.GetApplyGroupAudit() , pmsg.sender)
		case ss.SS_PROTO_TYPE_FETCH_AUDIT_GROUP_REQ:
			RecvFetchAuditGroupReq(pconfig , ss_msg.GetFetchAuditReq() , pmsg.sender)
		case ss.SS_PROTO_TYPE_ENTER_GROUP_REQ:
			RecvEnterGroupReq(pconfig , ss_msg.GetEnterGroupReq() , pmsg.sender)
		case ss.SS_PROTO_TYPE_LOAD_GROUP_REQ:
			RecvLoadGroupReq(pconfig , ss_msg.GetLoadGroupReq() , pmsg.sender)
		case ss.SS_PROTO_TYPE_SEND_CHAT_REQ:
			RecvSendChatReq(pconfig , ss_msg.GetSendChatReq() , pmsg.sender)
		case ss.SS_PROTO_TYPE_SAVE_GROUP_REQ:
			RecvSaveChatGroupReq(pconfig , ss_msg.GetSaveGroupReq() , pmsg.sender)
		case ss.SS_PROTO_TYPE_FETCH_CHAT_REQ:
			RecvFetchChatReq(pconfig , ss_msg.GetFetchChatReq() , pmsg.sender)
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
