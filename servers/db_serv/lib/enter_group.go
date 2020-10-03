package lib

import (
	"fmt"
	"schat/proto/ss"
	"schat/servers/comm"
)

func RecvEnterGroupReq(pconfig *Config , preq *ss.MsgEnterGroupReq , from int) {
	var _func_ = "<RecvEnterGroupReq>"
	log := pconfig.Comm.Log

	log.Debug("%s uid:%d grp_id:%d from:%d" , _func_ , preq.Uid , preq.GrpId , from)
	//sync
	go func() {
		uid := preq.Uid
		grp_id := preq.GrpId

		//head
		phead := pconfig.RedisClient.AllocSyncCmdHead()
		if phead == nil {
			log.Err("%s alloc head fail! uid:%d grp_id:%d" , _func_ , uid , grp_id)
			return
		}
		defer pconfig.RedisClient.FreeSyncCmdHead(phead)

		//Get Group Info
		tab_name := fmt.Sprintf(FORMAT_TAB_GROUP_INFO_PREFIX + "%d" , grp_id)
		res , err := pconfig.RedisClient.RedisExeCmdSync(phead , "HGET" , tab_name , "name")
		if err != nil {
			log.Err("%s query group failed! err:%v uid:%d grp_id:%d" , _func_ , err ,grp_id , uid)
			return
		}
		if res == nil {
			log.Err("%s group not exist! uid:%d grp_id:%d" , _func_ , uid , grp_id)
			SendEnterGroupRsp(pconfig , preq , "" , from , 1)
			return
		}
		grp_name , err := comm.Conv2String(res)
		if err != nil {
			log.Err("%s group name convert failed!! uid:%d grp_id:%d res:%v err:%v" , _func_ , uid , grp_id , res , err)
			return
		}

		//Add Member
		tab_name = fmt.Sprintf(FORMAT_TAB_GROUP_MEMBERS + "%d" , grp_id)
		res , err = pconfig.RedisClient.RedisExeCmdSync(phead , "SADD" , tab_name , uid)
		if err != nil {
			log.Err("%s add member failed! err:%v uid:%d grp_id:%d" , _func_ , err ,grp_id , uid)
			return
		}
		result , err := comm.Conv2Int(res)
		if err != nil {
			log.Err("%s sadd res convert failed!! uid:%d grp_id:%d res:%v err:%v" , _func_ , uid , grp_id , res , err)
			return
		}
		log.Debug("%s sadd result:%d uid:%d grp_id:%d" , _func_ , result , uid , grp_id)


		//Resp
		SendEnterGroupRsp(pconfig , preq , grp_name , from , 0)
	}()
}

//ret:0:ok 1:no-exist
func SendEnterGroupRsp(pconfig *Config , preq *ss.MsgEnterGroupReq , grp_name string , target_serv int , ret int32) {
	var _func_ = "<SendEnterGroupRsp>"
	log := pconfig.Comm.Log

	//ss
	prsp := new(ss.MsgEnterGroupRsp)
	prsp.Uid = preq.Uid
	prsp.GrpId = preq.GrpId
	prsp.GrpName = grp_name
	prsp.Result = ret
	prsp.Occupy = preq.Occupy

	var ss_msg ss.SSMsg
	err := comm.FillSSPkg(&ss_msg , ss.SS_PROTO_TYPE_ENTER_GROUP_RSP , prsp)
	if err != nil {
		log.Err("%s gen ss failed! uid:%d grp_id:%d ret:%d err:%v" , _func_ , preq.Uid , preq.GrpId , ret , err)
		return
	}

	//Send
	SendToServ(pconfig , target_serv , &ss_msg)
}