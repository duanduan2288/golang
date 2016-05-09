// Copyright 2013 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author：polaris	studygolang@gmail.com

package service

import (
	"errors"
	"html/template"
	"net/url"
	"strconv"
	"strings"

	"logger"
	"model"
	"util"
)

var NotModifyAuthorityErr = errors.New("没有修改权限")

// 发布主题。入topics和topics_ex库
func PublishTopic(user map[string]interface{}, form url.Values) (err error) {
	uid := user["uid"].(int)

	topic := model.NewTopic()

	if form.Get("tid") != "" {
		err = topic.Where("tid=?", form.Get("tid")).Find()
		if err != nil {
			logger.Errorln("Publish Topic find error:", err)
			return
		}

		isAdmin := false
		if _, ok := user["isadmin"]; ok {
			isAdmin = user["isadmin"].(bool)
		}
		if topic.Uid != uid && !isAdmin {
			err = NotModifyAuthorityErr
			return
		}

		_, err = ModifyTopic(user, form)
		if err != nil {
			logger.Errorln("Publish Topic error:", err)
			return
		}
	} else {

		util.ConvertAssign(topic, form)

		topic.Uid = uid
		topic.Ctime = util.TimeNow()

		var tid int
		tid, err = topic.Insert()

		if err != nil {
			logger.Errorln("Publish Topic error:", err)
			return
		}

		// 存扩展信息
		topicEx := model.NewTopicEx()
		topicEx.Tid = tid
		_, err = topicEx.Insert()
		if err != nil {
			logger.Errorln("Insert TopicEx error:", err)
			return
		}

		// 给 被@用户 发系统消息
		ext := map[string]interface{}{
			"objid":   tid,
			"objtype": model.TYPE_TOPIC,
			"uid":     user["uid"],
			"msgtype": model.MsgtypePublishAtMe,
		}
		go SendSysMsgAtUsernames(form.Get("usernames"), ext)

		// 发布主题，活跃度+10
		go IncUserWeight("uid="+strconv.Itoa(uid), 10)
	}

	return
}

// 修改主题
// user 修改人的（有可能是作者或管理员）
func ModifyTopic(user map[string]interface{}, form url.Values) (errMsg string, err error) {
	uid := user["uid"].(int)
	form.Set("editor_uid", strconv.Itoa(uid))

	fields := []string{"title", "content", "nid", "editor_uid"}
	query, args := updateSetClause(form, fields)

	tid := form.Get("tid")

	err = model.NewTopic().Set(query, args...).Where("tid=" + tid).Update()
	if err != nil {
		logger.Errorf("更新主题 【%s】 信息失败：%s\n", tid, err)
		errMsg = "对不起，服务器内部错误，请稍后再试！"
		return
	}

	username := user["username"].(string)
	// 修改主题，活跃度+2
	go IncUserWeight("username="+username, 2)

	return
}

// 获得主题详细信息（包括详细回复）
// 为了避免转换，tid传string类型
func FindTopicByTid(tid string) (topicMap map[string]interface{}, replies []map[string]interface{}, err error) {
	condition := "tid=" + tid
	// 主题信息
	topic := model.NewTopic()
	err = topic.Where(condition).Find()
	if err != nil {
		logger.Errorln("topic service FindTopicByTid Error:", err)
		return
	}
	// 主题不存在
	if topic.Tid == 0 {
		err = errors.New("The topic of tid is not exists")
		return
	}

	topicMap = make(map[string]interface{})
	util.Struct2Map(topicMap, topic)

	// 解析内容中的 @
	topicMap["content"] = decodeTopicContent(topic)

	topicEx := model.NewTopicEx()
	err = topicEx.Where(condition).Find()
	if err != nil {
		logger.Errorln("topic service FindTopicByTid Error:", err)
		return
	}
	if topicEx.Tid == 0 {
		return
	}
	util.Struct2Map(topicMap, topicEx)
	// 节点名字
	topicMap["node"] = GetNodeName(topic.Nid)

	// 回复信息（评论）
	replies, owerUser, lastReplyUser := FindObjComments(tid, strconv.Itoa(model.TYPE_TOPIC), topic.Uid, topic.Lastreplyuid)
	topicMap["user"] = owerUser
	// 有人回复
	if topic.Lastreplyuid != 0 {
		topicMap["lastreplyusername"] = lastReplyUser.Username
	}

	if topic.EditorUid != 0 {
		topicMap["editor_username"] = FindUsernameByUid(topic.EditorUid)
	}

	return
}

// 获取单个 Topic 信息（用于编辑）
func FindTopic(tid string) *model.Topic {
	topic := model.NewTopic()
	err := topic.Where("tid=?", tid).Find()
	if err != nil {
		logger.Errorf("FindTopic [%s] error：%s\n", tid, err)
	}

	return topic
}

// 通过tid获得话题的所有者
func getTopicOwner(tid int) int {
	// 主题信息
	topic := model.NewTopic()
	err := topic.Where("tid=" + strconv.Itoa(tid)).Find()
	if err != nil {
		logger.Errorln("topic service getTopicOwner Error:", err)
		return 0
	}
	return topic.Uid
}

func decodeTopicContent(topic *model.Topic) string {
	// 安全过滤
	content := template.HTMLEscapeString(topic.Content)

	// 允许内嵌 Wide iframe
	content = util.EmbedWide(content)

	// @别人
	return parseAtUser(content)
}

// 获得主题列表页需要的数据
// 如果order为空，则默认排序方式（之所以用不定参数，是为了可以不传）
func FindTopics(page, pageNum int, where string, orderSlice ...string) (topics []map[string]interface{}, total int) {
	if pageNum == 0 {
		pageNum = PAGE_NUM
	}
	var offset = 0
	if page > 1 {
		offset = (page - 1) * pageNum
	}
	// 即使传了多个，也只取第一个
	order := "mtime DESC"
	if len(orderSlice) > 0 && orderSlice[0] != "" {
		order = orderSlice[0]
	}
	return FindTopicsByWhere(where, order, strconv.Itoa(offset)+","+strconv.Itoa(pageNum))
}

// 获取话题列表（分页），目前供后台使用
func FindTopicsByPage(conds map[string]string, curPage, limit int) ([]*model.Topic, int) {
	conditions := make([]string, 0, len(conds))
	for k, v := range conds {
		conditions = append(conditions, k+"="+v)
	}

	topic := model.NewTopic()

	limitStr := strconv.Itoa((curPage-1)*limit) + "," + strconv.Itoa(limit)
	topicList, err := topic.Where(strings.Join(conditions, " AND ")).Order("tid DESC").Limit(limitStr).
		FindAll()
	if err != nil {
		logger.Errorln("topic service FindTopicsByPage Error:", err)
		return nil, 0
	}

	total, err := topic.Count()
	if err != nil {
		logger.Errorln("topic service FindTopicsByPage COUNT Error:", err)
		return nil, 0
	}

	return topicList, total
}

// 获得某个节点下的主题列表（侧边栏推荐）
func FindTopicsByNid(nid, curTid string) (topics []*model.Topic) {
	var err error
	topics, err = model.NewTopic().Where("nid=" + nid + " and tid!=" + curTid).Limit("0,10").FindAll()
	if err != nil {
		logger.Errorln("topic service FindTopicsByNid Error:", err)
		return
	}
	return
}

// 获得社区最新公告（废弃）
func FindNoticeTopic() (topic *model.Topic) {
	topics, err := model.NewTopic().Where("nid=15").Limit("0,1").Order("mtime DESC").FindAll()
	if err != nil {
		logger.Errorln("topic service FindNoticeTopic Error:", err)
		return
	}
	if len(topics) > 0 {
		topic = topics[0]
	}
	return
}

func FindTopicsByWhere(where, order, limit string) (topics []map[string]interface{}, total int) {
	topicObj := model.NewTopic()
	if where != "" {
		topicObj.Where(where)
	}
	if order != "" {
		topicObj.Order(order)
	}
	if limit != "" {
		topicObj.Limit(limit)
	}
	topicList, err := topicObj.FindAll()
	if err != nil {
		logger.Errorln("topic service topicObj.FindAll Error:", err)
		return
	}
	// 获得总主题数
	total, err = topicObj.Count()
	if err != nil {
		logger.Errorln("topic service topicObj.Count Error:", err)
		return
	}
	count := len(topicList)
	tids := make([]int, count)
	uids := make([]int, 0, count)
	nids := make([]int, count)
	for i, topic := range topicList {
		tids[i] = topic.Tid
		uids = append(uids, topic.Uid)
		if topic.Lastreplyuid != 0 {
			uids = append(uids, topic.Lastreplyuid)
		}
		nids[i] = topic.Nid
	}

	// 获取扩展信息（计数）
	topicExList, err := model.NewTopicEx().Where("tid in(" + util.Join(tids, ",") + ")").FindAll()
	if err != nil {
		logger.Errorln("topic service NewTopicEx FindAll Error:", err)
		return
	}
	topicExMap := make(map[int]*model.TopicEx, len(topicExList))
	for _, topicEx := range topicExList {
		topicExMap[topicEx.Tid] = topicEx
	}

	userMap := GetUserInfos(uids)

	// 获取节点信息
	nodes := GetNodesName(nids)

	topics = make([]map[string]interface{}, count)
	for i, topic := range topicList {
		tmpMap := make(map[string]interface{})
		util.Struct2Map(tmpMap, topic)
		util.Struct2Map(tmpMap, topicExMap[topic.Tid])
		tmpMap["user"] = userMap[topic.Uid]
		// 有人回复
		if tmpMap["lastreplyuid"].(int) != 0 {
			tmpMap["lastreplyusername"] = userMap[tmpMap["lastreplyuid"].(int)].Username
		}
		tmpMap["node"] = nodes[tmpMap["nid"].(int)]
		topics[i] = tmpMap
	}
	return
}

// 获得最近的主题(如果uid!=0，则获取某个用户最近的主题)
func FindRecentTopics(uid int, limit string) []*model.Topic {
	cond := ""
	if uid != 0 {
		cond = "uid=" + strconv.Itoa(uid)
	}

	topics, err := model.NewTopic().Where(cond).Order("ctime DESC").Limit(limit).FindAll()
	if err != nil {
		logger.Errorln("topic service FindRecentTopics error:", err)
		return nil
	}
	for _, topic := range topics {
		topic.Node = GetNodeName(topic.Nid)
	}
	return topics
}

// 获得回复最多的10条主题(TODO:避免一直显示相同的)
func FindHotTopics() []map[string]interface{} {
	topicExList, err := model.NewTopicEx().Order("reply DESC").Limit("0,10").FindAll()
	if err != nil {
		logger.Errorln("topic service FindHotReplies error:", err)
		return nil
	}
	tidMap := make(map[int]int, len(topicExList))
	topicExMap := make(map[int]*model.TopicEx, len(topicExList))
	for _, topicEx := range topicExList {
		tidMap[topicEx.Tid] = topicEx.Tid
		topicExMap[topicEx.Tid] = topicEx
	}
	tids := util.MapIntKeys(tidMap)
	topics := FindTopicsByTids(tids)
	if topics == nil {
		return nil
	}

	uids := util.Models2Intslice(topics, "Uid")
	userMap := GetUserInfos(uids)

	result := make([]map[string]interface{}, len(topics))
	for i, topic := range topics {
		oneTopic := make(map[string]interface{})
		util.Struct2Map(oneTopic, topic)
		util.Struct2Map(oneTopic, topicExMap[topic.Tid])
		oneTopic["user"] = userMap[topic.Uid]
		result[i] = oneTopic
	}
	return result
}

// 获取多个主题详细信息
func FindTopicsByTids(tids []int) []*model.Topic {
	if len(tids) == 0 {
		return nil
	}
	inTids := util.Join(tids, ",")
	topics, err := model.NewTopic().Where("tid in(" + inTids + ")").FindAll()
	if err != nil {
		logger.Errorln("topic service FindTopicsByTids error:", err)
		return nil
	}
	return topics
}

// 获取多个主题详细信息
func FindTopicsByIds(ids []int) []*model.Topic {
	if len(ids) == 0 {
		return nil
	}
	inIds := util.Join(ids, ",")
	topics, err := model.NewTopic().Where("tid in(" + inIds + ")").FindAll()
	if err != nil {
		logger.Errorln("topic service FindTopicsByIds error:", err)
		return nil
	}
	return topics
}

// 提供给其他service调用（包内）
func getTopics(tids map[int]int) map[int]*model.Topic {
	topics := FindTopicsByTids(util.MapIntKeys(tids))
	topicMap := make(map[int]*model.Topic, len(topics))
	for _, topic := range topics {
		topicMap[topic.Tid] = topic
	}
	return topicMap
}

// 获得热门节点
func FindHotNodes() []map[string]interface{} {
	strSql := "SELECT nid, COUNT(1) AS topicnum FROM topics GROUP BY nid ORDER BY topicnum DESC LIMIT 10"
	rows, err := model.NewTopic().DoSql(strSql)
	if err != nil {
		logger.Errorln("topic service FindHotNodes error:", err)
		return nil
	}
	nodes := make([]map[string]interface{}, 0, 10)
	for rows.Next() {
		var nid, topicnum int
		err = rows.Scan(&nid, &topicnum)
		if err != nil {
			logger.Errorln("rows.Scan error:", err)
			continue
		}
		name := GetNodeName(nid)
		node := map[string]interface{}{
			"name": name,
			"nid":  nid,
		}
		nodes = append(nodes, node)
	}
	return nodes
}

// 话题总数
func TopicsTotal() (total int) {
	total, err := model.NewTopic().Count()
	if err != nil {
		logger.Errorln("topic service TopicsTotal error:", err)
	}
	return
}

// 安全过滤
func JSEscape(topics []*model.Topic) []*model.Topic {
	for i, topic := range topics {
		topics[i].Title = template.JSEscapeString(topic.Title)
		topics[i].Content = template.JSEscapeString(topic.Content)
	}
	return topics
}

// 话题回复（评论）
type TopicComment struct{}

// 更新该主题的回复信息
// cid：评论id；objid：被评论对象id；uid：评论者；cmttime：评论时间
func (self TopicComment) UpdateComment(cid, objid, uid int, cmttime string) {
	tid := strconv.Itoa(objid)
	// 更新最后回复信息
	stringBuilder := util.NewBuffer().Append("lastreplyuid=").AppendInt(uid).Append(",lastreplytime=").Append(cmttime)
	err := model.NewTopic().Set(stringBuilder.String()).Where("tid=" + tid).Update()
	if err != nil {
		logger.Errorln("更新主题最后回复人信息失败：", err)
	}
	// 更新回复数（TODO：暂时每次都更新表）
	err = model.NewTopicEx().Where("tid="+tid).Increment("reply", 1)
	if err != nil {
		logger.Errorln("更新主题回复数失败：", err)
	}
}

func (self TopicComment) String() string {
	return "topic"
}

// 实现 CommentObjecter 接口
func (self TopicComment) SetObjinfo(ids []int, commentMap map[int][]*model.Comment) {
	topics := FindTopicsByTids(ids)
	if len(topics) == 0 {
		return
	}

	for _, topic := range topics {
		objinfo := make(map[string]interface{})
		objinfo["title"] = topic.Title
		objinfo["uri"] = model.PathUrlMap[model.TYPE_TOPIC]
		objinfo["type_name"] = model.TypeNameMap[model.TYPE_TOPIC]

		for _, comment := range commentMap[topic.Tid] {
			comment.Objinfo = objinfo
		}
	}
}

// 主题喜欢
type TopicLike struct{}

// 更新该主题的喜欢数
// objid：被喜欢对象id；num: 喜欢数(负数表示取消喜欢)
func (self TopicLike) UpdateLike(objid, num int) {
	// 更新喜欢数（TODO：暂时每次都更新表）
	err := model.NewTopicEx().Where("tid=?", objid).Increment("like", num)
	if err != nil {
		logger.Errorln("更新主题喜欢数失败：", err)
	}
}

func (self TopicLike) String() string {
	return "topic"
}
