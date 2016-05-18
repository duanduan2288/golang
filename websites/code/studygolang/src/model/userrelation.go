package model

import (
	"logger"
	"util"
)

// 站内关注等信息
type UserRelation struct {
	Id         int    `json:"ID"`
	FromUserId int    `json:"from_user_id"`
	ToUserId   int    `json:"to_user_id"`
	RelType    int    `json:"rel_type"`
	Created    string `json:"create"`

	// 数据库访问对象
	*Dao
}

func NewUserRelation() *UserRelation {
	return &UserRelation{
		Dao: &Dao{tablename: "user_relation"},
	}
}

func (this *UserRelation) Insert() (int, error) {
	this.prepareInsertData()
	result, err := this.Dao.Insert()
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	return int(id), err
}

// 为了方便返回对象本身
func (this *UserRelation) Find(selectCol ...string) (*UserRelation, error) {
	return this, this.Dao.Find(this.colFieldMap(), selectCol...)
}

func (this *UserRelation) FindAll(selectCol ...string) ([]*UserRelation, error) {
	if len(selectCol) == 0 {
		selectCol = util.MapKeys(this.colFieldMap())
	}
	rows, err := this.Dao.FindAll(selectCol...)
	if err != nil {
		return nil, err
	}
	// TODO:
	userList := make([]*UserRelation, 0, 10)
	colNum := len(selectCol)
	for rows.Next() {
		userRelation := NewUserRelation()
		err = this.Scan(rows, colNum, userRelation.colFieldMap(), selectCol...)
		if err != nil {
			logger.Errorln("userRelation FindAll Scan Error:", err)
			continue
		}
		userList = append(userList, userRelation)
	}
	return userList, nil
}

func (this *UserRelation) Delete() error {
	err := this.Dao.Delete()
	return err
}

// 为了支持连写
func (this *UserRelation) Where(condition string, args ...interface{}) *UserRelation {
	this.Dao.Where(condition, args...)
	return this
}

// 为了支持连写
func (this *UserRelation) Set(clause string, args ...interface{}) *UserRelation {
	this.Dao.Set(clause, args...)
	return this
}

// 为了支持连写
func (this *UserRelation) Limit(limit string) *UserRelation {
	this.Dao.Limit(limit)
	return this
}

// 为了支持连写
func (this *UserRelation) Order(order string) *UserRelation {
	this.Dao.Order(order)
	return this
}

func (this *UserRelation) prepareInsertData() {
	this.columns = []string{"from_user_id", "to_user_id", "rel_type", "created"}
	this.colValues = []interface{}{this.FromUserId, this.ToUserId, this.RelType, this.Created}
}

func (this *UserRelation) colFieldMap() map[string]interface{} {
	return map[string]interface{}{
		"id":           &this.Id,
		"from_user_id": &this.FromUserId,
		"to_user_id":   &this.ToUserId,
		"rel_type":     &this.RelType,
		"created":      &this.Created,
	}
}
