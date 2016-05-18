// Copyright 2013 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author：polaris	studygolang@gmail.com

package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"filter"
	"logger"
	"service"
	"util"

	"github.com/studygolang/mux"
)

// 评论（或回复）
// uri: /comment/{objid:[0-9]+}.json
func CommentHandler(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	objid := vars["objid"]

	user, ok := filter.CurrentUser(req)
	if !ok {
		fmt.Fprint(rw, `{"errno":1,"error":"用户没有登录"}`)
		return
	}
	uid := user["uid"].(int)
	comment, err := service.PostComment(uid, util.MustInt(objid), req.Form)
	if nil != err {
		fmt.Fprint(rw, `{"errno":1,"error":`+err.Error()+`}`)
		return
	}
	buf, err := json.Marshal(comment)
	if err != nil {
		fmt.Fprint(rw, `{"errno":1,"error":"数据错误"}`)
		return
	}
	fmt.Fprint(rw, `{"errno":0,"data":`+string(buf)+`}`)

}

// 获取某对象的评论信息
// uri: /object/comments.json
func ObjectCommentsHandler(rw http.ResponseWriter, req *http.Request) {
	objid := req.FormValue("objid")
	objtype := req.FormValue("objtype")

	commentList, err := service.FindObjectComments(objid, objtype)

	uids := util.Models2Intslice(commentList, "Uid")
	users := service.GetUserInfos(uids)

	result := map[string]interface{}{
		"comments": commentList,
	}

	// json encode 不支持 map[int]...
	for uid, user := range users {
		result[strconv.Itoa(uid)] = user
	}

	buf, err := json.Marshal(result)

	if err != nil {
		logger.Errorln("[RecentCommentHandler] json.marshal error:", err)
		fmt.Fprint(rw, `{"ok": 0, "error":"解析json出错"}`)
		return
	}
	fmt.Fprint(rw, `{"ok": 1, "data":`+string(buf)+`}`)
}
