// Copyright 2013 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author：polaris	studygolang@gmail.com

package model

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	// _ "github.com/ziutek/mymysql/godrv"
	. "config"
	"logger"
	"reflect"
	"sort"
	"strings"
	"util"
)

type Dao struct {
	*sql.DB
	// 构造sql语句相关
	tablename string
	where     string
	whereVal  []interface{} // where条件对应中字段对应的值
	limit     string
	order     string
	// 插入需要
	columns   []string      // 需要插入数据的字段
	colValues []interface{} // 需要插入字段对应的值
	// 查询需要
	selectCols string // 想要查询那些字段，接在SELECT之后的，默认为"*"
}

func NewDao(tablename string) *Dao {
	return &Dao{tablename: tablename}
}

func (this *Dao) Open() (err error) {
	this.DB, err = sql.Open(Config["drive_name"], Config["dsn"])
	return
}

// Insert 插入数据
func (this *Dao) Insert() (sql.Result, error) {
	strSql := util.InsertSql(this)
	logger.Debugln("Insert sql:", strSql)
	err := this.Open()
	if err != nil {
		return nil, err
	}
	defer this.Close()
	stmt, err := this.Prepare(strSql)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	return stmt.Exec(this.ColValues()...)
}

// Update 更新数据
func (this *Dao) Update() error {
	strSql := util.UpdateSql(this)
	if strSql == "" {
		// 没有字段需要更新，当作更新成功
		logger.Errorln("no field need update")
		return nil
	}
	logger.Debugln("Update sql:", strSql)
	err := this.Open()
	if err != nil {
		return err
	}
	defer this.Close()
	result, err := this.Exec(strSql, append(this.colValues, this.whereVal...)...)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	logger.Debugf("成功更新了`%s`表 %d 条记录", this.tablename, affected)
	return nil
}

func (this *Dao) Delete() error {
	strSql := util.DeleteSql(this)
	logger.Debugln("Delete sql:", strSql)
	err := this.Open()
	if err != nil {
		return err
	}
	defer this.Close()
	result, err := this.Exec(strSql, append(this.colValues, this.whereVal...)...)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	logger.Debugf("成功删除了`%s`表 %d 条记录", this.tablename, affected)
	return nil
}

// Increment 增加/减少 某个字段的值
func (this *Dao) Increment(field string, num int) error {
	if num == 0 {
		return errors.New("dao Increment(`num`不能为0)")
	}
	where := this.where
	if where != "" {
		where = "WHERE " + where
	}
	setClause := fmt.Sprintf("`%s`=`%s`", field, field)
	if num > 0 {
		setClause += fmt.Sprintf("+%d", num)
	} else {
		setClause += fmt.Sprintf("-%d", -num)
	}
	strSql := fmt.Sprintf("UPDATE `%s` SET %s %s", this.tablename, setClause, where)
	logger.Debugln("Increment sql:", strSql)
	err := this.Open()
	if err != nil {
		return err
	}
	defer this.Close()
	result, err := this.Exec(strSql, this.whereVal...)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("dao Increment 没有更新任何数据！")
	}
	logger.Debugf("成功 increment `%s`表 %d 条记录", this.tablename, affected)
	return nil
}

// 获取总记录数
func (this *Dao) Count() (total int, err error) {
	strSql := util.CountSql(this)
	logger.Debugln("Count sql:", strSql)
	err = this.Open()
	if err != nil {
		return
	}
	defer this.Close()
	row := this.QueryRow(strSql, this.whereVal...)
	err = row.Scan(&total)
	return
}

// Find 查找单条数据
// colFieldMap 数据库表中列对应go中对象的字段
func (this *Dao) Find(colFieldMap map[string]interface{}, selectCol ...string) error {
	colNum := len(selectCol)
	if colNum == 0 || (colNum == 1 && selectCol[0] == "*") {
		selectCol = util.MapKeys(colFieldMap)
	}
	sort.Sort(sort.StringSlice(selectCol))
	this.selectCols = "`" + strings.Join(selectCol, "`,`") + "`"
	strSql := util.SelectSql(this)
	logger.Debugln("Find sql:", strSql)
	err := this.Open()
	if err != nil {
		return err
	}
	defer this.Close()
	row := this.QueryRow(strSql, this.whereVal...)
	scanInterface := make([]interface{}, 0, colNum)
	for _, column := range selectCol {
		scanInterface = append(scanInterface, colFieldMap[column])
	}
	err = row.Scan(scanInterface...)
	if err == sql.ErrNoRows {
		logger.Infoln("Find", strSql, ":no result ret")
		return nil
	}
	return err
}

// FindAll 查找多条数据
func (this *Dao) FindAll(selectCol ...string) (*sql.Rows, error) {
	sort.Sort(sort.StringSlice(selectCol))
	this.selectCols = "`" + strings.Join(selectCol, "`,`") + "`"
	strSql := util.SelectSql(this)
	logger.Debugln("FindAll sql:", strSql)
	logger.Debugln("FindAll bind params:", this.whereVal)
	err := this.Open()
	if err != nil {
		return nil, err
	}
	defer this.Close()
	return this.Query(strSql, this.whereVal...)
}

// 执行sql语句（查询语句）
func (this *Dao) DoSql(strSql string, args ...interface{}) (*sql.Rows, error) {
	err := this.Open()
	if err != nil {
		return nil, err
	}
	defer this.Close()
	return this.Query(strSql, args...)
}

// 用于FindAll中，具体model在遍历rows时调用（提取的公共代码）
func (this *Dao) Scan(rows *sql.Rows, colNum int, colFieldMap map[string]interface{}, selectCol ...string) error {
	scanInterface := make([]interface{}, 0, colNum)
	for _, column := range selectCol {
		scanInterface = append(scanInterface, colFieldMap[column])
	}
	return rows.Scan(scanInterface...)
}

// 执行 insert、update、delete 等操作
func (this *Dao) Exec(strSql string, args ...interface{}) (sql.Result, error) {
	err := this.Open()
	if err != nil {
		return nil, err
	}
	defer this.Close()
	return this.DB.Exec(strSql, args...)
}

// 持久化 entity 到数据库
func (this *Dao) Persist(entity interface{}) error {
	strSql, args, err := genPersistParams(entity)

	if err != nil {
		logger.Errorln("Persist error:", err)
		return err
	}

	logger.Debugln("Persist sql:", strSql, ";args:", args)

	err = this.Open()
	if err != nil {
		return err
	}
	defer this.Close()
	result, err := this.Exec(strSql, args...)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	logger.Debugf("成功更新了`%s`表 %d 条记录", this.tablename, affected)
	return nil
}

func genPersistParams(entity interface{}) (string, []interface{}, error) {
	strSql := "UPDATE `%s` SET %s WHERE %s"

	var (
		tablename string
		set       = make([]string, 0, 8)
		setArgs   = make([]interface{}, 0, 8)
		where     = make([]string, 0, 2)
		whereArgs = make([]interface{}, 0, 2)
	)

	entityType := reflect.TypeOf(entity)
	if entityType.Kind() != reflect.Ptr {
		return "", nil, fmt.Errorf("Persist(non-pointer %s)", entityType)
	}
	entityValue := reflect.Indirect(reflect.ValueOf(entity))
	if entityValue.Kind() != reflect.Struct {
		return "", nil, fmt.Errorf("Persist(non-struct %s)", entityValue)
	}
	entityType = entityValue.Type()
	fieldNum := entityType.NumField()
	for i := 0; i < fieldNum; i++ {
		fieldValue := entityValue.Field(i)

		// struct 字段的反射类型（StructField）
		fieldType := entityType.Field(i)

		if fieldType.Anonymous {
			tablenameField := (reflect.Indirect(fieldValue)).FieldByName("tablename")
			tablename = tablenameField.String()
			continue
		}
		// 非导出字段不处理
		if fieldType.PkgPath != "" {
			continue
		}
		tag := fieldType.Tag.Get("json")
		tags := strings.Split(tag, ",")

		pk := fieldType.Tag.Get("pk")

		// 字段本身的反射类型（field type）
		fieldValType := fieldType.Type
		switch fieldValType.Kind() {
		case reflect.Int:
			val := fieldValue.Int()

			if len(tags) > 1 && tags[1] == "omitempty" &&
				val == 0 {
				continue
			}

			if pk == "1" {
				whereArgs = append(whereArgs, val)
			} else {
				setArgs = append(setArgs, val)
			}
		case reflect.String:
			val := fieldValue.String()

			if len(tags) > 1 && tags[1] == "omitempty" &&
				val == "" {
				continue
			}

			if pk == "1" {
				whereArgs = append(whereArgs, val)
			} else {
				setArgs = append(setArgs, val)
			}
		default:
			// TODO：其他类型不处理
			continue
		}

		if pk == "1" {
			where = append(where, "`"+tag+"`"+"=?")
		} else {
			set = append(set, "`"+tags[0]+"`"+"=?")
		}
	}

	strSql = fmt.Sprintf(strSql, tablename, strings.Join(set, ","), strings.Join(where, " AND "))
	args := append(setArgs, whereArgs...)

	return strSql, args, nil
}

/*
// FindAllEx 查找多条数据
func (this *Dao) FindAllEx(selectCol ...string) (reflect.Value, error) {
    modelInterface, ok := ORM[this.tablename]
    if !ok {
        return nil, errors.New("No register orm rightly!")
    }
    modelType := reflect.TypeOf(modelInterface)
    modelValue := reflect.New(modelType)
    methodValue := modelValue.MethodByName("ColFieldMap")
    if !methodValue.IsValue() {
        return nil, errors.New("model struct should define method: ColFieldMap")
    }
    var colFieldMap map[string]interface{}
    methodValue.Call([]reflect.Value{reflect.ValueOf(colFieldMap)})
    if len(selectCol) == 0 {
        selectCol = util.MapKeys(colFieldMap)
    }
    sort.Sort(sort.StringSlice(selectCol))
    this.selectCols = strings.Join(selectCol, ",")
    strSql := util.SelectSql(this)
    err := this.Open()
    if err != nil {
        return nil, err
    }
    defer this.Close()
    stmt, err := this.Prepare(strSql)
    if err != nil {
        return nil, err
    }
    defer stmt.Close()
    rows, err := stmt.Query(this.whereVal...)
    if err != nil {
        return nil, err
    }

    modeValueList := reflect.MakeSlice(modelType, 0, 10)
    colNum := len(selectCol)
    for rows.Next() {
        modelValue = reflect.New(modelType)
        methodValue := modelValue.MethodByName("ColFieldMap")
        if !methodValue.IsValue() {
            return nil, errors.New("model struct should define method: ColFieldMap")
        }
        methodValue.Call([]reflect.Value{reflect.ValueOf(colFieldMap)})

        scanInterface := make([]interface{}, 0, colNum)
        for _, column := range selectCol {
            scanInterface = append(scanInterface, colFieldMap[column])
        }
        rows.Scan(scanInterface)
        modeValueList = reflect.Append(modeValueList, modelValue)
    }
    return modeValueList, nil
}
*/

func (this *Dao) Columns() []string {
	return this.columns
}

func (this *Dao) ColValues() []interface{} {
	return this.colValues
}

func (this *Dao) Where(condition string, args ...interface{}) {
	if len(args) == 0 {
		this._where(condition)
		return
	}

	this.where = condition
	this.whereVal = args
}

// 查询条件处理（TODO:暂时没有处理between）
// bug: aid='' 时，会自动去掉条件(FindAuthority)
// 兼容之前的写法
func (this *Dao) _where(condition string) {
	this.whereVal = make([]interface{}, 0, 5)
	stringBuilder := util.NewBuffer()
	conditions := SplitIn(condition, []string{" and ", " AND ", " or ", " OR "}...)
	for _, condition := range conditions {
		condition = strings.TrimSpace(condition)
		parts := SplitIn(condition, "=", "<", ">")
		if len(parts) >= 3 {
			// 处理不等于
			if strings.HasSuffix(parts[0], "!") {
				stringBuilder.Append("`" + strings.Trim(parts[0], "` !") + "` !")
			} else {
				stringBuilder.Append("`" + strings.Trim(parts[0], "` ") + "`")
			}
			stringBuilder.Append(strings.TrimSpace(parts[1]))
			if len(parts) > 3 {
				// 判断是不是 ">="或"<="
				if strings.ContainsAny(parts[2], "= & < & >") {
					stringBuilder.Append(strings.TrimSpace(parts[2]))
				}
				start := len(parts[0]) + len(parts[1]) + 1
				this.whereVal = append(this.whereVal, strings.TrimSpace(condition[start:]))
			} else {
				this.whereVal = append(this.whereVal, strings.TrimSpace(parts[2]))
			}
			stringBuilder.Append("?")
		} else {
			tmp := strings.ToUpper(parts[0])
			if tmp == "OR" || tmp == "AND" {
				stringBuilder.Append(" ").Append(tmp).Append(" ")
			} else {
				// 处理"in"语句（TODO:用正则处理？）
				if strings.ContainsAny(strings.ToLower(parts[0]), "in & ( & )") {
					ins := Split(parts[0], "(", ")")
					if len(ins) == 3 {
						inVals := strings.Split(ins[1], ",")
						for _, inVal := range inVals {
							this.whereVal = append(this.whereVal, inVal)
						}
						// in中有多少个值
						inLen := len(inVals)
						qms := strings.Repeat("?,", inLen)
						field := ins[0][:len(ins[0])-3]
						stringBuilder.Append("`" + strings.Trim(field, "` ") + "` in").Append("(").Append(qms[:len(qms)-1]).Append(")")
					}
				} else {
					stringBuilder.Append("`" + strings.Trim(parts[0], "` ") + "`")
				}
			}
		}
	}
	this.where = stringBuilder.String()
}

// 更新操作的SET部分
func (this *Dao) Set(clause string, args ...interface{}) {
	if len(args) == 0 {
		this._set(clause)
		return
	}

	this.columns = strings.Split(clause, ",")

	if len(this.columns) != len(args) {
		this.columns = nil
		return
	}

	this.colValues = args
}

// 兼容之前的写法
func (this *Dao) _set(clause string) {
	clauses := strings.Split(clause, ",")
	if len(clauses) >= len(strings.Split(clause, "=")) {
		this.columns = nil
		return
	}

	for _, clause := range clauses {
		parts := strings.Split(clause, "=")
		// 如果参数不合法，让执行的sql报错
		if len(parts) != 2 {
			this.columns = nil
			return
		}
		parts[0] = strings.TrimFunc(parts[0], func(r rune) bool {
			switch r {
			case ' ', '`':
				return true
			}
			return false
		})
		this.columns = append(this.columns, "`"+parts[0]+"`=?")
		this.colValues = append(this.colValues, strings.TrimSpace(parts[1]))
	}
}

func (this *Dao) SelectCols() string {
	if this.selectCols == "" {
		return "*"
	}
	return this.selectCols
}

func (this *Dao) GetWhere() string {
	return this.where
}

func (this *Dao) Order(order string) {
	this.order = order
}

func (this *Dao) GetOrder() string {
	return this.order
}

func (this *Dao) Limit(limit string) {
	this.limit = limit
}

func (this *Dao) GetLimit() string {
	return this.limit
}

func (this *Dao) Tablename() string {
	return this.tablename
}

func Split(s string, seps ...string) []string {
	count := len(seps)
	if count == 0 {
		return []string{s}
	}
	if count == 1 {
		return strings.Split(s, seps[0])
	}

	result := []string{}

	strSlice := strings.Split(s, seps[0])
	for _, str := range strSlice {
		if strings.TrimSpace(str) == "" {
			continue
		}
		result = append(result, Split(str, seps[1:]...)...)
	}
	return result
}

func SplitIn(s string, seps ...string) []string {
	count := len(seps)
	if count == 0 {
		return []string{s}
	}
	if count == 1 {
		tmpSlice := strings.Split(s, seps[0])
		count = len(tmpSlice)
		total := 2*count - 1
		tmpResult := make([]string, 0, total)
		for i := 0; i < count; i++ {
			if strings.TrimSpace(tmpSlice[i]) == "" {
				continue
			}
			tmpResult = append(tmpResult, tmpSlice[i])
			// 只有最后一个后面才不加入
			if i < count-1 {
				tmpResult = append(tmpResult, seps[0])
			}
		}
		return tmpResult
	}

	result := []string{}

	strSlice := strings.Split(s, seps[0])
	tmpCount := len(strSlice)
	for i, str := range strSlice {
		if strings.TrimSpace(str) == "" {
			continue
		}
		result = append(result, SplitIn(str, seps[1:]...)...)
		if i < tmpCount-1 {
			result = append(result, seps[0])
		}
	}
	return result
}
