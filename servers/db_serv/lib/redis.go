package lib

import (
	"schat/servers/comm"
)

/*
* ##tables desc##
* #global:uid +string+ ; global uid allocator
* #users:global:[name]  +hash+ name | pass | uid | salt  ; short info by user_name
* #user:[uid] +hash+  uid | name | age | sex  | addr | level | blob_info | head_url ; detail info of user by uid
* #user:login_lock:[uid] <string> valid_second
* #user:profile:[uid] +string+ <user_basic>
* #global:grp_id +string+ ; global group id allocator
* #group:[grp_id] +hash+ gid | name | master_uid | pass | salt | create_ts | msg_count | load_serv | blob_info
* #group:mem:[grp_id] +set+ <uid>
* #group:apply:[grp_id] +list+ <apply_uid|apply_name|apply_msg>
* #user:group:applied:[uid] +set+ <grp_id|grp_name>
* #user:group:audited:[uid] +list+ <grp_id|grp_name|result>
* #chat_msg:[group]:[index] +list+ <chat_msg encoded>
* #offline_info:[uid] +list+ <off_type|xxx...> //off_type:REFER SS_OFFLINE_INFO_TYPE_xx
* #visible_group_set +zset+  <grp_id|grp_name>
 */

/*--------------------INSTRUCTION--------------------
      *********GROUP TABLE*********
               group:grp_id
  group:mem:grp_id group:apply:grp_id chat_msg:group:index
*/

const (
	PASSWD_SALT_LEN    = 32
	LOGIN_LOCK_LIFE    = 20     //login lock life (second)
	CHAT_MSG_LIST_SIZE = 100000 //single tab of chat-msg size
	CHAT_MSG_DES_KEY   = "MikmiYua"

	FORMAT_TAB_GLOBAL_UID             = "global:uid"       // ++ string ++
	FORMAT_TAB_USER_GLOBAL            = "users:global:%s"  //users:global:[name]  ++ hash ++ name | pass | uid | salt
	FORMAT_TAB_USER_INFO_REFIX        = "user:"            // user:[uid] ++ hash ++ uid | name | age | sex  | addr | level | online_logic | blob_info | head_url
	FORMAT_TAB_USER_LOGIN_LOCK_PREFIX = "user:login_lock:" //user:login:[uid] +string+ valid_second
	FORMAT_TAB_USER_PREOFILE_PREFIX   = "user:profile:"    //user:profile:[uid] +string+ <user_basic>
	FORMAT_TAB_GLOBAL_GRPID           = "global:grp_id"    // +string+
	FORMAT_TAB_GROUP_INFO_PREFIX      = "group:"           // group:[grp_id] +hash+ gid | name | master_uid | pass | salt | create_ts | msg_count | load_serv
	// | blob_info
	FORMAT_TAB_GROUP_MEMBERS       = "group:mem:"          //group:mem:[grp_id] +set+ <uid>
	FORMAT_TAB_GROUP_APPLY_LIST    = "group:apply:"        // group:apply:[grp_id] +list+ <apply_uid|apply_name|apply_msg>
	FORMAT_TAB_USER_GROUP_APPLIED  = "user:group:applied:" //user:group:applied:[uid] +set+ <grp_id|grp_name>
	FORMAT_TAB_USER_GROUP_AUDITED  = "user:group:audited:" //user:group:audited:[uid] +list+ <grp_id|grp_name|result>
	FORMAT_TAB_CHAT_MSG_LIST       = "chat_msg:%d:%d"      //chat_msg:[group]:[index] +list+ <chat_msg encoded>
	FORMAT_TAB_OFFLINE_INFO_PREFIX = "offline_info:"       // offline_info:[uid] +list+ <off_type|xxx...> off_type:REFER SS_OFFLINE_INFO_TYPE_xx

	FORMAT_TAB_VISIBLE_GROUP_SET = "visible_group" //visible_group_set +zset+  <grp_id|grp_name>

	//Useful FIELD
	FIELD_USER_INFO_ONLINE_LOGIC = "online_logic"
	FIELD_GROUP_INFO_MSG_COUNT   = "msg_count"
	FILED_GROUP_INFO_NAME        = "name"
	FIELD_GROUP_BLOB_NAME        = "blob_info"
	FILED_USER_INFO_HEAD_URL     = "head_url"
)

func OpenRedis(pconfig *Config) *comm.RedisClient {
	var _func_ = "<OpenRedisAddr>"

	log := pconfig.Comm.Log
	pclient := comm.NewRedisClient(pconfig.Comm, pconfig.FileConfig.RedisAddr, pconfig.FileConfig.AuthPass,
		pconfig.FileConfig.MaxConn, pconfig.FileConfig.NormalConn)

	if pclient == nil {
		log.Err("%s fail!", _func_)
		return nil
	}

	return pclient
}

//init db info when first started
func InitRedisDb(arg interface{}) {
	var _func_ = "<InitRedisDb>"
	pconfig, ok := arg.(*Config)
	if !ok {
		return
	}
	log := pconfig.Comm.Log
	if pconfig.RedisClient == nil {
		log.Info("%s redis not open or client not inited!", _func_)
		return
	}

	log.Info("%s starts...", _func_)
	//init global uid
	pconfig.RedisClient.RedisExeCmd(pconfig.Comm, cb_init_global_uid, nil, "SETNX",
		FORMAT_TAB_GLOBAL_UID, pconfig.FileConfig.InitUid)
	//init global group id
	pconfig.RedisClient.RedisExeCmd(pconfig.Comm, cb_init_global_grpid, nil, "SETNX",
		FORMAT_TAB_GLOBAL_GRPID, pconfig.FileConfig.InitGrpId)

	return
}

//ResetRedis
func ResetRedis(pconfig *Config, old_config *FileConfig, new_config *FileConfig) {
	var _func_ = "<ResetRedis>"
	log := pconfig.Comm.Log

	var new_addr string = ""
	var new_auth string = ""
	var new_max int = 0
	var new_normal int = 0
	var reset = false

	//check should reset
	if old_config.RedisAddr != new_config.RedisAddr {
		new_addr = new_config.RedisAddr
		reset = true
	}

	if old_config.AuthPass != new_config.AuthPass {
		new_auth = new_config.AuthPass
		reset = true
	}

	if old_config.MaxConn != new_config.MaxConn {
		new_max = new_config.MaxConn
		reset = true
	}

	if old_config.NormalConn != new_config.NormalConn {
		new_normal = new_config.NormalConn
		reset = true
	}

	if reset {
		log.Info("%s will reset redis attr!", _func_)
		pconfig.RedisClient.Reset(new_addr, new_auth, new_max, new_normal)
		return
	}

	log.Info("%s nothing to do", _func_)
}

/*----------------static func--------------------*/
func cb_init_global_uid(pcomm_config *comm.CommConfig, result interface{}, cb_arg []interface{}) {
	var _func_ = "<cb_init_global_uid>"
	log := pcomm_config.Log

	//check error
	if err, ok := result.(error); ok {
		log.Err("%s failed! err:%v", _func_, err)
		return
	}

	//print
	log.Info("%s result:%v", _func_, result)
	return
}

func cb_init_global_grpid(pcomm_config *comm.CommConfig, result interface{}, cb_arg []interface{}) {
	var _func_ = "<cb_init_global_grpid>"
	log := pcomm_config.Log

	//check error
	if err, ok := result.(error); ok {
		log.Err("%s failed! err:%v", _func_, err)
		return
	}

	//print
	log.Info("%s result:%v", _func_, result)
	return
}
