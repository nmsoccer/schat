package lib

import (
	"time"
)

type UserInfo struct {
	last_heart int64
	login_serv int
}

type WorldOnlineUsers struct {
	world_online int64
	user_map map[int64]*UserInfo
}

func GetUserInfo(pconfig *Config , uid int64) *UserInfo {
	if pconfig.world_online.user_map == nil {
		return nil
	}
	pinfo , ok := pconfig.world_online.user_map[uid]
	if !ok {
		return nil
	}
	return pinfo
}

func RecvLoginNotify(pconfig *Config , uid int64 , src_serv int) {
	var _func_ = "<RecvLoginNotify>"
	log := pconfig.Comm.Log

	//Get User Info
	curr_ts := time.Now().Unix()
	puser := GetUserInfo(pconfig , uid)
	if puser == nil {
		log.Debug("%s add online user. uid:%d" , _func_ , uid)
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

func RecvLogoutNotify(pconfig *Config , uid int64 , src_serv int) {
	var _func_ = "<RecvLogoutNotify>"
	log := pconfig.Comm.Log

	log.Debug("%s uid:%d from %d" , _func_ , uid , src_serv)
	delete(pconfig.world_online.user_map , uid)
	pconfig.world_online.world_online--
	if pconfig.world_online.world_online < 0 {
		pconfig.world_online.world_online = 0
	}
}