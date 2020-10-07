package lib

import (
	"fmt"
	"schat/proto/ss"
	"schat/servers/comm"
)

func RecvFileAddrNotify(pconfig *Config , pnotify *ss.MsgCommonNotify , src_serv int) {
	var _func_ = "<RecvFileAddrNotify>"
	log := pconfig.Comm.Log
	uid := pnotify.Uid

	for {
		//check config
		if pconfig.ServerInfo == nil || len(pconfig.ServerInfo.FileServList) == 0 {
			log.Err("%s file addr empty! uid:%d", _func_, uid)
			return
		}

		//fill info
		pall := pconfig.ServerInfo
		addr_count := len(pall.FileServList)
		idx := 0
		//create
		pnotify.Strs = make([]string , addr_count)
		var pinfo *FileServInfo
		for i:=0; i<addr_count; i++ {
			pinfo = pall.FileServList[i]
			if pinfo == nil {
				log.Err("%s file_serv_list<%d> nil pointer! please check!" , _func_ , i)
				continue
			}

			pnotify.Strs[idx] = fmt.Sprintf("%d:%s" , pinfo.ServIndex , pinfo.ServAddr)
			idx++
		}

		if idx == 0 {
			log.Err("%s no valid file_addr found! please check! uid:%d" , _func_ , uid)
			return
		}

		pnotify.IntV = int64(idx)
		break
	}

	//ss
	log.Debug("%s fetch:%d items! uid:%d strs:%v src_serv:%d" , _func_ , pnotify.IntV , uid , pnotify.Strs , src_serv)
	pss_msg , err := comm.GenDispMsg(ss.DISP_MSG_TARGET_LOGIC_SERVER , ss.DISP_MSG_METHOD_SPEC , ss.DISP_PROTO_TYPE_DISP_COMMON_NOTIFY ,
		src_serv , pconfig.ProcId , 0 , pnotify)
    if err != nil {
    	log.Err("%s gen ss failed! err:%v uid:%d" , _func_ , err , uid)
    	return
	}

	//to logic
	SendToDisp(pconfig , 0 , pss_msg)
}
