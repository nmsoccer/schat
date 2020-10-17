package lib

import (
	"fmt"
	"net/http"
	"time"
)

type HttpServer struct {
	pconfig   *Config
	serv_addr string
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
		http.ListenAndServe(hs.serv_addr, nil)
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
	resp_str := GenServerResponseStr(hs.pconfig, hs.pconfig.ServerInfo)
	if resp_str == "" {
		fmt.Fprintf(w, "{}")
		return
	}
	fmt.Fprintf(w, resp_str)
}
