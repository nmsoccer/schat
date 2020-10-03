package lib

import (
    "schat/proto/ss"
    "time"
    "schat/servers/comm"
)

func SendHeartBeatMsg(arg interface{}) {
	var pconfig *Config;
	var _func_ = "<SendHeartBeatMsg>";
	curr_ts := time.Now().Unix();
	pconfig , ok := arg.(*Config);
	if !ok {
		return;
	}
	log := pconfig.Comm.Log;

	//ss_msg
	var ss_msg ss.SSMsg
	pheart := new(ss.MsgHeartBeatReq)
	pheart.Ts = curr_ts
	err := comm.FillSSPkg(&ss_msg , ss.SS_PROTO_TYPE_HEART_BEAT_REQ , pheart)
	if err != nil {
		log.Err("%s gen ss_msg failed! err:%v" , _func_ , err)
	} else {
		SendToLogic(pconfig , &ss_msg)
	}
    
    
    //report
    pconfig.ReportServ.Report(comm.REPORT_PROTO_SERVER_HEART , curr_ts , "" , nil);
    pconfig.ReportServ.Report(comm.REPORT_PROTO_CONN_NUM , int64(pconfig.TcpServ.GetConnNum()) , "ClientConn" , nil);    
	return;
}

func RecvHeartBeatReq(pconfig *Config , preq *ss.MsgHeartBeatReq , from int) {
	var _func_ = "<RecvHeartBeatReq>";
	var log = pconfig.Comm.Log;
	var stats = pconfig.Comm.PeerStats;
	
	_ , ok := stats[from];
	if ok {
	    //log.Debug("%s update heartbeat server:%d %v --> %v" , _func_ , from , last , preq.GetTs());	
	} else {
		log.Debug("%s set  heartbeat server:%d %v" , _func_ , from , preq.GetTs());	
	}
	stats[from] = preq.GetTs();
}

func ReportSyncServer(arg interface{}) {
	var pconfig *Config;
	pconfig , ok := arg.(*Config);
	if !ok {
		return;
	}
	
	//msg
	pmsg := new(comm.SyncServerMsg);
	pmsg.StartTime = pconfig.Comm.StartTs;
	
	//send
	pconfig.ReportServ.Report(comm.REPORT_PROTO_SYNC_SERVER , 0 , "" , pmsg);
}

//ratio = int(valid_conn/max_conn * 100); if ratio<=30 load=0; if ratio>30 load=ratio
func NotifyLoad(arg interface{}) {
	pconfig , ok := arg.(*Config)
	if !ok {
		return
	}
	var _func_ = "<NotifyLoad>"
	log := pconfig.Comm.Log

	load := 0
	valid_conn := pconfig.TcpServ.GetConnNum()
	ratio := valid_conn*100/pconfig.FileConfig.MaxConn
	if ratio > 30 {
		load = ratio
	}

	//notify
	var ss_msg ss.SSMsg
	pnotify := new(ss.MsgCommonNotify)
	pnotify.NotifyType = ss.COMMON_NOTIFY_TYPE_NOTIFY_SERV_LOAD
	pnotify.Uid = int64(pconfig.ProcId)
	pnotify.IntV = int64(load)

    err := comm.FillSSPkg(&ss_msg , ss.SS_PROTO_TYPE_COMMON_NOTIFY , pnotify)
    if err != nil {
    	log.Err("%s gen ss failed! err:%v" , _func_ , err)
    	return
	}

	SendToLogic(pconfig , &ss_msg)
}
