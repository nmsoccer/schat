package lib

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"schat/servers/comm"
	"strconv"
	"strings"
	"sync"
	"time"
)

/*
//FILE_URL_TYPE
REFER comm.FILE_URL_T_xx
CHAT_FILE URL> 1:index:grp_id:file_name  | FILE_DIR> CHAT_PARENT_PATH/GROUP_ID/YYYYMM/ | FILE_NAME> YYYYMM_MD5.TYPE
HEAD_FILE URL> 2:index:sub_dir:file_name | FILE_DIR> HEAD_PARENT_PATH/SUB_DIR/UID/ | FILE_NAME> UID_MD5.TYPE
GROUP_FILE URL> 3:index:sub_dir:file_name | FILE_DIR> GROUP_HEAD_PARENT_PATH/SUB_DIR/GRPID/ | FILE_NAME> GRPID_MD5.TYPE
*/

const (
	UPLOAD_TMPL = "./html_tmpl/upload.html"

	//UPLOAD LABEL
	FORM_LABEL_URL_TYPE = "url_type"
	FORM_LABEL_UID      = "uid"
	FORM_LABEL_GRP_ID   = "grp_id"
	FORM_LABEL_TMP_ID   = "tmp_id"
	FORM_LABEL_UPLOAD   = "upload_file"
	FORM_LABEL_TOKEN    = "token"

	//UPLOAD RESULT
	UPLOAD_RESULT_SUCCESS = 0
	UPLOAD_RESULT_FAILED  = 1
	UPLOAD_RESULT_SIZE    = 2
	UPLOAD_RESULT_TYPE    = 3

	FILE_MSG_CHAN_SIZE = 1000
	MAX_FETCH_PER_TICK = 10

	//SUB_DIR --> type
	FILE_PARENT_DIR_CHATS = "chat"
	FILE_PARENT_DIR_HEADS = "head"
	FILE_PARENT_DIR_GROUP_HEADS = "g_head"

	//URL PATTERN
	URL_HEAD_CHATS = "/static"
	URL_HEAD_HEADS = "/head"
	URL_HEAD_G_HEADS = "/g_head"

	//FILE_MSG_TYPE
	FILE_MSG_EXIT              = 0 //exit
	FILE_MSG_UPLOAD            = 1 //server-->main upload one file
	FILE_MSG_UPLOAD_CHECK_FAIL = 2 //main-->server upload check err
	FILE_MSG_UPDATE_TOKEN      = 3 //update token

	//MAX_SUB_DIR
	MAX_HEAD_SUB_DIRS = 64 //uid % dirs
	MAX_GROUP_HEAD_SUB_DIRS = 61 //grp_id % dirs

    //HTTPS
    KEY_FILE_PATH = "./cfg/key.pem"
    CERT_FILE_PATH = "./cfg/cert.pem"
)

type FileMsg struct {
	msg_type int //refer FILE_MSG_XX
	uid      int64
	grp_id   int64
	url      string
	int_v    int64
	str_v    string
	file_type int //refer FILE_UPDATE_FILE
}

type suspect_item struct {
	ip string
	update_ts int64
	failed_times int32
}

type safe_watcher struct {
	sync.RWMutex
	suspect_map map[string]*suspect_item
}



type FileServer struct {
	//sync.Mutex
	sync.RWMutex
	pconfig *Config
	//config
	serv_index       int
	http_addr        string
	file_size        int
	parent_path      string
	chat_parent_path string
	head_parent_path string
	group_head_parent_path string
	//channel
	snd_chan  chan *FileMsg
	recv_chan chan *FileMsg
	//token
	last_token string
	curr_token string
	//watcher
	watcher safe_watcher
	last_check_watcher int64
}

type UploadResult struct {
	Result int    `json:"result"`
	Info   string `json:"info"`
	TmpId  int64  `json:"tmp_id"`
}

var chat_fs http.Handler //chat file server
var head_fs http.Handler //head file server
var group_head_fs http.Handler //group head file server

type AccMIME struct {
	sync.RWMutex
	inner_map map[string]int
}

func (pa *AccMIME) Accept(mime_type string) (int , bool) {
	pa.RLock()
	defer pa.RUnlock()

	file_type , ok := pa.inner_map[mime_type]
	if !ok {
		return -1 , false
	}
	return file_type , true
}
var accept_mime *AccMIME

type MiMeRefer struct {
	sync.RWMutex
	refer_map map[string]string
}
func (pm *MiMeRefer) GetEnding(file_type string) (string , bool){
	pm.RLock()
	defer pm.RUnlock()

	ending , ok := pm.refer_map[file_type]
	if !ok {
		return "" , false
	} else {
		return ending , true
	}
}
var local_mime *MiMeRefer


func init() {
	//mime refer
	local_mime = new(MiMeRefer)
	refer_map := make(map[string]string)
	refer_map["video/mp4"] = ".mp4"
	local_mime.refer_map = refer_map

	//accept mime
	accept_mime = new(AccMIME)
	inner_map := make(map[string] int)
	inner_map[".jpg"] = comm.FILE_TYPE_IMAGE
	inner_map[".jpeg"] = comm.FILE_TYPE_IMAGE
	inner_map[".png"] = comm.FILE_TYPE_IMAGE
	inner_map[".mp4"] = comm.FILE_TYPE_MP4
	//inner_map["mp4"] = true
	accept_mime.inner_map = inner_map
}


func StartFileServer(pconfig *Config) *FileServer {
	var _func_ = "<StartFileServer>"
	log := pconfig.Comm.Log

	//alloc
	fs := new(FileServer)
	fs.pconfig = pconfig
	fs.serv_index = pconfig.FileConfig.ServIndex
	fs.http_addr = pconfig.FileConfig.HttpAddr
	fs.file_size = pconfig.FileConfig.MaxFileSize
	fs.parent_path = pconfig.FileConfig.RealFilePath
	fs.chat_parent_path = fmt.Sprintf("%s/%s", fs.parent_path, FILE_PARENT_DIR_CHATS)
	fs.head_parent_path = fmt.Sprintf("%s/%s", fs.parent_path, FILE_PARENT_DIR_HEADS)
	fs.group_head_parent_path = fmt.Sprintf("%s/%s" , fs.parent_path , FILE_PARENT_DIR_GROUP_HEADS)
	fs.recv_chan = make(chan *FileMsg, FILE_MSG_CHAN_SIZE)
	fs.snd_chan = make(chan *FileMsg, FILE_MSG_CHAN_SIZE)
	fs.last_token = pconfig.NowToken
	fs.curr_token = pconfig.NowToken
	fs.watcher.suspect_map = make(map[string]*suspect_item)

	//new http_fs
	chat_fs = http.FileServer(http.Dir(fs.chat_parent_path))
	if chat_fs == nil {
		log.Err("%s ChatFileServer %s failed!", _func_, fs.chat_parent_path)
		return nil
	}

	head_fs = http.FileServer(http.Dir(fs.head_parent_path))
	if head_fs == nil {
		log.Err("%s HeadFileServer %s failed!", _func_, fs.head_parent_path)
		return nil
	}

	group_head_fs = http.FileServer(http.Dir(fs.group_head_parent_path))
	if group_head_fs == nil {
		log.Err("%s GroupHeadFileServer %s failed!", _func_, fs.group_head_parent_path)
		return nil
	}

	//start
	log.Info("%s addr:%s size:%d index:%d parent_path:%s chat_dir:%s head_dir:%s group_head_dir:%s", _func_, fs.http_addr, fs.file_size, fs.serv_index,
		fs.parent_path, fs.chat_parent_path, fs.head_parent_path , fs.group_head_parent_path)
	go fs.start_serv()
	return fs
}

//upper send to FileServer
func (fs *FileServer) Send(pmsg *FileMsg) bool {
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
func (fs *FileServer) Read(msg_list []*FileMsg) int {
	//set count
	count := len(fs.snd_chan)
	if count > len(msg_list) {
		count = len(msg_list)
	}

	//read
	for i := 0; i < count; i++ {
		msg_list[i] = <-fs.snd_chan
	}
	return count
}

func (fs *FileServer) Close() {
	pmsg := new(FileMsg)
	pmsg.msg_type = FILE_MSG_EXIT
	fs.recv_chan <- pmsg
}

func (fs *FileServer) Suspect() int {
	fs.watcher.RLock()
	defer fs.watcher.RUnlock()
	return len(fs.watcher.suspect_map)
}


/*-----------------------------------static func-------------------------------*/
func (fs *FileServer) start_serv() {
	log := fs.pconfig.Comm.Log

	//http server
	go func() {
		//reg handler
		http.Handle("/", http.HandlerFunc(fs.index_handler))
		http.Handle("/upload/", http.HandlerFunc(fs.upload_handler))
		http.Handle(URL_HEAD_CHATS + "/", http.HandlerFunc(fs.static_handler))
		http.Handle(URL_HEAD_HEADS + "/", http.HandlerFunc(fs.head_handler))
		http.Handle(URL_HEAD_G_HEADS + "/", http.HandlerFunc(fs.group_head_handler))

		//err := http.ListenAndServe(fs.http_addr, nil)
		err := http.ListenAndServeTLS(fs.http_addr , CERT_FILE_PATH , KEY_FILE_PATH , nil)
		if err != nil {
			log.Err("file_server start_serv at %s failed! err:%v", fs.http_addr, err)
		}
	}()

	//main proc
	for {
		//recv msg
		fs.recv_msg()
		//check
		fs.check_suspect()
		//sleep
		time.Sleep(10 * time.Millisecond) //10ms
	}
}

//@return:exit
func (fs *FileServer) recv_msg() bool {
	var _func_ = "<FileServer.recv_msg>"
	var pmsg *FileMsg
	log := fs.pconfig.Comm.Log

	//fetch msg
	for i := 0; i < MAX_FETCH_PER_TICK; i++ {
		select {
		case pmsg = <-fs.recv_chan:
			// next
		default:
			//empty
			return false
		}

		//handle msg
		switch pmsg.msg_type {
		case FILE_MSG_EXIT: //exit msg
			log.Info("%s detect exit flag! will exit...", _func_)
			return true
		case FILE_MSG_UPLOAD_CHECK_FAIL: //check fail
			log.Info("%s check result:%d! uid:%d grp_id:%d url:%s", _func_, pmsg.int_v, pmsg.uid, pmsg.grp_id, pmsg.url)
			fs.remove_file(pmsg.uid, pmsg.grp_id, pmsg.url)
		case FILE_MSG_UPDATE_TOKEN:
			fs.update_token(pmsg.str_v)
		default:
			//nothing
			log.Err("%s unkown file_msg type:%d", _func_, pmsg.msg_type)
		}

	}

	return false
}

func (fs *FileServer) update_token(new_token string) {
	fs.Lock()
	fs.pconfig.Comm.Log.Info("update token %s --> %s", fs.curr_token, new_token)
	fs.last_token = fs.curr_token
	fs.curr_token = new_token
	fs.Unlock()
}

//per 60s
func (fs *FileServer) check_suspect() {
	var _func_ = "<check_suspect>"
	log := fs.pconfig.Comm.Log
	defer func() {
		if err := recover(); err != nil {
			log.Fatal("%s recover from panic! err:%v" , _func_ , err)
		}
	}()


	curr_ts := time.Now().Unix()
	if fs.last_check_watcher + SUSPECT_CHECK_PEROID > curr_ts {
		return
	}
	fs.last_check_watcher = curr_ts
	//log.Debug("%s curr_ts:%d len:%d" , _func_ , curr_ts , len(fs.watcher.suspect_map))
	if len(fs.watcher.suspect_map) <= 0 {
		return
	}

	fs.watcher.RLock()
	//check expire
	strs := make([]string , len(fs.watcher.suspect_map))
	count := 0
	for _ , pinfo := range fs.watcher.suspect_map {
		if pinfo.update_ts + SUSPECT_JAIL_TIME >= curr_ts {	//less than 1h
			continue
		}

		if pinfo.update_ts + int64((pinfo.failed_times-SUSPECT_SAFE_FAIL)*SUSPECT_FAIL_PUNISH_SEC) < curr_ts {
			log.Debug("%s ip:%s expired! times:%d update:%d curr:%d" , _func_ , pinfo.ip , pinfo.failed_times ,
				pinfo.update_ts , curr_ts)
			strs[count] = pinfo.ip
			count++
		}

	}
	fs.watcher.RUnlock()

	//expire
	if count <= 0 {
		return
	}

	//clear
	fs.watcher.Lock()
	for i:=0; i<count; i++ {
		delete(fs.watcher.suspect_map , strs[i])
	}
	fs.watcher.Unlock()
}


/*------------------------HTTP HANDLE--------------------------------*/
func (fs *FileServer) file_safe_check(w http.ResponseWriter, r *http.Request , dir_level int) bool {
	var _func_ = "<file_safe_check>"
	log := fs.pconfig.Comm.Log
	mod := fs.pconfig.FileConfig.SafeMode
	var ok bool
	//no check
	if mod <= SAFE_MOD_NONE {
		return true
	}

	//check path
	if mod <= SAFE_MOD_PATH{
		return  fs.safe_check_path(w , r , dir_level)
	}

	//check token
	if mod <= SAFE_MOD_TOKEN {
		ok = fs.safe_check_path(w , r , dir_level)
		if !ok {
			return false
		}
		return fs.safe_check_token(w , r)
	}

	/*------- check ip ---------*/
	//get ip
	c_ip := ClientPublicIP(r)
	if len(c_ip) == 0 {
		c_ip = ClientIP(r)
	}
	if len(c_ip) == 0 {
		log.Err("%s get client ip failed! will let it go!" , _func_)
		return true
	}

	log.Debug("%s safe_mod:%d c_ip:%s" , _func_ , mod , c_ip)
	for {
		//path
		ok = fs.safe_check_path(w , r , dir_level)
		if !ok {
			break
		}

		//token
		ok = fs.safe_check_token(w , r)
		if !ok {
			break
		}

		//ip
		return fs.safe_check_ip(c_ip)
	}

	//check fail will at here
	//update suspect
	var pitem *suspect_item
	curr_ts := time.Now().Unix()

	fs.watcher.RLock()
	pitem , ok = fs.watcher.suspect_map[c_ip]
	fs.watcher.RUnlock()

	//add suspect
	if !ok {
		//add item
		pitem = new(suspect_item)
		pitem.ip = c_ip
		fs.watcher.Lock()
		fs.watcher.suspect_map[c_ip] = pitem
		fs.watcher.Unlock()
		log.Info("%s add suspect! ip:%s" , _func_ , c_ip)
	}

	//update item
	pitem.update_ts = curr_ts
	pitem.failed_times++
	log.Info("%s update suspect! ip:%s update_ts:%d times:%d" , _func_ , c_ip , pitem.update_ts , pitem.failed_times)
	return false
}


func (fs *FileServer) safe_check_path(w http.ResponseWriter, r *http.Request , dir_level int) bool {
	var _func_ = "<safe_check_path>"
	log := fs.pconfig.Comm.Log

	//mode
	if fs.pconfig.FileConfig.SafeMode < SAFE_MOD_PATH {
		return true
	}

	//check path
	path := path.Clean(r.URL.Path)
	path = strings.Trim(path , "/")
	strs := strings.Split(path , "/") //dir most dir_level
	if len(strs) <= dir_level {
		log.Info("%s not permitted! path:%s level:%d v:%v" , _func_ , path , len(strs) , strs)
		return false
	}
	return true
}


func (fs *FileServer) safe_check_token(w http.ResponseWriter, r *http.Request) bool {
	var _func_ = "<safe_check_token>"
	log := fs.pconfig.Comm.Log

	//mod
	if fs.pconfig.FileConfig.SafeMode < SAFE_MOD_TOKEN {
		return true
	}

	//get token
	uid_v := r.FormValue(FORM_LABEL_UID)
	u_token := r.FormValue(FORM_LABEL_TOKEN)
	if len(uid_v) <=0 || len(u_token)<=0 {
		log.Err("%s uid or u_token nil!" , _func_)
		return false
	}
	uid , err := strconv.ParseInt(uid_v , 10 ,64)
	if err != nil {
		log.Err("%s convert uid failed! err:%v uid:%s" , _func_ , err , uid)
		return false
	}

	fs.RLock()
	curr_token := fs.curr_token
	last_token := fs.last_token
	fs.RUnlock()

	//compare token
	curr_real_token := comm.CalcUserToken(uid , curr_token)
	last_real_token := comm.CalcUserToken(uid , last_token)

	if u_token != curr_real_token && u_token != last_real_token {
		log.Err("%s token not match! uid:%s u_token:%s now:%s last:%s", _func_, uid, u_token, curr_real_token ,
			last_real_token)
		return false
	}

	return true
}

func (fs *FileServer) safe_check_ip(ip string) bool {
	var _func_ = "<safe_check_ip>"
	log := fs.pconfig.Comm.Log

	//if block
	fs.watcher.RLock()
	pitem , ok := fs.watcher.suspect_map[ip]
	defer fs.watcher.RUnlock()

	if !ok {
		log.Debug("%s not in suspect! ip:%s" , _func_ , ip)
		return true
	}

	if pitem.failed_times <= SUSPECT_SAFE_FAIL {
		log.Debug("%s in safe times! ip:%s times:%d" , _func_ , ip , pitem.failed_times)
		return true
	}

	curr_ts := time.Now().Unix()
	block_sec := (pitem.failed_times-SUSPECT_SAFE_FAIL)*SUSPECT_FAIL_PUNISH_SEC
	block_ts := curr_ts + int64(block_sec)
	//if in block period
	if block_ts < curr_ts {
		log.Debug("%s expire block time! ip:%s times:%d curr_ts:%d block_sec:%d block_ts:%d" , _func_ , ip , pitem.failed_times,
			curr_ts , block_sec , block_ts)
		return true
	}

	log.Debug("%s in block time! ip:%s times:%d curr_ts:%d block_sec:%d block_ts:%d" , _func_ , ip , pitem.failed_times,
		curr_ts , block_sec , block_ts)
	return false
}



func (fs *FileServer) index_handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "index!")
}

func (fs *FileServer) static_handler(w http.ResponseWriter, r *http.Request) {
	// /static/group_id/yymm/xx.jpeg
	if !fs.file_safe_check(w , r , 3) {
		http.NotFound(w , r)
		return
	}

	//real handle
	http.StripPrefix(URL_HEAD_CHATS, chat_fs).ServeHTTP(w, r)
}

func (fs *FileServer) head_handler(w http.ResponseWriter, r *http.Request) {
	// /head/group_id/sub_dir/uid/xx.jpeg
	if !fs.file_safe_check(w , r , 3) {
		http.NotFound(w , r)
		return
	}

	//real handle
	http.StripPrefix(URL_HEAD_HEADS, head_fs).ServeHTTP(w, r)
}

func (fs *FileServer) group_head_handler(w http.ResponseWriter, r *http.Request) {
	// /head/group_id/sub_dir/uid/xx.jpeg
	if !fs.file_safe_check(w , r , 3) {
		http.NotFound(w , r)
		return
	}

	//real handle
	http.StripPrefix(URL_HEAD_G_HEADS, group_head_fs).ServeHTTP(w, r)
}


func (fs *FileServer) upload_handler(w http.ResponseWriter, r *http.Request) {
	if !fs.safe_check_token(w, r) {
		http.NotFound(w, r)
		return
	}
	fmt.Printf("upload handler\n")
	if r.Method == "GET" {
		fs.upload_handle_get(w, r)
		return
	}

	if r.Method == "POST" {
		fs.upload_handle_post(w, r)
		return
	}

}

func (fs *FileServer) upload_handle_get(w http.ResponseWriter, r *http.Request) {
	var _func_ = "<upload_handle_get>"
	log := fs.pconfig.Comm.Log
	//template
	tmpl, err := template.ParseFiles(UPLOAD_TMPL)
	if err != nil {
		log.Err("%s parse template %s failed! err:%v", _func_, UPLOAD_TMPL, err)
		fmt.Fprintf(w, "parse error!")
		return
	}

	//output
	tmpl.Execute(w, nil)
}

func convert_upload_result(result int, info string, tmp_id int64) string {
	var res UploadResult
	res.Result = result
	res.Info = info
	res.TmpId = tmp_id
	enc, err := json.Marshal(&res)
	if err != nil {
		return ""
	}
	return string(enc)
}

func (fs *FileServer) upload_handle_post(w http.ResponseWriter, r *http.Request) {
	var _func_ = "<upload_handle_post>"
	log := fs.pconfig.Comm.Log
	defer func() {
		if err := recover(); err != nil {
			log.Fatal("%s recovering from panic! err:%v", _func_, err)
		}
	}()

	//url type
	a_url_type := r.PostFormValue(FORM_LABEL_URL_TYPE)
	if a_url_type == "" {
		log.Err("%s url empty!", _func_)
		fmt.Fprintf(w, convert_upload_result(UPLOAD_RESULT_FAILED, "arg empty", 0))
		return
	}

	//uid
	a_uid := r.PostFormValue(FORM_LABEL_UID)
	if a_uid == "" {
		log.Err("%s uid empty!", _func_)
		fmt.Fprintf(w, convert_upload_result(UPLOAD_RESULT_FAILED, "arg empty", 0))
		return
	}

	//grp_id
	a_grp_id := r.PostFormValue(FORM_LABEL_GRP_ID)
	if a_grp_id == "" {
		log.Err("%s grp_id empty! uid:%s", _func_, a_uid)
		fmt.Fprintf(w, convert_upload_result(UPLOAD_RESULT_FAILED, "arg empty", 0))
		return
	}

	//tmp_id
	a_tmp_id := r.PostFormValue(FORM_LABEL_TMP_ID)
	if a_tmp_id == "" {
		log.Err("%s tmp_id empty! uid:%s", _func_, a_tmp_id)
		fmt.Fprintf(w, convert_upload_result(UPLOAD_RESULT_FAILED, "arg empty", 0))
		return
	}

	url_type, err := strconv.Atoi(a_url_type)
	if err != nil {
		log.Err("%s convert url_type failed! uid:%s err:%v", _func_, a_url_type, err)
		fmt.Fprintf(w, convert_upload_result(UPLOAD_RESULT_FAILED, "arg error", 0))
		return
	}

	uid, err := strconv.ParseInt(a_uid, 10, 64)
	if err != nil {
		log.Err("%s convert uid failed! uid:%s err:%v", _func_, a_uid, err)
		fmt.Fprintf(w, convert_upload_result(UPLOAD_RESULT_FAILED, "arg error", 0))
		return
	}

	grp_id, err := strconv.ParseInt(a_grp_id, 10, 64)
	if err != nil {
		log.Err("%s convert grp_id failed! grp_id:%s err:%v", _func_, a_grp_id, err)
		fmt.Fprintf(w, convert_upload_result(UPLOAD_RESULT_FAILED, "arg error", 0))
		return
	}

	tmp_id, err := strconv.ParseInt(a_tmp_id, 10, 64)
	if err != nil {
		log.Err("%s convert grp_id failed! tmp_id:%s err:%v", _func_, a_tmp_id, err)
		fmt.Fprintf(w, convert_upload_result(UPLOAD_RESULT_FAILED, "arg error", 0))
		return
	}

	switch url_type {
	case comm.FILE_URL_T_CHAT:
		fs.upload_chat_file(w, r, uid, grp_id, tmp_id)
	case comm.FILE_URL_T_HEAD:
		fs.upload_head_file(w, r, uid, 0, tmp_id)
	case comm.FILE_URL_T_GROUP:
		fs.upload_group_head_file(w , r , uid , grp_id , tmp_id)
	default:
		log.Err("%s url_type:%d illegal! uid:%d", _func_, url_type, uid)
		fmt.Fprintf(w, convert_upload_result(UPLOAD_RESULT_FAILED, "arg error", 0))
		return
	}

}

/*-----------------------------------DIFF URL TYPE--------------------------------
#
#
#
#
#
*/
//upload chat file
func (fs *FileServer) upload_chat_file(w http.ResponseWriter, r *http.Request, uid int64, grp_id int64, tmp_id int64) {
	var _func_ = "<FileServer.upload_chat_file>"
	log := fs.pconfig.Comm.Log
	defer func() {
		if err := recover(); err != nil {
			log.Fatal("%s recover from panic! uid:%d err:%v", _func_, uid , err)
		}
	}()

	log.Debug("%s will upload file! uid:%d grp_id:%d tmp_id:%d", _func_, uid, grp_id, tmp_id)
	//file
	file, _, err := r.FormFile(FORM_LABEL_UPLOAD)
	if err != nil {
		log.Err("%s form file faile! err:%v uid:%d", _func_, err, uid)
		fmt.Fprintf(w, convert_upload_result(UPLOAD_RESULT_FAILED, "file error", tmp_id))
		return
	}
	defer file.Close()

	//check size
	file_bytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Err("%s read file failed! err:%v uid:%d", _func_, err, uid)
		fmt.Fprintf(w, convert_upload_result(UPLOAD_RESULT_FAILED, "file error", tmp_id))
		return
	}
	if len(file_bytes) > fs.file_size {
		log.Err("%s file too large! %d:%d uid:%d", _func_, len(file_bytes), fs.file_size, uid)
		fmt.Fprintf(w, convert_upload_result(UPLOAD_RESULT_SIZE, "file error", tmp_id))
		return
	}


	//check type
	file_type := http.DetectContentType(file_bytes)

	//create dir FILE_PATH/GROUP_ID/YYYYMM/
	curr_ts := time.Now().Unix()
	year, month, _ := time.Unix(curr_ts, 0).Date()
	file_dir := fmt.Sprintf("%s/%d/%4d%02d/", fs.chat_parent_path, grp_id, year, month)
	err = os.MkdirAll(file_dir, 0766)
	if err != nil {
		log.Err("%s mkdir %s failed! uid:%d err:%v", _func_, file_dir, uid, err)
		fmt.Fprintf(w, convert_upload_result(UPLOAD_RESULT_SIZE, "sys error", tmp_id))
		return
	}

	//file_name:YYYYMM_MD5.TYPE
	md5_str := comm.EncMd5Bytes(file_bytes)

	file_endings, err := mime.ExtensionsByType(file_type)
	if err != nil {
		log.Err("%s extension failed! uid:%d err:%v", _func_, uid, err)
		fmt.Fprintf(w, convert_upload_result(UPLOAD_RESULT_SIZE, "file size error", tmp_id))
		return
	}
	if len(file_endings) <= 0 { //try inner mime type
		file_end , ok := local_mime.GetEnding(file_type)
		if !ok {
			log.Err("%s file_ending error! uid:%d ending:%v file_type:%s file_end:%v", _func_, uid, file_endings, file_type, file_end)
			fmt.Fprintf(w, convert_upload_result(UPLOAD_RESULT_TYPE, "file type error", tmp_id))
			return
		}

		file_endings = append(file_endings , file_end)
		log.Debug("%s inner mime refer matched! uid:%d file_type:%s file_endings:%v" , _func_ , uid , file_type , file_endings)
	}

	//check mime
	file_update_type , ok := accept_mime.Accept(file_endings[0])
	if ok == false {
		log.Err("%s mime not accept! uid:%d mime_type:%v", _func_, uid, file_endings[0])
		fmt.Fprintf(w, convert_upload_result(UPLOAD_RESULT_TYPE, "file type error", tmp_id))
		return
	}

	file_name := fmt.Sprintf("%4d%02d_%s_%s", year, month, md5_str, file_endings[0])
	//create file
	new_path := filepath.Join(file_dir, file_name)
	log.Debug("%s will locate on:%s uid:%d", _func_, new_path, uid)

	//check exist
	if comm.FileExist(new_path) {
		log.Debug("%s file:%s is exist already!", _func_, new_path)
	} else {
		//write
		new_file, err := os.Create(new_path)
		if err != nil {
			log.Err("%s create %s failed! err:%v uid:%d", _func_, new_path, err, uid)
			fmt.Fprintf(w, convert_upload_result(UPLOAD_RESULT_SIZE, "sys error", tmp_id))
			return
		}
		defer new_file.Close()

		_, err = new_file.Write(file_bytes)
		if err != nil {
			log.Err("%s write new file:%s failed! err:%v uid:%d", _func_, new_path, err, uid)
			fmt.Fprintf(w, convert_upload_result(UPLOAD_RESULT_SIZE, "sys error", tmp_id))
			return
		}
	}

	//FILE_URL: 1:index:group_id:file_name
	file_url := fmt.Sprintf("%d:%d:%d:%s", comm.FILE_URL_T_CHAT, fs.serv_index, grp_id, file_name)
	log.Info("%s upload success! uid:%d grp_id:%d file_url:%s", _func_, uid, grp_id, file_url)
	fmt.Fprintf(w, convert_upload_result(UPLOAD_RESULT_SUCCESS, "done", tmp_id))

	//notify
	if len(fs.snd_chan) >= cap(fs.snd_chan) {
		log.Err("%s snd channel full! will not check file! uid:%d grp_id:%d url:%s", _func_, uid, grp_id, file_url)
		return
	}
	pmsg := new(FileMsg)
	pmsg.msg_type = FILE_MSG_UPLOAD
	pmsg.uid = uid
	pmsg.grp_id = grp_id
	pmsg.url = file_url
	pmsg.int_v = tmp_id
	pmsg.file_type = file_update_type
	fs.snd_chan <- pmsg
}

//upload head file url> 2:index:sub_dir:file_name
func (fs *FileServer) upload_head_file(w http.ResponseWriter, r *http.Request, uid int64, grp_id int64, tmp_id int64) {
	var _func_ = "<FileServer.upload_head_file>"
	log := fs.pconfig.Comm.Log
	defer func() {
		if err := recover(); err != nil {
			log.Fatal("%s recover from panic! uid:%d", _func_, uid)
		}
	}()

	log.Debug("%s will upload file! uid:%d grp_id:%d tmp_id:%d", _func_, uid, grp_id, tmp_id)
	//file
	file, _, err := r.FormFile(FORM_LABEL_UPLOAD)
	if err != nil {
		log.Err("%s form file faile! err:%v uid:%d", _func_, err, uid)
		fmt.Fprintf(w, convert_upload_result(UPLOAD_RESULT_FAILED, "file error", tmp_id))
		return
	}
	defer file.Close()

	//check size
	file_bytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Err("%s read file failed! err:%v uid:%d", _func_, err, uid)
		fmt.Fprintf(w, convert_upload_result(UPLOAD_RESULT_FAILED, "file error", tmp_id))
		return
	}
	if len(file_bytes) > fs.file_size {
		log.Err("%s file too large! %d:%d uid:%d", _func_, len(file_bytes), fs.file_size, uid)
		fmt.Fprintf(w, convert_upload_result(UPLOAD_RESULT_SIZE, "file error", tmp_id))
		return
	}

	//check type
	file_type := http.DetectContentType(file_bytes)

	//create dir FILE_PATH/SUB_DIR/UID/
	sub_dir := uid % MAX_HEAD_SUB_DIRS
	file_dir := fmt.Sprintf("%s/%d/%d/", fs.head_parent_path, sub_dir, uid)
	err = os.MkdirAll(file_dir, 0766)
	if err != nil {
		log.Err("%s mkdir %s failed! uid:%d err:%v", _func_, file_dir, uid, err)
		fmt.Fprintf(w, convert_upload_result(UPLOAD_RESULT_SIZE, "sys error", tmp_id))
		return
	}

	//file_name:UID_MD5.TYPE
	md5_str := comm.EncMd5Bytes(file_bytes)

	file_endings, err := mime.ExtensionsByType(file_type)
	if err != nil {
		log.Err("%s extension failed! uid:%d err:%v", _func_, uid, err)
		fmt.Fprintf(w, convert_upload_result(UPLOAD_RESULT_SIZE, "file type error", tmp_id))
		return
	}
	//check mime
	file_update_type , ok := accept_mime.Accept(file_endings[0])
	if ok == false {
		log.Err("%s mime not accept! uid:%d mime_type:%v", _func_, uid, file_endings[0])
		fmt.Fprintf(w, convert_upload_result(UPLOAD_RESULT_TYPE, "file type error", tmp_id))
		return
	}
	//only for image
	if file_update_type != comm.FILE_TYPE_IMAGE {
		log.Err("%s only for image! uid:%d mime_type:%v", _func_, uid, file_endings[0])
		fmt.Fprintf(w, convert_upload_result(UPLOAD_RESULT_TYPE, "file type error", tmp_id))
		return
	}


	file_name := fmt.Sprintf("%d_%s_%s", uid, md5_str, file_endings[0])

	//create file
	new_path := filepath.Join(file_dir, file_name)
	log.Debug("%s will locate on:%s uid:%d", _func_, new_path, uid)

	//check exist
	if comm.FileExist(new_path) {
		log.Debug("%s file:%s is exist already!", _func_, new_path)
	} else {
		//write
		new_file, err := os.Create(new_path)
		if err != nil {
			log.Err("%s create %s failed! err:%v uid:%d", _func_, new_path, err, uid)
			fmt.Fprintf(w, convert_upload_result(UPLOAD_RESULT_SIZE, "sys error", tmp_id))
			return
		}
		defer new_file.Close()

		_, err = new_file.Write(file_bytes)
		if err != nil {
			log.Err("%s write new file:%s failed! err:%v uid:%d", _func_, new_path, err, uid)
			fmt.Fprintf(w, convert_upload_result(UPLOAD_RESULT_SIZE, "sys error", tmp_id))
			return
		}
	}

	//FILE_URL: 2:index:sub_dir:file_name
	file_url := fmt.Sprintf("%d:%d:%d:%s", comm.FILE_URL_T_HEAD, fs.serv_index, sub_dir, file_name)
	log.Info("%s upload success! uid:%d sub_dir:%d file_url:%s", _func_, uid, sub_dir, file_url)
	fmt.Fprintf(w, convert_upload_result(UPLOAD_RESULT_SUCCESS, "done", tmp_id))

	//notify
	if len(fs.snd_chan) >= cap(fs.snd_chan) {
		log.Err("%s snd channel full! will not check file! uid:%d grp_id:%d url:%s", _func_, uid, grp_id, file_url)
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


//upload group_head file url> 3:index:sub_dir:file_name
func (fs *FileServer) upload_group_head_file(w http.ResponseWriter, r *http.Request, uid int64, grp_id int64, tmp_id int64) {
	var _func_ = "<FileServer.upload_group_head_file>"
	log := fs.pconfig.Comm.Log
	defer func() {
		if err := recover(); err != nil {
			log.Fatal("%s recover from panic! uid:%d", _func_, uid)
		}
	}()

	log.Debug("%s will upload file! uid:%d grp_id:%d tmp_id:%d", _func_, uid, grp_id, tmp_id)
	//file
	file, _, err := r.FormFile(FORM_LABEL_UPLOAD)
	if err != nil {
		log.Err("%s form file faile! err:%v uid:%d grp_id:%d", _func_, err, uid , grp_id)
		fmt.Fprintf(w, convert_upload_result(UPLOAD_RESULT_FAILED, "file error", tmp_id))
		return
	}
	defer file.Close()

	//check size
	file_bytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Err("%s read file failed! err:%v uid:%d", _func_, err, uid)
		fmt.Fprintf(w, convert_upload_result(UPLOAD_RESULT_FAILED, "file error", tmp_id))
		return
	}
	if len(file_bytes) > fs.file_size {
		log.Err("%s file too large! %d:%d uid:%d", _func_, len(file_bytes), fs.file_size, uid)
		fmt.Fprintf(w, convert_upload_result(UPLOAD_RESULT_SIZE, "file error", tmp_id))
		return
	}

	//check type
	file_type := http.DetectContentType(file_bytes)

	//create dir FILE_PATH/SUB_DIR/UID/
	sub_dir := grp_id % MAX_GROUP_HEAD_SUB_DIRS
	file_dir := fmt.Sprintf("%s/%d/%d/", fs.group_head_parent_path, sub_dir, grp_id)
	err = os.MkdirAll(file_dir, 0766)
	if err != nil {
		log.Err("%s mkdir %s failed! uid:%d err:%v", _func_, file_dir, uid, err)
		fmt.Fprintf(w, convert_upload_result(UPLOAD_RESULT_SIZE, "sys error", tmp_id))
		return
	}

	//file_name:GRPID_MD5.TYPE
	md5_str := comm.EncMd5Bytes(file_bytes)

	file_endings, err := mime.ExtensionsByType(file_type)
	if err != nil {
		log.Err("%s extension failed! uid:%d err:%v", _func_, uid, err)
		fmt.Fprintf(w, convert_upload_result(UPLOAD_RESULT_SIZE, "file type error", tmp_id))
		return
	}
	//check mime
	file_update_type , ok := accept_mime.Accept(file_endings[0])
	if ok == false {
		log.Err("%s mime not accept! uid:%d mime_type:%v", _func_, uid, file_endings[0])
		fmt.Fprintf(w, convert_upload_result(UPLOAD_RESULT_TYPE, "file type error", tmp_id))
		return
	}
	//only for image
	if file_update_type != comm.FILE_TYPE_IMAGE {
		log.Err("%s only for image! uid:%d mime_type:%v", _func_, uid, file_endings[0])
		fmt.Fprintf(w, convert_upload_result(UPLOAD_RESULT_TYPE, "file type error", tmp_id))
		return
	}

	file_name := fmt.Sprintf("%d_%s_%s", grp_id, md5_str, file_endings[0])
	//create file
	new_path := filepath.Join(file_dir, file_name)
	log.Debug("%s will locate on:%s uid:%d grp_id:%d", _func_, new_path, uid , grp_id)

	//check exist
	if comm.FileExist(new_path) {
		log.Debug("%s file:%s is exist already!", _func_, new_path)
	} else {
		//write
		new_file, err := os.Create(new_path)
		if err != nil {
			log.Err("%s create %s failed! err:%v uid:%d", _func_, new_path, err, uid)
			fmt.Fprintf(w, convert_upload_result(UPLOAD_RESULT_SIZE, "sys error", tmp_id))
			return
		}
		defer new_file.Close()

		_, err = new_file.Write(file_bytes)
		if err != nil {
			log.Err("%s write new file:%s failed! err:%v uid:%d", _func_, new_path, err, uid)
			fmt.Fprintf(w, convert_upload_result(UPLOAD_RESULT_SIZE, "sys error", tmp_id))
			return
		}
	}

	//FILE_URL: 3:index:sub_dir:file_name
	file_url := fmt.Sprintf("%d:%d:%d:%s", comm.FILE_URL_T_GROUP, fs.serv_index, sub_dir, file_name)
	log.Info("%s upload success! uid:%d sub_dir:%d file_url:%s grp_id:%d", _func_, uid, sub_dir, file_url , grp_id)
	fmt.Fprintf(w, convert_upload_result(UPLOAD_RESULT_SUCCESS, "done", tmp_id))

	//notify
	if len(fs.snd_chan) >= cap(fs.snd_chan) {
		log.Err("%s snd channel full! will not check file! uid:%d grp_id:%d url:%s", _func_, uid, grp_id, file_url)
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


//url> url_type:xxx:xxx...
func (fs *FileServer) remove_file(uid int64, grp_id int64, url string) {
	var _func_ = "<FileServer.remove_file>"
	log := fs.pconfig.Comm.Log

	//get type
	url_type, err := comm.GetUrlType(url)
	if err != nil {
		log.Err("%s parse url_type failed! url:%s uid:%d err:%v", _func_, url, uid, err)
		return
	}

	//handle
	switch url_type {
	case comm.FILE_URL_T_CHAT:
		fs.remove_chat_file(uid, grp_id, url)
	case comm.FILE_URL_T_HEAD:
		fs.remove_head_file(uid, 0, url)
	case comm.FILE_URL_T_GROUP:
		fs.remove_group_head_file(uid , grp_id , url)
	default:
		log.Err("%s unhandled url_type:%d url:%s uid:%d", _func_, url_type, url, uid)
	}

}

//remove chat file
//URL> 1:index:grp_id:file_name  | FILE_DIR> CHAT_PARENT_PATH/GROUP_ID/YYYYMM/ | FILE_NAME> YYYYMM_MD5.TYPE
//exam: 1:1:5024:202010_fb6164acd88582da34857c5f1ffe07b9_.jpg
func (fs *FileServer) remove_chat_file(uid int64, grp_id int64, url string) {
	var _func_ = "<FileServer.remove_chat_file>"
	log := fs.pconfig.Comm.Log

	log.Info("%s uid:%d grp_id:%d uri:%s", _func_, uid, grp_id, url)
	//fetch YYYYMM
	uri_strs := strings.Split(url, ":")
	if len(uri_strs) != 4 {
		log.Err("%s illegal uri:%s uid:%d", _func_, url, grp_id)
		return
	}
	file_name := uri_strs[3]
	file_strs := strings.Split(file_name, "_")
	yymm := file_strs[0]
	if yymm == "" {
		log.Err("%s illegal file name:%s uid:%d", _func_, file_name, uid)
		return
	}

	//file path
	file_path := fmt.Sprintf("%s/%d/%s/%s", fs.chat_parent_path, grp_id, yymm, file_name)

	//del it
	err := os.Remove(file_path)
	if err != nil {
		log.Err("%s remove file %s failed! uid:%d err:%v", _func_, file_path, uid, err)
		return
	}

	log.Info("%s remove file %s success! uid:%d", _func_, file_path, uid)
}

//remove head file
//URL> 2:index:sub_dir:file_name | FILE_DIR> HEAD_PARENT_PATH/SUB_DIR/UID/ | FILE_NAME> UID_MD5.TYPE
//exam: 2:2:20:10004_f7d02aeae4c2e4faec47195cf7667608_.jpeg
func (fs *FileServer) remove_head_file(uid int64, grp_id int64, url string) {
	var _func_ = "<FileServer.remove_head_file>"
	log := fs.pconfig.Comm.Log

	log.Info("%s uid:%d grp_id:%d uri:%s", _func_, uid, grp_id, url)
	//fetch sub_dir and file_name
	uri_strs := strings.Split(url, ":")
	if len(uri_strs) != 4 {
		log.Err("%s illegal uri:%s uid:%d", _func_, url, uid)
		return
	}
	sub_dir := uri_strs[2]
	file_name := uri_strs[3]

	//file path
	file_path := fmt.Sprintf("%s/%s/%d/%s", fs.head_parent_path, sub_dir, uid, file_name)

	//del it
	err := os.Remove(file_path)
	if err != nil {
		log.Err("%s remove file %s failed! uid:%d err:%v", _func_, file_path, uid, err)
		return
	}

	log.Info("%s remove file %s success! uid:%d", _func_, file_path, uid)
}

//remove g_head file
//URL> 3:index:sub_dir:file_name | FILE_DIR> GROUP_HEAD_PARENT_PATH/SUB_DIR/GROUP_ID/ | FILE_NAME> GRPID_MD5.TYPE
//exam: 3:1:24:5026_e9095dc1eb1480aa045f0cc8048f2837_.jpeg
func (fs *FileServer) remove_group_head_file(uid int64, grp_id int64, url string) {
	var _func_ = "<FileServer.remove_group_head_file>"
	log := fs.pconfig.Comm.Log

	log.Info("%s uid:%d grp_id:%d uri:%s", _func_, uid, grp_id, url)
	//fetch sub_dir and file_name
	uri_strs := strings.Split(url, ":")
	if len(uri_strs) != 4 {
		log.Err("%s illegal uri:%s uid:%d", _func_, url, uid)
		return
	}
	sub_dir := uri_strs[2]
	file_name := uri_strs[3]

	//file path
	file_path := fmt.Sprintf("%s/%s/%d/%s", fs.group_head_parent_path, sub_dir, grp_id, file_name)

	//del it
	err := os.Remove(file_path)
	if err != nil {
		log.Err("%s remove file %s failed! uid:%d grp_id:%d err:%v", _func_, file_path, uid, grp_id , err)
		return
	}

	log.Info("%s remove file %s success! uid:%d grp_id:%d", _func_, file_path, uid , grp_id)
}