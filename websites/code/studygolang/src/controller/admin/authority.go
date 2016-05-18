// Copyright 2013 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author：polaris	studygolang@gmail.com

package admin

import (
	"encoding/json"
	"filter"
	"logger"
	"net/http"
	"service"
	"strconv"
)

// 所有权限（分页）
func AuthListHandler(rw http.ResponseWriter, req *http.Request) {

	curPage, limit := parsePage(req)
	total := len(service.Authorities)

	if total == 0 {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"datalist":   service.Authorities[(curPage-1)*limit : curPage*limit],
		"total":      total,
		"totalPages": (total + limit - 1) / limit,
		"page":       curPage,
		"limit":      limit,
	}
	req.Form.Set(filter.CONTENT_TPL_KEY, "/template/admin/authority/list.html,/template/admin/authority/query.html")
	filter.SetData(req, data)
}

func AuthQueryHandler(rw http.ResponseWriter, req *http.Request) {
	curPage, limit := parsePage(req)

	conds := parseConds(req, []string{"route", "name"})

	authorities, total := service.FindAuthoritiesByPage(conds, curPage, limit)

	if authorities == nil {
		logger.Errorln("[AuthQueryHandler]sql find error")
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"datalist":   authorities,
		"total":      total,
		"totalPages": (total + limit - 1) / limit,
		"page":       curPage,
		"limit":      limit,
	}

	// 设置内容模板
	req.Form.Set(filter.CONTENT_TPL_KEY, "/template/admin/authority/query.html")
	filter.SetData(req, data)
}

func NewAuthorityHandler(rw http.ResponseWriter, req *http.Request) {
	var data = make(map[string]interface{})

	if req.PostFormValue("submit") == "1" {
		user, _ := filter.CurrentUser(req)
		username := user["username"].(string)

		errmsg, err := service.SaveAuthority(req.PostForm, username)
		if err != nil {
			data["ok"] = 0
			data["error"] = errmsg
		} else {
			data["ok"] = 1
			data["error"] = "添加成功"
		}

	} else {
		menu1, menu2 := service.GetMenus()
		allmenu2, _ := json.Marshal(menu2)
		req.Form.Set(filter.CONTENT_TPL_KEY, "/template/admin/authority/new.html")
		data["allmenu1"] = menu1
		data["allmenu2"] = string(allmenu2)
	}

	filter.SetData(req, data)
}

func ModifyAuthorityHandler(rw http.ResponseWriter, req *http.Request) {
	var data = make(map[string]interface{})
	if req.PostFormValue("submit") == "1" {
		user, _ := filter.CurrentUser(req)
		username := user["username"].(string)
		errMsg, err := service.SaveAuthority(req.PostForm, username)
		if err != nil {
			data["ok"] = 0
			data["error"] = errMsg
		} else {
			data["ok"] = 1
			data["error"] = "修改成功"
		}
	} else {
		aid := req.FormValue("aid")
		if aid == "" {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		authority := service.FindAuthority(aid)
		if authority == nil {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		menu1, menu2 := service.GetMenus()
		allmenu2, _ := json.Marshal(menu2)
		req.Form.Set(filter.CONTENT_TPL_KEY, "/template/admin/authority/new.html")
		data["allmenu1"] = menu1
		data["allmenu2"] = string(allmenu2)
		data["authority"] = authority
	}

	filter.SetData(req, data)
}

func DelAuthorityHandler(rw http.ResponseWriter, req *http.Request) {
	var data = make(map[string]interface{})
	aid := req.FormValue("aid")

	if _, err := strconv.Atoi(aid); err != nil {
		data["ok"] = 0
		data["error"] = "aid不是整型"

		filter.SetData(req, data)
		return
	}

	if err := service.DelAuthority(aid); err != nil {
		data["ok"] = 0
		data["error"] = "删除失败！"
	} else {
		data["ok"] = 1
		data["msg"] = "删除成功！"
	}

	filter.SetData(req, data)
}
