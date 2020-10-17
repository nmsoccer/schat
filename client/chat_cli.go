package main

import (
	"bytes"
	"compress/zlib"
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	lnet "schat/lib/net"
	"schat/proto/cs"
	"strconv"
	"strings"
	"time"
)

const (
	TIME_EPOLL_BASE  int64 = 1577808000 //2020-01-01 00:00:00
	CMD_PING               = "ping"
	CMD_LOGIN              = "login"
	CMD_LOGOUT             = "logout"
	CMD_REG                = "reg"
	CMD_CREATE_GROUP       = "create"
	CMD_APPLY_GROUP        = "apply"
	CMD_AUDIT_GROUP        = "audit"
	CMD_CHAT               = "chat"
	CMD_QUIT               = "quit"
	CMD_HISTORY			   = "his"
	CMD_KICK               = "kick"
	CMD_QUERY_GROUP        = "grp_info"
	CMD_USER_PROFILE	   = "u_prof"
	CMD_GROUP_ATTR		   = "g_attr"
	CMD_GROUP_GROUND	   = "g_ground"

	BUFF_LEN = (10 * 1024)

	METHOD_INTERFACE = 1
	METHOD_COMMAND   = 2
)

var cmd_map map[string]string
var zlib_enable = true

//var buff_len int = 1024;
var exit_ch chan bool = make(chan bool, 1)

//var enc_type = lnet.NET_ENCRYPT_DES_ECB //des encrypt
var enc_type int8 = -1
var enc_block cipher.Block
var enc_key []byte

//flag
var help = flag.Bool("h", false, "show help")
var host *string = flag.String("a", "127.0.0.1", "server ip")
var port = flag.Int("p", 0, "server port")
var method = flag.Int("m", 0, "method 1:interace 2:command")
var cmd = flag.String("c", "", "cmd")
var keep = flag.Int("k", 0, "keepalive seconds if method=2")
var quiet = flag.Bool("q", false, "quiet")

var tcp_conn *net.TCPConn

func init() {
	cmd_map = make(map[string]string)
	//init cmd_map
	cmd_map[CMD_PING] = "ping to server"
	cmd_map[CMD_LOGIN] = "login <name> <pass> [version]"
	cmd_map[CMD_LOGOUT] = "logout"
	cmd_map[CMD_REG] = "register <name> <pass> <role_name> <sex:1|2> <addr>"
	cmd_map[CMD_CREATE_GROUP] = "create group <group_name> <group_pass>"
	cmd_map[CMD_APPLY_GROUP] = "apply group <group_id> <group_pass> <apply_msg>"
	cmd_map[CMD_AUDIT_GROUP] = "audit group apply <group_id><grp_name><apply_uid><audit 0|1>"
	cmd_map[CMD_CHAT] = "chat <chat_type><group_id><msg>" //type:0:text 1:img
	cmd_map[CMD_QUIT] = "quit group <group_id>"
	cmd_map[CMD_HISTORY] = "chat history <group_id><now_msg_id>"
	cmd_map[CMD_KICK] = "kick group member <group_id><member_id>"
	cmd_map[CMD_QUERY_GROUP] = "query group info <group_id>"
	cmd_map[CMD_USER_PROFILE] = "user profile [uid1] [uid2] ..."
	cmd_map[CMD_GROUP_ATTR] = "chg group attr <group_id><attr_id>"
	cmd_map[CMD_GROUP_GROUND] = "group ground <start_index>"
}

func v_print(format string, arg ...interface{}) {
	if !*quiet {
		fmt.Printf(format, arg...)
	}
}

func show_cmd() {
	fmt.Printf("----cmd----\n")
	for cmd, info := range cmd_map {
		fmt.Printf("[%s] %s\n", cmd, info)
	}
}

//generate local id
var seq uint16 = 1

func GenerateLocalId(wid int16, seq *uint16) int64 {
	var id int64 = 0
	curr_ts := time.Now().Unix() //
	diff := curr_ts - TIME_EPOLL_BASE

	*seq = *seq + 1
	if *seq >= 65530 {
		*seq = 1
	}

	id = ((int64(*seq) & 0xFFFF) << 47) | ((int64(wid) & 0xFFFF) << 31) | (diff & 0x7FFFFFFF)
	return id
}

func ValidConnection(conn *net.TCPConn) bool {
	//pack
	pkg_buff := make([]byte, 128)
	pkg_len := lnet.PackPkg(pkg_buff, []byte(lnet.CONN_VALID_KEY), lnet.PKG_OP_VALID)
	if pkg_len <= 0 {
		fmt.Printf("valid connection pack failed! pkg_len:%d\n", pkg_len)
		return false
	}

	//send
	_, err := conn.Write(pkg_buff[:pkg_len])
	if err != nil {
		fmt.Printf("send valid pkg failed! err:%v\n", err)
		return false
	}
	v_print("send valid success! pkg_len:%d valid_key:%s\n", pkg_len, lnet.CONN_VALID_KEY)
	return true
}

func RsaNegotiateDesKey(conn *net.TCPConn, inn_key []byte, rsa_pub []byte) bool {
	var _func_ = "<RsaNegotiateDesKey>"
	//encrypt by rsa
	encoded, err := lnet.RsaEncrypt(inn_key, rsa_pub)
	if err != nil {
		v_print("%s failed! err:%v\n", _func_, err)
		return false
	}

	//pack
	pkg_buff := make([]byte, len(encoded)+10)
	pkg_len := lnet.PackPkg(pkg_buff, encoded, lnet.PKG_OP_RSA_NEGO)
	if pkg_len <= 0 {
		fmt.Printf("valid connection pack failed! pkg_len:%d\n", pkg_len)
		return false
	}

	//send
	_, err = conn.Write(pkg_buff[:pkg_len])
	if err != nil {
		fmt.Printf("send valid pkg failed! err:%v\n", err)
		return false
	}
	v_print("send RsaEnc success! pkg_len:%d inn_key:%s\n", pkg_len, string(inn_key))
	return true
}

func RecvConnSpecPkg(tag uint8, data []byte) {
	var _func_ = "<RecvConnSpecPkg>"
	var err error
	//pkg option
	pkg_option := lnet.PkgOption(tag)
	switch pkg_option {
	case lnet.PKG_OP_ECHO:
		v_print("%s echo pkg! content:%s", _func_, string(data))
	case lnet.PKG_OP_VALID:
		enc_type = int8(data[0])
		v_print("%s valid pkg! enc_type:%d content:%s data:%v\n", _func_, enc_type, string(data), data)
		if enc_type == lnet.NET_ENCRYPT_DES_ECB {
			enc_key = make([]byte, 8)
			copy(enc_key, data[1:9])
			enc_block, err = des.NewCipher(enc_key)
			if err != nil {
				v_print("%s new des block failed! err:%v", _func_, err)
			}
			v_print("enc_key:%s\n", string(enc_key))
			break
		}
		if enc_type == lnet.NET_ENCRYPT_AES_CBC_128 {
			enc_key = make([]byte, 16)
			copy(enc_key, data[1:17])
			enc_block, err = aes.NewCipher(enc_key)
			if err != nil {
				v_print("%s new aes block failed! err:%v", _func_, err)
			}
			v_print("enc_key:%s\n", string(enc_key))
			break
		}
		if enc_type == lnet.NET_ENCRYPT_RSA {
			rsa_pub_key := make([]byte, len(data)-1)
			copy(rsa_pub_key, data[1:])
			//v_print("%s rsa_pub_key:%s\n" , _func_ , string(rsa_pub_key))

			//RSA ENC
			enc_key = []byte("12345678")
			ok := RsaNegotiateDesKey(tcp_conn, enc_key, rsa_pub_key)
			if !ok {
				enc_key = enc_key[:0] //clear
			}
		}
	case lnet.PKG_OP_RSA_NEGO:
		v_print("%s rsa_nego pkg! result:%s\n", _func_, string(data))
		if bytes.Compare(data[:2], []byte("ok")) == 0 {
			enc_block, err = des.NewCipher(enc_key)
			if err != nil {
				v_print("%s new des block by rsa-nego failed! err:%v", _func_, err)
			}
		} else {
			v_print("%s nego des key failed for:%s", _func_, string(data))
		}
	default:
		v_print("%s unkonwn option:%d\n", _func_, pkg_option)
	}
}

func DecryptRecv(src_data []byte) []byte {
	var _func_ = "<DecryptRecv>"
	var err error
	pkg_data := src_data
	if enc_block == nil || len(enc_key) <= 0 {
		v_print("%s new encrypt block nil! enc_type:%d", _func_, enc_type)
		return nil
	}
	switch enc_type {
	case lnet.NET_ENCRYPT_DES_ECB:
		pkg_data, err = lnet.DesDecrypt(enc_block, pkg_data, enc_key)
		if err != nil {
			v_print("%s des decrypt failed! err:%v", _func_, err)
			return nil
		}
	case lnet.NET_ENCRYPT_AES_CBC_128:
		pkg_data, err = lnet.AesDecrypt(enc_block, pkg_data, enc_key)
		if err != nil {
			v_print("%s des decrypt failed! err:%v", _func_, err)
			return nil
		}
	case lnet.NET_ENCRYPT_RSA:
		pkg_data, err = lnet.DesDecrypt(enc_block, pkg_data, enc_key)
		if err != nil {
			v_print("%s rsa_des decrypt failed! err:%v", _func_, err)
			return nil
		}
	default:
		v_print("%s illegal enc_type:%d", _func_, enc_type)
		return nil
	}

	return pkg_data
}

func RecvPkg(conn *net.TCPConn) {
	read_buff := make([]byte, BUFF_LEN)
	var recv_end int
	for {
		time.Sleep(10 * time.Millisecond)

		read_buff = read_buff[:cap(read_buff)]
		//read
		n, err := conn.Read(read_buff)
		if err != nil {
			fmt.Printf("read failed! err:%v\n", err)
			os.Exit(0)
		}
		recv_end = n
	_unpacking:
		//unpack
		tag, pkg_data, pkg_len := lnet.UnPackPkg(read_buff[:recv_end])
		if lnet.PkgOption(tag) != lnet.PKG_OP_NORMAL {
			v_print("read spec pkg. tag:%d pkg_len:%d pkg_option:%d\n", tag, pkg_len, lnet.PkgOption(tag))
			RecvConnSpecPkg(tag, pkg_data)
			continue
		}
		//v_print("recv pkg: tag:%d pkg_len:%d pkg_data:%v\n" , tag , pkg_len , string(pkg_data))
		//decrypt
		if enc_type != lnet.NET_ENCRYPT_NONE {
			pkg_data = DecryptRecv(pkg_data)
			if pkg_data == nil {
				return
			}
		}

		//uncompress
		if zlib_enable {
			b := bytes.NewReader(pkg_data)
			var out bytes.Buffer
			r, err := zlib.NewReader(b)
			if err != nil {
				fmt.Printf("uncompress data failed! err:%v", err)
				if *method == METHOD_COMMAND {
					exit_ch <- false
					return
				}
				continue
			}
			io.Copy(&out, r)
			pkg_data = out.Bytes()
		}

		//decode
		var gmsg cs.GeneralMsg
		err = cs.DecodeMsg(pkg_data, &gmsg)
		if err != nil {
			fmt.Printf("decode failed! err:%v\n", err)
			if *method == METHOD_COMMAND {
				exit_ch <- false
				return
			}
			continue
		}

		//switch rsp
		switch gmsg.ProtoId {
		case cs.CS_PROTO_PING_RSP:
			prsp, ok := gmsg.SubMsg.(*cs.CSPingRsp)
			if ok {
				curr_ts := time.Now().UnixNano() / 1000
				v_print("ping:%v ms crr_ts:%d req:%d\n", (curr_ts-prsp.TimeStamp)/1000, curr_ts, prsp.TimeStamp)

				if *method == METHOD_COMMAND {
					exit_ch <- true
					return
				}
			}
		case cs.CS_PROTO_LOGIN_RSP:
			prsp, ok := gmsg.SubMsg.(*cs.CSLoginRsp)
			if ok {
				if prsp.Result == 0 {
					v_print("login result:%d name:%s role_name:%s head_url:%s\n", prsp.Result, prsp.Name, prsp.Basic.Name , prsp.Basic.HeadUrl)
					v_print("uid:%v sex:%d addr:%s level:%d Exp:%d AllGroup:%d MasterGroup:%d\n", prsp.Basic.Uid, prsp.Basic.Sex, prsp.Basic.Addr,
						prsp.Basic.Level, prsp.Detail.Exp, prsp.Detail.ChatInfo.AllGroup, prsp.Detail.ChatInfo.MasterGroup)
					if prsp.Detail.ChatInfo.AllGroup > 0 {
						for grp_id, info := range prsp.Detail.ChatInfo.AllGroups {
							v_print("[%d] grp_id:%d name:%s last_read:%d enter:%d\n", grp_id, info.GroupId, info.GroupName, info.LastMsgId , info.EnterTs)
						}

						if prsp.Detail.ChatInfo.MasterGroup > 0 {
							v_print("master_groups:%v\n", prsp.Detail.ChatInfo.MasterGroups)
						}

					}
				} else {
					v_print("login result:%d name:%s\n", prsp.Result, prsp.Name)
				}
				if *method == METHOD_COMMAND {
					exit_ch <- true
					return
				}
			}
		case cs.CS_PROTO_LOGOUT_RSP:
			prsp, ok := gmsg.SubMsg.(*cs.CSLogoutRsp)
			if ok {
				v_print("logout result:%d msg:%s\n", prsp.Result, prsp.Msg)

				if *method == METHOD_COMMAND {
					exit_ch <- true
					return
				}
			}
		case cs.CS_PROTO_CREATE_GRP_RSP:
			prsp, ok := gmsg.SubMsg.(*cs.CSCreateGroupRsp)
			if ok {
				v_print("create group result:%d grp_name:%s grp_id:%d ts:%d mem_count:%d\n", prsp.Result, prsp.Name,
					prsp.GrpId, prsp.CreateTs, prsp.MemberCnt)

				if *method == METHOD_COMMAND {
					exit_ch <- true
					return
				}
			}
		case cs.CS_PROTO_APPLY_GRP_RSP:
			prsp, ok := gmsg.SubMsg.(*cs.CSApplyGroupRsp)
			if ok {
				v_print("apply group result:%d grp_name:%s grp_id:%d\n", prsp.Result, prsp.GrpName, prsp.GrpId)
				if *method == METHOD_COMMAND {
					exit_ch <- true
					return
				}
			}
		case cs.CS_PROTO_REG_RSP:
			prsp, ok := gmsg.SubMsg.(*cs.CSRegRsp)
			if ok {
				v_print("reg result:%d name:%s\n", prsp.Result, prsp.Name)

				if *method == METHOD_COMMAND {
					exit_ch <- true
					return
				}
			}
		case cs.CS_PROTO_APPLY_GRP_NOTIFY:
			prsp, ok := gmsg.SubMsg.(*cs.CSApplyGroupNotify)
			if ok {
				v_print("apply_grp_notify apply_name:%s apply_uid:%d apply_msg:%s grp_id:%d grp_name:%s\n",
					prsp.ApplyName, prsp.ApplyUid, prsp.ApplyMsg, prsp.GrpId, prsp.GrpName)
				if *method == METHOD_COMMAND {
					exit_ch <- true
					return
				}
			}
		case cs.CS_PROTO_SEND_CHAT_RSP:
			prsp, ok := gmsg.SubMsg.(*cs.CSSendChatRsp)
			if ok {
				v_print("send_chat_rsp temp_id:%d result:%d\n", prsp.TempMsgId, prsp.Result)
				if prsp.ChatMsg != nil {
					v_print("chat_msg:%v\n", prsp.ChatMsg)
				}
				if *method == METHOD_COMMAND {
					exit_ch <- true
					return
				}
			}
		case cs.CS_PROTO_SYNC_CHAT_LIST:
			prsp, ok := gmsg.SubMsg.(*cs.CSSyncChatList)
			if ok {
				v_print("sync_chat_list grp_id:%d count:%d type:%d\n" , prsp.GrpId , prsp.Count , prsp.SyncType)
				for i:=0; i<prsp.Count; i++ {
					pchat := prsp.ChatList[i]
					v_print("[%d]<%d>sender:%d name:%s content:%s time:%d type:%d grp_id:%d\n" , i , pchat.MsgId , pchat.SenderUid ,
						pchat.SenderName , pchat.Content , pchat.SendTs , pchat.ChatType , pchat.GrpId)
				}
				if *method == METHOD_COMMAND {
					exit_ch <- true
					return
				}
			}
		case cs.CS_PROTO_EXIT_GROUP_RSP:
			prsp, ok := gmsg.SubMsg.(*cs.CSExitGroupRsp)
			if ok {
				v_print("exit_group_rsp grp_id:%d grp_name:%s result:%d del_group:%d is_kick:%d\n", prsp.GrpId, prsp.GrpName , prsp.Result ,
					prsp.DelGroup , prsp.ByKick)
				if *method == METHOD_COMMAND {
					exit_ch <- true
					return
				}
			}
		case cs.CS_PROTO_COMMON_NOTIFY:
			prsp , ok := gmsg.SubMsg.(*cs.CSCommonNotify)
			if ok {
				v_print("common_notify type:%d grp_id:%d intv:%d strv:%s strs:%v\n" , prsp.NotifyType , prsp.GrpId , prsp.IntV , prsp.StrV ,
					prsp.StrS)
				if *method == METHOD_COMMAND {
					exit_ch <- true
					return
				}
			}
		case cs.CS_PROTO_SYNC_GROUP_INFO:
			prsp , ok := gmsg.SubMsg.(*cs.CSSyncGroupInfo)
			if ok {
				v_print("sync group info field:%d grp_id:%d\n" , prsp.Field , prsp.GrpId)
				if prsp.GrpInfo != nil {
					v_print("group_info:%v\n" , *prsp.GrpInfo)
				}
				if *method == METHOD_COMMAND {
					exit_ch <- true
					return
				}
			}
		case cs.CS_PROTO_FETCH_USER_PROFILE_RSP:
			prsp , ok := gmsg.SubMsg.(*cs.CSFetchUserProfileRsp)
			if ok {
				v_print("fetch user profiles count:%d\n" , len(prsp.Profiles))
				if prsp.Profiles != nil {
					for uid , pinfo := range prsp.Profiles {
						if pinfo != nil {
							v_print("<%d> %v\n", uid , *pinfo)
						} else {
							v_print("<%d> nil\n", uid)
						}
					}
				}
				if *method == METHOD_COMMAND {
					exit_ch <- true
					return
				}
			}
		case cs.CS_PROTO_CHG_GROUP_ATTR_RSP:
			prsp , ok := gmsg.SubMsg.(*cs.CSChgGroupAttrRsp)
			if ok {
				v_print("chg group attr result:%d attr:%d grp_id:%d\n" , prsp.Result , prsp.Attr , prsp.GrpId)
				if *method == METHOD_COMMAND {
					exit_ch <- true
					return
				}
			}
		case cs.CS_PROTO_GROUP_GROUND_RSP:
			prsp , ok := gmsg.SubMsg.(*cs.CSGroupGroundRsp)
			if ok {
				v_print("ground group count:%d\n" , prsp.Count)
				if prsp.Count>0 && len(prsp.ItemList)>0 {
					for i:=0; i<prsp.Count; i++ {
						v_print("<%d> grp_id:%d grp_name:%s\n" , i , prsp.ItemList[i].GrpId , prsp.ItemList[i].GrpName)
					}
				}
				if *method == METHOD_COMMAND {
					exit_ch <- true
					return
				}
			}
		default:
			fmt.Printf("illegal proto:%d\n", gmsg.ProtoId)
		}

		//check data
		if pkg_len < recv_end {
			copy(read_buff, read_buff[pkg_len:])
			recv_end = recv_end - pkg_len
			goto _unpacking
		}

		if *method == METHOD_COMMAND {
			exit_ch <- false
			return
		}

	}
}

//send pkg to server
func SendPkg(conn *net.TCPConn, cmd string) {
	var _func_ = "<SendPkg>"
	var gmsg cs.GeneralMsg
	var err error
	var enc_data []byte

	pkg_buff := make([]byte, BUFF_LEN)
	//parse cmd and arg
	args := strings.Split(cmd, " ")

	//encode msg
	switch args[0] {
	case CMD_PING:
		v_print("ping...\n")

		gmsg.ProtoId = cs.CS_PROTO_PING_REQ
		psub := new(cs.CSPingReq)
		psub.TimeStamp = time.Now().UnixNano() / 1000
		gmsg.SubMsg = psub

	case CMD_LOGIN: //login <name> <pass> [version]
		if len(args) < 3 {
			show_cmd()
			return
		}
		gmsg.ProtoId = cs.CS_PROTO_LOGIN_REQ
		psub := new(cs.CSLoginReq)
		psub.Name = args[1]
		psub.Device = "onepluse9"
		psub.Pass = args[2]
		if len(args) == 4 {
			psub.Version = args[3]
		}
		gmsg.SubMsg = psub
		v_print("login...name:%s pass:%s\n", psub.Name, psub.Pass)
	case CMD_LOGOUT:
		v_print("logout...\n")

		gmsg.ProtoId = cs.CS_PROTO_LOGOUT_REQ
		psub := new(cs.CSLogoutReq)
		psub.Uid = 0
		gmsg.SubMsg = psub
	case CMD_REG: //register <name> <pass> <role_name> <sex:1|2> <addr>
		if len(args) != 6 {
			show_cmd()
			return
		}
		gmsg.ProtoId = cs.CS_PROTO_REG_REQ
		psub := new(cs.CSRegReq)
		psub.Name = args[1]
		psub.Pass = args[2]
		psub.RoleName = args[3]
		sex_v, _ := strconv.Atoi(args[4])
		psub.Sex = uint8(sex_v)
		psub.Addr = args[5]
		v_print("reg... name:%s pass:%s sex:%d addr:%s\n", psub.Name, psub.Pass, psub.Sex, psub.Addr)

		gmsg.SubMsg = psub
	case CMD_CREATE_GROUP: //create_grp  <name> <pass>
		if len(args) != 3 {
			show_cmd()
			return
		}

		gmsg.ProtoId = cs.CS_PROTO_CREATE_GRP_REQ
		psub := new(cs.CSCreateGroupReq)
		psub.Name = args[1]
		psub.Pass = args[2]
		v_print("create group name:%s pass:%s\n", psub.Name, psub.Pass)
		gmsg.SubMsg = psub
	case CMD_APPLY_GROUP: //apply_group <grp_id> <grp_pass> <apply_msg>
		if len(args) != 4 {
			show_cmd()
			return
		}

		gmsg.ProtoId = cs.CS_PROTO_APPLY_GRP_REQ
		psub := new(cs.CSApplyGroupReq)
		psub.GrpId, _ = strconv.ParseInt(args[1], 10, 64)
		psub.Pass = args[2]
		psub.Msg = args[3]
		v_print("apply group grp_id:%d pass:%s\n", psub.GrpId, psub.Pass)
		gmsg.SubMsg = psub
	case CMD_AUDIT_GROUP: //<group_id><grp_name><apply_uid><audit 0|1>
		if len(args) != 5 {
			show_cmd()
			return
		}

		gmsg.ProtoId = cs.CS_PROTO_APPLY_GRP_AUDIT
		psub := new(cs.CSApplyGroupAudit)
		psub.GrpId, _ = strconv.ParseInt(args[1], 10, 64)
		psub.GrpName = args[2]
		psub.ApplyUid, _ = strconv.ParseInt(args[3], 10, 64)
		psub.Audit, _ = strconv.Atoi(args[4])
		v_print("audit group request:%v\n", *psub)
		gmsg.SubMsg = psub
	case CMD_CHAT: //send_chat <chat_type><group_id><msg>
		if len(args) != 4 {
			show_cmd()
			return
		}

		gmsg.ProtoId = cs.CS_PROTO_SEND_CHAT_REQ
		psub := new(cs.CSSendChatReq)
		psub.ChatType, _ = strconv.Atoi(args[1])
		psub.GrpId, _ = strconv.ParseInt(args[2], 10, 64)
		psub.Content = args[3]
		psub.TempMsgId = GenerateLocalId(1, &seq)
		v_print("send chat req:%v\n", *psub)
		gmsg.SubMsg = psub
	case CMD_QUIT:	// quit group <group_id>
		if len(args) != 2 {
			show_cmd()
			return
		}
		gmsg.ProtoId = cs.CS_PROTO_EXIT_GROUP_REQ
		psub := new(cs.CSExitGroupReq)
		psub.GrpId, _ = strconv.ParseInt(args[1], 10, 64)
		v_print("exit group req:%v\n", *psub)
		gmsg.SubMsg = psub
	case CMD_HISTORY: // chat history <group_id><now_msg_id>
		if len(args) != 3 {
			show_cmd()
			return
		}
		gmsg.ProtoId = cs.CS_PROTO_CHAT_HISTORY_REQ
		psub := new(cs.CSChatHistoryReq)
		psub.GrpId, _ = strconv.ParseInt(args[1], 10, 64)
		psub.NowMsgId , _ = strconv.ParseInt(args[2] , 10 , 64)
		v_print("chat history req:%v\n", *psub)
		gmsg.SubMsg = psub
	case CMD_KICK: //"kick group member <group_id><member_id>"
		if len(args) != 3 {
			show_cmd()
			return
		}
		gmsg.ProtoId = cs.CS_PROTO_KICK_GROUP_REQ
		psub := new(cs.CSKickGroupReq)
		psub.GrpId , _ = strconv.ParseInt(args[1] , 10 , 64)
		psub.KickUid , _ = strconv.ParseInt(args[2] , 10 , 64)
		v_print("kick group req:%v\n", *psub)
		gmsg.SubMsg = psub
	case CMD_QUERY_GROUP: //"query group info <group_id>"
		if len(args) != 2 {
			show_cmd()
			return
		}
		gmsg.ProtoId = cs.CS_PROTO_QUERY_GROUP_REQ
		psub := new(cs.CSQueryGroupReq)
		psub.GrpId , _ = strconv.ParseInt(args[1] , 10 , 64)
		v_print("query group req:%v\n", *psub)
		gmsg.SubMsg = psub
	case CMD_USER_PROFILE: //"user profile [uid1] [uid2] ..."
		if len(args) < 2 {
			show_cmd()
			return
		}
		gmsg.ProtoId = cs.CS_PROTO_FETCH_USER_PROFILE_REQ
		psub := new(cs.CSFetchUserProfileReq)
		psub.TargetList = make([]int64 , len(args)-1)
		idx := 0
		for i:=1; i<len(args); i++ {
			psub.TargetList[idx] , _ = strconv.ParseInt(args[i] , 10 , 64)
			idx++
		}
		v_print("fetch user profile req:%v\n" , *psub)
		gmsg.SubMsg = psub
	case CMD_GROUP_ATTR: //"chg group attr <group_id><attr_id>"
		if len(args) != 3 {
			show_cmd()
			return
		}
		gmsg.ProtoId = cs.CS_PROTO_CHG_GROUP_ATTR_REQ
		psub := new(cs.CSChgGroupAttrReq)
		psub.GrpId , _ = strconv.ParseInt(args[1] , 10 , 64)
		psub.Attr , _ = strconv.Atoi(args[2])
		v_print("chg group attr req:%v\n", *psub)
		gmsg.SubMsg = psub
	case CMD_GROUP_GROUND: //"group ground <start_index>"
		if len(args) != 2 {
			show_cmd()
			return
		}
		gmsg.ProtoId = cs.CS_PROTO_GROUP_GROUND_REQ
		psub := new(cs.CSGroupGroundReq)
		psub.StartIndex , _ = strconv.Atoi(args[1])
		v_print("group group req:%v\n", *psub)
		gmsg.SubMsg = psub
	default:
		fmt.Printf("illegal cmd:%s\n", cmd)
		show_cmd()
		return
	}

	//encode
	enc_data, err = cs.EncodeMsg(&gmsg)
	if err != nil {
		fmt.Printf("encode %s failed! err:%v\n", cmd, err)
		return
	}

	//compress
	if zlib_enable {
		var b bytes.Buffer
		w := zlib.NewWriter(&b)
		w.Write(enc_data)
		w.Close()
		enc_data = b.Bytes()
	}

	//encrypt
	if enc_type != lnet.NET_ENCRYPT_NONE {
		if enc_block == nil || len(enc_key) <= 0 {
			v_print("%s new encrypt block nil! enc_type:%d", _func_, enc_type)
			return
		}
		switch enc_type {
		case lnet.NET_ENCRYPT_DES_ECB:
			enc_data, err = lnet.DesEncrypt(enc_block, enc_data, enc_key)
			if err != nil {
				v_print("%s des encrypt failed! err:%v", _func_, err)
				return
			}
		case lnet.NET_ENCRYPT_AES_CBC_128:
			enc_data, err = lnet.AesEncrypt(enc_block, enc_data, enc_key)
			if err != nil {
				v_print("%s des encrypt failed! err:%v", _func_, err)
				return
			}
		case lnet.NET_ENCRYPT_RSA:
			enc_data, err = lnet.DesEncrypt(enc_block, enc_data, enc_key)
			if err != nil {
				v_print("%s rsa_des encrypt failed! err:%v", _func_, err)
				return
			}
		default:
			v_print("%s illegal enc_type:%d", _func_, enc_type)
			return
		}

	}

	//pack
	pkg_len := lnet.PackPkg(pkg_buff, enc_data, 0)
	if pkg_len < 0 {
		fmt.Printf("pack cmd:%s failed!\n", cmd)
		return
	}

	//send
	_, err = conn.Write(pkg_buff[:pkg_len])
	if err != nil {
		fmt.Printf("send cmd pkg failed! cmd:%s err:%v\n", cmd, err)
	} else {
		v_print("send cmd:%s success! \n", cmd)
	}
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("main panic! err:%v", err)
		}
	}()
	flag.Parse()
	if *port <= 0 || (*method != METHOD_INTERFACE && *method != METHOD_COMMAND) {
		flag.PrintDefaults()
		show_cmd()
		return
	}
	if *help {
		show_cmd()
		return
	}

	v_print("start client ...\n")
	start_us := time.Now().UnixNano() / 1000
	//server_addr := "localhost:18909";
	server_addr := *host + ":" + strconv.Itoa(*port)

	//init
	//exit_ch = make(chan bool, 1)

	//connect
	tcp_addr, err := net.ResolveTCPAddr("tcp4", server_addr)
	if err != nil {
		fmt.Printf("resolve addr:%s failed! err:%s\n", server_addr, err)
		return
	}

	conn, err := net.DialTCP("tcp4", nil, tcp_addr)
	if err != nil {
		fmt.Printf("connect %s failed! err:%v\n", server_addr, err)
		return
	}
	tcp_conn = conn
	defer conn.Close()

	//valid connection
	ok := ValidConnection(conn)
	if !ok {
		fmt.Printf("valid connection error!\n")
		return
	}

	rs := make([]byte, 128)
	//pack_buff := make([]byte , 128);
	//read
	go RecvPkg(conn)

	show_cmd()
	//check option
	switch *method {
	case METHOD_INTERFACE:
		for {
			rs = rs[:cap(rs)]
			fmt.Printf("please input:>>")
			n, _ := os.Stdin.Read(rs)
			rs = rs[:n-1] //trip last \n

			if string(rs) == "EXIT" {
				fmt.Println("byte...")
				exit_ch <- true
				break
			}

			//pkg_len := lnet.PackPkg(pack_buff, rs , (uint8)(*option));
			//fmt.Printf("read %d bytes and packed:%d\n", n , pkg_len);
			//n , _ = conn.Write(pack_buff[:pkg_len]);
			SendPkg(conn, string(rs))
			time.Sleep(50 * time.Millisecond)
		}

	case METHOD_COMMAND:
		if len(*cmd) <= 0 {
			flag.PrintDefaults()
			show_cmd()
			break
		}
		//start_us := time.Now().UnixNano()/1000;
		SendPkg(conn, *cmd)
		//time.Sleep(2 * time.Second)
		v := <-exit_ch
		end_us := time.Now().UnixNano() / 1000
		if v { //success
			fmt.Printf("cmd:0|start:%d|end:%d|cost:%d\n", start_us, end_us, (end_us - start_us))
		} else {
			fmt.Printf("cmd:-1|start:%d|end:%d|cost:%d\n", start_us, end_us, (end_us - start_us))
		}

		//keep alive
		time.Sleep(time.Duration(*keep) * time.Second)

	default:
		fmt.Printf("option:%d nothing tod!\n", *method)
	}

	return
}
