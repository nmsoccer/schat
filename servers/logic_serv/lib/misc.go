package lib

import (
	"schat/proto/ss"
	"schat/servers/comm"
	"strconv"
	"strings"
	"time"
)

func ParseOfflineInfo(pconfig *Config, uid int64, info string) {
	var _func_ = "<ParseOfflineInfo>"
	log := pconfig.Comm.Log

	var grp_id int64
	var ts int64

	//splice
	strs := strings.Split(info, "|")
	if len(strs) == 0 {
		log.Err("%s failed! info:%s empty! uid:%d", _func_, info, uid)
		return
	}

	//convert
	int_v, err := strconv.Atoi(strs[0])
	if err != nil {
		log.Err("%s failed! convert info_type failed! err:%v uid:%d info:%s", _func_, err, uid, info)
		return
	}

	//switch
	switch ss.SS_OFFLINE_INFO_TYPE(int_v) {
	case ss.SS_OFFLINE_INFO_TYPE_OFT_KICK_GROUP: //<type|grp_id|grp_name|kick_ts>
		if len(strs) != 4 {
			log.Err("%s failed! kick_group info len not match! err:%v uid:%d info:%s", _func_, err, uid, info)
			break
		}
		grp_id, err = strconv.ParseInt(strs[1], 10, 64)
		if err != nil {
			log.Err("%s convert grp_id failed! err:%v uid:%d info:%s", _func_, err, uid, info)
			break
		}
		ts, err = strconv.ParseInt(strs[3], 10, 64)
		if err != nil {
			log.Err("%s convert kick_ts failed! err:%v uid:%d info:%s", _func_, err, uid, info)
			break
		}
		//fake notify
		pnotify := new(ss.MsgCommonNotify)
		pnotify.Uid = uid
		pnotify.GrpId = grp_id
		pnotify.StrV = strs[2]
		pnotify.IntV = ts
		RecvKickGroupNotify(pconfig, pnotify)
	default:
		log.Err("%s unknown info_type:%d uid:%d info:%s", _func_, int_v, uid, info)
	}

}

func RecvUploadFileNotify(pconfig *Config, pnotify *ss.MsgCommonNotify, file_server int) {
	var _func_ = "<RecvUploadFileNotify>"
	log := pconfig.Comm.Log
	uid := pnotify.Uid
	url := pnotify.StrV
	file_type := pnotify.Occupy

	//url_type
	url_type, err := comm.GetUrlType(url)
	if err != nil {
		log.Err("%s parse url failed! url:%s uid:%d file_type:%d", _func_, url, uid , file_type)
		return
	}

	//switch
	switch url_type {
	case comm.FILE_URL_T_CHAT:
		UploadChatFileNotify(pconfig, pnotify, file_server)
	case comm.FILE_URL_T_HEAD:
		UploadHeadFileNotify(pconfig, pnotify, file_server)
	case comm.FILE_URL_T_GROUP:
		UploadGroupHeadFileNotify(pconfig , pnotify , file_server)
	default:
		log.Err("%s illegal url_type:%d uid:%d url:%s", _func_, url_type, uid, url)
	}

}

func UploadHeadFileNotify(pconfig *Config, pnotify *ss.MsgCommonNotify, file_server int) {
	var _func_ = "<UploadHeadFileNotify>"
	var puser_info *UserOnLine
	var err error
	var old_url string
	log := pconfig.Comm.Log
	uid := pnotify.Uid
	url := pnotify.StrV
	check_err := 0
	file_index := 0

	log.Debug("%s. uid:%d url:%s", _func_, uid, url)
	for {
		//check url
		file_index, err = comm.GetUrlIndex(url)
		if err != nil {
			log.Err("%s get url index failed! err:%v uid:%d url:%s file_index:%d", _func_, err, uid, url , file_index)
			return
		}

		//check online
		puser_info = GetUserInfo(pconfig, uid)
		if puser_info == nil {
			log.Err("%s not online!! uid:%d", _func_, uid)
			check_err = comm.FILE_UPT_CHECK_ONLINE
			break
		}

		//url
		old_url = puser_info.user_info.BasicInfo.HeadUrl
		if old_url == url {
			log.Info("%s head url not change! url:%s uid:%d", _func_, url, uid)
			return
		}

		//update url
		puser_info.user_info.BasicInfo.HeadUrl = url
		log.Info("%s update head url %s-->%s uid:%d", _func_, old_url, url, uid)
		if len(old_url) > 0 {
			check_err = comm.FILE_UPT_CHECK_DEL
		}
		break
	}

	//save
	if check_err==0 || check_err==comm.FILE_UPT_CHECK_DEL {
		//save profile
		SaveUserProfile(pconfig, uid)

		//to client
		pnotify2 := new(ss.MsgCommonNotify)
		pnotify2.NotifyType = ss.COMMON_NOTIFY_TYPE_NOTIFY_HEAD_URL
		pnotify2.GrpId = 0
		pnotify2.Uid = uid
		pnotify2.StrV = url

		//pack
		var ss_msg ss.SSMsg
		err := comm.FillSSPkg(&ss_msg , ss.SS_PROTO_TYPE_COMMON_NOTIFY , pnotify2)
		if err != nil {
			log.Err("%s gen ss_msg to client failed! uid:%d" , _func_ , uid)
			return
		}
		SendToConnect(pconfig , &ss_msg)

		if check_err == 0 {
			return
		}
	}


	//back to file serv
	pss_msg := new(ss.SSMsg)
	pnotify.IntV = int64(check_err)
	switch check_err {
	case comm.FILE_UPT_CHECK_ONLINE:
		pss_msg, err = comm.GenDispMsg(ss.DISP_MSG_TARGET_FILE_SERVER, ss.DISP_MSG_METHOD_SPEC, ss.DISP_PROTO_TYPE_DISP_COMMON_NOTIFY,
			file_server, pconfig.ProcId, 0, pnotify)
		if err != nil {
			log.Err("%s gen file ss msg failed! err:%v uid:%d  url:%s", _func_, err, uid, url)
			return
		}
	case comm.FILE_UPT_CHECK_DEL:
		SendDelOldFile(pconfig , old_url , uid , 0)
		return
		/*
		pnotify.StrV = old_url //del old head file
		//old_url index
		old_file_index, _ := comm.GetUrlIndex(old_url)
		//new and old url from same file server
		if old_file_index == file_index {
			log.Debug("%s new and old url from same file_index:%d uid:%d", _func_, file_index, uid)
			pss_msg, err = comm.GenDispMsg(ss.DISP_MSG_TARGET_FILE_SERVER, ss.DISP_MSG_METHOD_SPEC, ss.DISP_PROTO_TYPE_DISP_COMMON_NOTIFY,
				file_server, pconfig.ProcId, 0, pnotify)
			if err != nil {
				log.Err("%s gen file ss msg failed! err:%v uid:%d  old_url:%s", _func_, err, uid, old_url)
				return
			}
			break
		}

		log.Debug("%s new and old url  diff file_index %d:%d uid:%d to dir", _func_, old_file_index, file_index, uid)
		//old and new url from diff file server
		pnotify.Occupy = int64(old_file_index)
		pss_msg, err = comm.GenDispMsg(ss.DISP_MSG_TARGET_DIR_SERVER, ss.DISP_MSG_METHOD_RAND, ss.DISP_PROTO_TYPE_DISP_COMMON_NOTIFY,
			0, pconfig.ProcId, 0, pnotify)
		if err != nil {
			log.Err("%s gen dir ss msg failed! err:%v uid:%d  old_url:%s", _func_, err, uid, old_url)
			return
		}*/
	default:
		//nothing
		return
	}

	//send
	SendToDisp(pconfig, 0, pss_msg)
	return

}

func UploadChatFileNotify(pconfig *Config, pnotify *ss.MsgCommonNotify, file_server int) {
	var _func_ = "<UploadChatFileNotify>"
	var puser_info *UserOnLine
	log := pconfig.Comm.Log
	uid := pnotify.Uid
	grp_id := pnotify.GrpId
	url := pnotify.StrV
	file_type := pnotify.Occupy
	check_err := 0

	log.Debug("%s. uid:%d grp_id:%d url:%s tmp_id:%d file_type:%d", _func_, uid, grp_id, pnotify.StrV, pnotify.IntV , file_type)
	for {
		//check online
		puser_info = GetUserInfo(pconfig, uid)
		if puser_info == nil {
			log.Err("%s not online!! uid:%d", _func_, uid)
			check_err = comm.FILE_UPT_CHECK_ONLINE
			break
		}

		//check group
		pchat_info := puser_info.user_info.BlobInfo.ChatInfo
		if pchat_info.AllGroup <= 0 || pchat_info.AllGroups == nil {
			log.Err("%s enter no group!! uid:%d grp_id:%d", _func_, uid, grp_id)
			check_err = comm.FILE_UPT_CHECK_GROUP
			break
		}

		_, ok := pchat_info.AllGroups[grp_id]
		if !ok {
			log.Err("%s no in group!! uid:%d grp_id:%d", _func_, uid, grp_id)
			check_err = comm.FILE_UPT_CHECK_GROUP
			break
		}


		break
	}
	//check value
	if check_err != 0 {
		pnotify.IntV = int64(check_err)
		pss_msg, err := comm.GenDispMsg(ss.DISP_MSG_TARGET_FILE_SERVER, ss.DISP_MSG_METHOD_SPEC, ss.DISP_PROTO_TYPE_DISP_COMMON_NOTIFY,
			file_server, pconfig.ProcId, 0, pnotify)

		if err != nil {
			log.Err("%s gen -->file ss msg failed! err:%v uid:%d grp_id:%d url:%s", _func_, err, uid, grp_id, url)
			return
		}

		//send
		SendToDisp(pconfig, 0, pss_msg)
		return
	}

	//check passed
	//create req
	chat_type := ss.CHAT_MSG_TYPE_CHAT_TYPE_IMG //default image
	switch int(file_type) {
	case comm.FILE_TYPE_IMAGE:
		//nothing
	case comm.FILE_TYPE_MP4:
		chat_type = ss.CHAT_MSG_TYPE_CHAT_TYPE_MP4
	default:
		log.Err("%s illegal file_type:%d uid:%d" , _func_ , file_type , uid)
		return
	}
	log.Debug("%s check passed! will create chat_msg! uid:%d grp_id:%d url:%s file_type:%d chat_type:%d", _func_, uid, grp_id, url , file_type ,
		chat_type)

	preq := new(ss.MsgSendChatReq)
	preq.Uid = uid
	preq.TempId = pnotify.IntV
	pchat := new(ss.ChatMsg)
	pchat.GroupId = grp_id
	pchat.Content = url
	pchat.SenderUid = uid
	pchat.Sender = puser_info.user_info.BasicInfo.Name
	pchat.ChatType = chat_type
	preq.ChatMsg = pchat

	//ss
	pss_msg, err := comm.GenDispMsg(ss.DISP_MSG_TARGET_CHAT_SERVER, ss.DISP_MSG_METHOD_HASH, ss.DISP_PROTO_TYPE_DISP_SEND_CHAT_REQ,
		0, pconfig.ProcId, grp_id, preq)
	if err != nil {
		log.Err("%s gen send_chat ss failed! err:%v uid:%d grp_id:%d content:%s", _func_, err, uid, grp_id, url)
		return
	}

	//to chat_serv
	SendToDisp(pconfig, 0, pss_msg)
}

func UploadGroupHeadFileNotify(pconfig *Config, pnotify *ss.MsgCommonNotify, file_server int) {
	var _func_ = "<UploadGroupHeadFileNotify>"
	var puser_info *UserOnLine
	log := pconfig.Comm.Log
	uid := pnotify.Uid
	grp_id := pnotify.GrpId
	url := pnotify.StrV
	check_err := 0

	log.Debug("%s. uid:%d grp_id:%d url:%s tmp_id:%d", _func_, uid, grp_id, pnotify.StrV, pnotify.IntV)
	for {
		//check online
		puser_info = GetUserInfo(pconfig, uid)
		if puser_info == nil {
			log.Err("%s not online!! uid:%d", _func_, uid)
			check_err = comm.FILE_UPT_CHECK_ONLINE
			break
		}

		//check group
		pchat_info := puser_info.user_info.BlobInfo.ChatInfo
		if pchat_info.AllGroup <= 0 || pchat_info.AllGroups == nil {
			log.Err("%s enter no group!! uid:%d grp_id:%d", _func_, uid, grp_id)
			check_err = comm.FILE_UPT_CHECK_GROUP
			break
		}

		_, ok := pchat_info.AllGroups[grp_id]
		if !ok {
			log.Err("%s no in group!! uid:%d grp_id:%d", _func_, uid, grp_id)
			check_err = comm.FILE_UPT_CHECK_GROUP
			break
		}

		if pchat_info.MasterGroup <=0 || pchat_info.MasterGroups == nil {
			log.Err("%s owns no group!! uid:%d grp_id:%d", _func_, uid, grp_id)
			check_err = comm.FILE_UPT_CHECK_GROUP
			break
		}

		_ , ok = pchat_info.MasterGroups[grp_id]
		if !ok {
			log.Err("%s not own such group!! uid:%d grp_id:%d", _func_, uid, grp_id)
			check_err = comm.FILE_UPT_CHECK_GROUP
			break
		}

		break
	}
	//check value
	if check_err != 0 {
		pnotify.IntV = int64(check_err)
		pss_msg, err := comm.GenDispMsg(ss.DISP_MSG_TARGET_FILE_SERVER, ss.DISP_MSG_METHOD_SPEC, ss.DISP_PROTO_TYPE_DISP_COMMON_NOTIFY,
			file_server, pconfig.ProcId, 0, pnotify)

		if err != nil {
			log.Err("%s gen -->file ss msg failed! err:%v uid:%d grp_id:%d url:%s", _func_, err, uid, grp_id, url)
			return
		}

		//send
		SendToDisp(pconfig, 0, pss_msg)
		return
	}

	//check passed
	//create attr req
	log.Debug("%s check passed! will create chg attr req! uid:%d grp_id:%d url:%s", _func_, uid, grp_id, url)
	preq := new(ss.MsgChgGroupAttrReq)
	preq.Attr = ss.GROUP_ATTR_TYPE_GRP_ATTR_HEAD_URL
	preq.Uid = uid
	preq.GrpId = grp_id
	preq.StrV = url

	//ss
	pss_msg, err := comm.GenDispMsg(ss.DISP_MSG_TARGET_CHAT_SERVER, ss.DISP_MSG_METHOD_HASH, ss.DISP_PROTO_TYPE_DISP_CHG_GROUP_ATTR_REQ,
		0, pconfig.ProcId, grp_id, preq)
	if err != nil {
		log.Err("%s gen chg_attr_req ss failed! err:%v uid:%d grp_id:%d content:%s", _func_, err, uid, grp_id, url)
		return
	}

	//to chat_serv
	SendToDisp(pconfig, 0, pss_msg)
}

func SendDelOldFile(pconfig *Config , del_url string , uid int64 , grp_id int64) {
	var _func_ = "<SendDelOldFile>"
	log := pconfig.Comm.Log
	del_file_index, _ := comm.GetUrlIndex(del_url)

	log.Info("%s del_url:%s file_index:%d uid:%d grp_id:%d" , _func_ , del_url , del_file_index , uid , grp_id)
	//notify
	pnotify := new(ss.MsgCommonNotify)
	pnotify.Uid = uid
	pnotify.GrpId = grp_id
	pnotify.NotifyType = ss.COMMON_NOTIFY_TYPE_NOTIFY_UPLOAD_FILE
	pnotify.Occupy = int64(del_file_index)
	pnotify.StrV = del_url
	pnotify.IntV = comm.FILE_UPT_CHECK_DEL

	//ss_msg
	pss_msg, err := comm.GenDispMsg(ss.DISP_MSG_TARGET_DIR_SERVER, ss.DISP_MSG_METHOD_RAND, ss.DISP_PROTO_TYPE_DISP_COMMON_NOTIFY,
		0, pconfig.ProcId, 0, pnotify)
	if err != nil {
		log.Err("%s gen dir ss msg failed! err:%v uid:%d  del_url:%s", _func_, err, uid, del_url)
		return
	}

	//to dir
	SendToDisp(pconfig , 0 , pss_msg)
}


func RecvCommonQuery(pconfig *Config , preq *ss.MsgCommonQuery) {
	var _func_ = "<RecvCommonQuery>"
	log := pconfig.Comm.Log
	uid := preq.Uid

	//check online
	puser_info := GetUserInfo(pconfig , uid)
	if puser_info == nil {
		log.Err("%s offline! uid:%d type:%d" , _func_ , uid , preq.QueryType)
		return
	}

	//switch
	switch ss.SS_COMMON_QUERY_TYPE(preq.QueryType) {
	case ss.SS_COMMON_QUERY_TYPE_QRY_OWN_SNAP:
		log.Debug("%s query own group snap! uid:%d" , _func_ , uid)
		SyncUserGroupSnap(pconfig , uid)
	case ss.SS_COMMON_QUERY_TYPE_QRY_SET_HEART:
		puser_info.hearbeat = time.Now().Unix()
		log.Debug("%s set hearbeat! uid:%d heart:%d" , _func_ , uid , puser_info.hearbeat)
	default:
		log.Err("%s illegal query type:%d uid:%d" , _func_ , preq.QueryType , uid)
	}

}

func StoreNewMsgNotify(pconfig *Config , uid int64 , grp_id int64) {
	var _func_ = "<StoreNewMsgNotify>"
	log := pconfig.Comm.Log

	//check online
	puser_info := GetUserInfo(pconfig , uid)
	if puser_info == nil {
		log.Err("%s offline! uid:%d" , _func_ , uid)
		return
	}

	//set
	puser_info.new_chat_when_lost[grp_id] = true
    log.Debug("%s set grp_id:%d uid:%d" , _func_ , grp_id , uid)
}

//check new msg notify when connected
func CheckNewMsgNotify(pconfig *Config , uid int64) {
	var _func_ = "<StoreNewMsgNotify>"
	log := pconfig.Comm.Log

	//check online
	puser_info := GetUserInfo(pconfig , uid)
	if puser_info == nil {
		log.Err("%s offline! uid:%d" , _func_ , uid)
		return
	}

	if len(puser_info.new_chat_when_lost) <= 0 {
		return
	}

	//set
	for grp_id , _ := range puser_info.new_chat_when_lost {
		log.Debug("%s fetch chat from grp_id:%d uid:%d", _func_, grp_id, uid)
		SendFetchChatReq(pconfig , uid , grp_id)
	}

	puser_info.new_chat_when_lost = make(map[int64] bool) //new as clear
}
