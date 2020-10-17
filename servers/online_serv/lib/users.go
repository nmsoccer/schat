package lib

import (
	"schat/proto/ss"
	"time"
)

type UserInfo struct {
	last_heart int64
	login_serv int
}

type WorldOnlineUsers struct {
	world_online int64
	user_map     map[int64]*UserInfo
}

//get online user info
//@return: nil offline; else *user_info
func GetUserInfo(pconfig *Config, uid int64) *UserInfo {
	if pconfig.world_online.user_map == nil {
		return nil
	}
	pinfo, ok := pconfig.world_online.user_map[uid]
	if !ok {
		return nil
	}
	return pinfo
}

func RecvLoginNotify(pconfig *Config, uid int64, src_serv int) {
	var _func_ = "<RecvLoginNotify>"
	log := pconfig.Comm.Log

	//Get User Info
	curr_ts := time.Now().Unix()
	puser := GetUserInfo(pconfig, uid)
	if puser == nil {
		log.Info("%s add online user. uid:%d serv:%d", _func_, uid, src_serv)
		puser = new(UserInfo)
		puser.login_serv = src_serv
		puser.last_heart = curr_ts
		pconfig.world_online.user_map[uid] = puser
		pconfig.world_online.world_online++
		return
	}

	//update heart
	puser.last_heart = curr_ts
	puser.login_serv = src_serv
	return
}

func RecvLogoutNotify(pconfig *Config, uid int64, src_serv int) {
	var _func_ = "<RecvLogoutNotify>"
	log := pconfig.Comm.Log

	log.Info("%s uid:%d from %d", _func_, uid, src_serv)
	delete(pconfig.world_online.user_map, uid)
	pconfig.world_online.world_online--
	if pconfig.world_online.world_online < 0 {
		pconfig.world_online.world_online = 0
	}
}

func RecvBatchOnLineNotify(pconfig *Config, pnotify *ss.MsgCommonNotify, src_serv int) {
	var _func_ = "<RecvBatchOnLineNotify>"
	log := pconfig.Comm.Log

	if pnotify.Members == nil || len(pnotify.Members) == 0 {
		log.Err("%s no member report!", _func_)
		return
	}

	//set
	var uid int64
	for uid, _ = range pnotify.Members {
		RecvLoginNotify(pconfig, uid, src_serv)
	}

	//log.Debug("%s finish! count:%d src_serv:%d" , _func_ , len(pnotify.Members) , src_serv)
}
