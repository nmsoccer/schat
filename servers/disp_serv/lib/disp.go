package lib

import (
	"schat/proto/ss"
	"schat/servers/comm"
)

func RecvDispMsg(pconfig *Config , pdisp *ss.MsgDisp , from int , msg []byte) {
	var _func_ = "<RecvDispMsg>"
	log := pconfig.Comm.Log

	//Dispatch to Target
	//log.Debug("%s disp_proto:%d from:%d disp_from:%d target:%d spec:%d method:%d" , _func_ , pdisp.ProtoType , from , pdisp.FromServer , pdisp.Target ,
	//	pdisp.SpecServer , pdisp.Method)

	//to spec server
	if pdisp.Method == ss.DISP_MSG_METHOD_SPEC {
		if pdisp.SpecServer <= 0 {
			log.Err("%s fail! spec method but spec server not set!" ,_func_)
			return
		}
		SendToServ(pconfig , int(pdisp.SpecServer) , msg)
		return
	}

	//Dispatch target
	switch pdisp.Target {
	case ss.DISP_MSG_TARGET_LOGIC_SERVER:
		DispToLogicServ(pconfig , pdisp , msg)
	case ss.DISP_MSG_TARGET_CHAT_SERVER:
		DispToChastServ(pconfig , pdisp , msg)
	case ss.DISP_MSG_TARGET_ONLINE_SERVER:
		DispToOnlineServ(pconfig , pdisp , msg)
	case ss.DISP_MSG_TARGET_FILE_SERVER:
		DispToFileServ(pconfig , pdisp , msg)
	case ss.DISP_MSG_TARGET_DIR_SERVER:
		DispToDirServ(pconfig , pdisp , msg)
	default:
		log.Err("%s target:%d can not handle!" , _func_ , pdisp.Target)
	}

}

func DispToLogicServ(pconfig *Config , pdisp *ss.MsgDisp , msg []byte) {
	var _func_ = "<DispToLogicServ>"
	var target_serv int
	log := pconfig.Comm.Log

	//method
	switch pdisp.Method {
	case ss.DISP_MSG_METHOD_RAND:
		target_serv = comm.SelectProperServ(pconfig.Comm , comm.SELECT_METHOD_RAND , 0 , pconfig.FileConfig.LogicServList , pconfig.Comm.PeerStats ,
			comm.PERIOD_HEART_BEAT_DEFAULT/1000)
	case ss.DISP_MSG_METHOD_HASH:
		target_serv = comm.SelectProperServ(pconfig.Comm , comm.SELECT_METHOD_HASH , pdisp.HashV , pconfig.FileConfig.LogicServList , pconfig.Comm.PeerStats ,
			comm.PERIOD_HEART_BEAT_DEFAULT/1000)
	case ss.DISP_MSG_METHOD_ALL:
		for i:=0; i < len(pconfig.FileConfig.LogicServList); i++ {
			SendToServ(pconfig , pconfig.FileConfig.LogicServList[i] , msg)
		}
		return
	default:
		log.Err("%s method:%d illegal! proto:%d" , _func_ , pdisp.Method , pdisp.ProtoType)
		return
	}

    //check
    if target_serv <= 0 {
    	log.Err("%s fail! no proper target found! method:%d proto:%d hash:%d" , _func_ , pdisp.Method , pdisp.ProtoType ,
    		pdisp.HashV)
	}


	//send to target
	log.Debug("%s send to:%d method:%d hash:%d" , _func_ , target_serv , pdisp.Method , pdisp.HashV)
	SendToServ(pconfig , target_serv , msg)
}

func DispToChastServ(pconfig *Config , pdisp *ss.MsgDisp , msg []byte) {
	var _func_ = "<DispToChastServ>"
	var target_serv int
	log := pconfig.Comm.Log

	//method
	switch pdisp.Method {
	case ss.DISP_MSG_METHOD_RAND:
		target_serv = comm.SelectProperServ(pconfig.Comm , comm.SELECT_METHOD_RAND , 0 , pconfig.FileConfig.ChatServList , pconfig.Comm.PeerStats ,
			comm.PERIOD_HEART_BEAT_DEFAULT/1000)
	case ss.DISP_MSG_METHOD_HASH:
		target_serv = comm.SelectProperServ(pconfig.Comm , comm.SELECT_METHOD_HASH , pdisp.HashV , pconfig.FileConfig.ChatServList , pconfig.Comm.PeerStats ,
			comm.PERIOD_HEART_BEAT_DEFAULT/1000)
	case ss.DISP_MSG_METHOD_ALL:
		for i:=0; i < len(pconfig.FileConfig.ChatServList); i++ {
			SendToServ(pconfig , pconfig.FileConfig.ChatServList[i] , msg)
		}
		return
	default:
		log.Err("%s method:%d illegal! proto:%d" , _func_ , pdisp.Method , pdisp.ProtoType)
		return
	}

	//check
	if target_serv <= 0 {
		log.Err("%s fail! no proper target found! method:%d proto:%d hash:%d" , _func_ , pdisp.Method , pdisp.ProtoType ,
			pdisp.HashV)
	}

	//send to target
	log.Debug("%s send to:%d method:%d hash:%d" , _func_ , target_serv , pdisp.Method , pdisp.HashV)
	SendToServ(pconfig , target_serv , msg)
}

func DispToOnlineServ(pconfig *Config , pdisp *ss.MsgDisp , msg []byte) {
	var _func_ = "<DispToOnlineServ>"
	var target_serv int
	log := pconfig.Comm.Log

	//method
	switch pdisp.Method {
	case ss.DISP_MSG_METHOD_RAND:
		target_serv = comm.SelectProperServ(pconfig.Comm , comm.SELECT_METHOD_RAND , 0 , pconfig.FileConfig.OnlineServList , pconfig.Comm.PeerStats ,
			comm.PERIOD_HEART_BEAT_DEFAULT/1000)
	case ss.DISP_MSG_METHOD_HASH:
		target_serv = comm.SelectProperServ(pconfig.Comm , comm.SELECT_METHOD_HASH , pdisp.HashV , pconfig.FileConfig.OnlineServList , pconfig.Comm.PeerStats ,
			comm.PERIOD_HEART_BEAT_DEFAULT/1000)
	case ss.DISP_MSG_METHOD_ALL:
		for i:=0; i < len(pconfig.FileConfig.OnlineServList); i++ {
			SendToServ(pconfig , pconfig.FileConfig.OnlineServList[i] , msg)
		}
		return
	default:
		log.Err("%s method:%d illegal! proto:%d" , _func_ , pdisp.Method , pdisp.ProtoType)
		return
	}

	//check
	if target_serv <= 0 {
		log.Err("%s fail! no proper target found! method:%d proto:%d hash:%d" , _func_ , pdisp.Method , pdisp.ProtoType ,
			pdisp.HashV)
	}

	//send to target
	log.Debug("%s send to:%d method:%d hash:%d" , _func_ , target_serv , pdisp.Method , pdisp.HashV)
	SendToServ(pconfig , target_serv , msg)
}

func DispToFileServ(pconfig *Config , pdisp *ss.MsgDisp , msg []byte) {
	var _func_ = "<DispToFileServ>"
	var target_serv int
	log := pconfig.Comm.Log

	//method
	switch pdisp.Method {
	case ss.DISP_MSG_METHOD_RAND:
		target_serv = comm.SelectProperServ(pconfig.Comm , comm.SELECT_METHOD_RAND , 0 , pconfig.FileConfig.FileServList , pconfig.Comm.PeerStats ,
			comm.PERIOD_HEART_BEAT_DEFAULT/1000)
	case ss.DISP_MSG_METHOD_HASH:
		target_serv = comm.SelectProperServ(pconfig.Comm , comm.SELECT_METHOD_HASH , pdisp.HashV , pconfig.FileConfig.FileServList , pconfig.Comm.PeerStats ,
			comm.PERIOD_HEART_BEAT_DEFAULT/1000)
	case ss.DISP_MSG_METHOD_ALL:
		for i:=0; i < len(pconfig.FileConfig.FileServList); i++ {
			SendToServ(pconfig , pconfig.FileConfig.FileServList[i] , msg)
		}
		return
	default:
		log.Err("%s method:%d illegal! proto:%d" , _func_ , pdisp.Method , pdisp.ProtoType)
		return
	}

	//check
	if target_serv <= 0 {
		log.Err("%s fail! no proper target found! method:%d proto:%d hash:%d" , _func_ , pdisp.Method , pdisp.ProtoType ,
			pdisp.HashV)
	}

	//send to target
	log.Debug("%s send to:%d method:%d hash:%d" , _func_ , target_serv , pdisp.Method , pdisp.HashV)
	SendToServ(pconfig , target_serv , msg)
}

func DispToDirServ(pconfig *Config , pdisp *ss.MsgDisp , msg []byte) {
	var _func_ = "<DispToDirServ>"
	var target_serv int
	log := pconfig.Comm.Log

	//method
	switch pdisp.Method {
	case ss.DISP_MSG_METHOD_RAND:
		target_serv = comm.SelectProperServ(pconfig.Comm , comm.SELECT_METHOD_RAND , 0 , pconfig.FileConfig.DirServList , pconfig.Comm.PeerStats ,
			comm.PERIOD_HEART_BEAT_DEFAULT/1000)
	case ss.DISP_MSG_METHOD_HASH:
		target_serv = comm.SelectProperServ(pconfig.Comm , comm.SELECT_METHOD_HASH , pdisp.HashV , pconfig.FileConfig.DirServList , pconfig.Comm.PeerStats ,
			comm.PERIOD_HEART_BEAT_DEFAULT/1000)
	case ss.DISP_MSG_METHOD_ALL:
		for i:=0; i < len(pconfig.FileConfig.DirServList); i++ {
			SendToServ(pconfig , pconfig.FileConfig.DirServList[i] , msg)
		}
		return
	default:
		log.Err("%s method:%d illegal! proto:%d" , _func_ , pdisp.Method , pdisp.ProtoType)
		return
	}

	//check
	if target_serv <= 0 {
		log.Err("%s fail! no proper target found! method:%d proto:%d hash:%d" , _func_ , pdisp.Method , pdisp.ProtoType ,
			pdisp.HashV)
	}

	//send to target
	log.Debug("%s send to:%d method:%d hash:%d" , _func_ , target_serv , pdisp.Method , pdisp.HashV)
	SendToServ(pconfig , target_serv , msg)
}