package lib

import (
	"net"
	"net/http"
	"schat/proto/ss"
	"schat/servers/comm"
	"strings"
)

const (
	//TOKEN LEN
	FILE_SERV_TOKEN_LEN = 12

	PERIOD_UPDATE_TOKEN = 3600000 //1h update token
	PERIOD_SYNC_TOKEN   = 600000  //10min sync token to dir

	//SAFE MOD
	SAFE_MOD_NONE = 0
	SAFE_MOD_PATH = 1	//check path valid when query files
	SAFE_MOD_TOKEN = 2	//check token valid, will include above check when query files
	SAFE_MOD_IP = 3	//check ip valid,will include above check when query files. query failed times will record ip to blacklist

	//WATCHER
	SUSPECT_JAIL_TIME = 3600 //1h suspect ip in watcher if no fail again
	SUSPECT_SAFE_FAIL = 5 //most safe failed times
	SUSPECT_FAIL_PUNISH_SEC = 60 //overflow safe door, will block ip punish_sec per 1 failed access
	SUSPECT_CHECK_PEROID = 60
)

func UpdateServToken(arg interface{}) {
	var _func_ = "<UpdateServToken>"
	pconfig, ok := arg.(*Config)
	if !ok {
		return
	}
	log := pconfig.Comm.Log

	//update token
	new_token, err := comm.GenRandNumStr(FILE_SERV_TOKEN_LEN)
	if err != nil {
		log.Err("%s fail! rand str err:%v", _func_, err)
		return
	}

	pconfig.NowToken = new_token
	log.Info("%s update new token:%s", _func_, new_token)

	//to file_server
	pmsg := new(FileMsg)
	pmsg.msg_type = FILE_MSG_UPDATE_TOKEN
	pmsg.str_v = new_token
	pconfig.FileServer.Send(pmsg)

	//to dir serv
	SyncServToken(pconfig)
}

func SyncServToken(arg interface{}) {
	var _func_ = "<SyncServToken>"
	pconfig, ok := arg.(*Config)
	if !ok {
		return
	}
	log := pconfig.Comm.Log

	//to dir
	pnotify := new(ss.MsgCommonNotify)
	pnotify.NotifyType = ss.COMMON_NOTIFY_TYPE_NOTIFY_FILE_TOKEN
	pnotify.IntV = int64(pconfig.FileConfig.ServIndex)
	pnotify.StrV = pconfig.NowToken

	pss_msg, err := comm.GenDispMsg(ss.DISP_MSG_TARGET_DIR_SERVER, ss.DISP_MSG_METHOD_ALL, ss.DISP_PROTO_TYPE_DISP_COMMON_NOTIFY,
		0, pconfig.ProcId, 0, pnotify)
	if err != nil {
		log.Err("%s gen ss failed! err:%v", _func_, err)
		return
	}

	//send
	SendToDisp(pconfig, 0, pss_msg)
}

//dir --> file query
func RecvFileTokenNotify(pconfig *Config, pnotify *ss.MsgCommonNotify, dir_serv int) {
	var _func_ = "<RecvFileTokenNotify>"
	log := pconfig.Comm.Log
	serv_index := int(pnotify.IntV)

	//check index
	if serv_index != pconfig.FileConfig.ServIndex {
		log.Info("%s not my index! %d:%d abandon!", _func_, serv_index, pconfig.FileConfig.ServIndex)
		return
	}

	//back to dir
	log.Info("%s get token query from dir_serv:%d will send back! serv_index:%d", _func_, dir_serv, serv_index)
	SyncServToken(pconfig)
}

/* --------------code from gayhub -------------*/
// refer https://github.com/thinkeridea/go-extend/tree/master/exnet
// ClientIP 尽最大努力实现获取客户端 IP 的算法。
// 解析 X-Real-IP 和 X-Forwarded-For 以便于反向代理（nginx 或 haproxy）可以正常工作。
func ClientIP(r *http.Request) string {
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	ip := strings.TrimSpace(strings.Split(xForwardedFor, ",")[0])
	if ip != "" {
		return ip
	}

	ip = strings.TrimSpace(r.Header.Get("X-Real-Ip"))
	if ip != "" {
		return ip
	}

	if ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		return ip
	}

	return ""
}

// ClientPublicIP 尽最大努力实现获取客户端公网 IP 的算法。
// 解析 X-Real-IP 和 X-Forwarded-For 以便于反向代理（nginx 或 haproxy）可以正常工作。
func ClientPublicIP(r *http.Request) string {
	var ip string
	for _, ip = range strings.Split(r.Header.Get("X-Forwarded-For"), ",") {
		ip = strings.TrimSpace(ip)
		if ip != "" && !HasLocalIPddr(ip) {
			return ip
		}
	}

	ip = strings.TrimSpace(r.Header.Get("X-Real-Ip"))
	if ip != "" && !HasLocalIPddr(ip) {
		return ip
	}

	if ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		if !HasLocalIPddr(ip) {
			return ip
		}
	}

	return ""
}

// HasLocalIPddr 检测 IP 地址字符串是否是内网地址
func HasLocalIPddr(ip string) bool {
	return HasLocalIP(net.ParseIP(ip))
}

// HasLocalIP 检测 IP 地址是否是内网地址
// 通过直接对比ip段范围效率更高
func HasLocalIP(ip net.IP) bool {
	if ip.IsLoopback() {
		return true
	}

	ip4 := ip.To4()
	if ip4 == nil {
		return false
	}

	return ip4[0] == 10 || // 10.0.0.0/8
		(ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31) || // 172.16.0.0/12
		(ip4[0] == 169 && ip4[1] == 254) || // 169.254.0.0/16
		(ip4[0] == 192 && ip4[1] == 168) // 192.168.0.0/16
}

/* --------------code from gayhub end -------------*/