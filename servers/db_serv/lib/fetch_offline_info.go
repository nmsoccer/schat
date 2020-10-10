package lib

import (
	"fmt"
	"schat/proto/ss"
	"schat/servers/comm"
)

func RecvFetchOfflineInfoReq(pconfig *Config , preq *ss.MsgFetchOfflineInfoReq , from int) {
	var _func_ = "<RecvFetchOfflineInfoReq>"
	log := pconfig.Comm.Log
	uid := preq.Uid

	go func() {
		phead := pconfig.RedisClient.AllocSyncCmdHead()
		if phead == nil {
			log.Err("%s alloc head failed! uid:%d" , _func_ , uid)
			return
		}
		defer pconfig.RedisClient.FreeSyncCmdHead(phead)

		//rsp
		prsp := new(ss.MsgFetchOfflineInfoRsp)
		prsp.Uid = uid
		prsp.Result = ss.SS_COMMON_RESULT_FAILED


		//handle
		log.Debug("%s uid:%d  from:%d", _func_, uid, from)
		for {
			//get offline info
			tab_name := fmt.Sprintf(FORMAT_TAB_OFFLINE_INFO_PREFIX+"%d", uid)
			res, err := pconfig.RedisClient.RedisExeCmdSync(phead, "LRANGE", tab_name, 0, preq.FetchCount-1)
			if err != nil {
				log.Err("%s lrange %s failed! uid:%d  err:%v", _func_, tab_name, uid, err)
				break
			}

			//convert
			strs , err := comm.Conv2Strings(res)
			if err != nil {
				log.Err("%s convert res %v failed! uid:%d  err:%v", _func_, res, uid, err)
				break
			}

			//fill info
			prsp.FetchCount = int32(len(strs))
			prsp.InfoList = strs
			if prsp.FetchCount < preq.FetchCount {
				prsp.Complete = 1
			}
			prsp.Result = ss.SS_COMMON_RESULT_SUCCESS

			//trim list
			_ , err = pconfig.RedisClient.RedisExeCmdSync(phead , "LTRIM" , tab_name , preq.FetchCount , -1)
			if err != nil {
				log.Err("%s ltrim tab:%s %d:%d failed! err:%v uid:%d" , _func_ , tab_name , preq.FetchCount , -1 , err , uid)
			}
			break
		}

		//ss
		var ss_msg ss.SSMsg
		err := comm.FillSSPkg(&ss_msg , ss.SS_PROTO_TYPE_FETCH_OFFLINE_INFO_RSP , prsp)
		if err != nil {
			log.Err("%s gen ss failed! err:%v uid:%d" , _func_ , err ,uid)
			return
		}

		//back
		SendToServ(pconfig , from , &ss_msg)
	}()
}

