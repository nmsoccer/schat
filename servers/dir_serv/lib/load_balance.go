package lib

import (
	"encoding/json"
	"math/rand"
	"schat/proto/ss"
	"schat/servers/comm"
	"sync"
	"time"
)

const (
	MAX_PEER_SERV_LIVE = 30 //30 seconds
)

type FileServInfo struct {
	ServIndex int
	ServAddr  string
}

type ConnServInfo struct {
	ServProc int
	ServAddr string
	Weight   int  //[0 , 100]
	Load     int  //[0 , 100] ratio = int(valid_conn/max_conn * 100); if ratio<=30 load=0; if ratio>30 load=ratio
	LastHeart int64
}


type AllServerInfo struct {
	sync.RWMutex
    FileServList []*FileServInfo
    ConnServList []int
    ConnServMap  map[int]*ConnServInfo
}

type ServerResponse struct {
	FileServList []*FileServInfo `json:"file_serv"`
	ConnServ     string          `json:"conn_serv"`
}



func InitAllServerInfo(pconfig *Config , pall *AllServerInfo) bool {
	var _func_ = "<InitAllServerInfo>"
	log := pconfig.Comm.Log

	//file serv
    if len(pconfig.FileConfig.FileServIndex) != len(pconfig.FileConfig.FileServAddr) {
        log.Err("%s fail! FileServ len index not match addr! please check!" , _func_)
        return false
	}
	file_count := len(pconfig.FileConfig.FileServIndex)
	pall.FileServList = make([]*FileServInfo , file_count)
	for i:=0; i<file_count; i++ {
		pall.FileServList[i] = new(FileServInfo)
		pall.FileServList[i].ServIndex = pconfig.FileConfig.FileServIndex[i]
		pall.FileServList[i].ServAddr  = pconfig.FileConfig.FileServAddr[i]
	}
	log.Info("%s init file_serv finish! count:%d info:%v" , _func_ , file_count , pall.FileServList)

    //conn serv
	if len(pconfig.FileConfig.ConnServProc) != len(pconfig.FileConfig.ConnServAddr) || len(pconfig.FileConfig.ConnServProc) != len(pconfig.FileConfig.ConnServWeigth) {
		log.Err("%s fail! ConnServ len not match! please check!" , _func_)
		return false
	}
	conn_count := len(pconfig.FileConfig.ConnServProc)
	pall.ConnServMap = make(map[int]*ConnServInfo)
	pall.ConnServList = make([]int , conn_count)
    for i:=0; i<conn_count; i++ {
    	pinfo := new(ConnServInfo)
    	if pconfig.FileConfig.ConnServWeigth[i] < comm.MIN_LOAD_WEIGHT {
			pconfig.FileConfig.ConnServWeigth[i] = comm.MIN_LOAD_WEIGHT
		}
		if pconfig.FileConfig.ConnServWeigth[i] > comm.MAX_LOAD_WEIGHT {
			pconfig.FileConfig.ConnServWeigth[i] = comm.MAX_LOAD_WEIGHT
		}

    	pinfo.Weight = pconfig.FileConfig.ConnServWeigth[i]
    	pinfo.ServAddr = pconfig.FileConfig.ConnServAddr[i]
    	pinfo.ServProc = pconfig.FileConfig.ConnServProc[i]
    	pinfo.Load = comm.MIN_LOAD_WEIGHT
    	pall.ConnServMap[pinfo.ServProc] = pinfo

    	pall.ConnServList[i] = pinfo.ServProc
	}
	log.Info("%s init conn_serv finish! count:%d map:%v , list:%v" , _func_ , conn_count , pall.ConnServMap ,
		pall.ConnServList)
    return true
}

type rand_item struct {
	proc int
	score int
	addr string
}


func RandOneConn(pconfig *Config , pall *AllServerInfo) string {
	var _func_ = "<RandOneConn>"
	log := pconfig.Comm.Log
    curr_ts := time.Now().Unix()

	count := 0
	rand_list := make([]rand_item , len(pall.ConnServList))
	score := 0
	total_val := 0

	//gen rand list
	pall.RLock()
	for _ , info := range(pall.ConnServMap) {
		if info.Weight == 0 {
			log.Debug("%s proc:%d addr:%s weight 0!" , _func_ , info.ServProc , info.ServAddr)
			continue
		}
		if (curr_ts-info.LastHeart) > MAX_PEER_SERV_LIVE {
			log.Info("%s lose proc:%d addr:%s heart! last_heart:%d!" , _func_ , info.ServProc , info.ServAddr , info.LastHeart)
			continue
		}


		score = info.Weight - info.Load
		if score <= comm.MIN_LOAD_WEIGHT {
			score = comm.MIN_LOAD_WEIGHT + 1 //least 1 score
		}
		total_val += score
		rand_list[count].score = score
		rand_list[count].proc = info.ServProc
		rand_list[count].addr = info.ServAddr
		count++
	}
    pall.RUnlock()

	//check
	if total_val <= 0 {
		return ""
	}

	//rand
	rand_v := rand.Intn(total_val)
	calc_base := 0
	for i:=0; i<count; i++ {
		calc_base += rand_list[i].score
		if rand_v < calc_base {
			//bingo
			log.Debug("%s rand_v:%d total_v:%d base:%d rand_list:%v" , _func_ , rand_v , total_val , calc_base , rand_list)
			return rand_list[i].addr
		}
	}

	return ""
}



func GenServerResponseStr(pconfig *Config , pall *AllServerInfo) string{
	var _func_ = "<GenServerResponseStr>"
	log := pconfig.Comm.Log
    curr_ts := time.Now().Unix()

	var resp ServerResponse
	//all file info deleted<fill in query on user login>
	//resp.FileServList = pall.FileServList

	//rand connect
	switch len(pall.ConnServList) {
	case 0:
		//not connect
	case 1:
		pall.RLock()
		info := pall.ConnServMap[pall.ConnServList[0]]
		if info.Weight == 0 {
			log.Debug("%s proc:%d addr:%s weight 0!" , _func_ , info.ServProc , info.ServAddr)
			break
		}
		if (curr_ts-info.LastHeart) > MAX_PEER_SERV_LIVE {
			log.Info("%s lose proc:%d addr:%s heart! last_heart:%d!" , _func_ , info.ServProc , info.ServAddr , info.LastHeart)
			break
		}
		resp.ConnServ = info.ServAddr
		pall.RUnlock()
    default:
		resp.ConnServ = RandOneConn(pconfig , pall)
	}

	//enc
	enc_data , err := json.Marshal(&resp)
	if err != nil {
		log.Err("%s json marshall failed! err:%v" , _func_ , err)
		return ""
	}

	//return
    return string(enc_data)
}

func RecvLoadNotify(pconfig *Config , pnotify *ss.MsgCommonNotify) {
	var _func_ = "<RecvLoadNotify>"
	log := pconfig.Comm.Log

    //init
    curr_ts := time.Now().Unix()
    serv_id := int(pnotify.Uid)
    load := int(pnotify.IntV)

    if pconfig.ServerInfo == nil || pconfig.ServerInfo.ConnServMap==nil {
    	log.Err("%s conn server info nil!" , _func_)
    	return
	}

	//upate
	pconfig.ServerInfo.Lock()
    defer pconfig.ServerInfo.Unlock()

	pinfo , ok := pconfig.ServerInfo.ConnServMap[serv_id]
	if !ok {
		log.Err("%s conn server:%d not included!" , _func_ , serv_id)
		return
	}

    pinfo.LastHeart = curr_ts
    pinfo.Load = load
}