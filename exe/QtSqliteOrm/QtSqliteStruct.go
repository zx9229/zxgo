package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/zx9229/zxgo/zxre"
)

type QtSqliteStruct struct {
	QtClassName string //Qt中的类名.
	Tablename   string
	Fields      []*QtSqliteField
}

func newQtSqliteStruct() *QtSqliteStruct {
	newData := &QtSqliteStruct{"", "", make([]*QtSqliteField, 0)}
	return newData
}

func (self *QtSqliteStruct) parseContent(line string) error {

	if len(self.QtClassName) == 0 {
		patternStruct := "^struct[ \t]+(?P<QtClassName>[a-zA-Z0-9_:]+)[ \t]*//`(?P<Tablename>[a-zA-Z0-9_]+)`"
		allGroupMap := zxre.CalcAllGroupDict(patternStruct, line)
		if len(allGroupMap) != 1 {
			return errors.New(fmt.Sprintf("logical error, content=%v", line))
		}

		var ok bool

		if self.QtClassName, ok = allGroupMap[0]["QtClassName"]; !ok || len(self.QtClassName) == 0 {
			return errors.New(fmt.Sprintf("logical error, content=%v", line))
		}

		if self.Tablename, ok = allGroupMap[0]["Tablename"]; !ok || len(self.Tablename) == 0 {
			return errors.New(fmt.Sprintf("logical error, content=%v", line))
		}

	} else {
		if strings.HasPrefix(line, "{") {
			return nil
		}

		currField := newQtSqliteField()
		if err := currField.parseContent(line); err != nil {
			return err
		}

		self.Fields = append(self.Fields, currField)
	}
	return nil
}

func repeatContent(content string, num int) string {
	newContent := ""
	for i := 0; i < num; i++ {
		newContent += content
	}
	return newContent
}

func (self *QtSqliteStruct) calc_maxFieldLen() int {
	maxLen := 0
	for _, fieldObj := range self.Fields {
		if !fieldObj.SqliteValid {
			continue
		}
		curLen := len(fieldObj.QtDataName)
		if maxLen < curLen {
			maxLen = curLen
		}
	}
	return maxLen
}

func (self *QtSqliteStruct) generate_table_name() string {
	INDENT := repeatContent("    ", 1)
	DELIMITER := "\r\n"
	temp_str := ""

	content := ""
	content += INDENT + "static QString table_name()" + DELIMITER
	content += INDENT + "{" + DELIMITER
	temp_str = INDENT + `    return "%v";` + DELIMITER
	content += fmt.Sprintf(temp_str, self.Tablename)
	content += INDENT + "};" + DELIMITER

	return content
}

func (self *QtSqliteStruct) generate_create_table_sql_pk() string {
	content := ""

	fields4pk := make([]string, 0)
	for _, fieldObj := range self.Fields {
		if !fieldObj.SqliteValid {
			continue
		}
		if fieldObj.SqlitePk {
			fields4pk = append(fields4pk, fieldObj.QtDataName)
		}
	}

	if 0 < len(fields4pk) {
		content = fmt.Sprintf("PRIMARY KEY(%v)", strings.Join(fields4pk, ","))
	}

	return content
}

func (self *QtSqliteStruct) generate_create_table_sql() string {
	INDENT := repeatContent("    ", 1)
	DELIMITER := "\r\n"
	temp_str := ""

	maxFieldLen := self.calc_maxFieldLen()

	content := ""
	content += INDENT + "static QString create_table_sql()" + DELIMITER
	content += INDENT + "{" + DELIMITER
	content += INDENT + `    QString sql = QObject::tr("CREATE TABLE IF NOT EXISTS %1 (\` + DELIMITER
	temp_str = INDENT + `    %v,\` + DELIMITER
	for _, fieldObj := range self.Fields {
		if !fieldObj.SqliteValid {
			continue
		}
		content += fmt.Sprintf(temp_str, fieldObj.generate_create_table_sql_field(maxFieldLen))
	}
	temp_str = INDENT + `    %v  )").QString::arg(table_name());` + DELIMITER
	content += fmt.Sprintf(temp_str, self.generate_create_table_sql_pk())
	content += INDENT + "    return sql;" + DELIMITER
	content += INDENT + "};" + DELIMITER

	return content
}

func (self *QtSqliteStruct) generate_insert_sql() string {
	INDENT := repeatContent("    ", 1)
	DELIMITER := "\r\n"
	temp_str := ""

	content := ""
	content += INDENT + "QString insert_sql(bool insertOrReplace)" + DELIMITER
	content += INDENT + "{" + DELIMITER
	content += INDENT + `    QString sqlKeyword = insertOrReplace ? QObject::tr("INSERT OR REPLACE INTO") : QObject::tr("INSERT INTO");` + DELIMITER
	content += INDENT + `    QString strKey, strVal;` + DELIMITER
	temp_str = INDENT + `    if (this->idq_%s) { strKey += "%s,"; strVal += QObject::tr("'%%1',").QString::arg(this->%s); }` + DELIMITER
	for _, fieldObj := range self.Fields {
		if !fieldObj.SqliteValid {
			continue
		}
		content += fmt.Sprintf(temp_str, fieldObj.QtDataName, fieldObj.QtDataName, fieldObj.QtDataName)
	}
	content += INDENT + "    strKey.chop(1);" + DELIMITER
	content += INDENT + "    strVal.chop(1);" + DELIMITER
	content += INDENT + `    QString sql = QObject::tr("%1 %2(%3) VALUES(%4)").QString::arg(sqlKeyword).QString::arg(table_name()).QString::arg(strKey).QString::arg(strVal);` + DELIMITER
	content += INDENT + "    return sql;" + DELIMITER
	content += INDENT + "};" + DELIMITER

	return content
}

func (self *QtSqliteStruct) generate_insert_data() string {
	INDENT := repeatContent("    ", 1)
	DELIMITER := "\r\n"

	content := ""
	content += INDENT + "bool insert_data(QSqlQuery& query, bool insertOrReplace = false)" + DELIMITER
	content += INDENT + "{" + DELIMITER
	content += INDENT + `    return query.exec(insert_sql(insertOrReplace));` + DELIMITER
	content += INDENT + "};" + DELIMITER

	return content
}

func (self *QtSqliteStruct) generate_delete_sql() string {
	INDENT := repeatContent("    ", 1)
	DELIMITER := "\r\n"
	temp_str := ""

	content := ""
	content += INDENT + "QString delete_sql()" + DELIMITER
	content += INDENT + "{" + DELIMITER
	content += INDENT + `    QString sql = QObject::tr("DELETE FROM %1 WHERE 1=1 ").QString::arg(table_name());` + DELIMITER
	temp_str = INDENT + `    if (this->idq_%s) { sql += QObject::tr("AND %s='%%1' ").QString::arg(this->%s); }` + DELIMITER
	for _, fieldObj := range self.Fields {
		if !fieldObj.SqliteValid {
			continue
		}
		content += fmt.Sprintf(temp_str, fieldObj.QtDataName, fieldObj.QtDataName, fieldObj.QtDataName)
	}
	content += INDENT + "    return sql;" + DELIMITER
	content += INDENT + "};" + DELIMITER

	return content
}

func (self *QtSqliteStruct) generate_delete_data() string {
	INDENT := repeatContent("    ", 1)
	DELIMITER := "\r\n"

	content := ""
	content += INDENT + "bool delete_data(QSqlQuery& query)" + DELIMITER
	content += INDENT + "{" + DELIMITER
	content += INDENT + `    return query.exec(delete_sql());` + DELIMITER
	content += INDENT + "};" + DELIMITER

	return content
}

func (self *QtSqliteStruct) generate_get_data() string {
	INDENT := repeatContent("    ", 1)
	DELIMITER := "\r\n"
	temp_str := ""

	content := ""
	temp_str = INDENT + `static void get_data(QSqlQuery& query, QList<%s>& dataOut)` + DELIMITER
	content += fmt.Sprintf(temp_str, self.QtClassName)
	content += INDENT + "{" + DELIMITER
	content += INDENT + "    while (query.next())" + DELIMITER
	content += INDENT + "    {" + DELIMITER
	content += INDENT + "        bool isOk = false;" + DELIMITER
	temp_str = INDENT + "        %s currData;" + DELIMITER
	content += fmt.Sprintf(temp_str, self.QtClassName)
	content += INDENT + "        currData.flush_flag(true);" + DELIMITER
	temp_str = INDENT + "        %s" + DELIMITER
	for _, fieldObj := range self.Fields {
		if !fieldObj.SqliteValid {
			continue
		}
		allfieldLine := fieldObj.generate_get_data_field(INDENT, DELIMITER)
		for _, fieldLine := range allfieldLine {
			content += fmt.Sprintf(temp_str, fieldLine)
		}
	}
	content += INDENT + "        dataOut.append(currData);" + DELIMITER
	content += INDENT + "    }" + DELIMITER
	content += INDENT + "};" + DELIMITER

	return content
}

func (self *QtSqliteStruct) generate_query_sql() string {
	INDENT := repeatContent("    ", 1)
	DELIMITER := "\r\n"
	temp_str := ""

	content := ""
	content += INDENT + "QString query_sql()" + DELIMITER
	content += INDENT + "{" + DELIMITER
	content += INDENT + `    QString sql = QObject::tr("SELECT * FROM %1 WHERE 1=1 ").QString::arg(table_name());` + DELIMITER
	temp_str = INDENT + `    if (this->idq_%s) { sql += QObject::tr("AND %s='%%1' ").QString::arg(this->%s); }` + DELIMITER
	for _, fieldObj := range self.Fields {
		if !fieldObj.SqliteValid {
			continue
		}
		content += fmt.Sprintf(temp_str, fieldObj.QtDataName, fieldObj.QtDataName, fieldObj.QtDataName)
	}
	content += INDENT + "    return sql;" + DELIMITER
	content += INDENT + "};" + DELIMITER

	return content
}

func (self *QtSqliteStruct) generate_query_data() string {
	INDENT := repeatContent("    ", 1)
	DELIMITER := "\r\n"
	temp_str := ""

	content := ""
	temp_str = INDENT + `void query_data(QSqlQuery& query, QList<%s>& dataOut)` + DELIMITER
	content += fmt.Sprintf(temp_str, self.QtClassName)
	content += INDENT + "{" + DELIMITER
	content += INDENT + `    if (query.exec(query_sql()))` + DELIMITER
	content += INDENT + "    {" + DELIMITER
	content += INDENT + "        get_data(query, dataOut);" + DELIMITER
	content += INDENT + "    }" + DELIMITER
	content += INDENT + "};" + DELIMITER

	return content
}

func (self *QtSqliteStruct) generate_flush_flag() string {
	INDENT := repeatContent("    ", 1)
	DELIMITER := "\r\n"
	temp_str := ""

	content := ""
	content += INDENT + "void flush_flag(bool flagValue)" + DELIMITER
	content += INDENT + "{" + DELIMITER
	temp_str = INDENT + `    this->idq_%s = flagValue;` + DELIMITER
	for _, fieldObj := range self.Fields {
		if !fieldObj.SqliteValid {
			continue
		}
		content += fmt.Sprintf(temp_str, fieldObj.QtDataName)
	}
	content += INDENT + "};" + DELIMITER

	return content
}

func (self *QtSqliteStruct) generate_pk_equal() string {
	INDENT := repeatContent("    ", 1)
	DELIMITER := "\r\n"
	temp_str := ""

	fields4pk := make([]string, 0)
	for _, fieldObj := range self.Fields {
		if !fieldObj.SqliteValid {
			continue
		}
		if fieldObj.SqlitePk {
			fields4pk = append(fields4pk, fieldObj.QtDataName)
		}
	}

	content := ""
	temp_str = INDENT + `bool pk_equal(const %s& other) const` + DELIMITER
	content += fmt.Sprintf(temp_str, self.QtClassName)
	content += INDENT + "{" + DELIMITER
	if len(fields4pk) == 0 {
		content += INDENT + "    return false;" + DELIMITER
	} else {
		content += INDENT + "    if (" + DELIMITER
		temp_str = INDENT + "        (this->%s == other.%s) &&" + DELIMITER
		for _, field4pk := range fields4pk {
			content += fmt.Sprintf(temp_str, field4pk, field4pk)
		}
		content += INDENT + "        true)" + DELIMITER
		content += INDENT + "        return true;" + DELIMITER
		content += INDENT + "    else" + DELIMITER
		content += INDENT + "        return false;" + DELIMITER
	}
	content += INDENT + "};" + DELIMITER

	return content
}

func (self *QtSqliteStruct) generate_cxx_definition_members() string {
	INDENT := repeatContent("    ", 1)
	DELIMITER := "\r\n"

	content := ""
	tmpStr1 := INDENT + `%s %s;//%s` + DELIMITER
	tmpStr2 := INDENT + `bool idq_%s;` + DELIMITER
	for _, fieldObj := range self.Fields {
		content += fmt.Sprintf(tmpStr1, fieldObj.QtDataType, fieldObj.QtDataName, fieldObj.calc_SqliteSet())
		if !fieldObj.SqliteValid {
			continue
		}
		content += fmt.Sprintf(tmpStr2, fieldObj.QtDataName)
	}

	return content
}

func (self *QtSqliteStruct) generate_cxx_definition() string {
	DELIMITER := "\r\n"
	content := ""
	content += fmt.Sprintf("class %s", self.QtClassName) + DELIMITER
	content += "{" + DELIMITER
	content += "public:" + DELIMITER
	content += self.generate_cxx_definition_members()
	content += "public:" + DELIMITER
	content += self.generate_table_name()
	content += self.generate_create_table_sql()
	content += self.generate_insert_sql()
	content += self.generate_insert_data()
	content += self.generate_delete_sql()
	content += self.generate_delete_data()
	content += self.generate_get_data()
	content += self.generate_query_sql()
	content += self.generate_query_data()
	content += self.generate_flush_flag()
	content += self.generate_pk_equal()

	content += "};" + DELIMITER

	return content
}
