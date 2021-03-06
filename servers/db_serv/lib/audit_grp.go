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

	go func() {
		uid := preq.Uid
		grp_id := preq.GroupId
		log.Debug("%s grp_id:%d uid:%d apply_uid:%d from:%d", _func_, grp_id, uid, preq.ApplyUid, from)
		//pclient
		pclient := SelectRedisClient(pconfig , REDIS_OPT_W)
		if pclient == nil {
			log.Err("%s failed! no proper redis found! uid:%d grp_id:%d apply:%d" , _func_ , uid , grp_id ,
				preq.ApplyUid)
			return
		}
		//phead
		phead := pclient.AllocSyncCmdHead()
		if phead == nil {
			log.Err("%s alloc head failed! uid:%d", _func_, uid)
			return
		}
		defer pclient.FreeSyncCmdHead(phead)

		//v
		push_v := fmt.Sprintf("%d|%s|%d|%d", preq.GroupId, preq.GroupName, preq.Result , preq.Flag)
		tab_name := fmt.Sprintf(FORMAT_TAB_USER_GROUP_AUDITED+"%d", preq.ApplyUid)
		_ , err := pclient.RedisExeCmdSync(phead , "RPUSH" , tab_name , push_v)
		if err != nil {
			log.Err("%s exe RPUSH %s %s failed! err:%v uid:%d" , _func_ , tab_name , push_v , err , uid)
			return
		}

		//back to serv
		var ss_msg ss.SSMsg
		preq.FromDb = 1

		err = comm.FillSSPkg(&ss_msg , ss.SS_PROTO_TYPE_APPLY_GROUP_AUDIT , preq)
		if err != nil {
			log.Err("%s gen ss failed! err:%v uid:%d" , _func_ , err , uid)
			return
		}

		SendToServ(pconfig , from , &ss_msg)
	}()

}


func RecvFetchAuditGroupReq(pconfig *Config, preq *ss.MsgFetchAuditGrpReq, from int) {
	var _func_ = "<RecvFetchAuditGroupReq>"
	log := pconfig.Comm.Log

	log.Debug("%s uid:%d  from:%d", _func_, preq.Uid, from)
	//pclient
	pclient := SelectRedisClient(pconfig , REDIS_OPT_R)
	if pclient == nil {
		log.Err("%s failed! no proper redis found! uid:%d" , _func_ , preq.Uid)
		return
	}


	//get grp info
	cb_arg := []interface{}{preq, from}
	tab_name := fmt.Sprintf(FORMAT_TAB_USER_GROUP_AUDITED+"%d", preq.Uid)
	pclient.RedisExeCmd(pconfig.Comm, cb_range_user_audied, cb_arg, "LRANGE",
		tab_name, 0, preq.FetchCount-1)
}

/*-----------------------static func-----------------------*/
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
	prsp.AuditList = make([]*ss.MsgApplyGroupAudit, len(res))
	i := 0
	//parse
	for idx := 0; idx < len(res); idx++ { //content:grp_id|grp_name|result
		splits := strings.Split(res[idx], "|")
		if len(splits) != 4 {
			log.Err("%s split:%s illegal! uid:%d", _func_, res[idx], preq.Uid)
			continue
		}

		//conv apply_uid
		grp_id, err := strconv.ParseInt(splits[0], 10, 64)
		if err != nil {
			log.Err("%s split:%s conv grp_id failed! uid:%d err:%v", _func_, splits[0], preq.Uid, err)
			continue
		}

		audit_result, err := strconv.Atoi(splits[2])
		if err != nil {
			log.Err("%s split:%s conv audit_result failed! uid:%d err:%v", _func_, splits[2], preq.Uid, err)
			continue
		}

		aud_flag , err := strconv.Atoi(splits[3])
		if err != nil {
			log.Err("%s split:%s conv audit_flag failed! uid:%d err:%v", _func_, splits[3], preq.Uid, err)
			continue
		}


		//fill
		prsp.AuditList[i] = new(ss.MsgApplyGroupAudit)
		prsp.AuditList[i].GroupId = grp_id
		prsp.AuditList[i].GroupName = splits[1]
		prsp.AuditList[i].Result = ss.APPLY_GROUP_RESULT(audit_result)
		prsp.AuditList[i].Flag = int32(aud_flag)
		i++
	}
	prsp.FetchCount = int32(i)

	err = comm.FillSSPkg(&ss_msg, ss.SS_PROTO_TYPE_FETCH_AUDIT_GROUP_RSP, prsp)
	if err != nil {
		log.Err("%s gen ss failed! err:%v uid:%d", _func_, err, preq.Uid)
		return
	}
	SendToServ(pconfig, from, &ss_msg)

	//pclient
	pclient := SelectRedisClient(pconfig , REDIS_OPT_W)
	if pclient == nil {
		log.Err("%s failed! no proper redis found! uid:%d" , _func_ , preq.Uid)
		return
	}

	//del apply list
	tab_name := fmt.Sprintf(FORMAT_TAB_USER_GROUP_AUDITED+"%d", preq.Uid)
	pclient.RedisExeCmd(pconfig.Comm, nil, cb_arg, "LTRIM",
		tab_name, preq.FetchCount, -1)
}
