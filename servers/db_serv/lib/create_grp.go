package lib

import (
	"fmt"
	lnet "schat/lib/net"
	"schat/proto/ss"
	"schat/servers/comm"
	"time"
)

func RecvCreateGroupReq(pconfig *Config, preq *ss.MsgCreateGrpReq, from int) {
	var _func_ = "<RecvCreateGroupReq>"
	log := pconfig.Comm.Log
	uid := preq.Uid

	log.Info("%s uid:%d grp_name:%s", _func_, uid, preq.GrpName)
	//Sync
	go func() {
		//pclient
		pclient := SelectRedisClient(pconfig , REDIS_OPT_RW)
		if pclient == nil {
			log.Err("%s failed! no proper redis found! uid:%d" , _func_ , uid)
			return
		}
		//phead
		phead := pclient.AllocSyncCmdHead()
		if phead == nil {
			log.Err("%s alloc head failed! uid:%d", _func_, uid)
			return
		}
		defer pclient.FreeSyncCmdHead(phead)

		//Step1.Alloc grp id
		res, err := pclient.RedisExeCmdSync(phead, "INCRBY", FORMAT_TAB_GLOBAL_GRPID, pconfig.FileConfig.GrpIdIncr)
		if err != nil {
			log.Err("%s alloc group id failed! uid:%d err:%v", _func_, uid, err)
			return
		}
		grp_id, err := comm.Conv2Int64(res)
		if err != nil {
			log.Err("%s convert group_id failed! uid:%d err:%v res:%v", _func_, uid, err, res)
			SendCreateGroupErrRsp(pconfig, preq, from, ss.CREATE_GROUP_RESULT_CREATE_RET_DB_ERR)
			return
		}

		//generate salt
		salt, err := comm.GenRandStr(PASSWD_SALT_LEN)
		if err != nil {
			log.Err("%s generate salt failed! err:%v uid:%s", _func_, err, preq.Uid)
			SendCreateGroupErrRsp(pconfig, preq, from, ss.CREATE_GROUP_RESULT_CREATE_RET_DB_ERR)
			return
		}

		//enc pass
		enc_pass := comm.EncPassString(preq.GrpPass, salt)

		//Step2.Push One Msg
		var curr_ts = time.Now().Unix()
		var one_msg ss.ChatMsg
		one_msg.ChatType = ss.CHAT_MSG_TYPE_CHAT_TYPE_TEXT
		one_msg.SendTs = curr_ts
		one_msg.SenderUid = 0
		one_msg.Content = fmt.Sprintf("welcome to %s", preq.GrpName)
		one_msg.GroupId = grp_id

		//Pack
		coded, err := ss.Pack(&one_msg)
		if err != nil {
			log.Err("%s encode chat msg failed! uid:%d grp_id:%d err:%v", _func_, uid, grp_id, err)
			return
		}

		//Encrypt
		coded, err = lnet.DesEncrypt(nil, coded, []byte(CHAT_MSG_DES_KEY))
		if err != nil {
			log.Err("%s encrypt chat msg failed! uid:%d grp_id:%d err:%v", _func_, uid, grp_id, err)
			return
		}
		tab_name := fmt.Sprintf(FORMAT_TAB_CHAT_MSG_LIST, grp_id, 0)
		_, err = pclient.RedisExeCmdSync(phead, "RPUSH", tab_name, coded)
		if err != nil {
			log.Err("%s push one_msg faile! err:%v uid:%d", _func_, err, uid)
			SendCreateGroupErrRsp(pconfig, preq, from, ss.CREATE_GROUP_RESULT_CREATE_RET_DB_ERR)
			return
		}

		//Step3. Set group info
		tab_name = fmt.Sprintf(FORMAT_TAB_GROUP_INFO_PREFIX+"%d", grp_id)
		_, err = pclient.RedisExeCmdSync(phead, "HMSET", tab_name, "gid", grp_id,
			"name", preq.GrpName, "master_uid", preq.Uid, "pass", enc_pass, "salt", salt, "create_ts", curr_ts, "msg_count", 1, "load_serv",
			-1)
		if err != nil {
			log.Err("%s Set Group Info Failed! err:%v uid:%d grp_id:%d", _func_, err, uid, grp_id)
			SendCreateGroupErrRsp(pconfig, preq, from, ss.CREATE_GROUP_RESULT_CREATE_RET_DB_ERR)
			return
		}

		//res should always right
		log.Info("%s uid:%d grp_id:%d grp_name:%s create_ts:%d mem:%d", _func_, preq.Uid, grp_id, preq.GrpName,
			curr_ts, 1)

		//back to logic
		var ss_msg ss.SSMsg
		pCreateGroupRsp := new(ss.MsgCreateGrpRsp)
		pCreateGroupRsp.Uid = preq.Uid
		pCreateGroupRsp.Ret = ss.CREATE_GROUP_RESULT_CREATE_RET_SUCCESS
		pCreateGroupRsp.GrpName = preq.GrpName
		pCreateGroupRsp.GrpId = grp_id
		pCreateGroupRsp.CreateTs = curr_ts
		pCreateGroupRsp.MemCount = 1

		//pack
		err = comm.FillSSPkg(&ss_msg, ss.SS_PROTO_TYPE_CREATE_GROUP_RSP, pCreateGroupRsp)
		if err != nil {
			log.Err("%s gen ss failed! err:%v", _func_, err)
			return
		}

		//sendback
		SendToServ(pconfig, from, &ss_msg)
	}()
}

func SendCreateGroupErrRsp(pconfig *Config, preq *ss.MsgCreateGrpReq, target_serv int, result ss.CREATE_GROUP_RESULT) {
	var _func_ = "<SendCreateGroupRsp>"
	log := pconfig.Comm.Log

	log.Info("%s uid:%s target:%d grp_name:%s result:%d", _func_, preq.Uid, target_serv, preq.GrpName, result)
	if result == ss.CREATE_GROUP_RESULT_CREATE_RET_SUCCESS { //success on the other way
		return
	}
	//msg
	var ss_msg ss.SSMsg
	pCreateGroupRsp := new(ss.MsgCreateGrpRsp)
	pCreateGroupRsp.Uid = preq.Uid
	pCreateGroupRsp.Ret = result
	pCreateGroupRsp.GrpName = preq.GrpName

	//pack
	err := comm.FillSSPkg(&ss_msg, ss.SS_PROTO_TYPE_CREATE_GROUP_RSP, pCreateGroupRsp)
	if err != nil {
		log.Err("%s pack failed! err:%v", _func_, err)
		return
	}

	//sendback
	if ok := SendToServ(pconfig, target_serv, &ss_msg); !ok {
		log.Err("%s send to logic:%d failed!", _func_, target_serv)
	}
}

/*-----------------------static func-----------------------*/
