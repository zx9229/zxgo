package main

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/zx9229/zxgo/zxre"
)

type QtSqliteField struct {
	QtDataType         string   //Qt中的字段类型.
	QtDataName         string   //Qt中的字段名.
	SqliteValid        bool     //该字段是否在Sqlite中.
	ObjectTableName    bool     //(使用)对象中存储的tableName吗.
	SqliteType         string   //这一列在数据库中的类型(INTEGER,REAL,TEXT,BLOB).
	SqliteNotNull      bool     //这一列是(NULL)还是(NOT NULL)
	SqlitePk           bool     //这一列是(PRIMARY KEY)吗
	SqliteOtherOptions []string //sqlite的其他属性
}

func newQtSqliteField() *QtSqliteField {
	newData := &QtSqliteField{"", "", false, false, "", false, false, make([]string, 0)}
	return newData
}

func (self *QtSqliteField) calc_SqliteNotNull() string {
	if self.SqliteValid {
		if self.SqliteNotNull {
			return "NOT NULL"
		} else {
			return "NULL"
		}
	} else {
		return ""
	}
}

func (self *QtSqliteField) calc_SqlitePk() string {
	if self.SqlitePk {
		return "PRIMARY KEY"
	} else {
		return ""
	}
}

func (self *QtSqliteField) calc_SqliteSet() string {
	if self.SqliteValid {
		slice_ := make([]string, 0)
		slice_ = append(slice_, self.SqliteType)
		slice_ = append(slice_, self.calc_SqliteNotNull())
		if self.SqlitePk {
			slice_ = append(slice_, self.calc_SqlitePk())
		}
		slice_ = append(slice_, self.SqliteOtherOptions...)
		return "`" + strings.Join(slice_, ",") + "`"
	} else {
		return "``"
	}
}

func isSqliteDatatype(content string) bool {
	return (content == "INTEGER" || content == "REAL" || content == "TEXT" || content == "BLOB")
}

func isNullOrNotNull(content string) (matched bool, isNull bool) {
	matched, err := regexp.MatchString("^(NULL)|(NOT[ \t]+NULL)$", content)
	if err != nil {
		panic(fmt.Sprintf("content=%v,err=%v", content, err))
	}
	isNull = (content == "NULL")
	return
}

func isPrimaryKey(content string) bool {
	matched, err := regexp.MatchString("^(PK)|(PRIMARY[ \t]+KEY)$", content)
	if err != nil {
		panic(fmt.Sprintf("content=%v,err=%v", content, err))
	}
	return matched
}

func (self *QtSqliteField) parseSqliteOptions(content string) error {
	fields := strings.Split(content, ",")

	for _, field := range fields {
		field = strings.Trim(field, " \t")
		if len(field) == 0 {
			continue
		}

		if isSqliteDatatype(field) {
			self.SqliteType = field
		} else if matched, isNull := isNullOrNotNull(field); matched {
			self.SqliteNotNull = !isNull
		} else if isPrimaryKey(field) {
			self.SqlitePk = true
		} else {
			self.SqliteOtherOptions = append(self.SqliteOtherOptions, field)
		}
	}
	if len(self.SqliteType) == 0 {
		return errors.New("can not find SQLite Datatype")
	} else {
		return nil
	}
}

func (self *QtSqliteField) parseContent_1(line string) (matched bool, err error) {
	matched = false
	err = nil
	return
}

func (self *QtSqliteField) parseContent_2(line string) (matched bool, err error) {
	pattern := "^[ \t]*(?P<QtDataType>[a-zA-Z0-9_: ]+?)[ \t]+(?P<QtDataName>[a-zA-Z0-9_:]+);[ \t]*//`(?P<SqliteOptions>[a-zA-Z0-9_, \t]*?)`"
	allGroupMap := zxre.CalcAllGroupDict(pattern, line)
	if len(allGroupMap) != 1 {
		matched = false
		err = nil
		return
	}

	matched = true

	var ok bool

	if self.QtDataType, ok = allGroupMap[0]["QtDataType"]; !ok || len(self.QtDataType) == 0 {
		err = errors.New(fmt.Sprintf("can not find QtDataType, content=%v", line))
		return
	}

	if self.QtDataName, ok = allGroupMap[0]["QtDataName"]; !ok || len(self.QtDataName) == 0 {
		err = errors.New(fmt.Sprintf("can not find QtDataName, content=%v", line))
		return
	}

	if SqliteOptions, ok := allGroupMap[0]["SqliteOptions"]; !ok {
		err = errors.New(fmt.Sprintf("can not find SqliteOptions, content=%v", line))
		return
	} else {
		if len(SqliteOptions) == 0 {
			self.SqliteValid = false
			self.ObjectTableName = false
		} else if 0 <= strings.Index(SqliteOptions, "ZX_TABLENAME") {
			self.SqliteValid = false
			self.ObjectTableName = true
		} else {
			self.SqliteValid = true
			self.ObjectTableName = false
			if err = self.parseSqliteOptions(SqliteOptions); err != nil {
				err = errors.New(fmt.Sprintf("parseSqliteOptions, %v, content=%v", err, line))
			}
		}
	}

	return
}

func (self *QtSqliteField) parseContent(line string) error {
	if matched, err := self.parseContent_1(line); matched {
		return err
	}

	if matched, err := self.parseContent_2(line); matched {
		return err
	}

	return errors.New("can not parse it.")
}

func (self *QtSqliteField) generate_create_table_sql_field_without_pk(maxFieldLen int) string {
	format := `%-` + strconv.Itoa(maxFieldLen) + `s %-7s %8s %v`
	content := fmt.Sprintf(format, self.QtDataName, self.SqliteType, self.calc_SqliteNotNull(), strings.Join(self.SqliteOtherOptions, " "))
	return content
}

func (self *QtSqliteField) generate_create_table_sql_field_with_pk(maxFieldLen int) string {
	format := `%-` + strconv.Itoa(maxFieldLen) + `s %-7s %8s %s %v`
	content := fmt.Sprintf(format, self.QtDataName, self.SqliteType, self.calc_SqliteNotNull(), self.calc_SqlitePk(), strings.Join(self.SqliteOtherOptions, " "))
	return content
}

func (self *QtSqliteField) generate_get_data_field(INDENT string, DELIMITER string) []string {
	slice_ := make([]string, 0)
	line := ""
	if self.QtDataType == "QString" {
		line = fmt.Sprintf(`currData.%s = query.value("%s").toString();`, self.QtDataName, self.QtDataName)
		slice_ = append(slice_, line)
	} else if self.QtDataType == "bool" {
		line = fmt.Sprintf(`currData.%s = query.value("%s").toBool();`, self.QtDataName, self.QtDataName)
		slice_ = append(slice_, line)
	} else if self.QtDataType == "int" {
		line = fmt.Sprintf(`currData.%s = query.value("%s").toInt(&isOk);`, self.QtDataName, self.QtDataName)
		slice_ = append(slice_, line)
		line = fmt.Sprintf(`if (!isOk) { currData.idq_%s = false; currData.%s = 0; }`, self.QtDataName, self.QtDataName)
		slice_ = append(slice_, line)
	} else if self.QtDataType == "unsigned int" {
		line = fmt.Sprintf(`currData.%s = query.value("%s").toUInt(&isOk);`, self.QtDataName, self.QtDataName)
		slice_ = append(slice_, line)
		line = fmt.Sprintf(`if (!isOk) { currData.idq_%s = false; currData.%s = 0; }`, self.QtDataName, self.QtDataName)
		slice_ = append(slice_, line)
	} else if self.QtDataType == "long long" {
		line = fmt.Sprintf(`currData.%s = query.value("%s").toLongLong(&isOk);`, self.QtDataName, self.QtDataName)
		slice_ = append(slice_, line)
		line = fmt.Sprintf(`if (!isOk) { currData.idq_%s = false; currData.%s = 0; }`, self.QtDataName, self.QtDataName)
		slice_ = append(slice_, line)
	} else if self.QtDataType == "unsigned long long" {
		line = fmt.Sprintf(`currData.%s = query.value("%s").toULongLong(&isOk);`, self.QtDataName, self.QtDataName)
		slice_ = append(slice_, line)
		line = fmt.Sprintf(`if (!isOk) { currData.idq_%s = false; currData.%s = 0; }`, self.QtDataName, self.QtDataName)
		slice_ = append(slice_, line)
	} else if self.QtDataType == "float" {
		line = fmt.Sprintf(`currData.%s = query.value("%s").toFloat(&isOk);`, self.QtDataName, self.QtDataName)
		slice_ = append(slice_, line)
		line = fmt.Sprintf(`if (!isOk) { currData.idq_%s = false; currData.%s = 0; }`, self.QtDataName, self.QtDataName)
		slice_ = append(slice_, line)
	} else if self.QtDataType == "double" {
		line = fmt.Sprintf(`currData.%s = query.value("%s").toDouble(&isOk);`, self.QtDataName, self.QtDataName)
		slice_ = append(slice_, line)
		line = fmt.Sprintf(`if (!isOk) { currData.idq_%s = false; currData.%s = 0; }`, self.QtDataName, self.QtDataName)
		slice_ = append(slice_, line)
	} else {
		panic(fmt.Sprintf("QtDataType=%v,QtDataName=%v", self.QtDataType, self.QtDataName))
	}
	return slice_
}
