package lib

import (
	"fmt"
	"schat/proto/ss"
	"schat/servers/comm"
	"strconv"
)

//user login
func RecvUserLoginReq(pconfig *Config, preq *ss.MsgLoginReq, from int) {
	var _func_ = "<RecvUserLoginReq>"
	log := pconfig.Comm.Log

	log.Debug("%s user:%s pass:%s c_key:%d", _func_, preq.GetName(), preq.GetPass(), preq.GetCKey())
	//Sync Mod Must be In a routine
	go func() {
		//pclient
		pclient := SelectRedisClient(pconfig , REDIS_OPT_RW)
		if pclient == nil {
			log.Err("%s failed! no proper redis found! name:%s" , _func_ , preq.Name)
			return
		}
		//Get SyncHead
		phead := pclient.AllocSyncCmdHead()
		if phead == nil {
			log.Err("%s alloc synchead faileed! uid:%d", _func_, preq.Uid)
			return
		}
		defer pclient.FreeSyncCmdHead(phead)

		//check pass
		result, err := pclient.RedisExeCmdSync(phead, "HGETALL", fmt.Sprintf(FORMAT_TAB_USER_GLOBAL, preq.Name))
		if err != nil {
			log.Err("%s query pass failed! name:%s", _func_, preq.Name)
			return
		}
		ok := user_login_check_pass(pconfig, result, preq, from)
		if !ok {
			return
		}

		//lock temp
		log.Debug("%s try to lock login. uid:%d name:%s", _func_, preq.Uid, preq.Name)
		tab_name := fmt.Sprintf(FORMAT_TAB_USER_LOGIN_LOCK_PREFIX+"%d", preq.Uid)
		result, err = pclient.RedisExeCmdSync(phead, "SET", tab_name, preq.Uid, "EX",
			LOGIN_LOCK_LIFE, "NX")
		if err != nil {
			log.Err("%s lock login failed! name:%s uid:%d", _func_, preq.Name, preq.Uid)
			return
		}
		ok = user_login_lock(pconfig, result, preq, from)
		if !ok {
			return
		}

		//get user info
		log.Debug("%s ok! try to get user_info. uid:%d name:%s", _func_, preq.Uid, preq.Name)
		tab_name = fmt.Sprintf(FORMAT_TAB_USER_INFO_REFIX+"%d", preq.Uid)
		result, err = pclient.RedisExeCmdSync(phead, "HGETALL", tab_name)
		if err != nil {
			log.Err("%s get user info failed! name:%s uid:%d", _func_, preq.Name, preq.Uid)
			return
		}
		pss_msg, ok := user_login_get_info(pconfig, result, preq, from)
		if !ok || pss_msg == nil {
			return
		}

		//update online_logic
		log.Debug("%s update online logic! uid:%d", _func_, preq.Uid)
		tab_name = fmt.Sprintf(FORMAT_TAB_USER_INFO_REFIX+"%d", preq.Uid)
		result, err = pclient.RedisExeCmdSync(phead, "HSET", tab_name, FIELD_USER_INFO_ONLINE_LOGIC, from)
		if err != nil {
			log.Err("%s update user online_logic failed! name:%s uid:%d", _func_, preq.Name, preq.Uid)
			return
		}
		user_login_update_online(pconfig, result, preq, from, pss_msg)
		log.Info("%s finish! uid:%d name:%s", _func_, preq.Uid, preq.Name)
	}()

}

//user logout
func RecvUserLogoutReq(pconfig *Config, preq *ss.MsgLogoutReq, from int) {
	var _func_ = "<RecvUserLogoutReq>"
	log := pconfig.Comm.Log

	log.Debug("%s uid:%d reason:%d", _func_, preq.Uid, preq.Reason)
	//synchronise
	go func() {
		var err error
		var res interface{}
		//pclient
		pclient := SelectRedisClient(pconfig , REDIS_OPT_W)
		if pclient == nil {
			log.Err("%s failed! no proper redis found! uid:%d" , _func_ , preq.Uid)
			return
		}
		//Get SyncHead
		phead := pclient.AllocSyncCmdHead()
		if phead == nil {
			log.Err("%s alloc synchead faileed! uid:%d", _func_, preq.Uid)
			return
		}
		defer pclient.FreeSyncCmdHead(phead)

		//Exe Cmds
		user_tab := fmt.Sprintf(FORMAT_TAB_USER_INFO_REFIX+"%d", preq.Uid)
		if preq.UserInfo != nil && preq.Reason != ss.USER_LOGOUT_REASON_LOGOUT_OFFLINE_USER {
			puser_info := preq.UserInfo
			user_blob, err := ss.Pack(preq.UserInfo.BlobInfo)
			if err != nil {
				log.Err("%s save user_info failed! pack blob info fail! err:%v uid:%d", _func_, err, preq.Uid)
				return
			}
			res, err = pclient.RedisExeCmdSync(phead, "HMSET", user_tab, "addr",
				puser_info.BasicInfo.Addr, "level", puser_info.BasicInfo.Level, FIELD_USER_INFO_ONLINE_LOGIC, -1, "blob_info", string(user_blob),
				FILED_USER_INFO_HEAD_URL, puser_info.BasicInfo.HeadUrl , FIELD_USER_INFO_NAME , puser_info.BasicInfo.Name)
		} else { //only update online-logic
			res, err = pclient.RedisExeCmdSync(phead, "HSET", user_tab, FIELD_USER_INFO_ONLINE_LOGIC, -1)
		}

		//Get Result
		if err != nil {
			log.Err("%s failed! err:%v uid:%d reason:%d", _func_, err, preq.Uid, preq.Reason)
		} else {
			log.Info("%s done! res:%v uid:%d reason:%d", _func_, res, preq.Uid, preq.Reason)
		}

		//save profile
		if preq.UserInfo != nil && preq.Reason != ss.USER_LOGOUT_REASON_LOGOUT_OFFLINE_USER {
			//gen profile
			profile := new(ss.UserProfile)
			profile.Uid = preq.Uid
			profile.HeadUrl = preq.UserInfo.BasicInfo.HeadUrl
			profile.Name = preq.UserInfo.BasicInfo.Name
			profile.Level = preq.UserInfo.BasicInfo.Level
			if preq.UserInfo.BasicInfo.Sex {
				profile.Sex = comm.SEX_INT_MALE
			} else {
				profile.Sex = comm.SEX_INT_FEMALE
			}
			profile.Addr = preq.UserInfo.BasicInfo.Addr
			profile.UserDesc = preq.UserInfo.BlobInfo.UserDesc
			//pack
			enc_data, err := ss.Pack(profile)
			if err != nil {
				log.Err("%s pack profile failed! err:%v uid:%d", _func_, err, preq.Uid)
			} else {
				SaveUserProfile(pclient, phead, preq.Uid, string(enc_data))
			}
		}

		return
	}()

}

/*---------------------------------STATIC FUNC-----------------------------*/
//@return next_step
func user_login_check_pass(pconfig *Config, result interface{}, preq *ss.MsgLoginReq, from_serv int) bool {
	var _func_ = "<user_login_check_pass>"
	log := pconfig.Comm.Log

	//rsp
	var ss_msg ss.SSMsg
	prsp := new(ss.MsgLoginRsp)
	prsp.CKey = preq.CKey
	prsp.Name = preq.Name

	//do while 0
	for {
		//check result may need reg
		if result == nil {
			log.Info("%s no user:%s exist!", _func_, preq.Name)
			prsp.Result = ss.USER_LOGIN_RET_LOGIN_EMPTY
			break
		}

		//conv result
		sm, err := comm.Conv2StringMap(result)
		if err != nil {
			log.Err("%s conv result failed! err:%v", _func_, err)
			return false
		}

		pass := sm["pass"]
		uid := sm["uid"]
		salt := sm["salt"]
		log.Debug("%s try to check pass! uid:%s", _func_, uid)
		//check pass
		enc_pass := comm.EncPassString(preq.Pass, salt)
		if enc_pass != pass {
			log.Info("%s pass not matched! user:%s c_key:%v ", _func_, preq.Name, preq.CKey)
			prsp.Result = ss.USER_LOGIN_RET_LOGIN_PASS
			break
		}

		//sucess. get user info
		if preq.Uid == 0 { //default role
			//conv uid
			preq.Uid, err = strconv.ParseInt(uid, 10, 64)
			if err != nil {
				log.Err("%s conv uid failed! err:%v uid:%s", _func_, err, uid)
				prsp.Result = ss.USER_LOGIN_RET_LOGIN_ERR
				break
			}
		}
		return true
	}

	/*Back to Client*/
	//fill
	err := comm.FillSSPkg(&ss_msg, ss.SS_PROTO_TYPE_LOGIN_RSP, prsp)
	if err != nil {
		log.Err("%s gen ss failed! err:%v", _func_, err)
		return false
	}

	//send
	SendToServ(pconfig, from_serv, &ss_msg)
	return false
}

//lock login stat
//@return next_step
func user_login_lock(pconfig *Config, result interface{}, preq *ss.MsgLoginReq, from_serv int) bool {
	var _func_ = "<user_login_lock>"
	log := pconfig.Comm.Log

	//check result
	if result == nil { //in login process
		log.Err("%s is in login process! uid:%d name:%s", _func_, preq.Uid, preq.Name)
		//rsp
		var ss_msg ss.SSMsg
		prsp := new(ss.MsgLoginRsp)
		prsp.CKey = preq.CKey
		prsp.Name = preq.Name
		prsp.Result = ss.USER_LOGIN_RET_LOGIN_MULTI_ON

		if err := comm.FillSSPkg(&ss_msg, ss.SS_PROTO_TYPE_LOGIN_RSP, prsp); err != nil {
			log.Err("%s gen ss failed! err:%v name:%s", _func_, err, preq.Name)
		} else {
			SendToServ(pconfig, from_serv, &ss_msg)
		}
		return false
	}

	return true
}

//get user detail inf
//@return pss_msg(if success) , next_step
func user_login_get_info(pconfig *Config, result interface{}, preq *ss.MsgLoginReq, from_serv int) (*ss.SSMsg, bool) {
	var _func_ = "<user_login_get_info>"
	log := pconfig.Comm.Log

	/*create rsp */
	pss_msg := new(ss.SSMsg)
	prsp := new(ss.MsgLoginRsp)
	prsp.CKey = preq.CKey
	prsp.Name = preq.Name
	prsp.Uid = preq.Uid

	//do while 0
	for {
		//check result
		if result == nil {
			log.Err("%s no user detail:%s exist!", _func_, preq.Name)
			prsp.Result = ss.USER_LOGIN_RET_LOGIN_EMPTY
			break
		}

		//conv result
		sm, err := comm.Conv2StringMap(result)
		if err != nil {
			log.Err("%s conv result failed! err:%v", _func_, err)
			prsp.Result = ss.USER_LOGIN_RET_LOGIN_ERR
			break
		}

		//Get User Info
		prsp.UserInfo = new(ss.UserInfo)
		prsp.UserInfo.BasicInfo = new(ss.UserBasic)
		pbasic := prsp.UserInfo.BasicInfo
		puser_blob := new(ss.UserBlob)

		var uid int64
		var online_logic = -1

		pbasic.AccountName = preq.Name
		//uid
		if v, ok := sm["uid"]; ok {
			uid, err = strconv.ParseInt(v, 10, 64)
			if err != nil {
				log.Err("%s conv uid failed! err:%v uid:%s", _func_, err, v)
				prsp.Result = ss.USER_LOGIN_RET_LOGIN_ERR
				break
			}
		} else {
			log.Err("%s no uid found of user:%s", _func_, preq.Name)
			prsp.Result = ss.USER_LOGIN_RET_LOGIN_ERR
			break
		}

		if uid != preq.Uid {
			log.Err("%s fail! uid not match! %d<->%d", _func_, uid, preq.Uid)
			prsp.Result = ss.USER_LOGIN_RET_LOGIN_ERR
			break
		}
		pbasic.Uid = uid

		//online_logic
		if v, ok := sm["online_logic"]; ok {
			online_logic, err = strconv.Atoi(v)
			if err != nil {
				log.Err("%s conv online-logic failed! err:%v", _func_, err)
				prsp.Result = ss.USER_LOGIN_RET_LOGIN_ERR
			}
		} else {
			log.Err("%s conv online-logic not exist! uid:%d", _func_, uid)
			prsp.Result = ss.USER_LOGIN_RET_LOGIN_ERR
		}
		//check online
		//already login at other logic should kick first
		if online_logic >= 0 && online_logic != from_serv {
			log.Info("%s user:%s login at other logic server %d now logic:%d kick first!", _func_, preq.Name, online_logic,
				from_serv)
			prsp.Result = ss.USER_LOGIN_RET_LOGIN_MULTI_ON
			prsp.OnlineLogic = int32(online_logic)
			break
		}

		//role_name
		if v, ok := sm["name"]; ok {
			pbasic.Name = v
		} else {
			log.Err("%s no role_name found of uid:%d", _func_, uid)
			prsp.Result = ss.USER_LOGIN_RET_LOGIN_ERR
			break
		}

		//age
		if v, ok := sm["age"]; ok {
			age, err := strconv.Atoi(v)
			if err != nil {
				log.Err("%s conv age failed! err:%v age:%s", _func_, err, v)
				prsp.Result = ss.USER_LOGIN_RET_LOGIN_ERR
				break
			}
			pbasic.Age = int32(age)
		}

		//sex
		if v, ok := sm["sex"]; ok {
			sex, err := strconv.Atoi(v)
			if err != nil {
				log.Err("%s conv sex failed! err:%v sex:%s", _func_, err, v)
				prsp.Result = ss.USER_LOGIN_RET_LOGIN_ERR
				break
			}
			if sex == 1 {
				pbasic.Sex = true //male; false:female
			} else {
				pbasic.Sex = false
			}
		}

		//addr
		if v, ok := sm["addr"]; ok {
			pbasic.Addr = v
		}

		//level
		if v, ok := sm["level"]; ok {
			level, err := strconv.Atoi(v)
			if err != nil {
				log.Err("%s conv level failed! err:%v level:%s", _func_, err, v)
				prsp.Result = ss.USER_LOGIN_RET_LOGIN_ERR
				break
			}
			pbasic.Level = int32(level)
		}

		//blob info
		if v, ok := sm["blob_info"]; ok {
			err = ss.UnPack([]byte(v), puser_blob)
			if err != nil {
				log.Err("%s unpack user_blob failed! err:%v uid:%d", _func_, err, uid)
				prsp.Result = ss.USER_LOGIN_RET_LOGIN_ERR
				break
			}
		} else { //set blob_info default value
			puser_blob.Exp = 100
		}
		prsp.UserInfo.BlobInfo = puser_blob

		//head_url
		if v, ok := sm[FILED_USER_INFO_HEAD_URL]; ok {
			pbasic.HeadUrl = v
		}

		//Fullfill
		prsp.Result = ss.USER_LOGIN_RET_LOGIN_SUCCESS
		log.Debug("%s success! user:%s uid:%v", _func_, pbasic.Name, pbasic.Uid)

		//msg
		err = comm.FillSSPkg(pss_msg, ss.SS_PROTO_TYPE_LOGIN_RSP, prsp)
		if err != nil {
			log.Err("%s gen ss failed! err:%v", _func_, err)
			return nil, false
		}

		return pss_msg, true
	}

	/* Err will Back to Client*/
	//fill
	err := comm.FillSSPkg(pss_msg, ss.SS_PROTO_TYPE_LOGIN_RSP, prsp)
	if err != nil {
		log.Err("%s gen ss failed! err:%v", _func_, err)
		return nil, false
	}

	//send
	SendToServ(pconfig, from_serv, pss_msg)
	return nil, false
}

//cb_arg={0:preq 1:from_server 2:pss_msg}
func user_login_update_online(pconfig *Config, result interface{}, preq *ss.MsgLoginReq, from_serv int, pss_msg *ss.SSMsg) {
	var _func_ = "<user_login_update_online>"
	log := pconfig.Comm.Log
	uid := preq.Uid

	/*---------result handle--------------*/
	/*Get Result*/
	ret_code, err := comm.Conv2Int(result)
	if err != nil {
		log.Err("%s conv result failed! uid:%d err:%v", _func_, uid, err)
		return
	}
	log.Debug("%s ret_code:%d name:%s uid:%d", _func_, ret_code, preq.Name, uid)

	//Back to Client
	SendToServ(pconfig, from_serv, pss_msg)
}
