package lib

import (
	"schat/proto/ss"
	"schat/servers/comm"
)

const (
	RECV_FILE_MSG_TICK = 10
)
var recv_msg_list []*FileMsg

func init() {
	recv_msg_list = make([]*FileMsg , RECV_FILE_MSG_TICK)
}

func ReadFileMsg(pconfig *Config) {
	var count int

    //read
    count = pconfig.FileServer.Read(recv_msg_list)
    if count <= 0 {
    	return
	}

    //handle
    for i:=0; i<count; i++ {
        HandleFileMsg(pconfig , recv_msg_list[i])
	}
}

func HandleFileMsg(pconfig *Config , pmsg *FileMsg) {
	var _func_ = "<HandleFileMsg>"
	log := pconfig.Comm.Log

	log.Debug("%s type:%d uid:%d grp_id:%d url:%s" , _func_ , pmsg.msg_type , pmsg.uid , pmsg.grp_id , pmsg.url)
	switch pmsg.msg_type {
	case FILE_MSG_UPLOAD:
		//send to online check
        ReadUploadMsg(pconfig , pmsg)
	default:
		//nothing to do
	}
}

func ReadUploadMsg(pconfig *Config , pmsg *FileMsg) {
	var _func_ = "<ReadUploadMsg>"
	log := pconfig.Comm.Log

	//pack
	pnotify := new(ss.MsgCommonNotify)
	pnotify.NotifyType = ss.COMMON_NOTIFY_TYPE_NOTIFY_UPLOAD_FILE
	pnotify.GrpId = pmsg.grp_id
	pnotify.Uid = pmsg.uid
	pnotify.StrV = pmsg.url
	pnotify.IntV = pmsg.int_v

	pss_msg , err := comm.GenDispMsg(ss.DISP_MSG_TARGET_ONLINE_SERVER , ss.DISP_MSG_METHOD_RAND , ss.DISP_PROTO_TYPE_DISP_COMMON_NOTIFY , 0 ,
		pconfig.ProcId , 0 , pnotify)
	if err != nil {
		log.Err("%s gen ss failed! err:%v uid:%d" , _func_ , err , pmsg.uid)
		return
	}

    //to online_serv
    SendToDisp(pconfig , 0 , pss_msg)
}

func RecvUploadNotify(pconfig *Config , pnotify *ss.MsgCommonNotify) {
	var _func_ = "<ReadUploadMsg>"
	log := pconfig.Comm.Log

	//handle
    switch pnotify.IntV {
	case 1: //not online
	    log.Info("%s uid:%d not online! will del url:%s grp_id:%d" , _func_ , pnotify.Uid , pnotify.StrV , pnotify.GrpId)
	case 2: //not in group
		log.Info("%s uid:%d not in group!! will del url:%s grp_id:%d" , _func_ , pnotify.Uid , pnotify.StrV , pnotify.GrpId)
	default:
		log.Err("%s unkown value! uid:%d result:%d grp_id:%d url:%s" , _func_ , pnotify.Uid , pnotify.IntV , pnotify.GrpId , pnotify.StrV)
	    return
    }

    //msg
    pmsg := new(FileMsg)
    pmsg.msg_type = FILE_MSG_UPLOAD_CHECK_FAIL
    pmsg.int_v = pnotify.IntV
    pmsg.uid = pnotify.Uid
    pmsg.grp_id = pnotify.GrpId
    pmsg.url = pnotify.StrV
    ok := pconfig.FileServer.Send(pmsg)
    if !ok {
    	log.Err("%s FileServer Send Failed for Full! uid:%d url:%s grp_id:%d" , _func_ , pmsg.uid , pmsg.url , pmsg.grp_id)
    	return
	}
}

