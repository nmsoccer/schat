package lib

import (
	"fmt"
	"schat/proto/ss"
	"schat/servers/comm"
)

func RecvExitGroupReq(pconfig *Config , preq *ss.MsgExitGroupReq , from int) {
	if preq.DelGroup == 1 {
		do_del_group(pconfig , preq , from)
	} else {
		do_exit_group(pconfig , preq , from)
	}
}

func do_exit_group(pconfig *Config , preq *ss.MsgExitGroupReq , from int) {
	var _func_ = "<do_exit_group>"
	log := pconfig.Comm.Log
	uid := preq.Uid
	grp_id := preq.GrpId

	go func() {
		//head
		phead := pconfig.RedisClient.AllocSyncCmdHead()
		if phead == nil {
			log.Err("%s alloc head fail! uid:%d grp_id:%d" , _func_ , uid , grp_id)
			return
		}
		defer pconfig.RedisClient.FreeSyncCmdHead(phead)

		prsp := new(ss.MsgExitGroupRsp)
		prsp.GrpId = grp_id
		prsp.Uid = uid
		prsp.GrpName = preq.GrpName
		for {
			//check grp
			_ , result := GetGroupInfo(pconfig, phead, grp_id , FILED_GROUP_INFO_NAME)
            if result != ss.SS_COMMON_RESULT_SUCCESS {
            	prsp.Result = result
            	break
			}

			//exit group
			tab_name := fmt.Sprintf(FORMAT_TAB_GROUP_MEMBERS + "%d" , grp_id)
			_ , err := pconfig.RedisClient.RedisExeCmdSync(phead , "SREM" , tab_name , uid)
			if err != nil {
				log.Err("%s remove member failed! err:%v uid:%d grp_id:%d" , _func_ , err ,grp_id , uid)
				prsp.Result = ss.SS_COMMON_RESULT_FAILED
				break
			}
			log.Debug("%s srem done! uid:%d grp_id:%d" , _func_ , uid , grp_id)
			prsp.Result = ss.SS_COMMON_RESULT_SUCCESS
            break
		}

		//back
		var ss_msg ss.SSMsg
		err := comm.FillSSPkg(&ss_msg , ss.SS_PROTO_TYPE_EXIT_GROUP_RSP , prsp)
        if err != nil {
        	log.Err("%s gen ss failed! err:%v uid:%d grp_id:%d result:%d" , _func_ , err , uid , grp_id , prsp.Result)
        	return
		}

		SendToServ(pconfig , from , &ss_msg)
	}()

}

/*
*********GROUP TABLE*********
group:grp_id
group:mem:grp_id group:apply:grp_id chat_msg:group:index
*/
func do_del_group(pconfig *Config , preq *ss.MsgExitGroupReq , from int) {
	var _func_ = "<do_del_group>"
	log := pconfig.Comm.Log
	uid := preq.Uid
	grp_id := preq.GrpId

	go func() {
		//head
		phead := pconfig.RedisClient.AllocSyncCmdHead()
		if phead == nil {
			log.Err("%s alloc head fail! uid:%d grp_id:%d", _func_, uid, grp_id)
			return
		}
		defer pconfig.RedisClient.FreeSyncCmdHead(phead)

		prsp := new(ss.MsgExitGroupRsp)
		prsp.GrpId = grp_id
		prsp.Uid = uid
		prsp.GrpName = preq.GrpName
		prsp.DelGroup = 1
		for {
			//check grp
			res, result := GetGroupInfo(pconfig, phead, grp_id, FIELD_GROUP_INFO_MSG_COUNT)
			if result != ss.SS_COMMON_RESULT_SUCCESS {
				prsp.Result = result
				break
			}
			msg_count, err := comm.Conv2Int(res)
			if err != nil {
				log.Err("%s convert msg count failed! err:%v res:%v grp_id:%d uid:%d", _func_, err, res, grp_id, uid)
				prsp.Result = ss.SS_COMMON_RESULT_FAILED
				break
			}

			//del group:grp_id
			tab_name := fmt.Sprintf(FORMAT_TAB_GROUP_INFO_PREFIX + "%d" , grp_id)
			res , err = pconfig.RedisClient.RedisExeCmdSync(phead , "DEL" , tab_name)
			if err != nil {
				log.Err("%s del %s failed! err:%v grp_id:%d uid:%d" , _func_ , tab_name , err , grp_id , uid)
				prsp.Result = ss.SS_COMMON_RESULT_FAILED
				break
			}
			log.Info("%s del %s res:%v grp_id:%d uid:%d" , _func_ , tab_name , res , grp_id , uid)

            prsp.Result = ss.SS_COMMON_RESULT_SUCCESS //will always be success next
			//del group:apply:grp_id
			tab_name = fmt.Sprintf(FORMAT_TAB_GROUP_APPLY_LIST + "%d" , grp_id)
			res , err = pconfig.RedisClient.RedisExeCmdSync(phead , "DEL" , tab_name)
			if err != nil {
				log.Err("%s del %s failed! err:%v grp_id:%d uid:%d" , _func_ , tab_name , err , grp_id , uid)
			} else {
				log.Info("%s del %s res:%v grp_id:%d uid:%d", _func_, tab_name, res, grp_id, uid)
			}

			//del group:mem:grp_id
			tab_name = fmt.Sprintf(FORMAT_TAB_GROUP_MEMBERS + "%d" , grp_id)
			res , err = pconfig.RedisClient.RedisExeCmdSync(phead , "DEL" , tab_name)
			if err != nil {
				log.Err("%s del %s failed! err:%v grp_id:%d uid:%d" , _func_ , tab_name , err , grp_id , uid)
			} else {
				log.Info("%s del %s res:%v grp_id:%d uid:%d", _func_, tab_name, res, grp_id, uid)
			}

			//del chat_msg:group:index
    		if msg_count > 0 {
    			log.Info("%s will del chat msg! total:%d grp_id:%d uid:%d" , _func_ , msg_count , grp_id , uid)
    			tab_cnt := msg_count / CHAT_MSG_LIST_SIZE
    			for i:=0; i<=tab_cnt; i++ {
					tab_name = fmt.Sprintf(FORMAT_TAB_CHAT_MSG_LIST , grp_id , i)
					res , err = pconfig.RedisClient.RedisExeCmdSync(phead , "DEL" , tab_name)
					if err != nil {
						log.Err("%s del %s failed! err:%v grp_id:%d uid:%d" , _func_ , tab_name , err , grp_id , uid)
					} else {
						log.Info("%s del %s res:%v grp_id:%d uid:%d", _func_, tab_name, res, grp_id, uid)
					}
				}
			}

			break
		}

		//back
		var ss_msg ss.SSMsg
		err := comm.FillSSPkg(&ss_msg , ss.SS_PROTO_TYPE_EXIT_GROUP_RSP , prsp)
		if err != nil {
			log.Err("%s gen ss failed! err:%v uid:%d grp_id:%d result:%d" , _func_ , err , uid , grp_id , prsp.Result)
			return
		}

		SendToServ(pconfig , from , &ss_msg)
	}()
}