package comm

import (
	"errors"
	"fmt"
	"schat/proto/ss"
)


/*
Generate Disp Msg by arg. this function should be modified when new ss.DISP_PROTO_TYPE added
@disp_msg: Disp sub msg  like ss.MsgDispxxxx
@target:target server type
@method:choose spec target server method
@spec:spec target server will ignore @target and @method
@hash_v: if method is hash , this specify hash_value
@return:ss_msg , error
 */
func GenDispMsg(target ss.DISP_MSG_TARGET, method ss.DISP_MSG_METHOD, proto ss.DISP_PROTO_TYPE, spec int , sender int , hash_v int64 ,
	disp_msg interface{}) (*ss.SSMsg , error) {
	var ss_msg = new(ss.SSMsg)
	ss_msg.ProtoType = ss.SS_PROTO_TYPE_USE_DISP_PROTO
	body := new(ss.SSMsg_MsgDisp)
	body.MsgDisp = new(ss.MsgDisp)
	body.MsgDisp.ProtoType = proto
	body.MsgDisp.Method = method
	body.MsgDisp.Target = target
	body.MsgDisp.HashV = hash_v
	body.MsgDisp.SpecServer = int32(spec)
	body.MsgDisp.FromServer = int32(sender)
	ss_msg.MsgBody = body

	//create dis_body
	//switch proto
	switch proto {
	case ss.DISP_PROTO_TYPE_DISP_HELLO:
		disp_body := new(ss.MsgDisp_Hello)
		pv , ok := disp_msg.(*ss.MsgDispHello)
		if !ok {
			return nil , errors.New("not MsgDispHello")
		}
		disp_body.Hello = pv
		body.MsgDisp.DispBody = disp_body
	case ss.DISP_PROTO_TYPE_DISP_KICK_DUPLICATE_USER:
		disp_body := new(ss.MsgDisp_KickDupUser)
		pv , ok := disp_msg.(*ss.MsgDispKickDupUser)
		if !ok {
			return nil , errors.New("not MsgDispKickDupUser")
		}
		disp_body.KickDupUser = pv
		body.MsgDisp.DispBody = disp_body
	case ss.DISP_PROTO_TYPE_DISP_APPLY_GROUP_REQ:
		disp_body := new(ss.MsgDisp_ApplyGroupReq)
		pv , ok := disp_msg.(*ss.MsgApplyGroupReq)
		if !ok {
			return nil , errors.New("not MsgApplyGroupReq")
		}
		disp_body.ApplyGroupReq = pv
		body.MsgDisp.DispBody = disp_body
	case ss.DISP_PROTO_TYPE_DISP_APPLY_GROUP_RSP:
		disp_body := new(ss.MsgDisp_ApplyGroupRsp)
		pv , ok := disp_msg.(*ss.MsgApplyGroupRsp)
		if !ok {
			return nil , errors.New("not MsgApplyGroupRsp")
		}
		disp_body.ApplyGroupRsp = pv
		body.MsgDisp.DispBody = disp_body
	case ss.DISP_PROTO_TYPE_DISP_APPLY_GROUP_NOTIFY:
		disp_body := new(ss.MsgDisp_ApplyGroupNotify)
		pv , ok := disp_msg.(*ss.MsgApplyGroupNotify)
		if !ok {
			return nil , errors.New("not MsgApplyGroupNotify")
		}
		disp_body.ApplyGroupNotify = pv
		body.MsgDisp.DispBody = disp_body
	case ss.DISP_PROTO_TYPE_DISP_APPLY_GROUP_AUDIT:
		disp_body := new(ss.MsgDisp_ApplyGroupAudit)
		pv , ok := disp_msg.(*ss.MsgApplyGroupAudit)
		if !ok {
			return nil , errors.New("not MsgApplyGroupNotify")
		}
		disp_body.ApplyGroupAudit = pv
		body.MsgDisp.DispBody = disp_body
	case ss.DISP_PROTO_TYPE_DISP_COMMON_NOTIFY:
		disp_body := new(ss.MsgDisp_CommonNotify)
		pv , ok := disp_msg.(*ss.MsgCommonNotify)
		if !ok {
			return nil , errors.New("not MsgCommonNotify")
		}
		disp_body.CommonNotify = pv
		body.MsgDisp.DispBody = disp_body
	case ss.DISP_PROTO_TYPE_DISP_ENTER_GROUP_REQ:
		disp_body := new(ss.MsgDisp_EnterGroupReq)
		pv , ok := disp_msg.(*ss.MsgEnterGroupReq)
		if !ok {
			return nil , errors.New("not MsgEnterGroupReq")
		}
		disp_body.EnterGroupReq = pv
		body.MsgDisp.DispBody = disp_body
	case ss.DISP_PROTO_TYPE_DISP_ENTER_GROUP_RSP:
		disp_body := new(ss.MsgDisp_EnterGroupRsp)
		pv , ok := disp_msg.(*ss.MsgEnterGroupRsp)
		if !ok {
			return nil , errors.New("not MsgEnterGroupRsp")
		}
		disp_body.EnterGroupRsp = pv
		body.MsgDisp.DispBody = disp_body
	case ss.DISP_PROTO_TYPE_DISP_SEND_CHAT_REQ:
		disp_body := new(ss.MsgDisp_SendChatReq)
		pv , ok := disp_msg.(*ss.MsgSendChatReq)
		if !ok {
			return nil , errors.New("not MsgSendChatReq")
		}
		disp_body.SendChatReq = pv
		body.MsgDisp.DispBody = disp_body
	case ss.DISP_PROTO_TYPE_DISP_SEND_CHAT_RSP:
		disp_body := new(ss.MsgDisp_SendChatRsp)
		pv , ok := disp_msg.(*ss.MsgSendChatRsp)
		if !ok {
			return nil , errors.New("not MsgSendChatRsp")
		}
		disp_body.SendChatRsp = pv
		body.MsgDisp.DispBody = disp_body
	case ss.DISP_PROTO_TYPE_DISP_QUERY_GROUP_REQ:
		disp_body := new(ss.MsgDisp_QueryGroupReq)
		pv , ok := disp_msg.(*ss.MsgQueryGroupReq)
		if !ok {
			return nil , errors.New("not MsgQueryGroupReq")
		}
		disp_body.QueryGroupReq = pv
		body.MsgDisp.DispBody = disp_body
	case ss.DISP_PROTO_TYPE_DISP_SYNC_GROUP_INFO:
		disp_body := new(ss.MsgDisp_SyncGroupInfo)
		pv , ok := disp_msg.(*ss.MsgSyncGroupInfo)
		if !ok {
			return nil , errors.New("not MsgSyncGroupInfo")
		}
		disp_body.SyncGroupInfo = pv
		body.MsgDisp.DispBody = disp_body
	default:
		return nil , errors.New(fmt.Sprintf("disp proto:%d not handled" , proto))
	}

	return ss_msg , nil
}
