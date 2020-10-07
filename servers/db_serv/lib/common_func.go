package lib

import (
	"fmt"
	"schat/proto/ss"
	"schat/servers/comm"
)

//Get Group Name
//@return(res , SS_COMMON_RESULT)
func GetGroupInfo(pconfig *Config , phead *comm.SyncCmdHead , grp_id int64 , field string) (interface{} , ss.SS_COMMON_RESULT){
	var _func_ = "<GetGroupName>"
	log := pconfig.Comm.Log

	//sync query
	tab_name := fmt.Sprintf(FORMAT_TAB_GROUP_INFO_PREFIX + "%d" , grp_id)
	res , err := pconfig.RedisClient.RedisExeCmdSync(phead , "HGET" , tab_name , field)
	if err != nil {
		log.Err("%s query group failed! err:%v uid:%d grp_id:%d" , _func_ , err ,grp_id)
		return nil , ss.SS_COMMON_RESULT_FAILED
	}
	if res == nil {
		log.Err("%s group not exist! grp_id:%d" , _func_ , grp_id)
		return nil , ss.SS_COMMON_RESULT_NOEXIST
	}


	return res , ss.SS_COMMON_RESULT_SUCCESS
}
