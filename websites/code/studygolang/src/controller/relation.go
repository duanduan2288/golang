package controller

import (
	//"encoding/json"
	//"fmt"
	"net/http"
	//"strconv"

	"filter"
	"service"
	"util"
	//"github.com/studygolang/mux"
)

//关注 /relation/subscribe.json
func SubscribeHandler(rw http.ResponseWriter, req *http.Request) {
	var data = make(map[string]interface{})

	user, _ := filter.CurrentUser(req)

	from_user_id := user["uid"].(int)
	to_user_id := util.MustInt(req.PostFormValue("to_user_id"))

	isfans := service.IsFans(from_user_id, to_user_id)
	if !isfans {
		err := service.Subscribe(from_user_id, to_user_id)
		if err == nil {
			data["ok"] = 1
			data["error"] = "已关注"
		} else {
			data["ok"] = 0
			data["error"] = err.Error()
		}
	} else {
		data["ok"] = 1
		data["error"] = "已关注"
	}

	filter.SetData(req, data)
}

//取消关注/relation/unsubscribe.json
func UnsubscribeHandler(rw http.ResponseWriter, req *http.Request) {
	var data = make(map[string]interface{})

	user, _ := filter.CurrentUser(req)

	from_user_id := user["uid"].(int)
	to_user_id := util.MustInt(req.PostFormValue("to_user_id"))

	isfans := service.IsFans(from_user_id, to_user_id)
	if isfans {
		err := service.Unsubscribe(from_user_id, to_user_id)
		if err == nil {
			data["ok"] = 1
			data["error"] = "已取消关注"
		} else {
			data["ok"] = 0
			data["error"] = err.Error()
		}
	} else {
		data["ok"] = 1
		data["error"] = "未关注"
	}

	filter.SetData(req, data)
}
