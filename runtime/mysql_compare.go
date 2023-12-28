package runtime

import (
	"fmt"
	"reflect"
	"strings"
)

//DBCompareSQL sql
type DBCompareSQL struct {
	SQL   string
	Table string
}

// DBCompareResultItem 比较结果项
type DBCompareResultItem struct {
	Comment string
	SQLList []DBCompareSQL
}

// DBCompareIndex 索引比较结果项
type DBCompareIndex struct {
	Type         string
	Fields       []string
	FieldsString string
}

// DBCompareService 数据库比较
type DBCompareService struct {
	//新数据
	LeftList map[string]MySQLTable
	//老数据
	RightList map[string]MySQLTable

	tableCreateSQL []DBCompareSQL
	tableModifySQL []DBCompareSQL
	tableDropSQL   []DBCompareSQL

	fieldAddSQL      []DBCompareSQL
	fieldModifySQL   []DBCompareSQL
	fieldDropSQL     []DBCompareSQL
	fieldPositionSQL []DBCompareSQL

	indexAddSQL    []DBCompareSQL
	indexModifySQL []DBCompareSQL
	indexDropSQL   []DBCompareSQL

	tableCollateSQL []DBCompareSQL

	defaultCharacter     string
	compareFieldAttrList []string
}

func (c *DBCompareService) init() {
	c.tableCreateSQL = []DBCompareSQL{}
	c.tableModifySQL = []DBCompareSQL{}
	c.tableDropSQL = []DBCompareSQL{}

	c.fieldAddSQL = []DBCompareSQL{}
	c.fieldModifySQL = []DBCompareSQL{}
	c.fieldDropSQL = []DBCompareSQL{}
	c.fieldPositionSQL = []DBCompareSQL{}

	c.indexAddSQL = []DBCompareSQL{}
	c.indexModifySQL = []DBCompareSQL{}
	c.indexDropSQL = []DBCompareSQL{}

	c.tableCollateSQL = []DBCompareSQL{}

	c.defaultCharacter = "utf8mb4_general_ci"
	c.compareFieldAttrList = []string{"data_type", "is_nullable", "column_default", "column_comment", "extra" /*, "collation_name"*/}
}
func (c *DBCompareService) tableExist(name string, tableList map[string]MySQLTable) bool {
	_, err := tableList[name]
	return err
}
func (c *DBCompareService) isTable(table MySQLTable) bool {
	return table.TABLE_TYPE == "BASE TABLE"
}

// Compare 比较数据库
func (c *DBCompareService) Compare() string {
	c.init()
	// FilePutContents("./right.json", JSONEncode(c.RightList))
	// FilePutContents("./left.json", JSONEncode(c.LeftList))

	//删除表
	for key, val := range c.RightList {
		if !c.isTable(val) {
			continue
		}

		if !c.tableExist(key, c.LeftList) {
			c.tableDropSQL = append(c.tableDropSQL, DBCompareSQL{
				SQL:   "drop table `" + key + "`;",
				Table: key,
			})
		}
	}

	for key, val := range c.LeftList {
		if !c.isTable(val) {
			continue
		}

		//创建表
		if !c.tableExist(key, c.RightList) {
			c.tableCreateSQL = append(c.tableCreateSQL, DBCompareSQL{
				SQL:   c.getTableCreateSQL(val),
				Table: key,
			})
			continue
		}
		//修改表
		alterSQLList := c.getTableAlterCollateSQL(c.RightList[key], val)
		if len(alterSQLList) > 0 {
			c.tableModifySQL = append(c.tableModifySQL, DBCompareSQL{
				SQL:   fmt.Sprintf("alter table `%s` %s;", val.TABLE_NAME, Implode(",", alterSQLList)),
				Table: val.TABLE_NAME,
			})
		}
		//检查字段
		c.compareColumnList(val.COLUMN_LIST, c.RightList[key].COLUMN_LIST)
		//检查索引
		c.compareIndexList(val.TABLE_NAME, val.INDEX_LIST, c.RightList[key].INDEX_LIST)
	}

	return c.outputString()
}

func (c *DBCompareService) outputString() string {
	result := ""
	res := []DBCompareResultItem{
		{Comment: "删除表", SQLList: c.tableDropSQL},
		{Comment: "创建表", SQLList: c.tableCreateSQL},
		{Comment: "修改表", SQLList: c.tableModifySQL},
		{Comment: "创建字段", SQLList: c.fieldAddSQL},
		{Comment: "修改字段", SQLList: c.fieldModifySQL},
		{Comment: "修正字段位置", SQLList: c.fieldPositionSQL},
		{Comment: "删除字段", SQLList: c.fieldDropSQL},
		{Comment: "创建索引", SQLList: c.indexAddSQL},
		{Comment: "修改索引", SQLList: c.indexModifySQL},
		{Comment: "删除索引", SQLList: c.indexDropSQL},
		{Comment: "调整表", SQLList: c.tableCollateSQL},
	}
	for _, item := range res {
		if len(item.SQLList) > 0 {
			result += "-- " + item.Comment + "\n"
			var sqlList []string
			for _, sqlItem := range item.SQLList {
				sqlList = append(sqlList, sqlItem.SQL)
			}
			result += Implode("\n", sqlList) + "\n\n"
		}
	}
	result = strings.TrimSpace(result) + "\n\n"
	return result
}

func (c *DBCompareService) compareIndexList(tableName string, leftIndexes []MySQLIndex, rightIndexes []MySQLIndex) {
	leftIndexList := c.getIndexesList(leftIndexes)
	rightIndexList := c.getIndexesList(rightIndexes)
	for key, val := range leftIndexList {
		indexType := "index"
		if len(val.Type) > 0 {
			indexType = val.Type + " index"
		}

		if _, err := rightIndexList[key]; !err { //add
			c.indexAddSQL = append(c.indexAddSQL, DBCompareSQL{
				SQL:   fmt.Sprintf("alter table `%s` add %s `%s` (%s);", tableName, indexType, key, val.FieldsString),
				Table: tableName,
			})
		} else if val.FieldsString != rightIndexList[key].FieldsString { //modify
			c.indexModifySQL = append(c.indexModifySQL, DBCompareSQL{
				SQL:   fmt.Sprintf("alter table `%s` drop index `%s`,add %s `%s` (%s);", tableName, key, indexType, key, val.FieldsString),
				Table: tableName,
			})
		}
	}
	//drop
	for key := range rightIndexList {
		if _, err := leftIndexList[key]; !err {
			c.indexDropSQL = append(c.indexDropSQL, DBCompareSQL{
				SQL:   fmt.Sprintf("alter table `%s` drop index `%s`;", tableName, key),
				Table: tableName,
			})
		}
	}
}

func (c *DBCompareService) getColumnPositionSQL(columnList []MySQLColumn, columnName string) string {
	for key, val := range columnList {
		if columnName == val.COLUMN_NAME {
			if StringToInt(val.ORDINAL_POSITION) == 1 {
				return " first"
			}
			// FilePutContents("./mysql.txt", JSONEncode(map[string]interface{}{
			// 	"columnList": columnList,
			// 	"key":        key,
			// 	"columnName": columnName,
			// }))

			return fmt.Sprintf(" after `%s`", columnList[key-1].COLUMN_NAME)
		}
	}
	return ""
}

func getColumnNameList(columns []MySQLColumn) []string {
	var result []string
	for _, val := range columns {
		result = append(result, val.COLUMN_NAME)
	}
	return result
}

func (c *DBCompareService) compareColumnList(leftColumnList []MySQLColumn, rightColumnList []MySQLColumn) {
	leftColumnNameList := map[string]MySQLColumn{}
	rightColumnNameList := map[string]MySQLColumn{}
	for _, val := range leftColumnList {
		leftColumnNameList[val.COLUMN_NAME] = val
	}
	for _, val := range rightColumnList {
		rightColumnNameList[val.COLUMN_NAME] = val
	}
	modifyRightColumnList := rightColumnList

	for _, val := range leftColumnList {
		key := val.COLUMN_NAME
		//add
		if _, err := rightColumnNameList[key]; !err {
			//创建字段
			val.COLLATION_NAME = c.defaultCharacter
			c.fieldAddSQL = append(c.fieldAddSQL, DBCompareSQL{
				SQL:   fmt.Sprintf("alter table `%s` add column %s;", val.TABLE_NAME, c.getColumnSQL(val, false)),
				Table: val.TABLE_NAME,
			})
			modifyRightColumnList = append(modifyRightColumnList, val)

			continue
		}

		//modify
		var differentColumnList []string
		for _, attrName := range c.compareFieldAttrList {
			if reflect.ValueOf(val).FieldByName(StrToUpper(attrName)).String() != reflect.ValueOf(rightColumnNameList[key]).FieldByName(StrToUpper(attrName)).String() {
				differentColumnList = append(differentColumnList, attrName)
			}
		}
		if !InStringSlice("column_type", differentColumnList) && c.isFloatField(rightColumnNameList[key]) {
			differentColumnList = append(differentColumnList, "column_type")
		}
		if !c.isUtf8mb4Column(rightColumnNameList[key]) && c.isTextColumn(rightColumnNameList[key]) {
			differentColumnList = append(differentColumnList, "character_set_name")
		}

		if len(differentColumnList) > 0 {
			comment := "-- new:" + c.getColumnSQL(val, false) + "\n"
			comment += "-- old:" + c.getColumnSQL(rightColumnNameList[key], true) + "\n"

			positionSQL := c.getColumnPositionSQL(leftColumnList, key)

			val.COLLATION_NAME = c.defaultCharacter
			columnSql := c.getColumnSQL(val, false)

			c.fieldModifySQL = append(c.fieldModifySQL, DBCompareSQL{
				SQL:   fmt.Sprintf("%s alter table `%s` change column `%s` %s%s;", comment, val.TABLE_NAME, val.COLUMN_NAME, columnSql, positionSQL),
				Table: val.TABLE_NAME,
			})
		}
	}
	//drop
	for key, val := range rightColumnNameList {
		if _, err := leftColumnNameList[key]; !err {
			c.fieldDropSQL = append(c.fieldDropSQL, DBCompareSQL{
				SQL:   fmt.Sprintf("alter table `%s` drop column `%s`;", val.TABLE_NAME, val.COLUMN_NAME),
				Table: val.TABLE_NAME,
			})
			for k, v := range modifyRightColumnList {
				if v.COLUMN_NAME == key {
					modifyRightColumnList = append(modifyRightColumnList[:k], modifyRightColumnList[k+1:]...)
					break
				}
			}
		}
	}
	modifyTableName := ""
	var modifySQLList []string

	for {
		isSame := true
		modifyRightColumnTempList := []MySQLColumn{}
		for key, val := range leftColumnList {
			modifyRightColumnTempList = append(modifyRightColumnTempList, val)

			if modifyRightColumnList[key].COLUMN_NAME != val.COLUMN_NAME {
				isSame = false
				modifyTableName = val.TABLE_NAME

				positionSQL := c.getColumnPositionSQL(leftColumnList, val.COLUMN_NAME)
				val.COLLATION_NAME = c.defaultCharacter
				modifySQLList = append(modifySQLList, fmt.Sprintf("\tmodify column %s", c.getColumnSQL(val, false)+positionSQL))

				for k, v := range modifyRightColumnList {
					if k >= key && v.COLUMN_NAME != val.COLUMN_NAME {
						modifyRightColumnTempList = append(modifyRightColumnTempList, v)
					}
				}
				break
			}
		}
		modifyRightColumnList = modifyRightColumnTempList
		if isSame {
			if len(modifySQLList) > 0 {
				c.fieldPositionSQL = append(c.fieldPositionSQL, DBCompareSQL{
					SQL:   fmt.Sprintf("alter table `%s` \n%s;", modifyTableName, Implode(",\n", modifySQLList)),
					Table: modifyTableName,
				})
			}
			break
		}
	}

}

func (c *DBCompareService) getTableCreateSQL(table MySQLTable) string {
	var columnSQLList []string
	primaryKey := ""
	for _, column := range table.COLUMN_LIST {
		if primaryKey == "" && column.COLUMN_KEY == "PRI" {
			primaryKey = column.COLUMN_NAME
		}
		columnSQLList = append(columnSQLList, c.getColumnSQL(column, false))
	}

	columnSQLList = append(columnSQLList, c.getIndexSQL(table.INDEX_LIST, true)...)

	for key, val := range columnSQLList {
		columnSQLList[key] = "\n\t" + val
	}
	columnSQL := Implode(",", columnSQLList)
	result := fmt.Sprintf("create table `%s` (%s\n)", table.TABLE_NAME, columnSQL)
	createOptions := ""
	if len(table.CREATE_OPTIONS) > 0 {
		createOptions = " " + table.CREATE_OPTIONS
	}
	comment := ""
	if len(table.TABLE_COMMENT) > 0 {
		comment = fmt.Sprintf(" comment=\"%s\"", table.TABLE_COMMENT)
	}
	engine := table.ENGINE
	if StrToLower(engine) == "innodb" {
		engine = "MyISAM"
	}
	result += " engine=" + engine + " auto_increment=1 default collate=" + c.defaultCharacter + createOptions + comment + ";"
	return result
}

func (c *DBCompareService) getTableAlterCollateSQL(oldTable MySQLTable, newTable MySQLTable) []string {
	var result []string
	if oldTable.ENGINE != "MyISAM" {
		result = append(result, "engine=MyISAM")
	}

	if oldTable.TABLE_COLLATION != c.defaultCharacter {
		result = append(result, "collate="+c.defaultCharacter)
	}

	if oldTable.TABLE_COMMENT != newTable.TABLE_COMMENT {
		result = append(result, fmt.Sprintf("comment=\"%s\"", newTable.TABLE_COMMENT))
	}

	return result
}

func (c *DBCompareService) getColumnSQL(column MySQLColumn, origin bool) string {
	nullSQL := ""

	column.COLUMN_DEFAULT = strings.ReplaceAll(column.COLUMN_DEFAULT, "\n", "\\n")
	defaultSQL := "'" + TrimChar(column.COLUMN_DEFAULT, "'") + "'"
	if column.COLUMN_DEFAULT == "''" || IsEmpty(column.COLUMN_DEFAULT) {
		defaultSQL = "''"
	}
	if column.IS_NULLABLE == "NO" {
		nullSQL += " not null "
		if column.COLUMN_DEFAULT != "NULL" {
			nullSQL += " default " + defaultSQL
		}
	} else {
		if column.COLUMN_DEFAULT == "NULL" {
			nullSQL += " default null"
		} else {
			nullSQL += " default " + defaultSQL
		}
	}

	commentSQL := ""
	if len(column.COLUMN_COMMENT) > 0 {
		commentSQL = fmt.Sprintf(" comment \"%s\" ", c.addSlashes(column.COLUMN_COMMENT))
	}
	extraSQL := ""
	if len(column.EXTRA) > 0 && !InStringSlice(column.EXTRA, []string{"DEFAULT_GENERATED"}) {
		extraSQL = " " + column.EXTRA
	}
	characterSQL := ""
	if column.CHARACTER_SET_NAME != "NULL" {
		characterSQL = " collate " + column.COLLATION_NAME
	}
	if origin {

	} else {
		// if column.COLLATION_NAME != c.defaultCharacter && column.COLLATION_NAME != "NULL" {
		// 	characterSQL = " collate " + c.defaultCharacter
		// }
		if c.isFloatField(column) {
			oldColumnType := column.COLUMN_TYPE
			pos := StrPos(oldColumnType, "(")
			columnType := "decimal"
			if pos > 0 {
				columnType += oldColumnType[pos:]
			} else {
				columnType += "(10,2)"
			}
			column.COLUMN_TYPE = columnType
		}
	}
	if StrPos(column.COLUMN_TYPE, "(") < 0 && c.isNumberField(column) {
		column.COLUMN_TYPE += "(" + column.NUMERIC_PRECISION + ")"
	}
	return fmt.Sprintf("`%s` %s", column.COLUMN_NAME, column.COLUMN_TYPE) + characterSQL + nullSQL + extraSQL + commentSQL
}

func (c *DBCompareService) getIndexSQL(indexList []MySQLIndex, isCreate bool) []string {
	var indexSQLList []string
	var indexType string
	if isCreate {
		indexType = "key"
	} else {
		indexType = "index"
	}
	res := c.getIndexesList(indexList)
	for key, val := range res {
		indexSQLList = append(indexSQLList, Implode(" ", []string{val.Type, indexType})+fmt.Sprintf(" `%s` (%s)", key, val.FieldsString))
	}

	return indexSQLList
}
func (c *DBCompareService) getIndexesList(indexList []MySQLIndex) map[string]DBCompareIndex {
	result := map[string]DBCompareIndex{}
	for _, val := range indexList {
		if _, exist := result[val.INDEX_NAME]; !exist {
			result[val.INDEX_NAME] = DBCompareIndex{
				Type:         "",
				Fields:       []string{},
				FieldsString: "",
			}
		}
		indexType := ""
		if val.INDEX_NAME == "PRIMARY" {
			indexType = "primary"
		} else if val.NON_UNIQUE == "0" {
			indexType = "unique"
		} else if val.INDEX_TYPE == "FULLTEXT" {
			indexType = "fulltext"
		}

		item := result[val.INDEX_NAME]
		item.Type = indexType
		item.Fields = append(item.Fields, val.COLUMN_NAME)
		result[val.INDEX_NAME] = item
	}
	for key, val := range result {
		var fields []string
		for _, v := range val.Fields {
			fields = append(fields, "`"+v+"`")
		}
		val.FieldsString = Implode(",", fields)
		result[key] = val
	}
	return result
}

func (c *DBCompareService) isTextColumn(column MySQLColumn) bool {
	return InStringSlice(column.DATA_TYPE, []string{
		"char", "varchar", "text", "tinytext", "mediumtext", "longtext",
	})
}

func (c *DBCompareService) isUtf8mb4Column(column MySQLColumn) bool {
	return InStringSlice(column.COLLATION_NAME, []string{
		"", "NULL", c.defaultCharacter,
	})
}

func (c *DBCompareService) isFloatField(column MySQLColumn) bool {
	return InStringSlice(column.DATA_TYPE, []string{
		"float", "double",
	})
}

func (c *DBCompareService) isNumberField(column MySQLColumn) bool {
	return InStringSlice(column.DATA_TYPE, []string{
		"float", "double", "decimal", "int", "tinyint", "smallint", "mediumint", "bigint",
	})
}

func (c *DBCompareService) addSlashes(s string) string {
	s = strings.Replace(s, "\r", "\\r", -1)
	s = strings.Replace(s, "\n", "\\n", -1)
	s = strings.Replace(s, "\"", "\\\"", -1)
	return s
}

func (c *WebDeployService) cmdDBInfo() {
	cmd := "db-info"
	c.AddHelp(cmd, `获取数据库配置信息`)

	if argsAct != cmd {
		return
	}

	db := getRemoteDb()
	fmt.Println(db.ConfigContent)

	res := db.Query("show VARIABLES where variable_name='sql_mode';")
	if len(res) > 0 {
		if strings.Index(StrToLower(res[0]["Value"].(string)), "strict_trans_tables") >= 0 {
			ColorErrorText("mysql sql_mode需要禁用STRICT_TRANS_TABLES选项，否则插入数据时会引发异常")
		}
	}

	defer closeRemoteDb()
	RuntimeExit()
}
