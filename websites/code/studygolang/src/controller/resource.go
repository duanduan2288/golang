// Copyright 2013 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author：polaris	studygolang@gmail.com

package controller

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"filter"
	"github.com/studygolang/mux"
	"model"
	"service"
	"util"
)

// 在需要评论（喜欢）且要回调的地方注册评论（喜欢）对象
func init() {
	// 注册评论（喜欢）对象
	service.RegisterCommentObject(model.TYPE_RESOURCE, service.ResourceComment{})
	service.RegisterLikeObject(model.TYPE_RESOURCE, service.ResourceLike{})
}

// 资源索引页
// uri: /resources
func ResIndexHandler(rw http.ResponseWriter, req *http.Request) {
	util.Redirect(rw, req, "/resources/cat/1")
}

// 某个分类的资源列表
// uri: /resources/cat/{catid:[0-9]+}
func CatResourcesHandler(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	catid := vars["catid"]

	page, _ := strconv.Atoi(req.FormValue("p"))
	if page == 0 {
		page = 1
	}

	resources, total := service.FindResourcesByCatid(catid, page)
	pageHtml := service.GetPageHtml(page, total, req.URL.Path)

	req.Form.Set(filter.CONTENT_TPL_KEY, "/template/resources/index.html")
	filter.SetData(req, map[string]interface{}{"activeResources": "active", "resources": resources, "categories": service.AllCategory, "page": template.HTML(pageHtml), "curCatid": util.MustInt(catid)})
}

// 某个资源详细页
// uri: /resources/{id:[0-9]+}
func ResourceDetailHandler(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	resource, comments := service.FindResource(vars["id"])

	if len(resource) == 0 {
		util.Redirect(rw, req, "/resources")
		return
	}

	likeFlag := 0
	hadCollect := 0
	user, ok := filter.CurrentUser(req)
	if ok {
		uid := user["uid"].(int)
		id := resource["id"].(int)
		likeFlag = service.HadLike(uid, id, model.TYPE_RESOURCE)
		hadCollect = service.HadFavorite(uid, id, model.TYPE_RESOURCE)
	}

	service.Views.Incr(req, model.TYPE_RESOURCE, util.MustInt(vars["id"]))

	req.Form.Set(filter.CONTENT_TPL_KEY, "/template/resources/detail.html,/template/common/comment.html")
	filter.SetData(req, map[string]interface{}{"activeResources": "active", "resource": resource, "comments": comments, "likeflag": likeFlag, "hadcollect": hadCollect})
}

// 发布新资源
// uri: /resources/new{json:(|.json)}
func NewResourceHandler(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	title := req.PostFormValue("title")
	// 请求新建资源页面
	if title == "" || req.Method != "POST" || vars["json"] == "" {
		req.Form.Set(filter.CONTENT_TPL_KEY, "/template/resources/new.html")
		filter.SetData(req, map[string]interface{}{"activeResources": "active", "categories": service.AllCategory})
		return
	}

	errMsg := ""
	resForm := req.PostFormValue("form")
	if resForm == model.LinkForm {
		if req.PostFormValue("url") == "" {
			errMsg = "url不能为空"
		}
	} else {
		if req.PostFormValue("content") == "" {
			errMsg = "内容不能为空"
		}
	}
	if errMsg != "" {
		fmt.Fprint(rw, `{"ok": 0, "error":"`+errMsg+`"}`)
		return
	}

	user, _ := filter.CurrentUser(req)
	err := service.PublishResource(user, req.PostForm)
	if err != nil {
		fmt.Fprint(rw, `{"ok": 0, "error":"内部服务错误，请稍候再试！"}`)
		return
	}

	fmt.Fprint(rw, `{"ok": 1, "data":""}`)
}

// 修改資源
// uri: /resources/modify{json:(|.json)}
func ModifyResourceHandler(rw http.ResponseWriter, req *http.Request) {
	id := req.FormValue("id")
	if id == "" {
		util.Redirect(rw, req, "/resources")
		return
	}

	vars := mux.Vars(req)
	// 请求编辑資源页面
	if req.Method != "POST" || vars["json"] == "" {
		resource := service.FindResourceById(id)
		req.Form.Set(filter.CONTENT_TPL_KEY, "/template/resources/new.html")
		filter.SetData(req, map[string]interface{}{"resource": resource, "activeResources": "active", "categories": service.AllCategory})
		return
	}

	user, _ := filter.CurrentUser(req)
	err := service.PublishResource(user, req.PostForm)
	if err != nil {
		if err == service.NotModifyAuthorityErr {
			rw.WriteHeader(http.StatusForbidden)
			return
		}
		fmt.Fprint(rw, `{"ok": 0, "error":"内部服务错误！"}`)
		return
	}
	fmt.Fprint(rw, `{"ok": 1, "data":""}`)
}
