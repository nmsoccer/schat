package lib

import (
	"fmt"
	"schat/proto/ss"
	"schat/servers/comm"
	"strconv"
	"strings"
)

func RecvApplyGroupAudit(pconfig *Config , preq *ss.MsgApplyGroupAudit , from int) {
	var _func_ = "<RecvApplyGroupAudit>"
	log := pconfig.Comm.Log

	log.Debug("%s grp_id:%d uid:%d apply_uid:%d from:%d", _func_, preq.GroupId, preq.ApplyUid, preq.ApplyUid, from)
	push_v := fmt.Sprintf("%d|%s|%d" , preq.GroupId , preq.GroupName , preq.Result)

	tab_name := fmt.Sprintf(FORMAT_TAB_USER_GROUP_AUDITED+"%d" , preq.ApplyUid)
	pconfig.RedisClient.RedisExeCmd(pconfig.Comm , cb_push_audit_list , []interface{}{preq , from} , "RPUSH" , tab_name , push_v)
	return
}

func RecvFetchAuditGroupReq(pconfig *Config , preq *ss.MsgFetchAuditGrpReq , from int) {
	var _func_ = "<RecvFetchAuditGroupReq>"
	log := pconfig.Comm.Log

	log.Debug("%s uid:%d  from:%d" , _func_ , preq.Uid , from)
	//get grp info
	cb_arg := []interface{}{preq, from}
	tab_name := fmt.Sprintf(FORMAT_TAB_USER_GROUP_AUDITED+"%d", preq.Uid)
	pconfig.RedisClient.RedisExeCmd(pconfig.Comm , cb_range_user_audied , cb_arg , "LRANGE" ,
		tab_name , 0 , preq.FetchCount-1)
}


/*-----------------------static func-----------------------*/
//cg_arg = {preq , from , uid}
func cb_push_audit_list(comm_config *comm.CommConfig, result interface{}, cb_arg []interface{}) {
	var _func_ = "<cb_push_audit_list>"
	log := comm_config.Log

	/*---------mostly common logic--------------*/
	//get config
	pconfig, ok := comm_config.ServerCfg.(*Config)
	if !ok {
		log.Err("%s failed! convert config fail!", _func_)
		return
	}

	//conv arg
	preq, ok := cb_arg[0].(*ss.MsgApplyGroupAudit)
	if !ok {
		log.Err("%s conv req failed!", _func_)
		return
	}


	/*---------result handle--------------*/
	//check result
	err, ok := result.(error)
	if ok {
		log.Err("%s exe failed! err:%v uid:%s", _func_, err, preq.Uid)
		return
	}

	//get result
	res, err := comm.Conv2Int(result)
	if err != nil {
		log.Err("%s conv result failed! err:%v uid:%s", _func_, err, preq.Uid)
		return
	}
	log.Debug("%s audited_len:%d apply_uid:%d", _func_, res, preq.ApplyUid)


	//check online
	tab_name := fmt.Sprintf(FORMAT_TAB_USER_INFO_REFIX + "%d" , preq.ApplyUid)
	pconfig.RedisClient.RedisExeCmd(pconfig.Comm, cb_audit_grp_online_logic, cb_arg, "HGET", tab_name , FIELD_USER_INFO_ONLINE_LOGIC)
}

//cg_arg = {preq , from}
func cb_audit_grp_online_logic(comm_config *comm.CommConfig, result interface{}, cb_arg []interface{}) {
	var _func_ = "<cb_audit_grp_online_logic>"
	log := comm_config.Log

	/*---------mostly common logic--------------*/
	//get config
	pconfig, ok := comm_config.ServerCfg.(*Config)
	if !ok {
		log.Err("%s failed! convert config fail!", _func_)
		return
	}

	//conv arg
	preq, ok := cb_arg[0].(*ss.MsgApplyGroupAudit)
	if !ok {
		log.Err("%s conv req failed!", _func_)
		return
	}

	from, ok := cb_arg[1].(int)
	if !ok {
		log.Err("%s conv from failed!", _func_)
		return
	}

	/*---------result handle--------------*/
	//check result
	err, ok := result.(error)
	if ok {
		log.Err("%s exe failed! err:%v uid:%d", _func_, err, preq.ApplyUid)
		return
	}

	//get result
	res, err := comm.Conv2String(result)
	if err != nil {
		log.Err("%s conv result failed! err:%v uid:%d", _func_, err, preq.ApplyUid)
		return
	}
	if len(res) == 0 {
		log.Err("%s online_logic empty! uid:%d" , _func_ , preq.ApplyUid)
		return
	}
	log.Debug("%s res:%s uid:%d", _func_, res, preq.ApplyUid)

	//check online_logic
	online_logic , err := strconv.Atoi(res)
	if err != nil {
		log.Err("%s conv integer failed! uid:%d res:%s" , _func_ , preq.ApplyUid , res)
		return
	}

	if online_logic <= 0 {
		log.Debug("%s offline! uid:%d" , _func_ , preq.ApplyUid)
		return
	}

	//common notify to applied user
	var ss_msg ss.SSMsg
	pnotify := new(ss.MsgCommonNotify)
	pnotify.Uid = preq.ApplyUid
	pnotify.NotifyType = ss.COMMON_NOTIFY_TYPE_NOTIFY_NEW_AUDIT
	pnotify.IntV = int64(online_logic)

	err = comm.FillSSPkg(&ss_msg , ss.SS_PROTO_TYPE_COMMON_NOTIFY , pnotify)
	if err != nil {
		log.Err("%s gen ss failed! uid:%d" , _func_ , preq.ApplyUid)
	}

	//send to chat_serv
	SendToServ(pconfig , from , &ss_msg)
}


//cg_arg = {preq , from}
func cb_range_user_audied(comm_config *comm.CommConfig, result interface{}, cb_arg []interface{}) {
	var _func_ = "<cb_range_user_audied>"
	log := comm_config.Log

	/*---------mostly common logic--------------*/
	//get config
	pconfig, ok := comm_config.ServerCfg.(*Config)
	if !ok {
		log.Err("%s failed! convert config fail!", _func_)
		return
	}

	//conv arg
	preq, ok := cb_arg[0].(*ss.MsgFetchAuditGrpReq)
	if !ok {
		log.Err("%s conv req failed!", _func_)
		return
	}

	from, ok := cb_arg[1].(int)
	if !ok {
		log.Err("%s conv from failed! uid:%d", _func_, preq.Uid)
		return
	}

	/*---------result handle--------------*/
	//check result
	err, ok := result.(error)
	if ok {
		log.Err("%s exe failed! err:%v uid:%s", _func_, err, preq.Uid)
		return
	}

	//get result
	res, err := comm.Conv2Strings(result)
	if err != nil {
		log.Err("%s conv result failed! err:%v uid:%s", _func_, err, preq.Uid)
		return
	}
	log.Debug("%s add result:%v uid:%d", _func_, res, preq.Uid)

	//fill info
	var ss_msg ss.SSMsg
	prsp := new(ss.MsgFetchAuditGrpRsp)
	prsp.Uid = preq.Uid
	if int32(len(res)) < preq.FetchCount {
		prsp.Complete = 1
	}
	prsp.AuditList = make([]*ss.MsgApplyGroupAudit , len(res))
	i := 0
	//parse
	for idx:=0 ; idx<len(res); idx++{ //content:grp_id|grp_name|result
		splits := strings.Split(res[idx] , "|")
		if len(splits) != 3 {
			log.Err("%s split:%s illegal! uid:%d" , _func_ , res[idx] , preq.Uid)
			continue
		}

		//conv apply_uid
		grp_id , err := strconv.ParseInt(splits[0] , 10 ,64)
		if err != nil {
			log.Err("%s split:%s conv grp_id failed! uid:%d err:%v" , _func_ , splits[0] , preq.Uid , err)
			continue
		}

		audit_result , err := strconv.Atoi(splits[2])
		if err != nil {
			log.Err("%s split:%s conv audit_result failed! uid:%d err:%v" , _func_ , splits[2] , preq.Uid , err)
			continue
		}


		//fill
		prsp.AuditList[i] = new(ss.MsgApplyGroupAudit)
		prsp.AuditList[i].GroupId = grp_id
		prsp.AuditList[i].GroupName = splits[1]
		prsp.AuditList[i].Result = ss.APPLY_GROUP_RESULT(audit_result)
		i++
	}
	prsp.FetchCount = int32(i)


	err = comm.FillSSPkg(&ss_msg , ss.SS_PROTO_TYPE_FETCH_AUDIT_GROUP_RSP , prsp)
	if err != nil {
		log.Err("%s gen ss failed! err:%v uid:%d" , _func_ , err , preq.Uid)
		return
	}

	SendToServ(pconfig , from , &ss_msg)

	//del apply list
	tab_name := fmt.Sprintf(FORMAT_TAB_USER_GROUP_AUDITED+"%d" , preq.Uid)
	pconfig.RedisClient.RedisExeCmd(pconfig.Comm , nil , cb_arg , "LTRIM" ,
		tab_name , preq.FetchCount , -1)
}