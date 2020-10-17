package lib

import (
	"fmt"
	"schat/proto/ss"
	"schat/servers/comm"
)

//Get Group Name
//@return(res , SS_COMMON_RESULT)
func GetGroupInfo(pconfig *Config, phead *comm.SyncCmdHead, grp_id int64, field string) (interface{}, ss.SS_COMMON_RESULT) {
	var _func_ = "<GetGroupName>"
	log := pconfig.Comm.Log

	//sync query
	tab_name := fmt.Sprintf(FORMAT_TAB_GROUP_INFO_PREFIX+"%d", grp_id)
	res, err := pconfig.RedisClient.RedisExeCmdSync(phead, "HGET", tab_name, field)
	if err != nil {
		log.Err("%s query group failed! err:%v uid:%d grp_id:%d", _func_, err, grp_id)
		return nil, ss.SS_COMMON_RESULT_FAILED
	}
	if res == nil {
		log.Err("%s group not exist! grp_id:%d", _func_, grp_id)
		return nil, ss.SS_COMMON_RESULT_NOEXIST
	}

	return res, ss.SS_COMMON_RESULT_SUCCESS
}

//RM Group Member
func RemGroupMember(pconfig *Config, phead *comm.SyncCmdHead, uid int64, grp_id int64) ss.SS_COMMON_RESULT {
	var _func_ = "<RemGroupMember>"
	log := pconfig.Comm.Log

	//exit group
	tab_name := fmt.Sprintf(FORMAT_TAB_GROUP_MEMBERS+"%d", grp_id)
	_, err := pconfig.RedisClient.RedisExeCmdSync(phead, "SREM", tab_name, uid)
	if err != nil {
		log.Err("%s remove member failed! err:%v uid:%d grp_id:%d", _func_, err, grp_id, uid)
		return ss.SS_COMMON_RESULT_FAILED
	}

	log.Debug("%s tab:%s srem done! uid:%d grp_id:%d", _func_, tab_name, uid, grp_id)
	return ss.SS_COMMON_RESULT_SUCCESS
}

//Append offline_info
//@return: list_len , result
func AppendOfflineInfo(pconfig *Config, phead *comm.SyncCmdHead, uid int64, info string) (int, ss.SS_COMMON_RESULT) {
	var _func_ = "<AppendOfflineInfo>"
	log := pconfig.Comm.Log

	//handle
	tab_name := fmt.Sprintf(FORMAT_TAB_OFFLINE_INFO_PREFIX+"%d", uid)
	res, err := pconfig.RedisClient.RedisExeCmdSync(phead, "RPUSH", tab_name, info)
	if err != nil {
		log.Err("%s rpush %s failed! uid:%d info:%s err:%v", _func_, tab_name, uid, info, err)
		return 0, ss.SS_COMMON_RESULT_FAILED
	}

	int_v, err := comm.Conv2Int(res)
	if err != nil {
		log.Err("%s convert res:%v failed! err:%v uid:%d info:%s", _func_, res, err, uid, info)
		return 0, ss.SS_COMMON_RESULT_FAILED
	}

	return int_v, ss.SS_COMMON_RESULT_SUCCESS
}

//save user profile
func SaveUserProfile(pconfig *Config, phead *comm.SyncCmdHead, uid int64, profile string) ss.SS_COMMON_RESULT {
	var _func_ = "<SaveUserProfile>"
	log := pconfig.Comm.Log

	//save profile
	tab_name := fmt.Sprintf(FORMAT_TAB_USER_PREOFILE_PREFIX+"%d", uid)
	_, err := pconfig.RedisClient.RedisExeCmdSync(phead, "SET", tab_name, profile)
	if err != nil {
		log.Err("%s set failed! err:%v uid:%d profile:%s", _func_, err, uid, profile)
		return ss.SS_COMMON_RESULT_FAILED
	}

	//log.Debug("%s tab:%s set done! uid:%d profile:%s" , _func_ , tab_name , uid , profile)
	return ss.SS_COMMON_RESULT_SUCCESS
}
