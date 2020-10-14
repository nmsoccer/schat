package lib

import (
	"schat/proto/ss"
	"schat/servers/comm"
	"strconv"
	"strings"
)

func ParseOfflineInfo(pconfig *Config , uid int64 , info string) {
	var _func_ = "<ParseOfflineInfo>"
	log := pconfig.Comm.Log

	var grp_id int64
	var ts     int64

	//splice
	strs := strings.Split(info , "|")
	if len(strs) == 0 {
		log.Err("%s failed! info:%s empty! uid:%d" , _func_ , info , uid)
		return
	}

	//convert
	int_v , err := strconv.Atoi(strs[0])
	if err != nil {
		log.Err("%s failed! convert info_type failed! err:%v uid:%d info:%s" , _func_ , err , uid , info)
		return
	}

	//switch
	switch ss.SS_OFFLINE_INFO_TYPE(int_v) {
	case ss.SS_OFFLINE_INFO_TYPE_OFT_KICK_GROUP: //<type|grp_id|grp_name|kick_ts>
	    if len(strs) != 4 {
			log.Err("%s failed! kick_group info len not match! err:%v uid:%d info:%s" , _func_ , err , uid , info)
			break
		}
		grp_id , err = strconv.ParseInt(strs[1] , 10 , 64)
		if err != nil {
			log.Err("%s convert grp_id failed! err:%v uid:%d info:%s" , _func_ , err , uid , info)
			break
		}
		ts , err = strconv.ParseInt(strs[3] , 10 , 64)
		if err != nil {
			log.Err("%s convert kick_ts failed! err:%v uid:%d info:%s" , _func_ , err , uid , info)
			break
		}
		//fake notify
		pnotify := new(ss.MsgCommonNotify)
		pnotify.Uid = uid
		pnotify.GrpId = grp_id
		pnotify.StrV = strs[2]
		pnotify.IntV = ts
		RecvKickGroupNotify(pconfig , pnotify)
	default:
		log.Err("%s unknown info_type:%d uid:%d info:%s" , _func_ , int_v , uid , info)
	}


}

func RecvUploadFileNotify(pconfig *Config , pnotify *ss.MsgCommonNotify , file_server int) {
	var _func_ = "<RecvUploadFileNotify>"
	log := pconfig.Comm.Log
	uid := pnotify.Uid	
	url := pnotify.StrV
	
	//url_type
	url_type , err := comm.GetUrlType(url)
	if err != nil {
		log.Err("%s parse url failed! url:%s uid:%d" , _func_ , url , uid)
		return
	}
	
	//switch
	switch url_type {
	case comm.FILE_URL_T_CHAT:
		UploadChatFileNotify(pconfig , pnotify , file_server)
	case comm.FILE_URL_T_HEAD:
		UploadHeadFileNotify(pconfig , pnotify , file_server)
	default:
		log.Err("%s illegal url_type:%d uid:%d url:%s" , _func_ , url_type , uid , url)
	}

}

func UploadHeadFileNotify(pconfig *Config , pnotify *ss.MsgCommonNotify , file_server int) {
	var _func_ = "<UploadHeadFileNotify>"
	var puser_info *UserOnLine
	log := pconfig.Comm.Log
	uid := pnotify.Uid
	url := pnotify.StrV
	check_err := 0

	log.Debug("%s. uid:%d url:%s", _func_, uid, pnotify.StrV)
	for {
		//check online
		puser_info = GetUserInfo(pconfig, uid)
		if puser_info == nil {
			log.Err("%s not online!! uid:%d", _func_, uid)
			check_err = comm.FILE_UPT_CHECK_ONLINE
			break
		}

		//url
		old_url := puser_info.user_info.BasicInfo.HeadUrl
		if old_url == url {
			log.Info("%s head url not change! url:%s uid:%d" , _func_ , url , uid)
			return
		}

		//update url
		puser_info.user_info.BasicInfo.HeadUrl = url
		log.Info("%s update head url %s-->%s uid:%d" , _func_ , old_url , url , uid)
		if len(old_url) > 0 {
			check_err = comm.FILE_UPT_CHECK_DEL
			pnotify.StrV = old_url //del old head file
		}
		break
	}

	if check_err == 0 {
		return
	}

	//save profile
	SaveUserProfile(pconfig , uid)

	//back to file serv
	pnotify.IntV = int64(check_err)
	pss_msg , err := comm.GenDispMsg(ss.DISP_MSG_TARGET_FILE_SERVER , ss.DISP_MSG_METHOD_SPEC , ss.DISP_PROTO_TYPE_DISP_COMMON_NOTIFY ,
		file_server , pconfig.ProcId , 0 , pnotify)
	if err != nil {
		log.Err("%s gen file ss msg failed! err:%v uid:%d  url:%s" , _func_ , err , uid , url)
		return
	}

	//send
	SendToDisp(pconfig , 0 , pss_msg)
	return

}



func UploadChatFileNotify(pconfig *Config , pnotify *ss.MsgCommonNotify , file_server int) {
	var _func_ = "<UploadChatFileNotify>"
	var puser_info *UserOnLine
	log := pconfig.Comm.Log
	uid := pnotify.Uid
	grp_id := pnotify.GrpId
	url := pnotify.StrV
	check_err := 0

	log.Debug("%s. uid:%d grp_id:%d url:%s tmp_id:%d", _func_, uid, grp_id, pnotify.StrV , pnotify.IntV)
	for {
		//check online
		puser_info = GetUserInfo(pconfig, uid)
		if puser_info == nil {
			log.Err("%s not online!! uid:%d" , _func_ , uid)
			check_err = comm.FILE_UPT_CHECK_ONLINE
			break
		}

		//check group
		pchat_info := puser_info.user_info.BlobInfo.ChatInfo
		if pchat_info.AllGroup<=0 || pchat_info.AllGroups == nil {
			log.Err("%s enter no group!! uid:%d grp_id:%d" , _func_ , uid , grp_id)
			check_err = comm.FILE_UPT_CHECK_GROUP
			break
		}

		_ , ok := pchat_info.AllGroups[grp_id]
		if !ok {
			log.Err("%s no in group!! uid:%d grp_id:%d" , _func_ , uid , grp_id)
			check_err = comm.FILE_UPT_CHECK_GROUP
			break
		}

		break
	}
	//check value
	if check_err != 0 {
		pnotify.IntV = int64(check_err)
		pss_msg , err := comm.GenDispMsg(ss.DISP_MSG_TARGET_FILE_SERVER , ss.DISP_MSG_METHOD_SPEC , ss.DISP_PROTO_TYPE_DISP_COMMON_NOTIFY ,
			file_server , pconfig.ProcId , 0 , pnotify)

		if err != nil {
			log.Err("%s gen -->file ss msg failed! err:%v uid:%d grp_id:%d url:%s" , _func_ , err , uid , grp_id , url)
			return
		}

		//send
		SendToDisp(pconfig , 0 , pss_msg)
		return
	}

	//check passed
	//create req
	log.Debug("%s check passed! will create chat_msg! uid:%d grp_id:%d url:%s" , _func_ , uid , grp_id , url)
	preq := new(ss.MsgSendChatReq)
	preq.Uid = uid
	preq.TempId = pnotify.IntV
	pchat := new(ss.ChatMsg)
	pchat.GroupId = grp_id
	pchat.Content = url
	pchat.SenderUid = uid
	pchat.Sender = puser_info.user_info.BasicInfo.Name
	pchat.ChatType = ss.CHAT_MSG_TYPE_CHAT_TYPE_IMG
	preq.ChatMsg = pchat

	//ss
	pss_msg , err := comm.GenDispMsg(ss.DISP_MSG_TARGET_CHAT_SERVER , ss.DISP_MSG_METHOD_HASH , ss.DISP_PROTO_TYPE_DISP_SEND_CHAT_REQ ,
		0 , pconfig.ProcId , grp_id , preq)
	if err != nil {
		log.Err("%s gen send_chat ss failed! err:%v uid:%d grp_id:%d content:%s" , _func_ , err , uid , grp_id , url)
		return
	}

	//to chat_serv
	SendToDisp(pconfig , 0 , pss_msg)
}