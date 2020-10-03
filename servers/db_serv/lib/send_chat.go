package lib

import (
	"crypto/des"
	"fmt"
	lnet "schat/lib/net"
	"schat/proto/ss"
	"schat/servers/comm"
	"time"
)

func RecvSendChatReq(pconfig *Config , preq *ss.MsgSendChatReq , from int) {
	var _func_ = "<RecvSendChatReq>"
	log := pconfig.Comm.Log

	//check arg
	if preq.ChatMsg == nil {
		log.Err("%s chat msg nil! uid:%d" , _func_ , preq.Uid)
		return
	}

	//save
    go func() {
    	uid := preq.Uid
    	grp_id := preq.ChatMsg.GroupId
    	curr_ts := time.Now().Unix()
    	phead := pconfig.RedisClient.AllocSyncCmdHead()
    	if phead == nil {
    		log.Err("%s alloc head failed! uid:%d grp_id:%d" , _func_ , uid , grp_id)
    		return
		}
		defer pconfig.RedisClient.FreeSyncCmdHead(phead)

		log.Debug("%s uid:%d grp_id:%d raw_msg_id:%d tem_id:%d" , _func_ , uid , grp_id , preq.ChatMsg.MsgId , preq.TempId)
    	//Fill Chat Msg
    	preq.ChatMsg.SendTs = curr_ts

    	//Pack
    	coded , err := ss.Pack(preq.ChatMsg)
    	if err != nil {
    		log.Err("%s encode chat msg failed! uid:%d grp_id:%d err:%v" , _func_ , uid , grp_id , err)
    		return
		}

		//Encrypt
		coded , err = lnet.DesEncrypt(nil , coded , []byte(CHAT_MSG_DES_KEY))
		if err != nil {
			log.Err("%s encrypt chat msg failed! uid:%d grp_id:%d err:%v" , _func_ , uid , grp_id , err)
			return
		}

    	//Append to Chat List
		index := int((preq.ChatMsg.MsgId-1) / CHAT_MSG_LIST_SIZE)
		offset := 0
    	for {
			tab_name := fmt.Sprintf(FORMAT_TAB_CHAT_MSG_LIST, grp_id, index)
			res, err := pconfig.RedisClient.RedisExeCmdSync(phead, "RPUSH", tab_name, coded)
			if err != nil {
				log.Err("%s push chat msg failed! uid:%d grp_id:%d err:%v", _func_, uid, grp_id, err)
				return
			}
			offset, err = comm.Conv2Int(res)
			if err != nil {
				log.Err("%s push convert res failed! uid:%d grp_id:%d err:%v res:%v", _func_, uid, grp_id, err, res)
				return
			}
			offset -= 1 //offset = count-1
			if offset >= CHAT_MSG_LIST_SIZE { //[0 , list_size-1]
				log.Info("%s chat_msg full! will rotate chat_msg pool! index:%d grp_id:%d" , _func_ , index , grp_id)
				index++
				continue
			}
			break
		}

		//real_location = index*size + offset; msg_id = real_location + 1
		//Set And Pack
		preq.ChatMsg.MsgId =  int64(offset + index*CHAT_MSG_LIST_SIZE + 1) //real msg id
		var ss_msg ss.SSMsg
		prsp := new(ss.MsgSendChatRsp)
		prsp.ChatMsg = preq.ChatMsg
		prsp.Uid = preq.Uid
		prsp.TempId = preq.TempId
		prsp.Occupy = preq.Occupy
		prsp.Result = ss.SEND_CHAT_RESULT_SEND_CHAT_SUCCESS

		err = comm.FillSSPkg(&ss_msg , ss.SS_PROTO_TYPE_SEND_CHAT_RSP , prsp)
		if err != nil {
			log.Err("%s gen ss pkg failed! uid:%d err:%v" , _func_ , uid , err)
			return
		}

		//Back
		SendToServ(pconfig , from , &ss_msg)
	}()

}

func RecvFetchChatReq(pconfig *Config , preq *ss.MsgFetchChatReq , from int) {
	var _func_ = "<RecvSendChatReq>"
	log := pconfig.Comm.Log

	//Sync
	go func() {
		uid := preq.Uid
		grp_id := preq.GrpId
		phead := pconfig.RedisClient.AllocSyncCmdHead()
		if phead == nil {
			log.Err("%s alloc head failed! uid:%d grp_id:%d" , _func_ , uid , grp_id)
			return
		}
		defer pconfig.RedisClient.FreeSyncCmdHead(phead)

		//Handle
		start_msg_id := preq.LatestMsgId + 1
		index := int((start_msg_id-1) / CHAT_MSG_LIST_SIZE)
		offset := int((start_msg_id-1) % CHAT_MSG_LIST_SIZE)
		end_offset := offset + int(preq.FetchCount)-1
		last_offset := end_offset
		if end_offset >= CHAT_MSG_LIST_SIZE {
			end_offset = CHAT_MSG_LIST_SIZE-1
		}

		result := ss.SS_COMMON_RESULT_SUCCESS
		prsp := new(ss.MsgFetchChatRsp)
		prsp.GrpId = grp_id
		prsp.Uid = uid

		for {
			//Check Len
			tab_name := fmt.Sprintf(FORMAT_TAB_CHAT_MSG_LIST, grp_id, index)
			res , err := pconfig.RedisClient.RedisExeCmdSync(phead , "LLEN" , tab_name)
			if err != nil {
				log.Err("%s LLEN %s failed! uid:%d grp_id:%d err:%v", _func_, tab_name , uid, grp_id, err)
				return
			}
			list_len , err := comm.Conv2Int(res)
			if err != nil {
				log.Err("%s llen convert res failed! uid:%d grp_id:%d err:%v res:%v", _func_, uid, grp_id, err, res)
				return
			}
			if list_len == 0 {
				log.Debug("%s not exist! uid:%d grp_id:%d tab:%s" , _func_ , uid , grp_id , tab_name)
				result = ss.SS_COMMON_RESULT_NOEXIST
				break
			}

			//Get Msg
			res, err = pconfig.RedisClient.RedisExeCmdSync(phead, "LRANGE", tab_name, offset, end_offset)
			if err != nil {
				log.Err("%s lrange chat msg failed! uid:%d grp_id:%d err:%v", _func_, uid, grp_id, err)
				return
			}
			strs, err := comm.Conv2Strings(res)
			if err != nil {
				log.Err("%s lrange convert res failed! uid:%d grp_id:%d err:%v res:%v", _func_, uid, grp_id, err, res)
				return
			}
			if len(strs) == 0 { //no more data
				log.Debug("%s no more! tab:%s start_msg_id:%d uid:%d grp_id:%d", _func_, tab_name, start_msg_id, uid, grp_id)
				prsp.Complete = 1
				break
			}

			//fill info
			prsp.ChatList = make([]*ss.ChatMsg , len(strs))
			idx := 0
			var msg_id int64 = 0
			enc_block , err := des.NewCipher([]byte(CHAT_MSG_DES_KEY))
			if err != nil {
				log.Err("%s new des cipher for key:%v failed! err:%v" , _func_ , CHAT_MSG_DES_KEY , err)
				return
			}

			for i:=0; i<len(strs); i++ {
				pmsg := new(ss.ChatMsg)
				msg_id = int64(offset+i + index*CHAT_MSG_LIST_SIZE +1)
				//decrypt
				out_data , err := lnet.DesDecrypt(enc_block , []byte(strs[i]) , []byte(CHAT_MSG_DES_KEY))
				if err != nil {
					log.Err("%s  decrypt chat_msg failed! err:%v msg_id:%d uid:%d grp_id:%d" , _func_ , err ,
						msg_id , uid , grp_id)
					continue
				}

				//unpack
				err = ss.UnPack(out_data , pmsg)
				if err != nil {
					log.Err("%s decode chat_msg failed! err:%v msg_id:%d uid:%d grp_id:%d" , _func_ , err ,
						 msg_id , uid , grp_id)
					continue
				}

				//msg
				pmsg.MsgId = msg_id
				prsp.ChatList[idx]=pmsg
				idx++
			}
			prsp.FetchCount = int32(idx)

			//check complete
			if len(strs) < int(preq.FetchCount) && (last_offset <= CHAT_MSG_LIST_SIZE-1) {
				prsp.Complete = 1 //no more data
			}

			break
		}

		//Pack
		prsp.Result = result
		var ss_msg ss.SSMsg
		err := comm.FillSSPkg(&ss_msg , ss.SS_PROTO_TYPE_FETCH_CHAT_RSP , prsp)
		if err != nil {
			log.Err("%s gen ss pkg failed! err:%v uid:%d grp_id:%d" , _func_ , err , uid , grp_id)
			return
		}

		//Send
		log.Debug("%s latest_read:%d read start_msg_id:%d count:%d offset:%d end_offset:%d complete:%d" , _func_ , preq.LatestMsgId ,
			start_msg_id , prsp.FetchCount , offset , end_offset , prsp.Complete)
		SendToServ(pconfig , from , &ss_msg)
	}()
}
