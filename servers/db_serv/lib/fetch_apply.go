package lib

import (
	"fmt"
	"schat/proto/ss"
	"schat/servers/comm"
	"strconv"
	"strings"
)

func RecvFetchApplyGroupReq(pconfig *Config , preq *ss.MsgFetchApplyGrpReq , from int) {
	var _func_ = "<RecvFetchApplyGroupReq>"
	log := pconfig.Comm.Log

	log.Debug("%s uid:%d  from:%d" , _func_ , preq.Uid , from)
	//get grp info
	cb_arg := []interface{}{preq, from}
	tab_name := fmt.Sprintf(FORMAT_TAB_USER_GROUP_APPLIED+"%d", preq.Uid)
	//tab_name := "user:group:applied:10004"
	pconfig.RedisClient.RedisExeCmd(pconfig.Comm , cb_rand_user_appied , cb_arg , "SRANDMEMBER" , tab_name)
}

//empty means error or no data or complete
func SendFetchApplyGroupEmpty(pconfig *Config , preq *ss.MsgFetchApplyGrpReq , from int , complete int) {
	var _func_ = "<SendFetchApplyGroupEmpty>"
	log := pconfig.Comm.Log

	//ss
	var ss_msg ss.SSMsg
	prsp := new(ss.MsgFetchApplyGrpRsp)
	prsp.Complete = int32(complete)
	prsp.Uid = preq.Uid
	prsp.FetchCount = 0

	err := comm.FillSSPkg(&ss_msg , ss.SS_PROTO_TYPE_FETCH_APPLY_GROUP_RSP , prsp)
	if err != nil {
		log.Err("%s gen ss fail! err:%v uid:%d" , _func_ , err , preq.Uid)
		return
	}

	//send
	SendToServ(pconfig , from , &ss_msg)
}



/*-----------------------static func-----------------------*/
//cg_arg = {preq , from}
func cb_rand_user_appied(comm_config *comm.CommConfig, result interface{}, cb_arg []interface{}) {
	var _func_ = "<cb_rand_user_appied>"
	log := comm_config.Log

	/*---------mostly common logic--------------*/
	//get config
	pconfig, ok := comm_config.ServerCfg.(*Config)
	if !ok {
		log.Err("%s failed! convert config fail!", _func_)
		return
	}

	//conv arg
	preq, ok := cb_arg[0].(*ss.MsgFetchApplyGrpReq)
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
		log.Err("%s exe failed! err:%v uid:%d", _func_, err, preq.Uid)
		return
	}

	//get result
	if result == nil {
		log.Debug("%s empty appied! uid:%d" , _func_ , preq.Uid)
		SendFetchApplyGroupEmpty(pconfig , preq , from , 1)
		return
	}

	res, err := comm.Conv2String(result)
	if err != nil {
		log.Err("%s conv result failed! err:%v uid:%d", _func_, err, preq.Uid)
		return
	}
	log.Debug("%s add result:%s uid:%d", _func_, res, preq.Uid)

	//no data
	if len(res) == 0 {
		log.Debug("%s no more data! uid:%d" , _func_ , preq.Uid)
		SendFetchApplyGroupEmpty(pconfig , preq , from , 1)
		return
	}


	//split grp_id|grp_name
	splits := strings.Split(res , "|")
	if len(splits) != 2 {
		log.Err("%s split failed! res:%s uid:%d" , _func_ , res , preq.Uid)
		return
	}

	//conv
	grp_id , err := strconv.ParseInt(splits[0] , 10 , 64)
	if err != nil {
		log.Err("%s conv grp_id failed! err:%v res:%s uid:%d" , _func_ , err , res , preq.Uid)
		return
	}
	grp_name := splits[1]

	//Fetch Apply List
	tab_name := fmt.Sprintf(FORMAT_TAB_GROUP_APPLY_LIST+"%d" , grp_id)
	pconfig.RedisClient.RedisExeCmd(pconfig.Comm , cb_range_apply_list , append(cb_arg , grp_id , grp_name) , "LRANGE" ,
		tab_name , 0 , preq.FetchCount-1)
}


//cg_arg = {preq , from , grp_id , grp_name}
func cb_range_apply_list(comm_config *comm.CommConfig, result interface{}, cb_arg []interface{}) {
	var _func_ = "<cb_range_apply_list>"
	log := comm_config.Log

	/*---------mostly common logic--------------*/
	//get config
	pconfig, ok := comm_config.ServerCfg.(*Config)
	if !ok {
		log.Err("%s failed! convert config fail!", _func_)
		return
	}

	//conv arg
	preq, ok := cb_arg[0].(*ss.MsgFetchApplyGrpReq)
	if !ok {
		log.Err("%s conv req failed!", _func_)
		return
	}

	from, ok := cb_arg[1].(int)
	if !ok {
		log.Err("%s conv from failed! uid:%d", _func_, preq.Uid)
		return
	}

	grp_id, ok := cb_arg[2].(int64)
	if !ok {
		log.Err("%s conv grp_id failed! uid:%d", _func_, preq.Uid)
		return
	}

	grp_name, ok := cb_arg[3].(string)
	if !ok {
		log.Err("%s conv grp_name failed! uid:%d", _func_, preq.Uid)
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
	prsp := new(ss.MsgFetchApplyGrpRsp)
	prsp.Uid = preq.Uid
	prsp.Complete = 0
	prsp.NotifyList = make([]*ss.MsgApplyGroupNotify , len(res))
	i := 0
		//parse
	for idx:=0 ; idx<len(res); idx++{ //content:apply_uid|apply_name|apply_msg
		splits := strings.Split(res[idx] , "|")
		if len(splits) != 3 {
			log.Err("%s split:%s illegal! uid:%d" , _func_ , res[idx] , preq.Uid)
			continue
		}

		//conv apply_uid
		apply_uid , err := strconv.ParseInt(splits[0] , 10 ,64)
		if err != nil {
			log.Err("%s split:%s conv grp_id failed! uid:%d err:%v" , _func_ , splits[0] , preq.Uid , err)
			continue
		}

		//fill
		prsp.NotifyList[i] = new(ss.MsgApplyGroupNotify)
		prsp.NotifyList[i].GroupName = grp_name
		prsp.NotifyList[i].GroupId = grp_id
		prsp.NotifyList[i].ApplyName = splits[1]
		prsp.NotifyList[i].ApplyUid = apply_uid
		prsp.NotifyList[i].ApplyMsg = splits[2]
		prsp.NotifyList[i].MasterUid = preq.Uid
		i++
	}
	prsp.FetchCount = int32(i)

	err = comm.FillSSPkg(&ss_msg , ss.SS_PROTO_TYPE_FETCH_APPLY_GROUP_RSP , prsp)
	if err != nil {
		log.Err("%s gen ss failed! err:%v uid:%d" , _func_ , err , preq.Uid)
		return
	}

	SendToServ(pconfig , from , &ss_msg)

	//del apply list
	tab_name := fmt.Sprintf(FORMAT_TAB_GROUP_APPLY_LIST+"%d" , grp_id)
	pconfig.RedisClient.RedisExeCmd(pconfig.Comm , nil , cb_arg , "LTRIM" ,
		tab_name , preq.FetchCount , -1)

	//del appied
    if int32(len(res)) < preq.FetchCount {
    	log.Debug("%s no more apply of group:%d uid:%d" , _func_ , grp_id , preq.Uid)
		tab_name = fmt.Sprintf(FORMAT_TAB_USER_GROUP_APPLIED+"%d", preq.Uid)
		pconfig.RedisClient.RedisExeCmd(pconfig.Comm , nil , cb_arg , "SREM" , tab_name ,
			fmt.Sprintf("%d|%s" , grp_id , grp_name))
	}


}