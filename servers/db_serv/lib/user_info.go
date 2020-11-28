package lib

import (
	"fmt"
	"schat/proto/ss"
	"schat/servers/comm"
)

//update user req
//only change password
func RecvUpdateUserReq(pconfig *Config, preq *ss.MsgUpdateUserReq, from int) {
	var _func_ = "<RecvUpdateUserReq>"
	log := pconfig.Comm.Log
	uid := preq.Uid

	log.Debug("%s uid:%d pass:%s", _func_, uid , preq.Passwd)
	//Sync Mod Must be In a routine
	go func() {
		//pclient
		pclient := SelectRedisClient(pconfig, REDIS_OPT_RW)
		if pclient == nil {
			log.Err("%s failed! no proper redis found! uid:%d", _func_, uid)
			return
		}
		//Get SyncHead
		phead := pclient.AllocSyncCmdHead()
		if phead == nil {
			log.Err("%s alloc synchead faileed! uid:%d", _func_, uid)
			return
		}
		defer pclient.FreeSyncCmdHead(phead)

		//handle
		for {
			//check pass
			if len(preq.Passwd) <=0 {
				log.Err("%s password:%s illegal! uid:%d" , _func_ , preq.Passwd , uid)
				SendUpdateUserRsp(pconfig , preq , from , ss.SS_COMMON_RESULT_FAILED)
				break
			}

			//generate salt
			salt, err := comm.GenRandStr(PASSWD_SALT_LEN)
			if err != nil {
				log.Err("%s generate salt failed! err:%v uid:%d", _func_, err, uid)
				SendUpdateUserRsp(pconfig, preq, from, ss.SS_COMMON_RESULT_FAILED)
				break
			}

			//enc pass
			enc_pass := comm.EncPassString(preq.Passwd, salt)

			//exe cmd
			tab_name := fmt.Sprintf(FORMAT_TAB_USER_GLOBAL, preq.AccountName)
			_ , err = pclient.RedisExeCmdSync(phead , "HMSET" , tab_name , "pass" , enc_pass , "salt" , salt)
			if err != nil {
				log.Err("%s hmset %s pass failed! err:%v uid:%d" , _func_ , tab_name , err , uid)
				SendUpdateUserRsp(pconfig, preq, from, ss.SS_COMMON_RESULT_FAILED)
				break
			}

			SendUpdateUserRsp(pconfig, preq, from, ss.SS_COMMON_RESULT_SUCCESS)
			break
		}

		//finish

	}()
}

func SendUpdateUserRsp(pconfig *Config , preq *ss.MsgUpdateUserReq , from int , result ss.SS_COMMON_RESULT) {
	var _func_ = "<SendUpdateUserRsp>"
	log := pconfig.Comm.Log
	uid := preq.Uid

	log.Info("%s uid:%d result:%d" , _func_ , uid , result)
	//ss_msg
	var ss_msg ss.SSMsg
	prsp := new(ss.MsgUpdateUserRsp)
	prsp.Uid = uid
	prsp.Passwd = preq.Passwd
	prsp.Result = result
	prsp.Addr = preq.Addr
	prsp.Desc = preq.Desc
	prsp.RoleName = preq.RoleName

	err := comm.FillSSPkg(&ss_msg , ss.SS_PROTO_TYPE_UPDATE_USER_RSP , prsp)
	if err != nil {
		log.Err("%s gen ss failed! uid:%d" , _func_ , uid)
		return
	}

	//to logic
	SendToServ(pconfig , from , &ss_msg)
}
