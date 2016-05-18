// Copyright 2014 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author：polaris	studygolang@gmail.com

package controller

import (
	"fmt"
	"net/http"
	"strconv"

	"filter"
	"model"
	"service"
	"util"

	"github.com/studygolang/mux"
)

// 在需要评论（喜欢）且要回调的地方注册评论（喜欢）对象
func init() {
	// 注册评论（喜欢）对象
	service.RegisterCommentObject(model.TYPE_PROJECT, service.ProjectComment{})
	service.RegisterLikeObject(model.TYPE_PROJECT, service.ProjectLike{})
}

// 开源项目列表页
// uri: /projects
func ProjectsHandler(rw http.ResponseWriter, req *http.Request) {

	lastId := req.FormValue("lastid")
	if lastId == "" {
		lastId = "0"
	}
	limit := 20
	projects := service.FindProjects(lastId, "25")

	num := len(projects)

	if num == 0 {
		if lastId == "0" {
			util.Redirect(rw, req, "/")
			return
		} else {
			util.Redirect(rw, req, "/projects")
			return
		}
	}

	//上一页下一页
	var (
		prevId, nextId     int
		has_prev, has_next bool
	)
	if lastId != "0" {
		prevId, _ := strconv.Atoi(lastId)

		if prevId-projects[0].Id > 5 {
			has_prev = false
		} else {
			prevId += limit
			has_prev = true
		}
	}

	if num < limit {
		nextId = projects[num-1].Id
	} else {
		projects = projects[:limit]
		nextId = projects[limit-1].Id
		has_next = true
	}

	pageInfo := map[string]interface{}{
		"has_next": has_next,
		"has_prev": has_prev,
		"prev_id":  prevId,
		"next_id":  nextId,
	}
	// 获取当前用户喜欢对象信息
	user, ok := filter.CurrentUser(req)
	var likeFlags map[int]int
	if ok {
		uid := user["uid"].(int)
		likeFlags, _ = service.FindUserLikeObjects(uid, model.TYPE_PROJECT, projects[0].Id, nextId)
	}

	req.Form.Set(filter.CONTENT_TPL_KEY, "/template/projects/list.html")

	// 设置模板数据
	filter.SetData(req, map[string]interface{}{"projects": projects, "activeProjects": "active", "page": pageInfo, "likeflags": likeFlags})
}

// 新建项目
// uri: /project/new{json:(|.json)}
func NewProjectHandler(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	name := req.FormValue("name")

	if vars["json"] == "" || name == "" || req.Method != "POST" {
		project := model.NewOpenProject()
		filter.SetData(req, map[string]interface{}{"project": project})
		req.Form.Set(filter.CONTENT_TPL_KEY, "/template/projects/new.html")
		return
	}
	user, _ := filter.CurrentUser(req)
	err := service.PublishProject(user, req.Form)
	if err != nil {
		fmt.Fprint(rw, `{"ok":0,"error":`+err.Error()+`}`)
		return
	}

	fmt.Fprint(rw, `{"ok":1,"msg":"发布成功"}`)
}

// 修改项目
// uri: /project/modify{json:(|.json)}
func ModifyProjectHandler(rw http.ResponseWriter, req *http.Request) {
	id := req.FormValue("id")
	vars := mux.Vars(req)
	if id == "" {
		util.Redirect(rw, req, "/projects")
		return
	}

	if vars["json"] == "" || req.Method != "POST" {

		project := service.FindProject(id)

		filter.SetData(req, map[string]interface{}{"project": project})
		req.Form.Set(filter.CONTENT_TPL_KEY, "/template/projects/new.html")
		return
	}

	user, _ := filter.CurrentUser(req)
	err := service.PublishProject(user, req.Form)
	if err != nil {
		fmt.Fprint(rw, `{"ok":0,"error":`+err.Error()+`}`)
		return
	}

	fmt.Fprint(rw, `{"ok":1,"msg":"修改成功"}`)
}

// 项目详情
// uri: /p/{uniq}
func ProjectDetailHandler(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id := vars["uniq"]
	project := service.FindProject(id)
	if project == nil {
		util.Redirect(rw, req, "/projects")
	}
	hadLike := 0
	hadCollect := 0
	user, ok := filter.CurrentUser(req)
	if ok {
		uid := user["uid"].(int)
		hadLike = service.HadLike(uid, project.Id, model.TYPE_PROJECT)
		hadCollect = service.HadFavorite(uid, project.Id, model.TYPE_PROJECT)
	}

	service.Views.Incr(req, model.TYPE_PROJECT, project.Id)

	project.Viewnum++

	filter.SetData(req, map[string]interface{}{"project": project, "likeflag": hadLike, "hadcollect": hadCollect})
	req.Form.Set(filter.CONTENT_TPL_KEY, "/template/projects/detail.html,/template/common/comment.html")
}

// 检测 uri 对应的项目是否存在(验证，true表示不存在；false表示存在)
// uri: /project/uri.json
func ProjectUriHandler(rw http.ResponseWriter, req *http.Request) {
	uri := req.FormValue("uri")
	if uri == "" {
		fmt.Fprint(rw, `true`)
		return
	}

	if service.ProjectUriExists(uri) {
		fmt.Fprint(rw, `false`)
		return
	}
	fmt.Fprint(rw, `true`)
}
