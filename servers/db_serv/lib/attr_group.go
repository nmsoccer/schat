package lib

import (
	"errors"
	"fmt"
	"schat/proto/ss"
	"schat/servers/comm"
	"strconv"
	"strings"
)

func RecvChgGroupAttrReq(pconfig *Config, preq *ss.MsgChgGroupAttrReq, from_serv int) {
	var _func_ = "<RecvChgGroupAttrReq>"
	log := pconfig.Comm.Log
	uid := preq.Uid
	grp_id := preq.GrpId
	attr := preq.Attr

	log.Info("%s uid:%d grp_id:%d attr:%d", _func_, uid, grp_id, attr)
	//Sync
	go func() {
		phead := pconfig.RedisClient.AllocSyncCmdHead()
		if phead == nil {
			log.Err("%s alloc head failed! uid:%d", _func_, uid)
			return
		}
		defer pconfig.RedisClient.FreeSyncCmdHead(phead)

		//rsp
		prsp := new(ss.MsgChgGroupAttrRsp)
		prsp.Attr = attr
		prsp.GrpId = grp_id
		prsp.Uid = uid
		prsp.Occupy = preq.Occupy
		prsp.Result = ss.SS_COMMON_RESULT_FAILED

		//switch
		switch attr {
		case ss.GROUP_ATTR_TYPE_GRP_ATTR_VISIBLE:
			prsp.Result = set_group_attr_visible(pconfig, phead, preq)
		case ss.GROUP_ATTR_TYPE_GRP_ATTR_INVISIBLE:
			prsp.Result = set_group_attr_invisible(pconfig, phead, preq)
		default:
			log.Err("%s unkown attr:%d uid:%d grp_id:%d", _func_, attr, uid, grp_id)
		}

		//ss
		var ss_msg ss.SSMsg
		err := comm.FillSSPkg(&ss_msg, ss.SS_PROTO_TYPE_CHG_GROUP_ATTR_RSP, prsp)
		if err != nil {
			log.Err("%s gen ss failed! err:%v uid:%d grp_id:%d attr:%d", _func_, err, uid, grp_id, attr)
			return
		}

		//send
		SendToServ(pconfig, from_serv, &ss_msg)
	}()
}

func RecvGroupGroundReq(pconfig *Config, preq *ss.MsgGroupGroudReq, from int) {
	var _func_ = "<RecvChgGroupAttrReq>"
	log := pconfig.Comm.Log
	uid := preq.Uid

	//Sync
	go func() {
		phead := pconfig.RedisClient.AllocSyncCmdHead()
		if phead == nil {
			log.Err("%s alloc head failed! uid:%d", _func_, uid)
			return
		}
		defer pconfig.RedisClient.FreeSyncCmdHead(phead)

		//rsp
		prsp := new(ss.MsgGroupGroudRsp)
		prsp.Uid = uid

		//query
		end_index := preq.StartIndex + preq.Count - 1

		for {
			//exe
			res, err := pconfig.RedisClient.RedisExeCmdSync(phead, "ZRANGE", FORMAT_TAB_VISIBLE_GROUP_SET, preq.StartIndex,
				end_index)
			if err != nil {
				log.Err("%s zrange %d:%d failed! err:%v uid:%d", _func_, preq.StartIndex, end_index, err, uid)
				break
			}
			//convert
			strs, err := comm.Conv2Strings(res)
			if err != nil {
				log.Err("%s convert res failed! err:%v uid:%d", _func_, err, uid)
				break
			}

			if strs == nil || len(strs) == 0 {
				log.Debug("%s empty! start:%d uid:%d", _func_, preq.StartIndex, uid)
				break
			}

			//fill
			prsp.ItemList = make([]*ss.GroupGroudItem, len(strs))
			var pitem *ss.GroupGroudItem
			var idx int32 = 0
			for i := 0; i < len(strs); i++ {
				pitem, err = parse_ground_string(strs[i])
				if err != nil {
					log.Err("%s parse %s failed! err:%v uid:%d", _func_, strs[i], err, uid)
					continue
				}

				prsp.ItemList[idx] = pitem
				idx++
			}
			prsp.Count = idx
			break
		}

		//ss
		var ss_msg ss.SSMsg
		err := comm.FillSSPkg(&ss_msg, ss.SS_PROTO_TYPE_GROUP_GROUND_RSP, prsp)
		if err != nil {
			log.Err("%s gen ss failed! err:%v uid:%d", _func_, err, uid)
			return
		}

		//to logic
		SendToServ(pconfig, from, &ss_msg)
	}()
}

/*-------------------static----------------------*/
func set_group_attr_visible(pconfig *Config, phead *comm.SyncCmdHead, preq *ss.MsgChgGroupAttrReq) ss.SS_COMMON_RESULT {
	var _func_ = "<set_group_attr_visible>"
	log := pconfig.Comm.Log
	uid := preq.Uid
	grp_id := preq.GrpId

	//gen value <grp_id|grp_name>
	item := gen_ground_string(grp_id, preq.StrV)

	//zadd
	_, err := pconfig.RedisClient.RedisExeCmdSync(phead, "ZADD", FORMAT_TAB_VISIBLE_GROUP_SET, preq.IntV, item)
	if err != nil {
		log.Err("%s zadd failed! err:%v item:%s uid:%d", _func_, err, item, uid)
		return ss.SS_COMMON_RESULT_FAILED
	}

	return ss.SS_COMMON_RESULT_SUCCESS
}

func set_group_attr_invisible(pconfig *Config, phead *comm.SyncCmdHead, preq *ss.MsgChgGroupAttrReq) ss.SS_COMMON_RESULT {
	var _func_ = "<set_group_attr_invisible>"
	log := pconfig.Comm.Log
	uid := preq.Uid
	grp_id := preq.GrpId

	//gen value <grp_id|grp_name>
	item := fmt.Sprintf("%d|%s", grp_id, preq.StrV)

	//zrem
	_, err := pconfig.RedisClient.RedisExeCmdSync(phead, "ZREM", FORMAT_TAB_VISIBLE_GROUP_SET, item)
	if err != nil {
		log.Err("%s zrem failed! err:%v item:%s uid:%d", _func_, err, item, uid)
		return ss.SS_COMMON_RESULT_FAILED
	}

	return ss.SS_COMMON_RESULT_SUCCESS
}

//<grp_id|grp_name>
func gen_ground_string(grp_id int64, grp_name string) string {
	return fmt.Sprintf("%d|%s", grp_id, grp_name)
}

func parse_ground_string(str string) (*ss.GroupGroudItem, error) {
	var err error

	//splice
	strs := strings.Split(str, "|")
	if len(strs) < 2 {
		return nil, errors.New("length illegal")
	}

	//parse
	pitem := new(ss.GroupGroudItem)
	pitem.GrpId, err = strconv.ParseInt(strs[0], 10, 64)
	if err != nil {
		return nil, errors.New("convert grp_id failed")
	}
	pitem.GrpName = strs[1]
	return pitem, nil
}
