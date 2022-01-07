package open_im_sdk

import (
	"encoding/json"
)

type OnGroupListener interface {
	OnMemberEnter(groupId string, memberList string)
	OnMemberLeave(groupId string, member string)
	OnMemberInvited(groupId string, opUser string, memberList string)
	OnMemberKicked(groupId string, opUser string, memberList string)
	OnGroupCreated(groupId string)
	OnGroupInfoChanged(groupId string, groupInfo string)
	OnReceiveJoinApplication(groupId string, member string, opReason string)
	OnApplicationProcessed(groupId string, opUser string, AgreeOrReject int32, opReason string)
}

func (u *UserRelated) SetGroupListener(callback OnGroupListener) {
	if callback == nil {
		sdkLog("callback null")
		return
	}
	u.listener = callback
	sdkLog("SetGroupListener ", callback)
}

func (u *UserRelated) CreateGroup(callback Base, groupBaseInfo string, memberList string, operationID string) {
	if callback == nil {
		sdkLog("callback is nil")
		return
	}
	go func() {
		NewInfo(operationID, "CreateGroup args: ", groupBaseInfo, memberList)
		var unmarshalCreateGroupBaseInfoParam CreateGroupBaseInfoParam
		u.jsonUnmarshalAndArgsValidate(groupBaseInfo, &unmarshalCreateGroupBaseInfoParam, callback, operationID)
		var unmarshalCreateGroupMemberRoleParam CreateGroupMemberRoleParam
		u.jsonUnmarshalAndArgsValidate(memberList, &unmarshalCreateGroupMemberRoleParam, callback, operationID)
		result := u.createGroup(unmarshalCreateGroupBaseInfoParam, unmarshalCreateGroupMemberRoleParam, operationID)
		callback.OnSuccess(structToJsonString(result))
		NewInfo(operationID, "CreateGroup callback: ", structToJsonString(result))
	}()
}

func (u *UserRelated) JoinGroup(groupId, message string, callback Base) {
	if callback == nil {
		sdkLog("callback is nil")
		return
	}
	go func() {
		sdkLog(".................joinGroup begin ...............", groupId, message)
		err := u.joinGroup(groupId, message)
		if err != nil {
			sdkLog("............joinGroup failed............ ", groupId, message, err.Error())
			callback.OnError(ErrCodeGroup, err.Error())
			return
		}
		sdkLog("...........joinGroup end, callback............... ", groupId, message)
		callback.OnSuccess(DeFaultSuccessMsg)
	}()
}

func (u *UserRelated) QuitGroup(groupId string, callback Base) {
	if callback == nil {
		sdkLog("callback is nil")
		return
	}
	go func() {
		sdkLog("............quitGroup begin...............")
		err := u.quitGroup(groupId)
		if err != nil {
			sdkLog(".........quitGroup failed.............", groupId, err.Error())
			callback.OnError(ErrCodeGroup, err.Error())
			return
		}
		sdkLog("..........quitGroup end callback...........", groupId)
		callback.OnSuccess(DeFaultSuccessMsg)
	}()
}

func (u *UserRelated) GetJoinedGroupList(callback Base) {
	if callback == nil {
		sdkLog("Base callback is nil")
		return
	}
	go func() {
		groupInfoList, err := u.getJoinedGroupListFromLocal()
		if err != nil {
			callback.OnError(ErrCodeGroup, err.Error())
			return
		}
		sdkLog("getJoinedGroupListFromLocal, ", groupInfoList)

		for i, v := range groupInfoList {
			g, err := u._getGroupInfoByGroupID(v.GroupId)
			if err != nil {
				sdkLog("findLocalGroupOwnerByGroupId failed,  ", err.Error(), v.GroupId)
				continue
			}
			sdkLog("findLocalGroupOwnerByGroupId ", v.GroupId, g.OwnerUserID)
			v.OwnerId = g.OwnerUserID
			sdkLog("getLocalGroupMemberNumByGroupId ", v.GroupId, g.MemberCount)
			v.MemberCount = uint32(g.MemberCount)
			groupInfoList[i] = v
		}

		jsonGroupInfoList, err := json.Marshal(groupInfoList)
		if err != nil {
			sdkLog("marshal failed, ", err.Error())
			callback.OnError(ErrCodeGroup, err.Error())
			return
		}
		sdkLog("jsonGroupInfoList, ", string(jsonGroupInfoList))
		callback.OnSuccess(string(jsonGroupInfoList))
	}()
}

func (u *UserRelated) GetGroupsInfo(groupIdList string, callback Base) {
	if callback == nil {
		sdkLog("Base callback is nil")
		return
	}
	go func() {
		var sctgroupIdList []string
		err := json.Unmarshal([]byte(groupIdList), &sctgroupIdList)
		if err != nil {
			callback.OnError(ErrCodeGroup, err.Error())
			return
		}

		groupsInfoList, err := u.getGroupsInfo(sctgroupIdList)
		if err != nil {
			callback.OnError(ErrCodeGroup, err.Error())
			return
		}
		jsonList, err := json.Marshal(groupsInfoList)
		if err != nil {
			callback.OnError(ErrCodeGroup, err.Error())
			return
		}
		callback.OnSuccess(string(jsonList))
	}()

}

func (u *UserRelated) SetGroupInfo(jsonGroupInfo string, callback Base) {
	if callback == nil {
		sdkLog("callback is nil")
		return
	}
	go func() {
		sdkLog("............SetGroupInfo begin...................")
		var newGroupInfo setGroupInfoReq
		err := json.Unmarshal([]byte(jsonGroupInfo), &newGroupInfo)
		if err != nil {
			sdkLog("............Unmarshal failed................", err.Error())
			callback.OnError(ErrCodeGroup, err.Error())
			return
		}
		err = u.setGroupInfo(newGroupInfo)
		if err != nil {
			sdkLog("..........setGroupInfo failed........... ", err.Error())
			callback.OnError(ErrCodeGroup, err.Error())
			return
		}
		sdkLog(".........setGroupInfo end, callback...............", jsonGroupInfo)
		callback.OnSuccess(DeFaultSuccessMsg)
	}()
}

func (u *UserRelated) GetGroupMemberList(groupId string, filter int32, next int32, callback Base) {
	if callback == nil {
		sdkLog("Base callback is nil")
		return
	}
	go func() {
		n, groupMemberResult, err := u.getGroupMemberListFromLocal(groupId, filter, next)
		if err != nil {
			callback.OnError(ErrCodeGroup, err.Error())
			return
		}
		sdkLog("getGroupMemberListFromLocal ", groupId, filter, next, groupMemberResult)

		var result getGroupMemberListResult
		if groupMemberResult == nil {
			groupMemberResult = make([]groupMemberFullInfo, 0)
		}
		result.Data = groupMemberResult
		result.NextSeq = n
		jsonGroupMemberResult, err := json.Marshal(result)
		if err != nil {
			sdkLog("marshal failed, ", err.Error())
			callback.OnError(ErrCodeGroup, err.Error())
			return
		}
		sdkLog("jsonGroupMemberResult ", string(jsonGroupMemberResult))
		callback.OnSuccess(string(jsonGroupMemberResult))
	}()
}

func (u *UserRelated) GetGroupMembersInfo(groupId string, userList string, callback Base) {
	if callback == nil {
		sdkLog("Base callback is nil")
		return
	}
	go func() {
		var sctmemberList []string
		err := json.Unmarshal([]byte(userList), &sctmemberList)
		if err != nil {
			callback.OnError(ErrCodeGroup, err.Error())
			return
		}
		sdkLog("GetGroupMembersInfo args: ", groupId, userList)
		r, err := u.getGroupMembersInfoFromLocal(groupId, sctmemberList)
		if err != nil {
			sdkLog("getGroupMembersInfoFromLocal failed", groupId, err.Error())
			callback.OnError(ErrCodeGroup, err.Error())
			return
		}
		sdkLog("getGroupMembersInfoFromLocal, ", groupId, sctmemberList, r)

		jsonResult, err := json.Marshal(r)
		if err != nil {
			sdkLog("marshal failed, ", err.Error())
			callback.OnError(ErrCodeGroup, err.Error())
			return
		}
		sdkLog("jsonResult", string(jsonResult))
		callback.OnSuccess(string(jsonResult))
	}()
}

func (u *UserRelated) KickGroupMember(groupId string, reason string, userList string, callback Base) {
	if callback == nil {
		sdkLog("callback null")
		return
	}
	go func() {
		var sctMemberList []string
		err := json.Unmarshal([]byte(userList), &sctMemberList)
		if err != nil {
			sdkLog("unmarshal failed, ", err.Error(), userList)
			callback.OnError(ErrCodeGroup, err.Error())
			return
		}
		r, err := u.kickGroupMember(groupId, sctMemberList, reason)
		if err != nil {
			callback.OnError(ErrCodeGroup, err.Error())
			return
		}
		sdkLog("kickGroupMember, ", groupId, sctMemberList, reason)

		jsonResult, err := json.Marshal(r)
		if err != nil {
			sdkLog("marshal failed, ", err.Error())
			callback.OnError(ErrCodeGroup, err.Error())
			return
		}
		sdkLog("KickGroupMember, req: ", groupId, userList, reason, "resp: ", string(jsonResult))
		callback.OnSuccess(string(jsonResult))
	}()
}

func (u *UserRelated) TransferGroupOwner(groupId, userId string, callback Base) {
	if callback == nil {
		return
	}
	go func() {
		err := u.transferGroupOwner(groupId, userId)
		if err != nil {
			callback.OnError(ErrCodeGroup, err.Error())
			return
		}
		u.syncGroupRequest()
		u.syncGroupMemberByGroupId(groupId)
		callback.OnSuccess(DeFaultSuccessMsg)
	}()
}

func (u *UserRelated) InviteUserToGroup(groupId, reason string, userList string, callback Base) {
	if callback == nil {
		sdkLog("callbak null")
		return
	}
	go func() {
		var sctUserList []string
		err := json.Unmarshal([]byte(userList), &sctUserList)
		if err != nil {
			sdkLog("unmarshal failed, ", err.Error())
			callback.OnError(ErrCodeGroup, err.Error())
			return
		}

		result, err := u.inviteUserToGroup(groupId, reason, sctUserList)
		if err != nil {
			callback.OnError(ErrCodeGroup, err.Error())
			return
		}
		sdkLog("inviteUserToGroup, ", groupId, reason, sctUserList)

		jsonResult, err := json.Marshal(result)
		if err != nil {
			sdkLog("marshal failed, ", err.Error())
			callback.OnError(ErrCodeGroup, err.Error())
			return
		}
		sdkLog("InviteUserToGroup, req: ", groupId, reason, userList, "resp: ", string(jsonResult))

		callback.OnSuccess(string(jsonResult))
	}()

}

func (u *UserRelated) GetGroupApplicationList(callback Base) {
	if callback == nil {
		return
	}
	go func() {
		r, err := u.getGroupApplicationList()
		if err != nil {
			sdkLog("getGroupApplicationList faild, ", err.Error())
			callback.OnError(ErrCodeGroup, err.Error())
			return
		}
		jsonResult, err := json.Marshal(r)
		if err != nil {
			sdkLog("getGroupApplicationList faild, ", err.Error())
			callback.OnError(ErrCodeGroup, err.Error())
			return
		}
		callback.OnSuccess(string(jsonResult))
		return
	}()
}

/*
func (u *UserRelated) TsetGetGroupApplicationList(callback Base) string {
	if callback == nil {
		return ""
	}

	r, err := u.getGroupApplicationList()
	if err != nil {
		sdkLog("getGroupApplicationList faild, ", err.Error())
		callback.OnError(ErrCodeGroup, err.Error())
		return ""
	}
	jsonResult, err := json.Marshal(r)
	if err != nil {
		sdkLog("getGroupApplicationList faild, ", err.Error())
		callback.OnError(ErrCodeGroup, err.Error())
		return ""
	}
	callback.OnSuccess(string(jsonResult))
	return string(jsonResult)

}*/

func (u *UserRelated) AcceptGroupApplication(application, reason string, callback Base) {
	if callback == nil {
		return
	}
	go func() {
		var sctApplication GroupReqListInfo
		err := json.Unmarshal([]byte(application), &sctApplication)
		if err != nil {
			sdkLog("Unmarshal, ", err.Error())
			callback.OnError(ErrCodeGroup, err.Error())
			return
		}

		var access accessOrRefuseGroupApplicationReq
		access.OperationID = operationIDGenerator()
		access.GroupId = sctApplication.GroupID
		access.FromUser = sctApplication.FromUserID
		access.FromUserNickName = sctApplication.FromUserNickname
		access.FromUserFaceUrl = sctApplication.FromUserFaceUrl
		access.ToUser = sctApplication.ToUserID
		access.ToUserNickname = sctApplication.ToUserNickname
		access.ToUserFaceUrl = sctApplication.ToUserFaceUrl
		access.AddTime = sctApplication.AddTime
		access.RequestMsg = sctApplication.RequestMsg
		access.HandledMsg = reason
		access.Type = sctApplication.Type
		access.HandleStatus = 2
		access.HandleResult = 1

		err = u.acceptGroupApplication(&access)
		if err != nil {
			sdkLog("acceptGroupApplication, ", err.Error())
			callback.OnError(ErrCodeGroup, err.Error())
			return
		}
		u.syncGroupRequest()
		u.syncGroupMemberByGroupId(sctApplication.GroupID)
		callback.OnSuccess(DeFaultSuccessMsg)
	}()
}

func (u *UserRelated) RefuseGroupApplication(application, reason string, callback Base) {
	if callback == nil {
		return
	}
	go func() {
		var sctApplication GroupReqListInfo
		err := json.Unmarshal([]byte(application), &sctApplication)
		if err != nil {
			sdkLog("Unmarshal, ", err.Error())
			callback.OnError(ErrCodeGroup, err.Error())
			return
		}

		var access accessOrRefuseGroupApplicationReq
		access.OperationID = operationIDGenerator()
		access.GroupId = sctApplication.GroupID
		access.FromUser = sctApplication.FromUserID
		access.FromUserNickName = sctApplication.FromUserNickname
		access.FromUserFaceUrl = sctApplication.FromUserFaceUrl
		access.ToUser = sctApplication.ToUserID
		access.ToUserNickname = sctApplication.ToUserNickname
		access.ToUserFaceUrl = sctApplication.ToUserFaceUrl
		access.AddTime = sctApplication.AddTime
		access.RequestMsg = sctApplication.RequestMsg
		access.HandledMsg = reason
		access.Type = sctApplication.Type
		access.HandleStatus = 2
		access.HandleResult = 0

		err = u.refuseGroupApplication(&access)
		if err != nil {
			sdkLog("refuseGroupApplication, ", err.Error())
			callback.OnError(ErrCodeGroup, err.Error())
			return
		}
		u.syncGroupRequest()
		callback.OnSuccess(DeFaultSuccessMsg)
	}()

}
