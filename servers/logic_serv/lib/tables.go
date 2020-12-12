package lib

import (
	"encoding/json"
	"schat/servers/comm"
	tables "schat/servers/logic_serv/table_desc"
)

//TABLE_FILE_NAME
const (
	TAB_FILE_ITEM        = "item.json"
	TAB_FILE_CHAT_CONFIG = "chat_config.json"
)

func RegistTableMap(pconfig *Config) bool {
	var _func_ = "<RegistTableMap>"
	var log = pconfig.Comm.Log

	pconfig.TableMap = make(comm.TableMap)
	tab_map := pconfig.TableMap

	/*table/xx.json <-> table_desc/xx.go*/
	//item.json
	tab_map[TAB_FILE_ITEM] = new(tables.ItemJson)
	if _, ok := tab_map[TAB_FILE_ITEM]; !ok {
		log.Err("%s failed! new %s failed!", _func_, TAB_FILE_ITEM)
		return false
	}

	//chat_config.json
	tab_map[TAB_FILE_CHAT_CONFIG] = new(tables.ChatConfigJson)
	if _, ok := tab_map[TAB_FILE_CHAT_CONFIG]; !ok {
		log.Err("%s failed! new %s failed!", _func_, TAB_FILE_CHAT_CONFIG)
		return false
	}

	return true
}

func ChatConfig2Str(pconfig *Config) string {
	var _func_ = "<ChatConfig2Str>"
	log := pconfig.Comm.Log

	//Chat Config
	tab_map := pconfig.TableMap
	pv , ok := tab_map[TAB_FILE_CHAT_CONFIG]
	if !ok {
		log.Err("%s no config:%s loaded!" , _func_ , TAB_FILE_CHAT_CONFIG)
		return ""
	}

	//convert
	pchat_cfg  , ok := pv.(*tables.ChatConfigJson)
	if !ok {
		log.Err("%s convert config:%s failed!" , _func_ , TAB_FILE_CHAT_CONFIG)
		return ""
	}

	//dump to string
	bs , err := json.Marshal(pchat_cfg)
	if err != nil {
		log.Err("%s json marshall failed! err:%v" , _func_ , err)
		return ""
	}

	return string(bs)
}
