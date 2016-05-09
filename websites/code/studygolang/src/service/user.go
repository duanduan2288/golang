// Copyright 2013 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author：polaris	studygolang@gmail.com

package service

import (
	"errors"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
	"time"

	"logger"
	"model"
	"util"
)

var DefaultAvatars = []string{
	"gopher_aqua.jpg", "gopher_boy.jpg", "gopher_brown.jpg", "gopher_gentlemen.jpg",
	"gopher_strawberry.jpg", "gopher_strawberry_bg.jpg", "gopher_teal.jpg",
	"gopher01.png", "gopher02.png", "gopher03.png", "gopher04.png",
	"gopher05.png", "gopher06.png", "gopher07.png", "gopher08.png",
	"gopher09.png", "gopher10.png", "gopher11.png", "gopher12.png",
	"gopher13.png", "gopher14.png", "gopher15.png", "gopher16.png",
	"gopher17.png", "gopher18.png", "gopher19.png", "gopher20.png",
	"gopher21.png", "gopher22.png", "gopher23.png", "gopher24.png",
	"gopher25.png", "gopher26.png", "gopher27.png", "gopher28.png",
}

func CreateUser(form url.Values) (errMsg string, err error) {
	if EmailExists(form.Get("email")) {
		err = errors.New("该邮箱已注册过")
		return
	}
	if UsernameExists(form.Get("username")) {
		err = errors.New("用户名已存在")
		return
	}
	// 存用户基本信息，产生自增长UID
	user := model.NewUser()
	err = util.ConvertAssign(user, form)
	if err != nil {
		logger.Errorln("user ConvertAssign error", err)
		errMsg = err.Error()
		return
	}
	user.Ctime = util.TimeNow()

	// 随机给一个默认头像
	user.Avatar = DefaultAvatars[rand.Intn(len(DefaultAvatars))]
	uid, err := user.Insert()
	if err != nil {
		errMsg = "内部服务器错误"
		logger.Errorln(errMsg, "：", err)
		return
	}

	// 存用户登录信息
	userLogin := model.NewUserLogin()
	err = util.ConvertAssign(userLogin, form)
	if err != nil {
		errMsg = err.Error()
		logger.Errorln("CreateUser error:", err)
		return
	}
	userLogin.Uid = uid
	_, err = userLogin.Insert()
	if err != nil {
		errMsg = "内部服务器错误"
		logger.Errorln(errMsg, "：", err)
		return
	}

	// 存用户角色信息
	userRole := model.NewUserRole()
	// 默认为初级会员
	userRole.Roleid = Roles[len(Roles)-1].Roleid
	userRole.Uid = uid
	if _, err = userRole.Insert(); err != nil {
		logger.Errorln("userRole insert Error:", err)
	}

	// 存用户活跃信息，初始活跃+2
	userActive := model.NewUserActive()
	userActive.Uid = uid
	userActive.Username = user.Username
	userActive.Avatar = user.Avatar
	userActive.Email = user.Email
	userActive.Weight = 2
	if _, err = userActive.Insert(); err != nil {
		logger.Errorln("UserActive insert Error:", err)
	}
	return
}

// 修改用户资料
func UpdateUser(form url.Values) (errMsg string, err error) {
	fields := []string{"name", "open", "city", "company", "github", "weibo", "website", "monlog", "introduce"}
	setClause := GenSetClause(form, fields)
	username := form.Get("username")
	err = model.NewUser().Set(setClause).Where("username=" + username).Update()
	if err != nil {
		logger.Errorf("更新用户 【%s】 信息失败：%s", username, err)
		errMsg = "对不起，服务器内部错误，请稍后再试！"
		return
	}

	// 修改用户资料，活跃度+1
	go IncUserWeight("username="+username, 1)

	return
}

// UpdateUserStatus 更新用户状态
func UpdateUserStatus(uid, status int) {
	setClause := "status=" + strconv.Itoa(status)
	err := model.NewUser().Set(setClause).Where("uid=?", uid).Update()
	if err != nil {
		logger.Errorf("更新用户 【%s】 状态失败：%s", uid, err)
		return
	}

	return
}

// ActivateUser 激活用户
func ActivateUser(email string) error {
	setClause := "status=" + strconv.Itoa(model.StatusAudit)
	err := model.NewUser().Set(setClause).Where("email=?", email).Update()
	if err != nil {
		logger.Errorf("激活用户 【%s】 失败：%s", email, err)
		return err
	}

	return nil
}

// 邮件订阅或取消订阅
func EmailSubscribe(uid, unsubscribe int) {
	err := model.NewUser().Set("unsubscribe=?", unsubscribe).Where("uid=?", uid).Update()
	if err != nil {
		logger.Errorln("Email Subscribe Error:", err)
	}
}

// 更换头像
func ChangeAvatar(uid int, avatar string) (err error) {
	err = model.NewUser().Set("avatar=?", avatar).Where("uid=?", uid).Update()
	if err == nil {
		err = model.NewUserActive().Set("avatar=?", avatar).Where("uid=?", uid).Update()
	}

	return
}

// 通过邮箱获取用户信息
func FindUserByEmail(email string) *model.User {
	user := model.NewUser()
	err := user.Where("email=?", email).Find()
	if err != nil {
		logger.Errorln("FindUserByEmail error:", err)
	}

	return user
}

// 获取当前登录用户信息（常用信息）
func FindCurrentUser(username string) (user map[string]interface{}, err error) {
	userInfo := model.NewUser()
	err = userInfo.Where("username=" + username).Find()
	if err != nil {
		logger.Errorf("获取用户 %s 信息失败：%s", username, err)
		return
	}
	if userInfo.Uid == 0 {
		logger.Infof("用户 %s 不存在！", username)
		return
	}
	user = map[string]interface{}{
		"uid":      userInfo.Uid,
		"username": userInfo.Username,
		"email":    userInfo.Email,
		"avatar":   userInfo.Avatar,
		"status":   userInfo.Status,
	}

	// 获取未读消息数
	user["msgnum"] = FindNotReadMsgNum(userInfo.Uid)

	// 获取角色信息
	userRoleList, err := model.NewUserRole().Where("uid=" + strconv.Itoa(userInfo.Uid)).FindAll()
	if err != nil {
		logger.Errorf("获取用户 %s 角色 信息失败：%s", username, err)
		return
	}
	for _, userRole := range userRoleList {
		if userRole.Roleid <= model.AdminMinRoleId {
			// 是管理员
			user["isadmin"] = true
		}
	}

	RecordLoginTime(username)

	return
}

// IsNormalUser 判断是否是正常的用户
func IsNormalUser(userStatus interface{}) bool {
	if userStatus == nil {
		return true
	}

	if userStatus.(int) > model.StatusRefuse {
		return false
	}

	return true
}

// 判断指定的用户名是否存在
func UsernameExists(username string) bool {
	userLogin := model.NewUserLogin()
	if err := userLogin.Where("username=" + username).Find("uid"); err != nil {
		logger.Errorln("service UsernameExists error:", err)
		return false
	}
	if userLogin.Uid != 0 {
		return true
	}
	return false
}

// 判断指定的邮箱（email）是否存在
func EmailExists(email string) bool {
	userLogin := model.NewUserLogin()
	if err := userLogin.Where("email=" + email).Find("uid"); err != nil {
		logger.Errorln("service EmailExists error:", err)
		return false
	}
	if userLogin.Uid != 0 {
		return true
	}
	return false
}

// 获取单个用户信息
func FindUserByUsername(username string) *model.User {
	return findUserByUniq("username", username)
}

// 通过UID获取用户名
func FindUsernameByUid(uid int) string {
	user := model.NewUser()
	err := user.Where("uid=" + strconv.Itoa(uid)).Find()
	if err != nil {
		logger.Errorf("获取用户 %s 信息失败：%s", uid, err)
		return ""
	}
	if user.Uid == 0 {
		return ""
	}

	return user.Username
}

// 获取单个用户信息
func FindUserByUID(uid string) *model.User {
	return findUserByUniq("uid", uid)
}

// 通过唯一键（uid或username）获取用户信息
func findUserByUniq(field, val string) *model.User {
	user := model.NewUser()
	err := user.Where(field + "=" + val).Find()
	if err != nil {
		logger.Errorf("获取用户 %s 信息失败：%s", val, err)
		return nil
	}
	if user.Uid == 0 {
		return nil
	}

	// 获取用户角色信息
	userRoleList, err := model.NewUserRole().
		Order("roleid ASC").Where("uid="+strconv.Itoa(user.Uid)).FindAll("uid", "roleid")
	if err != nil {
		logger.Errorf("获取用户 %s 角色 信息失败：%s", val, err)
		return nil
	}

	if roleNum := len(userRoleList); roleNum > 0 {
		user.Roleids = make([]int, roleNum)
		user.Rolenames = make([]string, roleNum)

		for i, userRole := range userRoleList {
			user.Roleids[i] = userRole.Roleid
			user.Rolenames[i] = Roles[userRole.Roleid-1].Name
		}
	}

	return user
}

// 获得活跃用户
func FindActiveUsers(start, num int) []*model.UserActive {
	activeUsers, err := model.NewUserActive().Order("weight DESC").Limit(strconv.Itoa(start) + "," + strconv.Itoa(num)).FindAll()
	if err != nil {
		logger.Errorln("user service FindActiveUsers error:", err)
		return nil
	}
	return activeUsers
}

// 最新加入会员
func FindNewUsers(start, num int) []*model.User {
	users, err := model.NewUser().Order("ctime DESC").Limit(strconv.Itoa(start) + "," + strconv.Itoa(num)).FindAll([]string{"uid", "username", "email", "avatar", "ctime"}...)
	if err != nil {
		logger.Errorln("user service FindNewUsers error:", err)
		return nil
	}
	return users
}

func FindUsersByPage(conds map[string]string, curPage, limit int) ([]*model.User, int) {
	conditions := make([]string, 0, len(conds))
	for k, v := range conds {
		conditions = append(conditions, k+"="+v)
	}

	user := model.NewUser()

	limitStr := strconv.Itoa((curPage-1)*limit) + "," + strconv.Itoa(limit)
	userList, err := user.Where(strings.Join(conditions, " AND ")).Limit(limitStr).
		FindAll()
	if err != nil {
		logger.Errorln("user service FindUsersByPage Error:", err)
		return nil, 0
	}

	total, err := user.Count()
	if err != nil {
		logger.Errorln("user service FindUsersByPage COUNT Error:", err)
		return nil, 0
	}

	return userList, total
}

// 获取 @ 的 suggest 列表
func GetUserMentions(term string, limit int) []map[string]string {
	term = "%" + term + "%"
	userActives, err := model.NewUserActive().Where("username like ?", term).Limit(strconv.Itoa(limit)).Order("mtime DESC").FindAll("email", "username", "avatar")
	if err != nil {
		logger.Errorln("user service GetUserMentions Error:", err)
		return nil
	}

	users := make([]map[string]string, len(userActives))
	for i, userActive := range userActives {
		user := make(map[string]string, 2)
		user["username"] = userActive.Username
		user["avatar"] = util.Gravatar(userActive.Avatar, userActive.Email, 20)
		users[i] = user
	}

	return users
}

var (
	ErrUsername = errors.New("用户名不存在")
	ErrPasswd   = errors.New("密码错误")
)

// 登录；成功返回用户登录信息(user_login)
func Login(username, passwd string) (*model.UserLogin, error) {
	userLogin := model.NewUserLogin()
	err := userLogin.Where("username=" + username + " OR email=" + username).Find()
	if err != nil {
		logger.Errorf("用户 %s 登录错误：%s", username, err)
		return nil, errors.New("内部错误，请稍后再试！")
	}
	// 校验用户
	if userLogin.Uid == 0 {
		logger.Infof("用户名 %s 不存在", username)
		return nil, ErrUsername
	}

	// 检验用户是否审核通过，暂时只有审核通过的才能登录
	userInfo := model.NewUser()
	err = userInfo.Where("uid=?", userLogin.Uid).Find()
	if err != nil {
		logger.Infof("用户名 %s 不存在", username)
		return nil, ErrUsername
	}
	if userInfo.Status != model.StatusAudit {
		logger.Infof("用户 %s 状态不是审核通过：%d", username, userInfo.Status)
		var errMap = map[int]error{
			model.StatusNoAudit: errors.New("您的账号未激活，请到注册邮件中进行激活操作！"),
			model.StatusRefuse:  errors.New("您的账号审核拒绝"),
			model.StatusFreeze:  errors.New("您的账号因为非法发布信息已被冻结，请联系管理员！"),
			model.StatusOutage:  errors.New("您的账号因为非法发布信息已被停号，请联系管理员！"),
		}
		return nil, errMap[userInfo.Status]
	}

	passcode := userLogin.GetPasscode()
	md5Passwd := util.Md5(passwd + passcode)
	logger.Debugf("passwd: %s, passcode: %s, md5passwd: %s, dbpasswd: %s", passwd, passcode, md5Passwd, userLogin.Passwd)
	if md5Passwd != userLogin.Passwd {
		logger.Infof("用户名 %s 填写的密码错误", username)
		return nil, ErrPasswd
	}

	// 登录，活跃度+1
	go IncUserWeight("uid="+strconv.Itoa(userLogin.Uid), 1)

	RecordLoginTime(username)

	return userLogin, nil
}

// 记录用户最后登录时间
func RecordLoginTime(username string) error {
	userLogin := model.NewUserLogin()
	err := userLogin.Set("login_time=" + time.Now().Format("2006-01-02 15:04:05")).Where("username=" + username).Update()
	if err != nil {
		logger.Errorf("记录用户 %s 登录时间错误：%s", username, err)
	}
	return err
}

// 更新用户密码（用户名或email）
func UpdatePasswd(username, passwd string) (string, error) {
	userLogin := model.NewUserLogin()
	passwd = userLogin.GenMd5Passwd(passwd)
	err := userLogin.Set("passwd=" + passwd + ",passcode=" + userLogin.GetPasscode()).Where("username=" + username + " OR email=" + username).Update()
	if err != nil {
		logger.Errorf("用户 %s 更新密码错误：%s", username, err)
		return "对不起，内部服务错误！", err
	}
	return "", nil
}

// 获取用户信息
func GetUserInfos(uids []int) map[int]*model.User {
	if len(uids) == 0 {
		return nil
	}
	// 获取用户信息
	inUids := util.Join(uids, ",")
	users, err := model.NewUser().Where("uid in(" + inUids + ")").FindAll()
	if err != nil {
		logger.Errorln("user service GetUserInfos Error:", err)
		return map[int]*model.User{}
	}
	userMap := make(map[int]*model.User, len(users))
	for _, user := range users {
		userMap[user.Uid] = user
	}
	return userMap
}

// 会员总数
func CountUsers() int {
	total, err := model.NewUserLogin().Count()
	if err != nil {
		logger.Errorln("user service CountUsers error:", err)
		return 0
	}
	return total
}

// 增加或减少用户活跃度
func IncUserWeight(where string, weight int) {
	if err := model.NewUserActive().Where(where).Increment("weight", weight); err != nil {
		logger.Errorln("UserActive update Error:", err)
	}
}

func DecrUserWeight(where string, divide int) {
	if divide <= 0 {
		return
	}

	strSql := "UPDATE user_active SET weight=weight/" + strconv.Itoa(divide) + " WHERE " + where
	if result, err := model.NewUserActive().Exec(strSql); err != nil {
		logger.Errorln("UserActive update Error:", err)
	} else {
		n, _ := result.RowsAffected()
		logger.Debugln(strSql, "affected num:", n)
	}
}

// 获取 loginTime 之前没有登录的用户
func FindNotLoginUsers(loginTime string) (userList []*model.UserLogin, err error) {
	userLogin := model.NewUserLogin()
	userList, err = userLogin.Where("login_time<" + loginTime).FindAll()
	return
}

func AllocUserRoles(uid int, roleids []string) error {
	userRole := model.NewUserRole()
	userRole.Uid = uid

	for _, roleId := range roleids {
		userRole.Roleid, _ = strconv.Atoi(roleId)
		if userRole.Roleid == 0 {
			continue
		}

		userRole.Insert()
	}

	return nil
}
