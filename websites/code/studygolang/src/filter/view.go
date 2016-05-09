// Copyright 2013 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author：polaris	studygolang@gmail.com

package filter

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"config"
	"github.com/gorilla/context"
	"github.com/studygolang/mux"
	"logger"
	"service"
	"util"
	"util/version"
)

// 自定义模板函数
var funcMap = template.FuncMap{
	// 获取gravatar头像
	"gravatar": util.Gravatar,
	// 转为前端显示需要的时间格式
	"formatTime": func(i interface{}) string {
		ctime, ok := i.(string)
		if !ok {
			return ""
		}
		t, _ := time.Parse("2006-01-02 15:04:05", ctime)
		return t.Format(time.RFC3339) + "+08:00"
	},
	"substring": util.Substring,
	"add": func(nums ...interface{}) int {
		total := 0
		for _, num := range nums {
			if n, ok := num.(int); ok {
				total += n
			}
		}
		return total
	},
	"explode": func(s, sep string) []string {
		return strings.Split(s, sep)
	},
	"noescape": func(s string) template.HTML {
		return template.HTML(s)
	},
}

// 保存模板路径的key
const CONTENT_TPL_KEY = "__content_tpl"

// 页面展示 过滤器
type ViewFilter struct {
	commonHtmlFiles []string // 通用的html文件
	baseTplName     string   // 第一个基础模板的名称
	isBackView      bool     // 是否是后端 view 过滤器

	// "继承"空实现
	*mux.EmptyFilter
}

func NewViewFilter(isBackView bool, files ...string) *ViewFilter {
	viewFilter := new(ViewFilter)
	if len(files) == 0 {
		// 默认使用前端通用模板
		viewFilter.commonHtmlFiles = []string{config.ROOT + "/template/common/layout.html"}
		viewFilter.baseTplName = "layout.html"
	} else {
		viewFilter.commonHtmlFiles = files
		viewFilter.baseTplName = filepath.Base(files[0])
	}

	viewFilter.isBackView = isBackView

	return viewFilter
}

func (this *ViewFilter) PreFilter(rw http.ResponseWriter, req *http.Request) bool {
	logger.Debugln(req.RequestURI)

	// ajax请求头设置
	if strings.HasSuffix(req.URL.Path, ".json") || req.FormValue("format") == "json" {
		setData(req, formatkey, "json")
		rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	} else if strings.HasSuffix(req.URL.Path, ".html") {
		setData(req, formatkey, "ajaxhtml")
	}

	return true
}

// 在逻辑处理完之后，最后展示页面
func (this *ViewFilter) PostFilter(rw http.ResponseWriter, req *http.Request) bool {
	data := GetData(req)

	format := "html"
	if formatInter := getData(req, formatkey); formatInter != nil {
		format = formatInter.(string)
	}

	switch format {
	case "json":
		if len(data) != 0 {
			result, err := json.Marshal(data)
			if err != nil {
				logger.Errorf("json.Marshal error：[%q] %s\n", req.RequestURI, err)
				return false
			}
			fmt.Fprint(rw, string(result))
		}
	case "ajaxhtml":
		contentHtml := req.FormValue(CONTENT_TPL_KEY)
		if contentHtml == "" {
			return true
		}

		contentHtml = "/template/admin/common_query.html," + contentHtml
		contentHtmls := strings.Split(contentHtml, ",")
		for i, contentHtml := range contentHtmls {
			contentHtmls[i] = config.ROOT + strings.TrimSpace(contentHtml)
		}

		tpl, err := template.New("common_query.html").Funcs(funcMap).ParseFiles(contentHtmls...)
		if err != nil {
			logger.Errorf("解析模板出错（ParseFiles）：[%q] %s\n", req.RequestURI, err)
			return false
		}

		err = tpl.Execute(rw, data)
		if err != nil {
			logger.Errorf("执行模板出错（Execute）：[%q] %s\n", req.RequestURI, err)
			return false
		}

	default:
		contentHtml := req.FormValue(CONTENT_TPL_KEY)
		if contentHtml == "" {
			return true
		}
		contentHtmls := strings.Split(contentHtml, ",")
		for i, contentHtml := range contentHtmls {
			contentHtmls[i] = config.ROOT + strings.TrimSpace(contentHtml)
		}

		if !this.isBackView {
			// TODO: 旧模板还未完成的页面
			if strings.HasPrefix(req.RequestURI, "/wiki") {
				this.commonHtmlFiles = []string{config.ROOT + "/template/common/base.html"}
				this.baseTplName = "base.html"
			} else {
				this.commonHtmlFiles = []string{config.ROOT + "/template/common/layout.html"}
				this.baseTplName = "layout.html"
			}
		}

		// 为了使用自定义的模板函数，首先New一个以第一个模板文件名为模板名。
		// 这样，在ParseFiles时，新返回的*Template便还是原来的模板实例
		tpl, err := template.New(this.baseTplName).Funcs(funcMap).ParseFiles(append(this.commonHtmlFiles, contentHtmls...)...)
		if err != nil {
			logger.Errorf("解析模板出错（ParseFiles）：[%q] %s\n", req.RequestURI, err)
			return false
		}
		// 如果没有定义css和js模板，则定义之
		if jsTpl := tpl.Lookup("js"); jsTpl == nil {
			tpl.Parse(`{{define "js"}}{{end}}`)
		}
		if jsTpl := tpl.Lookup("css"); jsTpl == nil {
			tpl.Parse(`{{define "css"}}{{end}}`)
		}

		// 当前用户信息
		me, _ := CurrentUser(req)
		data["me"] = me

		if this.isBackView {
			if menu1, menu2, curMenu1 := service.GetUserMenu(me["uid"].(int), req.RequestURI); menu2 != nil {
				data["menu1"] = menu1
				data["menu2"] = menu2
				data["uri"] = req.RequestURI
				data["cur_menu1"] = curMenu1
			}
		}

		// websocket主机
		data["wshost"] = config.Config["wshost"]
		data["build"] = map[string]string{
			"version": version.Version,
			"date":    version.Date,
		}

		err = tpl.Execute(rw, data)
		if err != nil {
			logger.Errorf("执行模板出错（Execute）：[%q] %s\n", req.RequestURI, err)
			return false
		}
	}

	return true
}

type viewKey int

const (
	datakey   viewKey = 0
	formatkey viewKey = 1 // 存 希望返回的数据格式，如 "html", "json" 等
)

func GetData(req *http.Request) map[string]interface{} {
	data := getData(req, datakey)
	if data == nil {
		return make(map[string]interface{})
	}

	return data.(map[string]interface{})
}

func SetData(req *http.Request, data map[string]interface{}) {
	setData(req, datakey, data)
}

func getData(req *http.Request, viewkey viewKey) interface{} {
	if rv := context.Get(req, viewkey); rv != nil {
		// 获取之后立马删除
		context.Delete(req, viewkey)
		return rv
	}
	return nil
}

func setData(req *http.Request, viewkey viewKey, data interface{}) {
	context.Set(req, viewkey, data)
}
