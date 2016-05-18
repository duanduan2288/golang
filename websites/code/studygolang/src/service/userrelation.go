// Copyright 2013 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author：polaris   studygolang@gmail.com

package service

import (
	//"fmt"
	//"html/template"
	"logger"
	"model"
	//"net/url"
	//"regexp"
	//"strconv"
	//"time"
	"util"
)

//获取某个人的粉丝列表
func FindUserFans(uid string) map[int]*model.User {
	fansList, err := model.NewUserRelation().Where("to_user_id=" + uid).FindAll()
	if err != nil {
		logger.Errorln("fanslist service FindUserFans Error:", err)
		return nil
	}

	uids := util.Models2Intslice(fansList, "FromUserTo")

	// 获得用户信息
	fans := GetUserInfos(uids)
	return fans
}

// 关注
func Subscribe(from_user_id, to_user_id int) (err error) {

	userRelation := model.NewUserRelation()

	userRelation.FromUserId = from_user_id
	userRelation.ToUserId = to_user_id
	userRelation.RelType = 1
	userRelation.Created = util.TimeNow()
	_, err = userRelation.Insert()
	return
}

//取消关注
func Unsubscribe(from_user_id, to_user_id int) (err error) {
	err = model.NewUserRelation().Where("to_user_id=? and from_user_id=? and rel_type=1", to_user_id, from_user_id).Delete()
	return err
}

//是否已经关注
func IsFans(from_user_id, to_user_id int) bool {
	userRelation, err := model.NewUserRelation().Where("to_user_id=? and from_user_id=? and rel_type=1", to_user_id, from_user_id).Find()
	if err != nil {
		return false
	}

	if userRelation.Id > 0 {
		return true
	}

	return false
}
