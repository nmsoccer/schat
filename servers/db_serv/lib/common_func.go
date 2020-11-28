package lib

import (
	"crypto/des"
	"fmt"
	lnet "schat/lib/net"
	"schat/proto/ss"
	"schat/servers/comm"
)

//Get Group Name
//@return(res , SS_COMMON_RESULT)
func GetGroupInfo(pclient *comm.RedisClient , phead *comm.SyncCmdHead, grp_id int64, field string) (interface{}, ss.SS_COMMON_RESULT) {
	var _func_ = "<GetGroupName>"
	log := pclient.GetLog()

	//sync query
	tab_name := fmt.Sprintf(FORMAT_TAB_GROUP_INFO_PREFIX+"%d", grp_id)
	res, err := pclient.RedisExeCmdSync(phead, "HGET", tab_name, field)
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
func RemGroupMember(pclient *comm.RedisClient, phead *comm.SyncCmdHead, uid int64, grp_id int64) ss.SS_COMMON_RESULT {
	var _func_ = "<RemGroupMember>"
	log := pclient.GetLog()

	//exit group
	tab_name := fmt.Sprintf(FORMAT_TAB_GROUP_MEMBERS+"%d", grp_id)
	_, err := pclient.RedisExeCmdSync(phead, "SREM", tab_name, uid)
	if err != nil {
		log.Err("%s remove member failed! err:%v uid:%d grp_id:%d", _func_, err, grp_id, uid)
		return ss.SS_COMMON_RESULT_FAILED
	}

	log.Debug("%s tab:%s srem done! uid:%d grp_id:%d", _func_, tab_name, uid, grp_id)
	return ss.SS_COMMON_RESULT_SUCCESS
}

//Append offline_info
//@return: list_len , result
func AppendOfflineInfo(pclient *comm.RedisClient, phead *comm.SyncCmdHead, uid int64, info string) (int, ss.SS_COMMON_RESULT) {
	var _func_ = "<AppendOfflineInfo>"
	log := pclient.GetLog()

	//handle
	tab_name := fmt.Sprintf(FORMAT_TAB_OFFLINE_INFO_PREFIX+"%d", uid)
	res, err := pclient.RedisExeCmdSync(phead, "RPUSH", tab_name, info)
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
func SaveUserProfile(pclient *comm.RedisClient, phead *comm.SyncCmdHead, uid int64, profile string) ss.SS_COMMON_RESULT {
	var _func_ = "<SaveUserProfile>"
	log := pclient.GetLog()

	//save profile
	tab_name := fmt.Sprintf(FORMAT_TAB_USER_PROFILE_PREFIX+"%d", uid)
	_, err := pclient.RedisExeCmdSync(phead, "SET", tab_name, profile)
	if err != nil {
		log.Err("%s set failed! err:%v uid:%d profile:%s", _func_, err, uid, profile)
		return ss.SS_COMMON_RESULT_FAILED
	}

	//log.Debug("%s tab:%s set done! uid:%d profile:%s" , _func_ , tab_name , uid , profile)
	return ss.SS_COMMON_RESULT_SUCCESS
}

//save group profile
func SaveGroupProfile(pclient *comm.RedisClient, phead *comm.SyncCmdHead, grp_id int64, profile string) ss.SS_COMMON_RESULT {
	var _func_ = "<SaveGroupProfile>"
	log := pclient.GetLog()

	//save profile
	tab_name := fmt.Sprintf(FORMAT_TAB_GROUP_PROFILE_PREFIX+"%d", grp_id)
	_, err := pclient.RedisExeCmdSync(phead, "SET", tab_name, profile)
	if err != nil {
		log.Err("%s set failed! err:%v grp_id:%d profile:%s", _func_, err, grp_id, profile)
		return ss.SS_COMMON_RESULT_FAILED
	}

	//log.Debug("%s tab:%s set done! uid:%d profile:%s" , _func_ , tab_name , uid , profile)
	return ss.SS_COMMON_RESULT_SUCCESS
}

//del group profile
func DelGroupProfile(pclient *comm.RedisClient, phead *comm.SyncCmdHead, grp_id int64) ss.SS_COMMON_RESULT {
	var _func_ = "<SaveGroupProfile>"
	log := pclient.GetLog()

	//save profile
	tab_name := fmt.Sprintf(FORMAT_TAB_GROUP_PROFILE_PREFIX+"%d", grp_id)
	_, err := pclient.RedisExeCmdSync(phead, "DEL", tab_name)
	if err != nil {
		log.Err("%s set failed! err:%v grp_id:%d", _func_, err, grp_id)
		return ss.SS_COMMON_RESULT_FAILED
	}

	log.Debug("%s tab:%s del done!" , _func_ , tab_name)
	return ss.SS_COMMON_RESULT_SUCCESS
}

//del visible
func InvisibleGroup(pclient *comm.RedisClient, phead *comm.SyncCmdHead, uid int64 , grp_id int64) ss.SS_COMMON_RESULT {
	var _func_ = "<InvisibleGroup>"
	log := pclient.GetLog()

	//gen value <grp_id>
	item := fmt.Sprintf("%d", grp_id)

	//zrem
	_, err := pclient.RedisExeCmdSync(phead, "ZREM", FORMAT_TAB_VISIBLE_GROUP_SET, item)
	if err != nil {
		log.Err("%s zrem failed! err:%v item:%s uid:%d grp_id:%d", _func_, err, item, uid , grp_id)
		return ss.SS_COMMON_RESULT_FAILED
	}

	log.Info("%s uid:%d grp_id:%d" , _func_ , uid , grp_id)
	return ss.SS_COMMON_RESULT_SUCCESS
}

//Decrypt and Unpack Chat
func UnpackChat(pconfig *Config , uid int64 , db_chat string) *ss.ChatMsg {
	var _func_ = "<DecyUnpackChat>"
	log := pconfig.Comm.Log

	//Decrypt
	enc_block, err := des.NewCipher([]byte(CHAT_MSG_DES_KEY))
	if err != nil {
		log.Err("%s new des cipher for key:%v failed! err:%v", _func_, CHAT_MSG_DES_KEY, err)
		return nil
	}

	//decrypt
	out_data, err := lnet.DesDecrypt(enc_block, []byte(db_chat), []byte(CHAT_MSG_DES_KEY))
	if err != nil {
		log.Err("%s  decrypt chat_msg failed! err:%v uid:%d", _func_, err, uid)
		return nil
	}

	//unpack
	pmsg := new(ss.ChatMsg)
	err = ss.UnPack(out_data, pmsg)
	if err != nil {
		log.Err("%s decode chat_msg failed! err:%v uid:%d", _func_, err, uid)
		return nil
	}

	return pmsg
}

//Decrypt and Unpack Chat
func PackChat(pconfig *Config , uid int64 , pchat *ss.ChatMsg) []byte {
	var _func_ = "<PackChat>"
	log := pconfig.Comm.Log

	//arg
	if pchat == nil {
		log.Err("%s chat nil!" , _func_)
		return nil
	}

	//Pack
	coded, err := ss.Pack(pchat)
	if err != nil {
		log.Err("%s encode chat msg failed! uid:%d err:%v", _func_, uid, err)
		return nil
	}

	//Encrypt
	coded, err = lnet.DesEncrypt(nil, coded, []byte(CHAT_MSG_DES_KEY))
	if err != nil {
		log.Err("%s encrypt chat msg failed! uid:%d  err:%v", _func_, uid, err)
		return nil
	}

	return coded
}