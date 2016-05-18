// Copyright 2013 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author：polaris	studygolang@gmail.com

package main

import (
	"api"
	"config"
	. "controller"
	"controller/admin"
	"filter"

	"github.com/studygolang/mux"
)

func initRouter() *mux.Router {
	// 登录校验过滤器
	loginFilter := new(filter.LoginFilter)
	loginFilterChain := mux.NewFilterChain(loginFilter)

	// 敏感词过滤器
	sensitiveFilter := new(filter.SensitiveFilter)
	sensitiveFilterChain := mux.NewFilterChain(sensitiveFilter)

	router := mux.NewRouter()
	// 所有的页面都需要先检查用户cookie是否存在，以便在没登录时自动登录
	cookieFilter := new(filter.CookieFilter)
	// 大部分handler都需要页面展示
	frontViewFilter := filter.NewViewFilter(false)
	// 表单校验过滤器（配置了验证规则就会执行）
	formValidateFilter := new(filter.FormValidateFilter)

	fontFilterChan := mux.NewFilterChain([]mux.Filter{cookieFilter, formValidateFilter, frontViewFilter}...)
	router.FilterChain(fontFilterChan)

	router.HandleFunc("/", IndexHandler)
	router.HandleFunc("/wr", WRHandler)
	router.HandleFunc("/pkgdoc", PkgdocHandler)

	router.HandleFunc("/topics{view:(|/popular|/no_reply|/last)}", TopicsHandler)
	router.HandleFunc("/topics/{tid:[0-9]+}", TopicDetailHandler)
	router.HandleFunc("/topics/new{json:(|.json)}", NewTopicHandler).AppendFilterChain(loginFilterChain).AppendFilterChain(sensitiveFilterChain)
	router.HandleFunc("/topics/modify{json:(|.json)}", ModifyTopicHandler).AppendFilterChain(loginFilterChain).AppendFilterChain(sensitiveFilterChain)
	// 某个节点下的话题
	router.HandleFunc("/topics/node{nid:[0-9]+}", NodesHandler)

	// 注册
	router.HandleFunc("/account/register{json:(|.json)}", RegisterHandler)
	router.HandleFunc("/account/activate", ActivateHandler)
	router.HandleFunc("/account/send_activate_email.json", SendActivateEmailHandler)
	// 登录
	router.HandleFunc("/account/login{json:(|.json)}", LoginHandler)
	router.HandleFunc("/account/logout", LogoutHandler)

	router.HandleFunc("/account/edit{json:(|.json)}", AccountEditHandler).AppendFilterChain(loginFilterChain)
	router.HandleFunc("/account/changepwd.json", ChangePwdHandler).AppendFilterChain(loginFilterChain)
	router.HandleFunc("/account/change_avatar.json", ChangeAvatarHandler).AppendFilterChain(loginFilterChain)

	router.HandleFunc("/account/forgetpwd", ForgetPasswdHandler)
	router.HandleFunc("/account/resetpwd", ResetPasswdHandler)

	// 用户相关
	router.HandleFunc("/users", UsersHandler)
	router.HandleFunc("/user/{username:\\w+}", UserHomeHandler)
	// 邮件退订页面
	router.HandleFunc("/user/email/unsubscribe{json:(|.json)}", EmailUnsubHandler)

	// 网友博文
	router.HandleFunc("/articles", ArticlesHandler)
	router.HandleFunc("/articles/{id:[0-9]+}", ArticleDetailHandler)

	// 技术晨读
	router.HandleFunc("/readings", ReadingsHandler)
	router.HandleFunc("/readings/{id:[0-9]+}", IReadingHandler)

	// 搜索
	router.HandleFunc("/search", SearchHandler)

	// 项目
	router.HandleFunc("/project/new{json:(|.json)}", NewProjectHandler).AppendFilterChain(loginFilterChain)
	router.HandleFunc("/project/modify{json:(|.json)}", ModifyProjectHandler).AppendFilterChain(loginFilterChain)
	router.HandleFunc("/p/{uniq}", ProjectDetailHandler)
	router.HandleFunc("/projects", ProjectsHandler)
	router.HandleFunc("/project/uri.json", ProjectUriHandler)

	// wiki
	router.HandleFunc("/wiki", WikisHandler)
	router.HandleFunc("/wiki/new{json:(|.json)}", NewWikiPageHandler).AppendFilterChain(loginFilterChain)
	router.HandleFunc("/wiki/{uri}", WikiContentHandler)

	// WIDE 内嵌 iframe
	router.HandleFunc("/wide/playground", PlaygroundHandler)

	// 酷站
	router.HandleFunc("/sites", SitesHandler)
	// 资源
	router.HandleFunc("/resources", ResIndexHandler)
	router.HandleFunc("/resources/cat/{catid:[0-9]+}", CatResourcesHandler)
	router.HandleFunc("/resources/{id:[0-9]+}", ResourceDetailHandler)
	router.HandleFunc("/resources/new{json:(|.json)}", NewResourceHandler).AppendFilterChain(loginFilterChain).AppendFilterChain(sensitiveFilterChain)
	router.HandleFunc("/resources/modify{json:(|.json)}", ModifyResourceHandler).AppendFilterChain(loginFilterChain).AppendFilterChain(sensitiveFilterChain)

	// 评论
	router.HandleFunc("/comment/{objid:[0-9]+}.json", CommentHandler).AppendFilterChain(loginFilterChain)
	router.HandleFunc("/object/comments.json", ObjectCommentsHandler)

	// 喜欢
	router.HandleFunc("/like/{objid:[0-9]+}.json", LikeHandler).AppendFilterChain(loginFilterChain)

	//关注
	router.HandleFunc("/relation/subscribe.json", SubscribeHandler)
	router.HandleFunc("/relation/unsubscribe.json", UnsubscribeHandler)

	// 收藏
	router.HandleFunc("/favorite/{objid:[0-9]+}.json", FavoriteHandler).AppendFilterChain(loginFilterChain)
	router.HandleFunc("/favorites/{username}", SomeoneFavoritesHandler)

	// 消息相关
	router.HandleFunc("/message/send{json:(|.json)}", SendMessageHandler).AppendFilterChain(loginFilterChain)
	router.HandleFunc("/message/{msgtype:(system|inbox|outbox)}", MessageHandler).AppendFilterChain(loginFilterChain)
	router.HandleFunc("/message/delete.json", DeleteMessageHandler).AppendFilterChain(loginFilterChain)

	/////////////////// 异步请求 开始///////////////////////
	// 某节点下其他帖子
	router.HandleFunc("/topics/others/{nid:[0-9]+}_{tid:[0-9]+}.json", OtherTopicsHandler)
	// 统计信息
	router.HandleFunc("/websites/stat.json", StatHandler)
	// 社区最新公告或go最新动态
	router.HandleFunc("/dymanics/recent.json", RecentDymanicHandler)
	// 热门节点
	router.HandleFunc("/nodes/hot.json", HotNodesHandler)
	// 最新帖子
	router.HandleFunc("/topics/recent.json", RecentTopicHandler)
	// 最新博文
	router.HandleFunc("/articles/recent.json", RecentArticleHandler)
	// 最新项目
	router.HandleFunc("/projects/recent.json", RecentProjectHandler)
	// 最新资源
	router.HandleFunc("/resources/recent.json", RecentResourceHandler)
	// 最新评论
	router.HandleFunc("/comments/recent.json", RecentCommentHandler)
	// 活跃会员
	router.HandleFunc("/users/active.json", ActiveUserHandler)
	// 新会员
	router.HandleFunc("/users/newest.json", NewestUserHandler)
	// 最新晨读
	router.HandleFunc("/readings/recent.json", RecentReadingHandler)
	// @ 某人 suggest（登录用户才能@）
	router.HandleFunc("/at/users.json", AtUsersHandler).AppendFilterChain(loginFilterChain)

	// 文件上传（图片）
	router.HandleFunc("/upload/image.json", UploadImageHandler).AppendFilterChain(loginFilterChain)
	/////////////////// 异步请求 结束 ///////////////////////

	// 管理后台权限检查过滤器
	adminFilter := new(filter.AdminFilter)
	backViewFilter := filter.NewViewFilter(true, config.ROOT+"/template/admin/common.html")
	adminFilterChain := mux.NewFilterChain([]mux.Filter{loginFilter, adminFilter, formValidateFilter, backViewFilter}...)
	// admin 子系统
	router.FilterChain(adminFilterChain).HandleFunc("/admin", admin.IndexHandler).AppendFilterChain(loginFilterChain) // 支持"/admin访问"
	subrouter := router.PathPrefix("/admin").Subrouter()
	// 所有后台需要的过滤器链
	subrouter.FilterChain(adminFilterChain)

	///////////////// 用户管理 ////////////////////////
	// 权限（路由）管理
	subrouter.HandleFunc("/user/auth/list", admin.AuthListHandler)
	subrouter.HandleFunc("/user/auth/query.html", admin.AuthQueryHandler)
	subrouter.HandleFunc("/user/auth/new", admin.NewAuthorityHandler)
	subrouter.HandleFunc("/user/auth/modify", admin.ModifyAuthorityHandler)
	subrouter.HandleFunc("/user/auth/del", admin.DelAuthorityHandler)

	// 角色 管理
	subrouter.HandleFunc("/user/role/list", admin.RoleListHandler)
	subrouter.HandleFunc("/user/role/query.html", admin.RoleQueryHandler)
	subrouter.HandleFunc("/user/role/new", admin.NewRoleHandler)
	subrouter.HandleFunc("/user/role/modify", admin.ModifyRoleHandler)
	subrouter.HandleFunc("/user/role/del", admin.DelRoleHandler)

	// 用户 管理
	subrouter.HandleFunc("/user/user/list", admin.UserListHandler)
	subrouter.HandleFunc("/user/user/query.html", admin.UserQueryHandler)
	subrouter.HandleFunc("/user/user/detail", admin.UserDetailHandler)
	subrouter.HandleFunc("/user/user/modify", admin.UserModifyHandler)

	///////////////// 社区管理 //////////////////////////
	// 帖子管理
	subrouter.HandleFunc("/community/topic/list", admin.TopicListHandler)
	subrouter.HandleFunc("/community/topic/query.html", admin.TopicQueryHandler)
	subrouter.HandleFunc("/community/topic/modify", admin.ModifyTopicHandler)
	subrouter.HandleFunc("/community/topic/del", admin.DelTopicHandler)
	// 修改评论内容
	subrouter.HandleFunc("/community/comment/modify", admin.ModifyCommentHandler)
	subrouter.HandleFunc("/community/comment/del", admin.DelCommentHandler)

	// 开源项目管理
	subrouter.HandleFunc("/community/project/list", admin.ProjectListHandler)
	subrouter.HandleFunc("/community/project/query.html", admin.ProjectQueryHandler)
	subrouter.HandleFunc("/community/project/update_status", admin.UpdateProjectStatusHandler)

	///////////////// 抓取管理 //////////////////////////
	// 文章管理
	subrouter.HandleFunc("/crawl/article/list", admin.ArticleListHandler)
	subrouter.HandleFunc("/crawl/article/query.html", admin.ArticleQueryHandler)
	subrouter.HandleFunc("/crawl/article/modify", admin.ModifyArticleHandler)
	subrouter.HandleFunc("/crawl/article/new", admin.CrawlArticleHandler)
	subrouter.HandleFunc("/crawl/article/del", admin.DelArticleHandler)
	// 规则管理
	subrouter.HandleFunc("/crawl/rule/list", admin.RuleListHandler)
	subrouter.HandleFunc("/crawl/rule/query.html", admin.RuleQueryHandler)
	subrouter.HandleFunc("/crawl/rule/new", admin.NewRuleHandler)
	subrouter.HandleFunc("/crawl/rule/modify", admin.ModifyRuleHandler)
	subrouter.HandleFunc("/crawl/rule/del", admin.DelRuleHandler)

	// 晨读管理
	subrouter.HandleFunc("/reading/list", admin.ReadingListHandler)
	subrouter.HandleFunc("/reading/query.html", admin.ReadingQueryHandler)
	subrouter.HandleFunc("/reading/publish", admin.PublishReadingHandler)

	// 工具类
	subrouter.HandleFunc("/tool/sitemap", admin.GenSitemapHandler)

	apirouter := router.PathPrefix("/api").Subrouter()
	apirouter.HandleFunc("/user/login", api.LoginHandler)
	apirouter.HandleFunc("/blog/category/all", api.BlogCategoryHandler)

	// 错误处理handler
	router.FilterChain(fontFilterChan).HandleFunc("/noauthorize", NoAuthorizeHandler) // 无权限handler
	// 404（寻找遗失儿童页面）
	router.FilterChain(fontFilterChan).HandleFunc("/404", LossChildrenHandler)

	// 找不到页面
	router.FilterChain(fontFilterChan).HandleFunc("/{*}", NotFoundHandler)

	return router
}
