package lib

import (
	"fmt"
	"schat/proto/ss"
	"schat/servers/comm"
)

func RecvUpdateChatReq(pconfig *Config, preq *ss.MsgUpdateChatReq, from int) {
	var _func_ = "<RecvSendChatReq>"
	log := pconfig.Comm.Log

	if preq.MsgId<=0 || preq.GrpId<=0 {
		log.Err("%s arg illegal! msg_id:%d grp_id:%d uid:%d" , _func_ , preq.MsgId , preq.GrpId , preq.Uid)
		return
	}

	//handle
	go func() {
		uid := preq.Uid
		grp_id := preq.GrpId
		//pclient
		pclient := SelectRedisClient(pconfig, REDIS_OPT_RW)
		if pclient == nil {
			log.Err("%s failed! no proper redis found! uid:%d grp_id:%d", _func_, uid, grp_id)
			return
		}
		//phead
		phead := pclient.AllocSyncCmdHead()
		if phead == nil {
			log.Err("%s alloc head failed! uid:%d grp_id:%d", _func_, uid, grp_id)
			return
		}
		defer pclient.FreeSyncCmdHead(phead)

		log.Debug("%s uid:%d grp_id:%d msg_id:%d type:%d", _func_, uid, grp_id, preq.MsgId, preq.UpdateType)
		//handle
		switch preq.UpdateType {
		case ss.UPDATE_CHAT_TYPE_UPT_CHAT_CANCEL:
			cancel_chat(pconfig , pclient , phead , preq , from)
		default:
			log.Err("%s unhandle type:%d uid:%d grp_id:%d" , _func_ , preq.UpdateType , uid , grp_id)
		}

	}()
}

/*--------------static func--------------*/
func send_upt_chat_resp(pconfig *Config , preq *ss.MsgUpdateChatReq , result ss.SS_COMMON_RESULT , from int , pold *ss.ChatMsg) {
	var _func_ = "<send_upt_chat_resp>"
	log := pconfig.Comm.Log

	log.Debug("%s uid:%d grp_id:%d result:%d type:%d msg_id:%d" , _func_ , preq.Uid , preq.GrpId , result , preq.UpdateType , preq.MsgId)
	//ss_msg
	var ss_msg ss.SSMsg
	prsp := new(ss.MsgUpdateChatRsp)
	prsp.Uid = preq.Uid
	prsp.GrpId = preq.GrpId
	prsp.MsgId = preq.MsgId
	prsp.UpdateType = preq.UpdateType
	prsp.Result = result

	if pold != nil {
		prsp.SrcType = pold.ChatType
		prsp.SrcContent = pold.Content
	}
	err := comm.FillSSPkg(&ss_msg , ss.SS_PROTO_TYPE_UPDATE_CHAT_RSP , prsp)
	if err != nil {
		log.Err("%s gen ss failed! err:%v uid:%d" , _func_ , err , prsp.Uid)
		return
	}

	//to logic
	SendToServ(pconfig , from , &ss_msg)
}

//cancel chat
func cancel_chat(pconfig *Config , pclient *comm.RedisClient , phead *comm.SyncCmdHead , preq *ss.MsgUpdateChatReq , from int) {
	var _func_ = "<cancel_chat>"
	log := pconfig.Comm.Log
	uid := preq.Uid
	grp_id := preq.GrpId
	msg_id := preq.MsgId

	//parse msg_id
	index := int((msg_id - 1) / CHAT_MSG_LIST_SIZE)
	offset := int((msg_id - 1) % CHAT_MSG_LIST_SIZE)
	result := ss.SS_COMMON_RESULT_FAILED
	pold  := new(ss.ChatMsg)
	for {
		//Get Msg
		tab_name := fmt.Sprintf(FORMAT_TAB_CHAT_MSG_LIST, grp_id, index)
		res, err := pclient.RedisExeCmdSync(phead, "LINDEX", tab_name, offset)
		if err != nil {
			log.Err("%s lindex chat msg failed! uid:%d grp_id:%d err:%v tab:%s msg_id:%d", _func_, uid, grp_id, err, tab_name,
				msg_id)
			result = ss.SS_COMMON_RESULT_FAILED
			break
		}
		str, err := comm.Conv2String(res)
		if err != nil {
			log.Err("%s lindex convert res failed! uid:%d grp_id:%d err:%v res:%v", _func_, uid, grp_id, err, res)
			result = ss.SS_COMMON_RESULT_FAILED
			break
		}

		if len(str) == 0 {
			log.Err("%s result empty! uid:%d grp_id:%d err:%v res:%v msg_id:%d", _func_, uid, grp_id, err, res , msg_id)
			result = ss.SS_COMMON_RESULT_NOEXIST
			break
		}

		//Decrypt and Unpack
		pchat := UnpackChat(pconfig , uid , str)
		if pchat == nil {
			log.Err("%s unpack chat fail! uid:%d grp_id:%d err:%v res:%v msg_id:%d", _func_, uid, grp_id, err, res , msg_id)
			result = ss.SS_COMMON_RESULT_FAILED
			break
		}

		//Check Sender
		if pchat.SenderUid != uid {
			log.Err("%s not send this chat! uid:%d grp_id:%d err:%v res:%v msg_id:%d sender:%d", _func_, uid, grp_id, err, res , msg_id ,
				pchat.SenderUid)
			result = ss.SS_COMMON_RESULT_PERMISSION
			break
		}

		//Reset Chat
		pold.ChatType = pchat.ChatType
		pold.Content = pchat.Content
		pchat.ChatType = ss.CHAT_MSG_TYPE_CHAT_TYPE_TEXT //reset
		pchat.Content = ""
		pchat.ChatFlag = ss.CHAT_MSG_FLAG_CHAT_FLAG_CANCELED

		//Pack and Encrypt
        enc_chat := PackChat(pconfig , uid , pchat)
        if enc_chat == nil {
			log.Err("%s pack chat fail! uid:%d grp_id:%d err:%v res:%v msg_id:%d", _func_, uid, grp_id, err, res , msg_id)
			result = ss.SS_COMMON_RESULT_FAILED
			break
		}

		//STORE CHAT
		_ , err = pclient.RedisExeCmdSync(phead, "LSET", tab_name, offset , string(enc_chat))
		if err != nil {
			log.Err("%s lset chat msg failed! uid:%d grp_id:%d err:%v tab:%s msg_id:%d", _func_, uid, grp_id, err, tab_name,
				msg_id)
			result = ss.SS_COMMON_RESULT_FAILED
			break
		}

		result = ss.SS_COMMON_RESULT_SUCCESS
		break
	}

	log.Info("%s result:%d uid:%d grp_id:%d msg_id:%d" , _func_ , result , uid , grp_id , msg_id)
	//back to resp
	send_upt_chat_resp(pconfig , preq , result , from , pold)
}
