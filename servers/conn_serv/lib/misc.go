package lib

import (
	"schat/proto/cs"
	"schat/proto/ss"
)

func RecvCommonNotify(pconfig *Config , pnotify *ss.MsgCommonNotify) {
	var _func_ = "<RecvCommonNotify>"
	log := pconfig.Comm.Log
    uid := pnotify.Uid

    //c_key
    c_key := GetClientKey(pconfig , uid)
    if c_key == 0 {
    	log.Err("%s offline! uid:%d type:%d" , _func_ , uid , pnotify.NotifyType)
    	return
	}

	//cs
	pv , err := cs.Proto2Msg(cs.CS_PROTO_COMMON_NOTIFY)
	if err != nil {
		log.Err("%s get msg fail! uid:%d err:%v" , _func_ , uid , err)
		return
	}
	pmsg , ok := pv.(*cs.CSCommonNotify)
	if !ok {
		log.Err("%s not CSCommonNotify! uid:%d" , _func_ , uid)
		return
	}

	//proto
	switch pnotify.NotifyType {
	case ss.COMMON_NOTIFY_TYPE_NOTIFY_FILE_ADDR:
		if pnotify.IntV==0 || pnotify.Strs==nil || len(pnotify.Strs)==0 {
			log.Err("%s file addr empty! uid:%d" , _func_ , uid)
			return
		}
		pmsg.NotifyType = cs.COMMON_NOTIFY_T_FILE_ADDR
		pmsg.StrS = pnotify.Strs
		pmsg.IntV = pnotify.IntV
 		log.Debug("%s file_addr:%v uid:%d" , _func_ , pmsg.StrS , uid)
	case ss.COMMON_NOTIFY_TYPE_NOTIFY_ADD_MEMBER:
		if pnotify.GrpId==0 || pnotify.StrV=="" || pnotify.IntV==0{
			log.Err("%s add member arg nil! uid:%d" , _func_ , uid)
			return
		}
		pmsg.NotifyType = cs.COMMON_NOTIFY_T_ADD_MEM
		pmsg.GrpId = pnotify.GrpId
		pmsg.StrV = pnotify.StrV
		pmsg.IntV = pnotify.IntV
		log.Debug("%s add member id:%d grp_id:%d grp_name:%s" , _func_ , pmsg.IntV , pmsg.GrpId , pmsg.StrV)
	case ss.COMMON_NOTIFY_TYPE_NOTIFY_DEL_MEMBER:
		if pnotify.GrpId==0 || pnotify.StrV=="" || pnotify.IntV==0{
			log.Err("%s del member arg nil! uid:%d" , _func_ , uid)
			return
		}
		pmsg.NotifyType = cs.COMMON_NOTIFY_T_DEL_MEM
		pmsg.GrpId = pnotify.GrpId
		pmsg.StrV = pnotify.StrV
		pmsg.IntV = pnotify.IntV
		log.Debug("%s del member id:%d grp_id:%d grp_name:%s" , _func_ , pmsg.IntV , pmsg.GrpId , pmsg.StrV)
	default:
		log.Err("%s unknown proto:%d uid:%d" , _func_ , pnotify.NotifyType , uid)
		return
	}

	//send
	SendToClient(pconfig , c_key , cs.CS_PROTO_COMMON_NOTIFY , pmsg)
}
