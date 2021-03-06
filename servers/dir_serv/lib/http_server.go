package lib

import (
	"fmt"
	"net/http"
	"time"
)

const (
	FORM_QUERY_KEY = "query_key"

	//HTTPS
	KEY_FILE_PATH = "./cfg/key.pem"
	CERT_FILE_PATH = "./cfg/cert.pem"
)

type HttpServer struct {
	pconfig   *Config
	serv_addr string
	query_key string
}

//var all_server_info AllServerInfo
//var pall = &all_server_info

func StartHttpServer(pconfig *Config) *HttpServer {
	var _func_ = "<StartHttpServer>"
	log := pconfig.Comm.Log

	//new hs
	hs := new(HttpServer)
	hs.pconfig = pconfig
	hs.serv_addr = pconfig.FileConfig.HttpAddr
	hs.query_key = pconfig.FileConfig.QueryKey
	log.Info("%s at %s finish!", _func_, hs.serv_addr)

	go hs.start_serv()
	return hs
}

func (hs *HttpServer) start_serv() {
	//register handler
	go func() {
		http.Handle("/", http.HandlerFunc(hs.index_handler))
		http.Handle("/query", http.HandlerFunc(hs.query_handler))

		//listen
		http.ListenAndServeTLS(hs.serv_addr, CERT_FILE_PATH , KEY_FILE_PATH , nil)
	}()

	//main proc
	for {
		//
		time.Sleep(10 * time.Millisecond)
	}
}

func (hs *HttpServer) index_handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "index!")
}

func (hs *HttpServer) query_handler(w http.ResponseWriter, r *http.Request) {
	var _func_ = "<query_handler>"
	log := hs.pconfig.Comm.Log

	//get key
	query_key := r.FormValue(FORM_QUERY_KEY)
	if len(query_key)<=0 {
		log.Err("%s query_key not valid!" , _func_)
		fmt.Fprintf(w, "{}")
		return
	}

	if query_key != hs.query_key {
		log.Err("%s query_key not match!client:%s server:%s" , _func_ , query_key , hs.query_key)
		fmt.Fprintf(w, "{}")
		return
	}


	resp_str := GenServerResponseStr(hs.pconfig, hs.pconfig.ServerInfo)
	if resp_str == "" {
		fmt.Fprintf(w, "{}")
		return
	}
	fmt.Fprintf(w, resp_str)
}
