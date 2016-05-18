// Copyright 2013 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author：polaris	studygolang@gmail.com

package admin

import (
	"filter"
	"net/http"
	"service"
	"strconv"
)

// 修改评论（只允许后台管理员修改评论内容）
func ModifyCommentHandler(rw http.ResponseWriter, req *http.Request) {
	var data = make(map[string]interface{})

	errMsg, err := service.ModifyComment(req.PostFormValue("cid"), req.PostFormValue("content"))
	if err != nil {
		data["ok"] = 0
		data["error"] = errMsg
	} else {
		data["ok"] = 1
		data["msg"] = "修改成功"
	}

	filter.SetData(req, data)
}

// 删除评论
func DelCommentHandler(rw http.ResponseWriter, req *http.Request) {
	var data = make(map[string]interface{})
	cid := req.FormValue("cid")

	if _, err := strconv.Atoi(cid); err != nil {
		data["ok"] = 0
		data["error"] = "不是整数"
	} else {
		errMsg, err := service.DelComment(cid, 2)
		if err != nil {
			data["ok"] = 0
			data["error"] = errMsg
		} else {
			data["ok"] = 1
			data["error"] = "删除成功"
		}
	}
	filter.SetData(req, data)
}
