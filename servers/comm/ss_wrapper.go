package comm

import (
	"errors"
	"fmt"
	"schat/proto/ss"
)

/*
Fill SSMsg By ProtoType and MsgBody This is a helper for wrapper pkg
@pmsg: ss.Msg*** defined in SSMsg.msg_body
@pss_msg: fill info of this ss_msg
@return: error
*/
func FillSSPkg(ss_msg *ss.SSMsg, proto ss.SS_PROTO_TYPE, pmsg interface{}) error {
	ss_msg.ProtoType = proto

	switch proto {
	case ss.SS_PROTO_TYPE_HEART_BEAT_REQ:
		body := new(ss.SSMsg_HeartBeatReq)
		pv, ok := pmsg.(*ss.MsgHeartBeatReq)
		if !ok {
			return errors.New("not MsgHeartBeatReq")
		}
		body.HeartBeatReq = pv
		ss_msg.MsgBody = body
	case ss.SS_PROTO_TYPE_PING_REQ:
		body := new(ss.SSMsg_PingReq)
		pv, ok := pmsg.(*ss.MsgPingReq)
		if !ok {
			return errors.New("not MsgPingReq")
		}
		body.PingReq = pv
		ss_msg.MsgBody = body
	case ss.SS_PROTO_TYPE_PING_RSP:
		body := new(ss.SSMsg_PingRsp)
		pv, ok := pmsg.(*ss.MsgPingRsp)
		if !ok {
			return errors.New("not MsgPingRsp")
		}
		body.PingRsp = pv
		ss_msg.MsgBody = body
	case ss.SS_PROTO_TYPE_LOGIN_REQ:
		body := new(ss.SSMsg_LoginReq)
		pv, ok := pmsg.(*ss.MsgLoginReq)
		if !ok {
			return errors.New("not MsgLoginReq")
		}
		body.LoginReq = pv
		ss_msg.MsgBody = body
	case ss.SS_PROTO_TYPE_LOGIN_RSP:
		body := new(ss.SSMsg_LoginRsp)
		pv, ok := pmsg.(*ss.MsgLoginRsp)
		if !ok {
			return errors.New("not MsgLoginRsp")
		}
		body.LoginRsp = pv
		ss_msg.MsgBody = body
	case ss.SS_PROTO_TYPE_LOGOUT_REQ:
		body := new(ss.SSMsg_LogoutReq)
		pv, ok := pmsg.(*ss.MsgLogoutReq)
		if !ok {
			return errors.New("not MsgLogoutReq")
		}
		body.LogoutReq = pv
		ss_msg.MsgBody = body
	case ss.SS_PROTO_TYPE_LOGOUT_RSP:
		body := new(ss.SSMsg_LogoutRsp)
		pv, ok := pmsg.(*ss.MsgLogoutRsp)
		if !ok {
			return errors.New("not MsgLogoutRsp")
		}
		body.LogoutRsp = pv
		ss_msg.MsgBody = body
	case ss.SS_PROTO_TYPE_REG_REQ:
		body := new(ss.SSMsg_RegReq)
		pv, ok := pmsg.(*ss.MsgRegReq)
		if !ok {
			return errors.New("not MsgRegReq")
		}
		body.RegReq = pv
		ss_msg.MsgBody = body
	case ss.SS_PROTO_TYPE_REG_RSP:
		body := new(ss.SSMsg_RegRsp)
		pv, ok := pmsg.(*ss.MsgRegRsp)
		if !ok {
			return errors.New("not MsgRegRsp")
		}
		body.RegRsp = pv
		ss_msg.MsgBody = body
	case ss.SS_PROTO_TYPE_CREATE_GROUP_REQ:
		body := new(ss.SSMsg_CreateGroupReq)
		pv, ok := pmsg.(*ss.MsgCreateGrpReq)
		if !ok {
			return errors.New("not MsgCreateGrpReq")
		}
		body.CreateGroupReq = pv
		ss_msg.MsgBody = body
	case ss.SS_PROTO_TYPE_CREATE_GROUP_RSP:
		body := new(ss.SSMsg_CreateGroupRsp)
		pv, ok := pmsg.(*ss.MsgCreateGrpRsp)
		if !ok {
			return errors.New("not MsgCreateGrpRsp")
		}
		body.CreateGroupRsp = pv
		ss_msg.MsgBody = body
	case ss.SS_PROTO_TYPE_APPLY_GROUP_REQ:
		body := new(ss.SSMsg_ApplyGroupReq)
		pv, ok := pmsg.(*ss.MsgApplyGroupReq)
		if !ok {
			return errors.New("not MsgApplyGroupReq")
		}
		body.ApplyGroupReq = pv
		ss_msg.MsgBody = body
	case ss.SS_PROTO_TYPE_APPLY_GROUP_RSP:
		body := new(ss.SSMsg_ApplyGroupRsp)
		pv, ok := pmsg.(*ss.MsgApplyGroupRsp)
		if !ok {
			return errors.New("not MsgApplyGroupRsp")
		}
		body.ApplyGroupRsp = pv
		ss_msg.MsgBody = body
	case ss.SS_PROTO_TYPE_APPLY_GROUP_NOTIFY:
		body := new(ss.SSMsg_ApplyGroupNotify)
		pv, ok := pmsg.(*ss.MsgApplyGroupNotify)
		if !ok {
			return errors.New("not MsgApplyGroupNotify")
		}
		body.ApplyGroupNotify = pv
		ss_msg.MsgBody = body
	case ss.SS_PROTO_TYPE_APPLY_GROUP_AUDIT:
		body := new(ss.SSMsg_ApplyGroupAudit)
		pv, ok := pmsg.(*ss.MsgApplyGroupAudit)
		if !ok {
			return errors.New("not MsgApplyGroupAudit")
		}
		body.ApplyGroupAudit = pv
		ss_msg.MsgBody = body
	case ss.SS_PROTO_TYPE_FETCH_APPLY_GROUP_REQ:
		body := new(ss.SSMsg_FetchApplyReq)
		pv, ok := pmsg.(*ss.MsgFetchApplyGrpReq)
		if !ok {
			return errors.New("not MsgFetchApplyGrpReq")
		}
		body.FetchApplyReq = pv
		ss_msg.MsgBody = body
	case ss.SS_PROTO_TYPE_FETCH_APPLY_GROUP_RSP:
		body := new(ss.SSMsg_FetchApplyRsp)
		pv, ok := pmsg.(*ss.MsgFetchApplyGrpRsp)
		if !ok {
			return errors.New("not MsgFetchApplyGrpRsp")
		}
		body.FetchApplyRsp = pv
		ss_msg.MsgBody = body
	case ss.SS_PROTO_TYPE_COMMON_NOTIFY:
		body := new(ss.SSMsg_CommonNotify)
		pv, ok := pmsg.(*ss.MsgCommonNotify)
		if !ok {
			return errors.New("not MsgCommonNotify")
		}
		body.CommonNotify = pv
		ss_msg.MsgBody = body
	case ss.SS_PROTO_TYPE_FETCH_AUDIT_GROUP_REQ:
		body := new(ss.SSMsg_FetchAuditReq)
		pv, ok := pmsg.(*ss.MsgFetchAuditGrpReq)
		if !ok {
			return errors.New("not MsgFetchAuditGrpReq")
		}
		body.FetchAuditReq = pv
		ss_msg.MsgBody = body
	case ss.SS_PROTO_TYPE_FETCH_AUDIT_GROUP_RSP:
		body := new(ss.SSMsg_FetchAuditRsp)
		pv, ok := pmsg.(*ss.MsgFetchAuditGrpRsp)
		if !ok {
			return errors.New("not MsgFetchAuditGrpReq")
		}
		body.FetchAuditRsp = pv
		ss_msg.MsgBody = body
	case ss.SS_PROTO_TYPE_ENTER_GROUP_REQ:
		body := new(ss.SSMsg_EnterGroupReq)
		pv, ok := pmsg.(*ss.MsgEnterGroupReq)
		if !ok {
			return errors.New("not MsgEnterGroupReq")
		}
		body.EnterGroupReq = pv
		ss_msg.MsgBody = body
	case ss.SS_PROTO_TYPE_ENTER_GROUP_RSP:
		body := new(ss.SSMsg_EnterGroupRsp)
		pv, ok := pmsg.(*ss.MsgEnterGroupRsp)
		if !ok {
			return errors.New("not MsgEnterGroupRsp")
		}
		body.EnterGroupRsp = pv
		ss_msg.MsgBody = body
	case ss.SS_PROTO_TYPE_LOAD_GROUP_REQ:
		body := new(ss.SSMsg_LoadGroupReq)
		pv, ok := pmsg.(*ss.MsgLoadGroupReq)
		if !ok {
			return errors.New("not MsgLoadGroupReq")
		}
		body.LoadGroupReq = pv
		ss_msg.MsgBody = body
	case ss.SS_PROTO_TYPE_LOAD_GROUP_RSP:
		body := new(ss.SSMsg_LoadGroupRsp)
		pv, ok := pmsg.(*ss.MsgLoadGroupRsp)
		if !ok {
			return errors.New("not MsgLoadGroupRsp")
		}
		body.LoadGroupRsp = pv
		ss_msg.MsgBody = body
	case ss.SS_PROTO_TYPE_SEND_CHAT_REQ:
		body := new(ss.SSMsg_SendChatReq)
		pv, ok := pmsg.(*ss.MsgSendChatReq)
		if !ok {
			return errors.New("not MsgSendChatReq")
		}
		body.SendChatReq = pv
		ss_msg.MsgBody = body
	case ss.SS_PROTO_TYPE_SEND_CHAT_RSP:
		body := new(ss.SSMsg_SendChatRsp)
		pv, ok := pmsg.(*ss.MsgSendChatRsp)
		if !ok {
			return errors.New("not MsgSendChatRsp")
		}
		body.SendChatRsp = pv
		ss_msg.MsgBody = body
	case ss.SS_PROTO_TYPE_SAVE_GROUP_REQ:
		body := new(ss.SSMsg_SaveGroupReq)
		pv, ok := pmsg.(*ss.MsgSaveGroupReq)
		if !ok {
			return errors.New("not MsgSaveGroupReq")
		}
		body.SaveGroupReq = pv
		ss_msg.MsgBody = body
	case ss.SS_PROTO_TYPE_SAVE_GROUP_RSP:
		body := new(ss.SSMsg_SaveGroupRsp)
		pv, ok := pmsg.(*ss.MsgSaveGroupRsp)
		if !ok {
			return errors.New("not MsgSaveGroupRsp")
		}
		body.SaveGroupRsp = pv
		ss_msg.MsgBody = body
	case ss.SS_PROTO_TYPE_FETCH_CHAT_REQ:
		body := new(ss.SSMsg_FetchChatReq)
		pv, ok := pmsg.(*ss.MsgFetchChatReq)
		if !ok {
			return errors.New("not MsgFetchChatReq")
		}
		body.FetchChatReq = pv
		ss_msg.MsgBody = body
	case ss.SS_PROTO_TYPE_FETCH_CHAT_RSP:
		body := new(ss.SSMsg_FetchChatRsp)
		pv, ok := pmsg.(*ss.MsgFetchChatRsp)
		if !ok {
			return errors.New("not MsgFetchChatRsp")
		}
		body.FetchChatRsp = pv
		ss_msg.MsgBody = body
	case ss.SS_PROTO_TYPE_SYNC_CHAT_LIST:
		body := new(ss.SSMsg_SyncChatList)
		pv, ok := pmsg.(*ss.MsgSyncChatList)
		if !ok {
			return errors.New("not MsgSyncChatList")
		}
		body.SyncChatList = pv
		ss_msg.MsgBody = body
	case ss.SS_PROTO_TYPE_EXIT_GROUP_REQ:
		body := new(ss.SSMsg_ExitGroupReq)
		pv, ok := pmsg.(*ss.MsgExitGroupReq)
		if !ok {
			return errors.New("not MsgExitGroupReq")
		}
		body.ExitGroupReq = pv
		ss_msg.MsgBody = body
	case ss.SS_PROTO_TYPE_EXIT_GROUP_RSP:
		body := new(ss.SSMsg_ExitGroupRsp)
		pv, ok := pmsg.(*ss.MsgExitGroupRsp)
		if !ok {
			return errors.New("not MsgExitGroupRsp")
		}
		body.ExitGroupRsp = pv
		ss_msg.MsgBody = body
	case ss.SS_PROTO_TYPE_KICK_GROUP_REQ:
		body := new(ss.SSMsg_KickGroupReq)
		pv, ok := pmsg.(*ss.MsgKickGroupReq)
		if !ok {
			return errors.New("not MsgKickGroupReq")
		}
		body.KickGroupReq = pv
		ss_msg.MsgBody = body
	case ss.SS_PROTO_TYPE_KICK_GROUP_RSP:
		body := new(ss.SSMsg_KickGroupRsp)
		pv, ok := pmsg.(*ss.MsgKickGroupRsp)
		if !ok {
			return errors.New("not MsgKickGroupRsp")
		}
		body.KickGroupRsp = pv
		ss_msg.MsgBody = body
	case ss.SS_PROTO_TYPE_FETCH_OFFLINE_INFO_REQ:
		body := new(ss.SSMsg_FetchOfflineInfoReq)
		pv, ok := pmsg.(*ss.MsgFetchOfflineInfoReq)
		if !ok {
			return errors.New("not MsgFetchOfflineInfoReq")
		}
		body.FetchOfflineInfoReq = pv
		ss_msg.MsgBody = body
	case ss.SS_PROTO_TYPE_FETCH_OFFLINE_INFO_RSP:
		body := new(ss.SSMsg_FetchOfflineInfoRsp)
		pv, ok := pmsg.(*ss.MsgFetchOfflineInfoRsp)
		if !ok {
			return errors.New("not MsgFetchOfflineInfoRsp")
		}
		body.FetchOfflineInfoRsp = pv
		ss_msg.MsgBody = body
	case ss.SS_PROTO_TYPE_QUERY_GROUP_REQ:
		body := new(ss.SSMsg_QueryGroupReq)
		pv, ok := pmsg.(*ss.MsgQueryGroupReq)
		if !ok {
			return errors.New("not MsgQueryGroupReq")
		}
		body.QueryGroupReq = pv
		ss_msg.MsgBody = body
	case ss.SS_PROTO_TYPE_SYNC_GROUP_INFO:
		body := new(ss.SSMsg_SyncGroupInfo)
		pv, ok := pmsg.(*ss.MsgSyncGroupInfo)
		if !ok {
			return errors.New("not MsgSyncGroupInfo")
		}
		body.SyncGroupInfo = pv
		ss_msg.MsgBody = body
	case ss.SS_PROTO_TYPE_FETCH_USER_PROFILE_REQ:
		body := new(ss.SSMsg_FetchUserProfileReq)
		pv, ok := pmsg.(*ss.MsgFetchUserProfileReq)
		if !ok {
			return errors.New("not MsgFetchUserProfileReq")
		}
		body.FetchUserProfileReq = pv
		ss_msg.MsgBody = body
	case ss.SS_PROTO_TYPE_FETCH_USER_PROFILE_RSP:
		body := new(ss.SSMsg_FetchUserProfileRsp)
		pv, ok := pmsg.(*ss.MsgFetchUserProfileRsp)
		if !ok {
			return errors.New("not MsgFetchUserProfileRsp")
		}
		body.FetchUserProfileRsp = pv
		ss_msg.MsgBody = body
	case ss.SS_PROTO_TYPE_SAVE_USER_PROFILE_REQ:
		body := new(ss.SSMsg_SaveUserProfileReq)
		pv, ok := pmsg.(*ss.MsgSaveUserProfileReq)
		if !ok {
			return errors.New("not MsgSaveUserProfileReq")
		}
		body.SaveUserProfileReq = pv
		ss_msg.MsgBody = body
	case ss.SS_PROTO_TYPE_CHG_GROUP_ATTR_REQ:
		body := new(ss.SSMsg_ChgGroupAttrReq)
		pv, ok := pmsg.(*ss.MsgChgGroupAttrReq)
		if !ok {
			return errors.New("not MsgChgGroupAttrReq")
		}
		body.ChgGroupAttrReq = pv
		ss_msg.MsgBody = body
	case ss.SS_PROTO_TYPE_CHG_GROUP_ATTR_RSP:
		body := new(ss.SSMsg_ChgGroupAttrRsp)
		pv, ok := pmsg.(*ss.MsgChgGroupAttrRsp)
		if !ok {
			return errors.New("not MsgChgGroupAttrRsp")
		}
		body.ChgGroupAttrRsp = pv
		ss_msg.MsgBody = body
	case ss.SS_PROTO_TYPE_GROUP_GROUND_REQ:
		body := new(ss.SSMsg_GroupGroundReq)
		pv, ok := pmsg.(*ss.MsgGroupGroudReq)
		if !ok {
			return errors.New("not MsgGroupGroudReq")
		}
		body.GroupGroundReq = pv
		ss_msg.MsgBody = body
	case ss.SS_PROTO_TYPE_GROUP_GROUND_RSP:
		body := new(ss.SSMsg_GroupGroundRsp)
		pv, ok := pmsg.(*ss.MsgGroupGroudRsp)
		if !ok {
			return errors.New("not MsgGroupGroudRsp")
		}
		body.GroupGroundRsp = pv
		ss_msg.MsgBody = body
	default:
		return errors.New(fmt.Sprintf("disp proto:%d not handled", proto))
	}

	return nil
}
