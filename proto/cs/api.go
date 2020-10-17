package cs

import (
	"encoding/json"
	"errors"
)

/*
This is a cs-proto using json format
*/

/*
CS PROTO ID
*/
const (
	//proto start
	CS_PROTO_START      = 0
	CS_PROTO_PING_REQ   = 1
	CS_PROTO_PING_RSP   = 2
	CS_PROTO_LOGIN_REQ  = 3
	CS_PROTO_LOGIN_RSP  = 4
	CS_PROTO_LOGOUT_REQ = 5
	CS_PROTO_LOGOUT_RSP = 6
	CS_PROTO_REG_REQ    = 7
	CS_PROTO_REG_RSP    = 8
	CS_PROTO_CREATE_GRP_REQ = 9
	CS_PROTO_CREATE_GRP_RSP  = 10
	CS_PROTO_APPLY_GRP_REQ = 11
	CS_PROTO_APPLY_GRP_RSP = 12
	CS_PROTO_APPLY_GRP_NOTIFY = 13
	CS_PROTO_APPLY_GRP_AUDIT  = 14
	CS_PROTO_SEND_CHAT_REQ = 15
	CS_PROTO_SEND_CHAT_RSP = 16
	CS_PROTO_SYNC_CHAT_LIST = 17
	CS_PROTO_EXIT_GROUP_REQ = 18
	CS_PROTO_EXIT_GROUP_RSP = 19
	CS_PROTO_CHAT_HISTORY_REQ = 20
	CS_PROTO_COMMON_NOTIFY    = 21
	CS_PROTO_KICK_GROUP_REQ  = 22
	CS_PROTO_QUERY_GROUP_REQ = 23
	CS_PROTO_SYNC_GROUP_INFO = 24
	CS_PROTO_FETCH_USER_PROFILE_REQ = 25
	CS_PROTO_FETCH_USER_PROFILE_RSP = 26
	CS_PROTO_CHG_GROUP_ATTR_REQ = 27
	CS_PROTO_CHG_GROUP_ATTR_RSP = 28
	CS_PROTO_GROUP_GROUND_REQ = 29
	CS_PROTO_GROUP_GROUND_RSP = 30
	//PS:new proto added should modify 'Proto2Msg' function
	//proto end = last + 1
	CS_PROTO_END = 31
)

/*
* GeneralMsg
 */
type GeneralMsg struct {
	ProtoId int         `json:"proto"`
	SubMsg  interface{} `json:"sub"`
}

type ProtoHead struct {
	ProtoId int         `json:"proto"`
	Sub     interface{} `json:"-"`
}

/*
* Encode GeneralMsg
* @return encoded_bytes , error
 */
func EncodeMsg(pmsg *GeneralMsg) ([]byte, error) {
	//proto
	if pmsg.ProtoId <= CS_PROTO_START || pmsg.ProtoId >= CS_PROTO_END {
		return nil, errors.New("proto_id illegal")
	}

	//encode
	return json.Marshal(pmsg)
}

/*
* Decode GeneralMsg
* @return
 */
func DecodeMsg(data []byte, pmsg *GeneralMsg) error {
	var proto_head ProtoHead
	var err error

	//decode proto
	err = json.Unmarshal(data, &proto_head)
	if err != nil {
		return err
	}

	//switch proto
	proto_id := proto_head.ProtoId
	psub, err := Proto2Msg(proto_id)
	if err != nil {
		return err
	}
	pmsg.SubMsg = psub

	//decode
	err = json.Unmarshal(data, pmsg)
	if err != nil {
		return err
	}

	return nil
}


/*
* Get ProtoMsg By Proto
 */
func Proto2Msg(proto_id int) (interface{}, error) {
	var pmsg interface{}

	//refer
	switch proto_id {
	case CS_PROTO_PING_REQ:
		pmsg = new(CSPingReq)
	case CS_PROTO_PING_RSP:
		pmsg = new(CSPingRsp)
	case CS_PROTO_LOGIN_REQ:
		pmsg = new(CSLoginReq)
	case CS_PROTO_LOGIN_RSP:
		pmsg = new(CSLoginRsp)
	case CS_PROTO_LOGOUT_REQ:
		pmsg = new(CSLogoutReq)
	case CS_PROTO_LOGOUT_RSP:
		pmsg = new(CSLogoutRsp)
	case CS_PROTO_REG_REQ:
		pmsg = new(CSRegReq)
	case CS_PROTO_REG_RSP:
		pmsg = new(CSRegRsp)
	case CS_PROTO_CREATE_GRP_REQ:
		pmsg = new(CSCreateGroupReq)
	case CS_PROTO_CREATE_GRP_RSP:
		pmsg = new(CSCreateGroupRsp)
	case CS_PROTO_APPLY_GRP_REQ:
		pmsg = new(CSApplyGroupReq)
	case CS_PROTO_APPLY_GRP_RSP:
		pmsg = new(CSApplyGroupRsp)
	case CS_PROTO_APPLY_GRP_NOTIFY:
		pmsg = new(CSApplyGroupNotify)
	case CS_PROTO_APPLY_GRP_AUDIT:
		pmsg = new(CSApplyGroupAudit)
	case CS_PROTO_SEND_CHAT_REQ:
		pmsg = new(CSSendChatReq)
	case CS_PROTO_SEND_CHAT_RSP:
		pmsg = new(CSSendChatRsp)
	case CS_PROTO_SYNC_CHAT_LIST:
		pmsg = new(CSSyncChatList)
	case CS_PROTO_EXIT_GROUP_REQ:
		pmsg = new(CSExitGroupReq)
	case CS_PROTO_EXIT_GROUP_RSP:
		pmsg = new(CSExitGroupRsp)
	case CS_PROTO_CHAT_HISTORY_REQ:
		pmsg = new(CSChatHistoryReq)
	case CS_PROTO_COMMON_NOTIFY:
		pmsg = new(CSCommonNotify)
	case CS_PROTO_KICK_GROUP_REQ:
		pmsg = new(CSKickGroupReq)
	case CS_PROTO_QUERY_GROUP_REQ:
		pmsg = new(CSQueryGroupReq)
	case CS_PROTO_SYNC_GROUP_INFO:
		pmsg = new(CSSyncGroupInfo)
	case CS_PROTO_FETCH_USER_PROFILE_REQ:
		pmsg = new(CSFetchUserProfileReq)
	case CS_PROTO_FETCH_USER_PROFILE_RSP:
		pmsg = new(CSFetchUserProfileRsp)
	case CS_PROTO_CHG_GROUP_ATTR_REQ:
		pmsg = new(CSChgGroupAttrReq)
	case CS_PROTO_CHG_GROUP_ATTR_RSP:
		pmsg = new(CSChgGroupAttrRsp)
	case CS_PROTO_GROUP_GROUND_REQ:
		pmsg = new(CSGroupGroundReq)
	case CS_PROTO_GROUP_GROUND_RSP:
		pmsg = new(CSGroupGroundRsp)
	default:
		return nil, errors.New("proto illegal!")
	}

	//return
	return pmsg, nil
}


/*-----------------------------------STATIC--------------------*/

