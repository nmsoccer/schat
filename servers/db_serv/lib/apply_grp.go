package lib

import (
	"fmt"
	"schat/proto/ss"
	"schat/servers/comm"
	"strconv"
)

func RecvApplyGroupReq(pconfig *Config, preq *ss.MsgApplyGroupReq, from int) {
	var _func_ = "<RecvApplyGroupReq>"
	log := pconfig.Comm.Log

	log.Debug("%s grp_id:%d uid:%d from_logic:%d from:%d", _func_, preq.GroupId, preq.ApplyUid, preq.Occupy, from)
	//pclient
	pclient := SelectRedisClient(pconfig , REDIS_OPT_R)
	if pclient == nil {
		log.Err("%s failed! no proper redis found! uid:%d grp_id:%d" , _func_ , preq.ApplyUid , preq.GroupId)
		return
	}

	//get grp info
	cb_arg := []interface{}{preq, from}
	tab_name := fmt.Sprintf(FORMAT_TAB_GROUP_INFO_PREFIX+"%d", preq.GroupId)
	pclient.RedisExeCmd(pconfig.Comm, cb_query_grp_info, cb_arg, "HMGET", tab_name, "name", "master_uid", "pass", "salt")
}

func SendApplyGroupRsp(pconfig *Config, preq *ss.MsgApplyGroupReq, from int, ret ss.APPLY_GROUP_RESULT) {
	var _func_ = "<SendApplyGroupRsp>"
	log := pconfig.Comm.Log

	//ss_msg
	var ss_msg ss.SSMsg
	pApplyGroupRsp := new(ss.MsgApplyGroupRsp)
	pApplyGroupRsp.ApplyUid = preq.ApplyUid
	pApplyGroupRsp.GroupId = preq.GroupId
	pApplyGroupRsp.GroupName = preq.GroupName
	pApplyGroupRsp.ApplyName = preq.ApplyName
	pApplyGroupRsp.Occupy = preq.Occupy
	pApplyGroupRsp.Result = ret

	//fill
	err := comm.FillSSPkg(&ss_msg, ss.SS_PROTO_TYPE_APPLY_GROUP_RSP, pApplyGroupRsp)
	if err != nil {
		log.Err("%s gen ss fail! err:%v", _func_, err)
		return
	}

	//back to from
	SendToServ(pconfig, from, &ss_msg)
}

/*-----------------------static func-----------------------*/
func cb_query_grp_info(comm_config *comm.CommConfig, result interface{}, cb_arg []interface{}) {
	var _func_ = "<cb_query_grp_info>"
	log := comm_config.Log

	/*---------mostly common logic--------------*/
	//get config
	pconfig, ok := comm_config.ServerCfg.(*Config)
	if !ok {
		log.Err("%s failed! convert config fail!", _func_)
		return
	}

	//conv arg
	preq, ok := cb_arg[0].(*ss.MsgApplyGroupReq)
	if !ok {
		log.Err("%s conv req failed!", _func_)
		return
	}

	from, ok := cb_arg[1].(int)
	if !ok {
		log.Err("%s conv from failed! uid:%d", _func_, preq.ApplyUid)
		return
	}

	/*---------result handle--------------*/
	//check result
	err, ok := result.(error)
	if ok {
		log.Err("%s exe failed! err:%v uid:%s", _func_, err, preq.ApplyUid)
		SendApplyGroupRsp(pconfig, preq, from, ss.APPLY_GROUP_RESULT_APPLY_GRP_ERR)
		return
	}

	//get result
	res, err := comm.Conv2Strings(result)
	if err != nil {
		log.Err("%s conv result failed! err:%v uid:%s", _func_, err, preq.ApplyUid)
		SendApplyGroupRsp(pconfig, preq, from, ss.APPLY_GROUP_RESULT_APPLY_GRP_ERR)
		return
	}
	log.Debug("%s grp_info:%v", _func_, res)

	//handle result
	if len(res) != 4 {
		log.Err("%s group info not enough! grp_id:%d res:%v", _func_, preq.GroupId, res)
		SendApplyGroupRsp(pconfig, preq, from, ss.APPLY_GROUP_RESULT_APPLY_GRP_ERR)
		return
	}
	if len(res[0]) == 0 || len(res[1]) == 0 || len(res[2]) == 0 || len(res[3]) == 0 {
		log.Err("%s group info illegal! grp_id:%d res:%v", _func_, preq.GroupId, res)
		SendApplyGroupRsp(pconfig, preq, from, ss.APPLY_GROUP_RESULT_APPLY_GRP_ERR)
		return
	}

	//res[name , master_uid , pass]
	grp_name := res[0]
	master_uid, err := strconv.ParseInt(res[1], 10, 64)
	if err != nil {
		log.Err("%s conv master_uid:%s failed! err:%v", _func_, res[1], err)
		SendApplyGroupRsp(pconfig, preq, from, ss.APPLY_GROUP_RESULT_APPLY_GRP_ERR)
		return
	}
	pass := res[2]
	salt := res[3]

	//valid pass
	enc_pass := comm.EncPassString(preq.Pass, salt)
	if enc_pass != pass {
		log.Err("%s group pass not match! grp_id:%d apply_uid:%d", _func_, preq.GroupId, preq.ApplyUid)
		SendApplyGroupRsp(pconfig, preq, from, ss.APPLY_GROUP_RESULT_APPLY_GRP_PASS)
		return
	}

	//pclient
	pclient := SelectRedisClient(pconfig , REDIS_OPT_R)
	if pclient == nil {
		log.Err("%s failed! no proper redis found! uid:%d grp_id:%d" , _func_ , preq.ApplyUid , preq.GroupId)
		return
	}


	//notify master if online
	preq.GroupName = grp_name
	tab_name := fmt.Sprintf(FORMAT_TAB_USER_INFO_REFIX+"%d", master_uid)
	pclient.RedisExeCmd(pconfig.Comm, cb_get_user_online_logic, append(cb_arg, grp_name, master_uid), "HGET", tab_name, "online_logic")
}

//cg_arg = {preq , from , grp_name , master_uid}
func cb_add_user_appied(comm_config *comm.CommConfig, result interface{}, cb_arg []interface{}) {
	var _func_ = "<cb_add_user_appied>"
	log := comm_config.Log

	/*---------mostly common logic--------------*/
	//get config
	pconfig, ok := comm_config.ServerCfg.(*Config)
	if !ok {
		log.Err("%s failed! convert config fail!", _func_)
		return
	}

	//conv arg
	preq, ok := cb_arg[0].(*ss.MsgApplyGroupReq)
	if !ok {
		log.Err("%s conv req failed!", _func_)
		return
	}

	from, ok := cb_arg[1].(int)
	if !ok {
		log.Err("%s conv from failed! uid:%d", _func_, preq.ApplyUid)
		return
	}

	grp_name, ok := cb_arg[2].(string)
	if !ok {
		log.Err("%s conv grp_name failed! uid:%d", _func_, preq.ApplyUid)
		return
	}

	master_uid, ok := cb_arg[3].(int64)
	if !ok {
		log.Err("%s conv master_uid failed! uid:%d", _func_, preq.ApplyUid)
		return
	}

	/*---------result handle--------------*/
	//check result
	err, ok := result.(error)
	if ok {
		log.Err("%s exe failed! err:%v uid:%s", _func_, err, preq.ApplyUid)
		SendApplyGroupRsp(pconfig, preq, from, ss.APPLY_GROUP_RESULT_APPLY_GRP_ERR)
		return
	}

	//get result
	res, err := comm.Conv2Int(result)
	if err != nil {
		log.Err("%s conv result failed! err:%v uid:%s", _func_, err, preq.ApplyUid)
		SendApplyGroupRsp(pconfig, preq, from, ss.APPLY_GROUP_RESULT_APPLY_GRP_ERR)
		return
	}
	log.Debug("%s add result:%d grp_name:%s master:%d", _func_, res, grp_name, master_uid)

	//done
	SendApplyGroupRsp(pconfig, preq, from, ss.APPLY_GROUP_RESULT_APPLY_GRP_DONE)
}

//cg_arg = {preq , from , grp_name , master_uid}
func cb_push_apply_list(comm_config *comm.CommConfig, result interface{}, cb_arg []interface{}) {
	var _func_ = "<cb_push_apply_list>"
	log := comm_config.Log

	/*---------mostly common logic--------------*/
	//get config
	pconfig, ok := comm_config.ServerCfg.(*Config)
	if !ok {
		log.Err("%s failed! convert config fail!", _func_)
		return
	}

	//conv arg
	preq, ok := cb_arg[0].(*ss.MsgApplyGroupReq)
	if !ok {
		log.Err("%s conv req failed!", _func_)
		return
	}

	from, ok := cb_arg[1].(int)
	if !ok {
		log.Err("%s conv from failed! uid:%d", _func_, preq.ApplyUid)
		return
	}

	grp_name, ok := cb_arg[2].(string)
	if !ok {
		log.Err("%s conv grp_name failed! uid:%d", _func_, preq.ApplyUid)
		return
	}

	master_uid, ok := cb_arg[3].(int64)
	if !ok {
		log.Err("%s conv master_uid failed! uid:%d", _func_, preq.ApplyUid)
		return
	}

	/*---------result handle--------------*/
	//check result
	err, ok := result.(error)
	if ok {
		log.Err("%s exe failed! err:%v uid:%s", _func_, err, preq.ApplyUid)
		SendApplyGroupRsp(pconfig, preq, from, ss.APPLY_GROUP_RESULT_APPLY_GRP_ERR)
		return
	}

	//get result
	res, err := comm.Conv2Int(result)
	if err != nil {
		log.Err("%s conv result failed! err:%v uid:%s", _func_, err, preq.ApplyUid)
		SendApplyGroupRsp(pconfig, preq, from, ss.APPLY_GROUP_RESULT_APPLY_GRP_ERR)
		return
	}
	log.Debug("%s apply_len:%d grp_name:%s master:%d", _func_, res, grp_name, master_uid)

	//client
	pclient := SelectRedisClient(pconfig , REDIS_OPT_W)
	if pclient == nil {
		log.Err("%s failed! no proper redis found! uid:%d grp_id:%d" , _func_ , preq.ApplyUid , preq.GroupId)
		SendApplyGroupRsp(pconfig, preq, from, ss.APPLY_GROUP_RESULT_APPLY_GRP_ERR)
		return
	}

	//add to user_group_applied
	var add_v string
	add_v = fmt.Sprintf("%d|%s", preq.GroupId, grp_name)
	tab_name := fmt.Sprintf(FORMAT_TAB_USER_GROUP_APPLIED+"%d", master_uid)
	pclient.RedisExeCmd(pconfig.Comm, cb_add_user_appied, cb_arg, "SADD", tab_name, add_v)
	return
}

//cg_arg = {preq , from , grp_name , master_uid}
func cb_get_user_online_logic(comm_config *comm.CommConfig, result interface{}, cb_arg []interface{}) {
	var _func_ = "<cb_get_user_online_logic>"
	log := comm_config.Log

	/*---------mostly common logic--------------*/
	//get config
	pconfig, ok := comm_config.ServerCfg.(*Config)
	if !ok {
		log.Err("%s failed! convert config fail!", _func_)
		return
	}

	//conv arg
	preq, ok := cb_arg[0].(*ss.MsgApplyGroupReq)
	if !ok {
		log.Err("%s conv req failed!", _func_)
		return
	}

	from, ok := cb_arg[1].(int)
	if !ok {
		log.Err("%s conv from failed!", _func_)
		return
	}

	grp_name, ok := cb_arg[2].(string)
	if !ok {
		log.Err("%s conv grp_name failed!", _func_)
		return
	}

	master_uid, ok := cb_arg[3].(int64)
	if !ok {
		log.Err("%s conv master_uid failed!", _func_)
		return
	}

	/*---------result handle--------------*/
	//check result
	err, ok := result.(error)
	if ok {
		log.Err("%s exe failed! err:%v uid:%d", _func_, err, master_uid)
		return
	}

	//get result
	res, err := comm.Conv2String(result)
	if err != nil {
		log.Err("%s conv result failed! err:%v uid:%d", _func_, err, master_uid)
		return
	}
	if len(res) == 0 {
		log.Err("%s online_logic empty! uid:%d", _func_, master_uid)
		return
	}
	log.Debug("%s res:%s uid:%d", _func_, res, master_uid)

	//check online_logic
	online_logic, err := strconv.Atoi(res)
	if err != nil {
		log.Err("%s conv integer failed! uid:%d res:%s", _func_, master_uid, res)
		return
	}

	if online_logic <= 0 {
		log.Debug("%s offline! will save to db. uid:%d", _func_, master_uid)
		pclient := SelectRedisClient(pconfig , REDIS_OPT_W)
		if pclient == nil {
			log.Err("%s failed! no proper redis found! uid:%d grp_id:%d" , _func_ , preq.ApplyUid , preq.GroupId)
			return
		}

		//append to apply_list
		var push_v string
		if len(preq.ApplyMsg) == 0 {
			preq.ApplyMsg = " "
		}
		push_v = fmt.Sprintf("%d|%s|%s", preq.ApplyUid, preq.ApplyName, preq.ApplyMsg)

		tab_name := fmt.Sprintf(FORMAT_TAB_GROUP_APPLY_LIST+"%d", preq.GroupId)
		pclient.RedisExeCmd(pconfig.Comm, cb_push_apply_list, cb_arg, "RPUSH", tab_name, push_v)
		return
	}

	//notify
	var ss_msg ss.SSMsg
	pnotify := new(ss.MsgApplyGroupNotify)
	pnotify.GroupId = preq.GroupId
	pnotify.MasterUid = master_uid
	pnotify.ApplyUid = preq.ApplyUid
	pnotify.ApplyName = preq.ApplyName
	pnotify.GroupName = grp_name
	pnotify.Occupy = make([]int64, 1)
	pnotify.Occupy[0] = int64(online_logic)
	pnotify.ApplyMsg = preq.ApplyMsg

	err = comm.FillSSPkg(&ss_msg, ss.SS_PROTO_TYPE_APPLY_GROUP_NOTIFY, pnotify)
	if err != nil {
		log.Err("%s gen ss failed! uid:%d", _func_, master_uid)
	}

	//send to chat_serv
	SendToServ(pconfig, from, &ss_msg)
}
