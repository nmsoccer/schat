package lib

import (
	"fmt"
	"schat/proto/ss"
	"schat/servers/comm"
	"strconv"
)

func RecvLoadGroupReq(pconfig *Config , preq *ss.MsgLoadGroupReq , from int) {
	var _func_ = "<RecvLoadGroupReq>"
	log := pconfig.Comm.Log

	log.Debug("%s uid:%d reason:%d" , _func_ , preq.Uid , preq.Reason)
	//basic check
	if preq.Reason == ss.LOAD_GROUP_REASON_LOAD_GRP_SEND_CHAT && preq.ChatMsg==nil {
		log.Err("%s no chat msg found! uid:%d" , _func_ , preq.Uid)
		return
	}

	//load
	go func() {
		var err error
		uid := preq.Uid
		grp_id := preq.GrpId
		phead := pconfig.RedisClient.AllocSyncCmdHead()
		result := ss.SS_COMMON_RESULT_FAILED
		if phead == nil {
			log.Err("%s alloc head failed! uid:%d" , _func_ , preq.Uid)
			return
		}
		defer pconfig.RedisClient.FreeSyncCmdHead(phead)

		//Handle
		pgroup := new(ss.GroupInfo)
		for {
			//load group info
			tab_name := fmt.Sprintf(FORMAT_TAB_GROUP_INFO_PREFIX+"%d", grp_id)
			res, err := pconfig.RedisClient.RedisExeCmdSync(phead, "HMGET", tab_name, "name", "master_uid",
				"create_ts", "msg_count")
			if err != nil {
				log.Err("%s load group info failed!err:%v uid:%d grp_id:%d", _func_, err, uid, grp_id)
				break
			}
			ret_v := load_group_basic_info(pconfig, res, preq, pgroup)
			if ret_v < 0 {
				if ret_v == -2 { //not exist
					result = ss.SS_COMMON_RESULT_NOEXIST
				}
				break
			}

			//load member
			tab_name = fmt.Sprintf(FORMAT_TAB_GROUP_MEMBERS+"%d", grp_id)
			res, err = pconfig.RedisClient.RedisExeCmdSync(phead, "SMEMBERS", tab_name)
			if err != nil {
				log.Err("%s load group members failed!err:%v uid:%d grp_id:%d", _func_, err, uid, grp_id)
				break
			}
			ok := load_group_member_info(pconfig, res, preq, pgroup)
			if !ok {
				break
			}

			//success
			result = ss.SS_COMMON_RESULT_SUCCESS
			break
		}

		//back to serv
		var ss_msg ss.SSMsg
		prsp := new(ss.MsgLoadGroupRsp)
		prsp.GrpId = preq.GrpId
		prsp.ChatMsg = preq.ChatMsg
		prsp.Uid = preq.Uid
		prsp.Reason = preq.Reason
		prsp.Occoupy = preq.Occoupy
		prsp.TempId = preq.TempId
		prsp.LoadResult = result
		prsp.GrpInfo = pgroup
		prsp.IntV = preq.IntV
		prsp.StrV = preq.StrV

		err = comm.FillSSPkg(&ss_msg , ss.SS_PROTO_TYPE_LOAD_GROUP_RSP , prsp)
		if err != nil {
			log.Err("%s gen ss failed!err:%v uid:%d grp_id:%d" , _func_ , err , uid , grp_id)
			return
		}

		SendToServ(pconfig , from , &ss_msg)
	}()
}


func RecvSaveChatGroupReq(pconfig *Config , preq *ss.MsgSaveGroupReq , from int) {
	var _func_ = "<RecvSaveChatGroupReq>"
	log := pconfig.Comm.Log
	grp_id := preq.GrpId

	//Sync
	go func() {
		phead := pconfig.RedisClient.AllocSyncCmdHead()
		if phead == nil {
			log.Err("%s alloc head failed! grp_id:%d" , _func_ , grp_id)
			return
		}
		defer pconfig.RedisClient.FreeSyncCmdHead(phead)

		prsp := new(ss.MsgSaveGroupRsp)
		prsp.GrpId = preq.GrpId
		var result = ss.SS_COMMON_RESULT_FAILED
		//Handle
		for {
			//Step1. Check Group Exist
			_ , result = GetGroupInfo(pconfig , phead , grp_id , FILED_GROUP_INFO_NAME)
			if result != ss.SS_COMMON_RESULT_SUCCESS {
				log.Err("%s check grp exist failed! grp_id:%d", _func_, grp_id)
				break
			}


			//Step2. Save Group Info
			tab_name := fmt.Sprintf(FORMAT_TAB_GROUP_INFO_PREFIX+"%d", grp_id)
			res , err := pconfig.RedisClient.RedisExeCmdSync(phead , "HMSET" , tab_name , "msg_count" , preq.MsgCount , "load_serv" ,
				preq.LoadServ)
			if err != nil {
				log.Err("%s hmset %s failed! err:%v grp_id:%d", _func_, tab_name, err, grp_id)
				break
			}
			/*
			str_v , err := comm.Conv2String(res)
			if err != nil {
				log.Err("%s convert hmset res failed! err:%v res:%v grp_id:%d" , _func_ , err , res , grp_id)
				break
			}
			if str_v != "OK" {
				log.Err("%s save failed! err:%v res:%v grp_id:%d" , _func_ , err , res , grp_id)
			}*/
			//log.Debug("%s save stat success! grp_id:%d" , _func_ , grp_id)
			result = ss.SS_COMMON_RESULT_SUCCESS

			//server exit no need further
			if preq.Reason == ss.SS_COMMON_REASON_REASON_EXIT {
				break
			}


			//step3. Get MemCount
			tab_name = fmt.Sprintf(FORMAT_TAB_GROUP_MEMBERS + "%d" , grp_id)
			res , err = pconfig.RedisClient.RedisExeCmdSync(phead , "SCARD" , tab_name)
			if err != nil {
				log.Err("%s scard %s failed! err:%v grp_id:%d", _func_, tab_name, err, grp_id)
				break
			}
			mem_count , err := comm.Conv2Int(res)
			if err != nil {
				log.Err("%s convert mem_count failed! err:%v grp_id:%d res:%v", _func_, err, grp_id , res)
				break
			}
			//count equal
			if mem_count == int(preq.MemCount) {
				break
			}

			log.Debug("%s mem_count not match will reload members! %d:%d grp_Id:%d" , _func_ , mem_count , preq.MemCount , grp_id)
			//count not match will sync
			prsp.MemberChged = 1
			if mem_count == 0 {	//no more member
				break
			}

			//step4. Load Members
			res , err = pconfig.RedisClient.RedisExeCmdSync(phead , "SMEMBERS" , tab_name)
			if err != nil {
				log.Err("%s smembers %s failed! err:%v grp_id:%d", _func_, tab_name, err, grp_id)
				break
			}
			strs, err := comm.Conv2Strings(res)
			if err != nil {
				log.Err("%s convert members failed!err:%v grp_id:%d", _func_, err, grp_id)
				break
			}
			if len(strs) == 0 {
				log.Info("%s empty member! grp_id:%d" , _func_ , grp_id)
				break
			}

			//Member
			prsp.Members = make(map[int64]int32)
			var member_uid int64 = 0
			for i:=0; i<len(strs); i++ {
				member_uid , err = strconv.ParseInt(strs[i] , 10 , 64)
				if err != nil {
					log.Err("%s convert member_uid failed!err:%v grp_id:%d str:%s", _func_, err, grp_id , strs[i])
					continue
				}
				prsp.Members[member_uid] = 1
			}
			break
		}

		//Pack
		var ss_msg ss.SSMsg
		prsp.Result = result

		err := comm.FillSSPkg(&ss_msg , ss.SS_PROTO_TYPE_SAVE_GROUP_RSP , prsp)
		if err != nil {
			log.Err("%s gen ss failed! err:%v grp_id:%d" , _func_ , err , grp_id)
			return
		}

		//Send
		SendToServ(pconfig , from , &ss_msg)
	}()

}


/*-----------------------static func-----------------------*/
//@return  -2:no group -1:err 0:success
func load_group_basic_info(pconfig *Config , res interface{} , preq *ss.MsgLoadGroupReq , pgroup *ss.GroupInfo) int {
	var _func_ = "<load_group_basic_info>"
	log := pconfig.Comm.Log
	uid := preq.Uid
	grp_id := preq.GrpId

	strs , err := comm.Conv2Strings(res)
	if err != nil {
		log.Err("%s convert group info failed!err:%v uid:%d grp_id:%d" , _func_ , err , uid , grp_id)
		return -1
	}
	if len(strs) != 4 {
		log.Err("%s group info filed num err! uid:%d grp_id:%d res:%v" , _func_ , uid , grp_id , strs)
		return -1
	}
	if len(strs[0]) == 0 {
		log.Err("%s group not exist! uid:%d grp_id:%d res:%v" , _func_ , uid , grp_id , strs)
		return -2
	}


	//group info
	pgroup.GroupId = grp_id
	pgroup.GroupName = strs[0]
	pgroup.MasterUid , err = strconv.ParseInt(strs[1] , 10 , 64)
	if err != nil {
		log.Err("%s convert master uid failed! err:%v uid:%d grp_id:%d" , _func_ , err , uid , grp_id)
		return -1
	}
	pgroup.CreatedTs , err = strconv.ParseInt(strs[2] , 10 , 64)
	if err != nil {
		log.Err("%s convert create_ts failed! err:%v uid:%d grp_id:%d" , _func_ , err , uid , grp_id)
		return -1
	}
	pgroup.LatestMsgId , err = strconv.ParseInt(strs[3] , 10 , 64)
	if err != nil {
		log.Err("%s convert msg_count failed! err:%v uid:%d grp_id:%d" , _func_ , err , uid , grp_id)
		return -1
	}

    return 0
}

//@return  ok
func load_group_member_info(pconfig *Config , res interface{} , preq *ss.MsgLoadGroupReq , pgroup *ss.GroupInfo) bool {
	var _func_ = "<load_group_member_info>"
	log := pconfig.Comm.Log
	uid := preq.Uid
	grp_id := preq.GrpId

	strs, err := comm.Conv2Strings(res)
	if err != nil {
		log.Err("%s convert group info failed!err:%v uid:%d grp_id:%d", _func_, err, uid, grp_id)
		return false
	}
	if len(strs) == 0 {
		log.Info("%s empty member! uid:%d grp_id:%d" , _func_ , uid , grp_id)
		return true
	}

	//Input Member
	pgroup.Members = make(map[int64]int32)
	var member_uid int64 = 0
	for i:=0; i<len(strs); i++ {
	    member_uid , err = strconv.ParseInt(strs[i] , 10 , 64)
	    if err != nil {
			log.Err("%s convert member_uid failed!err:%v uid:%d grp_id:%d", _func_, err, uid, grp_id)
			continue
		}
		pgroup.Members[member_uid] = 1
	}
    pgroup.MemCount = int32(len(pgroup.Members))

    return true
}