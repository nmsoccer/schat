package lib

import (
	"fmt"
	"schat/proto/ss"
	"schat/servers/comm"
)

func RecvFetchUserProfileReq(pconfig *Config, preq *ss.MsgFetchUserProfileReq, from_serv int) {
	var _func_ = "<RecvFetchUserProfileReq>"
	log := pconfig.Comm.Log
	uid := preq.Uid

	//sync
	go func() {
		//pclient
		pclient := SelectRedisClient(pconfig , REDIS_OPT_R)
		if pclient == nil {
			log.Err("%s failed! no proper redis found! uid:%d" , _func_ , uid)
			return
		}
		//head
		phead := pclient.AllocSyncCmdHead()
		if phead == nil {
			log.Err("%s alloc head fail! uid:%d", _func_, uid)
			return
		}
		defer pclient.FreeSyncCmdHead(phead)

		//args
		args := make([]interface{}, len(preq.TargetList))
		for i := 0; i < len(preq.TargetList); i++ {
			args[i] = fmt.Sprintf(FORMAT_TAB_USER_PREOFILE_PREFIX+"%d", preq.TargetList[i])
		}

		//resp
		prsp := new(ss.MsgFetchUserProfileRsp)
		prsp.Uid = uid
		prsp.Result = ss.SS_COMMON_RESULT_FAILED
		prsp.Profiles = make(map[int64]*ss.UserProfile)

		//exe
		for {
			//query
			res, err := pclient.RedisExeCmdSync(phead, "MGET", args...)
			if err != nil {
				log.Err("%s exe MGET failed! err:%v uid:%d", _func_, err, uid)
				break
			}

			//convert
			strs, err := comm.Conv2Strings(res)
			if err != nil {
				log.Err("%s convert res failed! err:%v uid:%d", _func_, err, uid)
				break
			}

			if len(strs) != len(preq.TargetList) {
				log.Err("%s length not match! %d:%d err:%v uid:%d", _func_, len(strs), len(preq.TargetList), err, uid)
				break
			}

			//fill info
			var tuid int64
			for i := 0; i < len(strs); i++ {
				tuid = preq.TargetList[i]
				prsp.Profiles[tuid] = nil
				//empty
				if len(strs[i]) == 0 {
					log.Debug("%s profile empty! uid:%d", _func_, tuid)
					continue
				}

				//fail
				profile := new(ss.UserProfile)
				err = ss.UnPack([]byte(strs[i]), profile)
				if err != nil {
					log.Err("%s unpack profile %d failed! err:%v", _func_, tuid, err)
					continue
				}

				//set
				prsp.Profiles[tuid] = profile
			}
			prsp.Result = ss.SS_COMMON_RESULT_SUCCESS
			log.Debug("%s fill %d target! uid:%d", _func_, len(strs), uid)
			break
		}

		//ss
		var ss_msg ss.SSMsg
		err := comm.FillSSPkg(&ss_msg, ss.SS_PROTO_TYPE_FETCH_USER_PROFILE_RSP, prsp)
		if err != nil {
			log.Err("%s gen ss failed! err:%v uid:%d", _func_, err, uid)
			return
		}

		//to logic
		SendToServ(pconfig, from_serv, &ss_msg)
	}()

}

func RecvSaveUserProfileReq(pconfig *Config, preq *ss.MsgSaveUserProfileReq) {
	var _func_ = "<RecvSaveUserProfileReq>"
	log := pconfig.Comm.Log
	uid := preq.Uid

	//check
	if preq.Profile == nil {
		log.Err("%s profile nil! uid:%d", _func_, uid)
		return
	}

	//sync
	go func() {
		//pclient
		pclient := SelectRedisClient(pconfig , REDIS_OPT_W)
		if pclient == nil {
			log.Err("%s failed! no proper redis found! uid:%d" , _func_ , uid)
			return
		}
		//head
		phead := pclient.AllocSyncCmdHead()
		if phead == nil {
			log.Err("%s alloc head fail! uid:%d", _func_, uid)
			return
		}
		defer pclient.FreeSyncCmdHead(phead)

		//pack
		enc_data, err := ss.Pack(preq.Profile)
		if err != nil {
			log.Err("%s pack profile failed! err:%v uid:%d", _func_, err, uid)
			return
		}

		//save
		result := SaveUserProfile(pclient, phead, uid, string(enc_data))
		log.Debug("%s result:%d uid:%d", _func_, result, uid)
	}()

}
