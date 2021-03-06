package lib

import (
	"fmt"
	"schat/proto/ss"
	"schat/servers/comm"
)

func RecvFileAddrNotify(pconfig *Config, pnotify *ss.MsgCommonNotify, src_serv int) {
	var _func_ = "<RecvFileAddrNotify>"
	log := pconfig.Comm.Log
	uid := pnotify.Uid

	for {
		//check config
		if pconfig.ServerInfo == nil || len(pconfig.ServerInfo.FileServMap) == 0 {
			log.Err("%s file addr empty! uid:%d", _func_, uid)
			return
		}

		//fill info
		pall := pconfig.ServerInfo
		addr_count := len(pall.FileServMap)
		idx := 0
		//create
		pnotify.Strs = make([]string, addr_count)
		var pinfo *FileServInfo
		var user_token = ""
		for _, pinfo = range pall.FileServMap {
			user_token = comm.CalcUserToken(uid , pinfo.Token)
			pnotify.Strs[idx] = fmt.Sprintf("%d|%s|%s", pinfo.ServIndex, user_token, pinfo.ServAddr)
			log.Debug("%s calc user_token:%s uid:%d server_token:%s" , _func_ , user_token , uid , pinfo.Token)
			idx++
		}

		if idx == 0 {
			log.Err("%s no valid file_addr found! please check! uid:%d", _func_, uid)
			return
		}

		pnotify.IntV = int64(idx)
		break
	}

	//ss
	log.Debug("%s fetch:%d items! uid:%d strs:%v src_serv:%d", _func_, pnotify.IntV, uid, pnotify.Strs, src_serv)
	pss_msg, err := comm.GenDispMsg(ss.DISP_MSG_TARGET_LOGIC_SERVER, ss.DISP_MSG_METHOD_SPEC, ss.DISP_PROTO_TYPE_DISP_COMMON_NOTIFY,
		src_serv, pconfig.ProcId, 0, pnotify)
	if err != nil {
		log.Err("%s gen ss failed! err:%v uid:%d", _func_, err, uid)
		return
	}

	//to logic
	SendToDisp(pconfig, 0, pss_msg)
}

func RecvFileTokenNotify(pconfig *Config, pnotify *ss.MsgCommonNotify, file_serv int) {
	var _func_ = "<RecvFileTokenNotify>"
	log := pconfig.Comm.Log
	serv_indx := int(pnotify.IntV)
	token := pnotify.StrV

	//check
	pall := pconfig.ServerInfo
	if pall.FileServMap == nil || len(pall.FileServMap) == 0 {
		log.Err("%s file serv map empty! serv_index:%d", _func_, serv_indx)
		return
	}

	//exist
	pinfo, ok := pall.FileServMap[serv_indx]
	if !ok {
		log.Err("%s file serv not exist! serv_index:%d", _func_, serv_indx)
		return
	}

	//update
	old_token := pinfo.Token
	pinfo.ServProc = file_serv
	pinfo.Token = token
	//log.Debug("%s update file token! serv_index:%d token:%s proc:%d " , _func_ , serv_indx , token , file_serv)

	//to logic serv
	if old_token == token {
		return
	}
	log.Info("%s token changed %s --> %s serv_index:%d", _func_, old_token, token, serv_indx)
}

//query file token when started or reload
func QueryFileServToken(arg interface{}) {
	pconfig, ok := arg.(*Config)
	if !ok {
		return
	}
	var _func_ = "<QueryFileServToken>"
	log := pconfig.Comm.Log

	//check info
	if pconfig.ServerInfo.FileServMap == nil || len(pconfig.ServerInfo.FileServMap) == 0 {
		log.Err("%s empty file serv!", _func_)
		return
	}

	//query file serv
	pnotify := new(ss.MsgCommonNotify)
	pnotify.NotifyType = ss.COMMON_NOTIFY_TYPE_NOTIFY_FILE_TOKEN
	for serv_index, pinfo := range pconfig.ServerInfo.FileServMap {
		if pinfo.ServProc > 0 && len(pinfo.Token) > 0 {
			continue
		}
		pnotify.IntV = int64(serv_index)
		pss_msg, err := comm.GenDispMsg(ss.DISP_MSG_TARGET_FILE_SERVER, ss.DISP_MSG_METHOD_ALL, ss.DISP_PROTO_TYPE_DISP_COMMON_NOTIFY,
			0, pconfig.ProcId, 0, pnotify)
		if err != nil {
			log.Err("%s gen ss failed! err:%v serv_index:%d", _func_, err, serv_index)
			continue
		}

		//Send
		SendToDisp(pconfig, 0, pss_msg)
		log.Info("%s serv_index:%d", _func_, serv_index)
	}

}

func RecvUploadFileNotify(pconfig *Config, pnotify *ss.MsgCommonNotify) {
	var _func_ = "<RecvUploadFileNotify>"
	log := pconfig.Comm.Log
	uid := pnotify.Uid
	grp_id := pnotify.GrpId
	url := pnotify.StrV
	file_index := int(pnotify.Occupy)

	log.Debug("%s. uid:%d url:%s file_index:%d grp_id:%d", _func_, uid, url, file_index , grp_id)
	//Get Info
	pinfo := GetFileServInfo(pconfig, file_index)
	if pinfo == nil || pinfo.ServProc <= 0 {
		log.Err("%s get file info failed! index:%d url:%s uid:%d", _func_, file_index, url, uid)
		return
	}

	//to file server
	pss_msg, err := comm.GenDispMsg(ss.DISP_MSG_TARGET_FILE_SERVER, ss.DISP_MSG_METHOD_SPEC, ss.DISP_PROTO_TYPE_DISP_COMMON_NOTIFY,
		pinfo.ServProc, pconfig.ProcId, 0, pnotify)
	if err != nil {
		log.Err("%s gen ss failed! err:%v url:%s uid:%d", _func_, err, url, uid)
		return
	}

	//send
	SendToDisp(pconfig, 0, pss_msg)
}

