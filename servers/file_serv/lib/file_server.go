package lib

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"schat/servers/comm"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	UPLOAD_TMPL = "./html_tmpl/upload.html"

	FORM_LABEL_UID = "uid"
	FORM_LABEL_GRP_ID = "grp_id"
	FORM_LABEL_TMP_ID = "tmp_id"
	FORM_LABEL_UPLOAD = "upload_file"


	UPLOAD_RESULT_SUCCESS = 0
	UPLOAD_RESULT_FAILED = 1
	UPLOAD_RESULT_SIZE   = 2

	FILE_MSG_CHAN_SIZE = 1000
	MAX_FETCH_PER_TICK = 10

	//FILE_MSG_TYPE
	FILE_MSG_EXIT   = 0     //exit
	FILE_MSG_UPLOAD = 1     //server-->main upload one file
	FILE_MSG_UPLOAD_CHECK_FAIL = 2 //main-->server upload check err
)


type FileMsg struct {
    msg_type int  //refer FILE_MSG_XX
    uid      int64
    grp_id   int64
    url      string
    int_v    int64
    str_v    string
}


type FileServer struct {
	sync.Mutex
	pconfig *Config
	//config
	serv_index     int
	http_addr      string
	file_size   int
	file_path  string
	//channel
    snd_chan  chan *FileMsg
    recv_chan  chan *FileMsg
}

type UploadResult struct {
	Result int `json:"result"`
	Info   string `json:"info"`
	TmpId  int64 `json:"tmp_id"`
}

var http_fs http.Handler //file server

func StartFileServer(pconfig *Config) *FileServer {
	var _func_ = "<StartFileServer>"
	log := pconfig.Comm.Log

    //alloc
    fs := new(FileServer)
    fs.pconfig = pconfig
    fs.serv_index = pconfig.FileConfig.ServIndex
    fs.http_addr = pconfig.FileConfig.HttpAddr
    fs.file_size = pconfig.FileConfig.MaxFileSize
    fs.file_path = pconfig.FileConfig.RealFilePath
    fs.recv_chan = make(chan *FileMsg , FILE_MSG_CHAN_SIZE)
	fs.snd_chan = make(chan *FileMsg , FILE_MSG_CHAN_SIZE)

    //new http_fs
    http_fs = http.FileServer(http.Dir(fs.file_path))
    if http_fs == nil {
    	log.Err("%s FileServer %s failed!" , _func_ , fs.file_path)
    	return nil
	}

    //start
    log.Info("%s addr:%s size:%d index:%d real_path:%s" , _func_ , fs.http_addr , fs.file_size , fs.serv_index ,
    	fs.file_path)
    go fs.start_serv()
    return fs
}


//upper send to FileServer
func (fs *FileServer) Send(pmsg *FileMsg) bool{
	//check full
	if len(fs.recv_chan) >= cap(fs.recv_chan) {
		//full
		return false
	}

	//send
	fs.recv_chan <- pmsg
	return true
}

//upper read from FileServer
func (fs *FileServer) Read(msg_list []*FileMsg) int{
	//set count
	count := len(fs.snd_chan)
	if count > len(msg_list) {
		count = len(msg_list)
	}

    //read
    for i:=0; i<count; i++ {
    	msg_list[i] = <- fs.snd_chan
	}
	return count
}

func (fs *FileServer) Close() {
	pmsg := new(FileMsg)
	pmsg.msg_type = FILE_MSG_EXIT
	fs.recv_chan <- pmsg
}


/*-----------------------------------static func-------------------------------*/
func (fs *FileServer) start_serv() {
	log := fs.pconfig.Comm.Log

	//http server
    go func() {
		//reg handler
		http.Handle("/" , http.HandlerFunc(fs.index_handler))
		http.Handle("/upload/" , http.HandlerFunc(fs.upload_handler))
		http.Handle("/static/" , http.HandlerFunc(fs.static_handler))

		err := http.ListenAndServe(fs.http_addr, nil)
		if err != nil {
			log.Err("file_server start_serv at %s failed! err:%v", fs.http_addr, err)
		}
	}()


	//main proc
	for {
		//recv msg
		fs.recv_msg()
        //sleep
		time.Sleep( 10 * time.Millisecond) //10ms
	}
}

//@return:exit
func (fs *FileServer) recv_msg() bool {
	var _func_ = "<FileServer.recv_msg>"
	var pmsg *FileMsg
	log := fs.pconfig.Comm.Log

	//fetch msg
	for i:=0; i<MAX_FETCH_PER_TICK; i++ {
		select {
		case pmsg = <- fs.recv_chan:
			// next
		default:
			//empty
			return false
		}

		//handle msg
		switch(pmsg.msg_type) {
		case FILE_MSG_EXIT: //exit msg
			log.Info("%s detect exit flag! will exit...", _func_)
			return true
		case FILE_MSG_UPLOAD_CHECK_FAIL: //check fail
			log.Info("%s check fail! uid:%d grp_id:%d url:%s" , _func_ , pmsg.uid , pmsg.grp_id , pmsg.url)
			fs.remove_file(pmsg.uid , pmsg.grp_id , pmsg.url)
		default:
		    //nothing
		}

	}

	return false
}


func (fs *FileServer) index_handler(w http.ResponseWriter , r *http.Request) {
	fmt.Fprintf(w , "index!")
}

func (fs *FileServer) static_handler(w http.ResponseWriter , r *http.Request) {
	//this code will shield web files directory if in debug mode ,could comment it
	path := strings.Trim(r.URL.Path , "/")
	res := strings.Split(path , "/") //dir most 3levels
	fmt.Printf("len res:%d res:%v\n" , len(res) , res)
	if len(res) <= 3 {
		fmt.Printf("dir not allowed!\n")
		http.NotFound(w , r)
		return
	}
	//end here

	//real handle
	http.StripPrefix("/static" , http_fs).ServeHTTP(w , r)
}

func (fs *FileServer) upload_handler(w http.ResponseWriter , r *http.Request) {
	fmt.Printf("upload handler\n")
	if r.Method == "GET" {
		fs.upload_handle_get(w , r)
		return
	}

	if r.Method == "POST" {
		fs.upload_handle_post(w , r)
		return
	}

}

func (fs *FileServer) upload_handle_get(w http.ResponseWriter , r *http.Request) {
	var _func_ = "<upload_handle_get>"
	log := fs.pconfig.Comm.Log

	//template
	tmpl, err := template.ParseFiles(UPLOAD_TMPL)
	if err != nil {
		log.Err("%s parse template %s failed! err:%v", _func_ , UPLOAD_TMPL, err)
		fmt.Fprintf(w, "parse error!")
		return
	}

	//output
	tmpl.Execute(w, nil)
}

func convert_upload_result(result int , info string , tmp_id int64) string {
	var res UploadResult
	res.Result = result
	res.Info = info
	res.TmpId = tmp_id
	enc , err := json.Marshal(&res)
	if err != nil {
		return ""
	}
	return string(enc)
}


func (fs *FileServer) upload_handle_post(w http.ResponseWriter , r *http.Request) {
	var _func_ = "<upload_handle_post>"
	log := fs.pconfig.Comm.Log

	defer func() {
	    if err := recover(); err != nil {
	    	log.Fatal("%s recovering from panic! err:%v" , _func_ , err)
		}
	}()


	//uid
	a_uid := r.PostFormValue(FORM_LABEL_UID)
	if a_uid == "" {
		log.Err("%s uid empty!" , _func_)
		fmt.Fprintf(w, convert_upload_result(UPLOAD_RESULT_FAILED , "arg empty" , 0))
		return
	}

	//grp_id
	a_grp_id := r.PostFormValue(FORM_LABEL_GRP_ID)
	if a_grp_id == "" {
		log.Err("%s grp_id empty! uid:%s" , _func_ , a_uid)
		fmt.Fprintf(w, convert_upload_result(UPLOAD_RESULT_FAILED , "arg empty" , 0))
		return
	}

	//tmp_id
	a_tmp_id := r.PostFormValue(FORM_LABEL_TMP_ID)
	if a_tmp_id == "" {
		log.Err("%s tmp_id empty! uid:%s" , _func_ , a_tmp_id)
		fmt.Fprintf(w, convert_upload_result(UPLOAD_RESULT_FAILED , "arg empty" , 0))
		return
	}

	uid , err := strconv.ParseInt(a_uid , 10 ,64)
	if err != nil {
		log.Err("%s convert uid failed! uid:%s err:%v" , _func_ , a_uid , err)
		fmt.Fprintf(w, convert_upload_result(UPLOAD_RESULT_FAILED , "arg error" , 0))
		return
	}

	grp_id , err := strconv.ParseInt(a_grp_id , 10 ,64)
	if err != nil {
		log.Err("%s convert grp_id failed! grp_id:%s err:%v" , _func_ , a_grp_id , err)
		fmt.Fprintf(w, convert_upload_result(UPLOAD_RESULT_FAILED , "arg error" , 0))
		return
	}

	tmp_id , err := strconv.ParseInt(a_tmp_id , 10 ,64)
	if err != nil {
		log.Err("%s convert grp_id failed! tmp_id:%s err:%v" , _func_ , a_tmp_id , err)
		fmt.Fprintf(w, convert_upload_result(UPLOAD_RESULT_FAILED , "arg error" , 0))
		return
	}

	log.Debug("%s will upload file! uid:%d grp_id:%d" , _func_ , uid , grp_id)
	//file
	file , _ , err := r.FormFile(FORM_LABEL_UPLOAD)
	if err != nil {
		log.Err("%s form file faile! err:%v uid:%d" , _func_ , err , uid)
		fmt.Fprintf(w, convert_upload_result(UPLOAD_RESULT_FAILED , "file error" , tmp_id))
		return
	}
	defer file.Close()

	//check size
	file_bytes , err := ioutil.ReadAll(file)
	if err != nil {
		log.Err("%s read file failed! err:%v uid:%d" , _func_ , err , uid)
		fmt.Fprintf(w, convert_upload_result(UPLOAD_RESULT_FAILED , "file error" , tmp_id))
		return
	}
	if len(file_bytes) > fs.file_size {
		log.Err("%s file too large! %d:%d uid:%d" , _func_ , len(file_bytes) , fs.file_size , uid)
		fmt.Fprintf(w, convert_upload_result(UPLOAD_RESULT_SIZE , "file error" , tmp_id))
		return
	}

	//check type
	file_type := http.DetectContentType(file_bytes)

	//create dir FILE_PATH/GROUP_ID/YYYYMM/
	curr_ts := time.Now().Unix()
	year , month , _ := time.Unix(curr_ts , 0).Date()
	file_dir := fmt.Sprintf("%s/%d/%4d%02d/" , fs.file_path , grp_id , year , month)
	err = os.MkdirAll(file_dir , 0766)
	if err != nil {
		log.Err("%s mkdir %s failed! uid:%d err:%v" , _func_ , file_dir , uid , err)
		fmt.Fprintf(w, convert_upload_result(UPLOAD_RESULT_SIZE , "sys error" , tmp_id))
		return
	}

	//file_name:YYYYMM_MD5.TYPE
	md5_str := comm.EncMd5Bytes(file_bytes)

	file_endings , err := mime.ExtensionsByType(file_type)
	if err != nil {
		log.Err("%s extension failed! uid:%d err:%v" , _func_ , uid , err)
		fmt.Fprintf(w, convert_upload_result(UPLOAD_RESULT_SIZE , "file type error" , tmp_id))
		return
	}
	file_name := fmt.Sprintf("%4d%02d_%s_%s" , year , month , md5_str , file_endings[0])

	//create file
	new_path := filepath.Join(file_dir , file_name)
	log.Debug("%s will locate on:%s uid:%d" , _func_ , new_path , uid)

	//check exist
    if comm.FileExist(new_path) {
    	log.Debug("%s file:%s is exist already!" , _func_ , new_path)
	} else {
		//write
		new_file, err := os.Create(new_path)
		if err != nil {
			log.Err("%s create %s failed! err:%v uid:%d", _func_, new_path, err, uid)
			fmt.Fprintf(w, convert_upload_result(UPLOAD_RESULT_SIZE, "sys error" , tmp_id))
			return
		}
		defer new_file.Close()

		_, err = new_file.Write(file_bytes)
		if err != nil {
			log.Err("%s write new file:%s failed! err:%v uid:%d", _func_, new_path, err, uid)
			fmt.Fprintf(w, convert_upload_result(UPLOAD_RESULT_SIZE, "sys error" , tmp_id))
			return
		}
	}

	file_url := fmt.Sprintf("%d:%d:%s" , grp_id , fs.serv_index , file_name)
	log.Info("%s upload success! uid:%d grp_id:%d file_url:%s" , _func_ , uid , grp_id , file_url)
	fmt.Fprintf(w , convert_upload_result(UPLOAD_RESULT_SUCCESS , "done" , tmp_id))

	//notify
	if len(fs.snd_chan) >= cap(fs.snd_chan) {
		log.Err("%s snd channel full! will not check file! uid:%d grp_id:%d url:%s" , _func_ , uid , grp_id , file_url)
		return
	}
	pmsg := new(FileMsg)
	pmsg.msg_type = FILE_MSG_UPLOAD
	pmsg.uid = uid
	pmsg.grp_id = grp_id
	pmsg.url = file_url
	pmsg.int_v = tmp_id
    fs.snd_chan <- pmsg
}

//remove file
//url> grp:index:YYYYMM_MD5_.type    exam: 5024:1:202010_540f35fe0a34e1f8deba6b6692315339_.jpg
func (fs *FileServer) remove_file(uid int64 , grp_id int64 , url string) {
	var _func_ = "<FileServer.remove_file>"
	log := fs.pconfig.Comm.Log

	log.Info("%s uid:%d grp_id:%d uri:%s" , _func_ , uid , grp_id , url)
    //fetch YYYYMM
    uri_strs := strings.Split(url , ":")
    if len(uri_strs) != 3 {
    	log.Err("%s illegal uri:%s uid:%d" , _func_ , url , grp_id)
    	return
	}
	file_name := uri_strs[2]
	file_strs := strings.Split(file_name , "_")
	yymm := file_strs[0]
	if yymm == "" {
		log.Err("%s illegal file name:%s uid:%d" , _func_ , file_name , uid)
		return
	}

    //file path
    file_path := fmt.Sprintf("%s/%d/%s/%s" , fs.file_path , grp_id , yymm , file_name)

    //del it
    err := os.Remove(file_path)
    if err != nil {
		log.Err("%s remove file %s failed! uid:%d err:%v" , _func_ , file_path , uid , err)
		return
	}

	log.Info("%s remove file %s success! uid:%d" , _func_ , file_path , uid)
}