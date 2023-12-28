package runtime

import (
	"database/sql"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"

	// 导入mysql
	_ "github.com/go-sql-driver/mysql"
)

// MySQLTable 表
type MySQLTable struct {
	TABLE_CATALOG   string
	TABLE_SCHEMA    string
	TABLE_NAME      string
	TABLE_TYPE      string
	ENGINE          string
	VERSION         string
	ROW_FORMAT      string
	TABLE_ROWS      string
	AVG_ROW_LENGTH  string
	DATA_LENGTH     string
	MAX_DATA_LENGTH string
	INDEX_LENGTH    string
	DATA_FREE       string
	AUTO_INCREMENT  string
	CREATE_TIME     string
	UPDATE_TIME     string
	CHECK_TIME      string
	TABLE_COLLATION string
	CHECKSUM        string
	CREATE_OPTIONS  string
	TABLE_COMMENT   string
	COLUMN_LIST     []MySQLColumn
	INDEX_LIST      []MySQLIndex

	PRIMARY_KEY        string //主键
	DELETE_FLAG_COLUMN string //删除标识列
	SORT_COLUMN        string //排序字段
}

type MySQLIndex struct {
	TABLE_CATALOG string
	TABLE_SCHEMA  string
	TABLE_NAME    string
	NON_UNIQUE    string
	INDEX_SCHEMA  string
	INDEX_NAME    string
	SEQ_IN_INDEX  string
	COLUMN_NAME   string
	COLLATION     string
	CARDINALITY   string
	SUB_PART      string
	PACKED        string
	NULLABLE      string
	INDEX_TYPE    string
	COMMENT       string
	INDEX_COMMENT string
}

type MySQLColumn struct {
	CHARACTER_MAXIMUM_LENGTH string
	CHARACTER_OCTET_LENGTH   string
	CHARACTER_SET_NAME       string
	COLLATION_NAME           string
	COLUMN_COMMENT           string
	COLUMN_DEFAULT           string
	COLUMN_KEY               string
	COLUMN_NAME              string
	COLUMN_TYPE              string
	DATA_TYPE                string
	DATETIME_PRECISION       string
	EXTRA                    string
	IS_NULLABLE              string
	NUMERIC_PRECISION        string
	NUMERIC_SCALE            string
	ORDINAL_POSITION         string
	PRIVILEGES               string
	TABLE_CATALOG            string
	TABLE_NAME               string
	TABLE_SCHEMA             string
}
type MySQLService struct {
	DB              *sql.DB
	DBUser          string
	DBPassword      string
	DBHost          string
	DBPort          string
	DBName          string
	DBPrefix        string
	Tables          []MySQLColumn
	LogGetTableList bool
	ConfigContent   string
}

func (c *MySQLService) PrintConnection() {
	//root:root@127.0.0.1:3306/haimingwei
	fmt.Println("mysql://" + c.DBUser + ":" + c.DBPassword + "@" + c.DBHost + ":" + c.DBPort + "/" + c.DBName)
}

// Query 执行sql
func (c *MySQLService) Query(s string) []map[string]interface{} {
	if c.DB == nil {
		c.DB, _ = sql.Open("mysql", c.DBUser+":"+c.DBPassword+"@tcp("+c.DBHost+":"+c.DBPort+")/"+c.DBName+"?charset=utf8")
	}

	rows, err := c.DB.Query(s)
	CheckErr(err)

	columns, err := rows.Columns()
	CheckErr(err)

	values := make([]sql.RawBytes, len(columns))

	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	var result []map[string]interface{}

	for rows.Next() {
		rows.Scan(scanArgs...)

		var value interface{}
		resultItem := make(map[string]interface{})

		for i, col := range values {
			if col == nil {
				//value = nil
				value = "NULL"
			} else {
				value = string(col)
			}
			resultItem[columns[i]] = value
		}
		result = append(result, resultItem)
	}
	return result
}

// IsSupportMyISAM 检查数据库是否支持MyISAM
func (c *MySQLService) IsSupportMyISAM() bool {
	res := c.Query("SHOW VARIABLES where variable_name='disabled_storage_engines';")
	if len(res) == 0 {
		//println("no disabled_storage_engines")
		return true
	}
	engineList := Explode(",", fmt.Sprintf("%v", res[0]["Value"]))
	//println("%v", JSONEncode(engineList))
	return !InStringSlice("myisam", engineList)
}

// GetTableList 获取表列表信息
func (c *MySQLService) GetTableList() map[string]MySQLTable {
	// c.PrintConnection()

	isMyISAM := c.IsSupportMyISAM()

	result := map[string]MySQLTable{}
	if c.DBName == "" {
		return nil
	}
	s := fmt.Sprintf("select * from information_schema.tables where table_schema='%s';", c.DBName)
	tableRes := c.Query(s)
	for key, tableData := range tableRes {
		table := &MySQLTable{
			PRIMARY_KEY: "id",
		}

		JSONConvert(tableData, table)

		if c.LogGetTableList {
			idx := key + 1
			total := len(tableRes)
			fmt.Printf("获取表(%d/%d)\r", idx, total)
			if idx >= total {
				fmt.Println("")
			}
		}

		//字段
		table.COLUMN_LIST = []MySQLColumn{}
		s = fmt.Sprintf("select * from information_schema.columns where table_schema='%s' and table_name='%s' order by ORDINAL_POSITION asc;", c.DBName, table.TABLE_NAME)
		columnRes := c.Query(s)
		for _, columnData := range columnRes {
			column := &MySQLColumn{}
			JSONConvert(columnData, column)
			column.COLUMN_DEFAULT = TrimChar(column.COLUMN_DEFAULT, "'")
			// column.COLUMN_TYPE = strings.Replace(column.COLUMN_TYPE, " unsigned", "", -1)
			table.COLUMN_LIST = append(table.COLUMN_LIST, *column)
			if column.COLUMN_KEY == "PRI" {
				table.PRIMARY_KEY = column.COLUMN_NAME
			}
			delete_flag_column_list := []string{"del_flag", "delete_flag", "is_delete", "deleted", "is_del"}
			if InStringSlice(column.COLUMN_NAME, delete_flag_column_list) {
				table.DELETE_FLAG_COLUMN = column.COLUMN_NAME
			}
			sort_column_list := []string{"sort", "ordid"}
			if InStringSlice(column.COLUMN_NAME, sort_column_list) {
				table.SORT_COLUMN = column.COLUMN_NAME
			}
		}

		//索引
		table.INDEX_LIST = []MySQLIndex{}
		s = fmt.Sprintf("select * from information_schema.STATISTICS where table_schema='%s' and table_name='%s';", c.DBName, table.TABLE_NAME)
		indexRes := c.Query(s)
		for _, indexData := range indexRes {
			index := &MySQLIndex{}
			JSONConvert(indexData, index)
			table.INDEX_LIST = append(table.INDEX_LIST, *index)
		}

		isRDS := false
		if !IsEmpty(c.ConfigContent) {
			isRDS = strings.Index(NewDBConfig(c.ConfigContent).HOST, "mysql.rds.aliyuncs.com") > 0
		}

		if !isMyISAM || isRDS {
			table.ENGINE = "MyISAM"
		}

		result[table.TABLE_NAME] = *table
	}

	return result
}

// GetTableList 获取表列表
func GetTableList(connection, tablePrefix string) map[string]MySQLTable {
	//root:root@127.0.0.1:3306/haimingwei
	temp := strings.Split(connection, "@")

	temp1 := strings.Split(temp[0], ":")
	userName := temp1[0]
	password := ""
	if len(temp1) > 1 {
		password = temp1[1]
	}

	temp2 := strings.Split(temp[1], "/")
	dbName := temp2[1]

	temp3 := strings.Split(temp2[0], ":")
	host := temp3[0]
	port := "3306"
	if len(temp3) > 0 {
		port = temp3[1]
	}

	ipList, _ := net.LookupIP(host)
	if len(ipList) == 0 {
		ErrorLog("host不存在")
	}
	host = string(ipList[0].To4().String())

	StepLog("连接" + host + "数据库：" + dbName)

	if host != "127.0.0.1" {
		ssh := NewSSHService(WebDeployConfigItem{
			ServerUser: userName,
			ServerHost: host,
			ServerPort: "22",
		})
		ssh.SetForwardRemoteAddr(host + ":" + port)
		cmd, forwardPort := ssh.MysqlForward()
		port = IntToString(forwardPort)
		defer cmd.Process.Kill()
	}
	if IsEmpty(tablePrefix) {
		tablePrefix = "ins_"
	}

	db := MySQLService{
		DBHost:          host,
		DBPort:          port,
		DBUser:          userName,
		DBPassword:      password,
		DBPrefix:        tablePrefix,
		DBName:          dbName,
		LogGetTableList: true,
	}

	result := db.GetTableList()
	return result
}

// DBConfig 数据库配置
type DBConfig struct {
	HOST   string
	PORT   string
	USER   string
	PWD    string
	NAME   string
	PREFIX string
}

// NewLocalDBConfig 实例化
func NewLocalDBConfig() *DBConfig {
	return NewDBConfig("")
}

// NewDBConfig 实例化
func NewDBConfig(data string) *DBConfig {
	c := DBConfig{}
	c.init(data)

	if IsEmpty(c.HOST) || StrToLower(c.HOST) == "localhost" {
		c.HOST = "127.0.0.1"
	}

	if IsEmpty(c.PORT) {
		c.PORT = "3306"
	}

	if IsEmpty(c.USER) {
		c.USER = "root"
	}

	return &c
}

func (c *DBConfig) init(data string) {

	if IsEmpty(data) {
		//tp3
		dbFile := Pwd() + "/data/config/db.php"
		dbConfigPath := ""

		if FileExists(dbFile) && IsEmpty(data) {
			dbConfigPath = dbFile
		}

		//tp5
		dbFile = Pwd() + "/data/config/database.php"
		if FileExists(dbFile) && IsEmpty(data) {
			dbConfigPath = dbFile
		}

		if !IsEmpty(dbConfigPath) {
			println(dbConfigPath)
			data = ExecutePHP("get_db_info", map[string]interface{}{
				"db_config_path": dbConfigPath,
			})
		}
		//tp6
		env_file := Pwd() + "/.env"
		if FileExists(env_file) && IsEmpty(data) {
			data = ExecutePHP("get_env_info", map[string]interface{}{
				"env_file": env_file,
			})
		}
	}

	if IsEmpty(data) {
		ErrorLog("无法获取数据库配置")
	}

	JSONService := JSONService{
		Data: []byte(data),
	}

	if !IsEmpty(JSONService.GetString("DB_HOST")) {
		//thinkphp 3.1版本
		c.HOST = JSONService.GetString("DB_HOST")
		c.PORT = JSONService.GetString("DB_PORT")
		c.USER = JSONService.GetString("DB_USER")
		c.PWD = JSONService.GetString("DB_PWD")
		c.NAME = JSONService.GetString("DB_NAME")
		c.PREFIX = JSONService.GetString("DB_PREFIX")
	} else {
		//thinkphp 5.1版本
		c.HOST = JSONService.GetString("hostname")
		c.PORT = JSONService.GetString("port")
		c.USER = JSONService.GetString("username")
		c.PWD = JSONService.GetString("password")
		c.NAME = JSONService.GetString("database")
		c.PREFIX = JSONService.GetString("prefix")
	}
}

func getLocalList() map[string]MySQLTable {
	StepLog("获取本地数据")

	dbConfig := NewLocalDBConfig()
	SubStepLog("连接本地数据库[" + dbConfig.NAME + "]")

	db := MySQLService{
		DBHost:          dbConfig.HOST,
		DBPort:          dbConfig.PORT,
		DBUser:          dbConfig.USER,
		DBPassword:      dbConfig.PWD,
		DBPrefix:        dbConfig.PREFIX,
		DBName:          dbConfig.NAME,
		LogGetTableList: true,
	}

	return db.GetTableList()
}

func getRemoteDBInfo() string {
	ssh := initSSHService()

	outShellFile := Pwd() + "/.git/deploy/server.sh"

	contentBytes, err := Asset("assets/scripts/get_remote_db_info.sh")
	CheckErr(err)

	content := ParseTemplate(string(contentBytes), map[string]interface{}{
		"ServerPath": currentDeployConfig.ServerPath,
	})
	FilePutContents(outShellFile, content)
	respContent := ssh.ShellQuiet(content)
	Rm(outShellFile)
	return respContent
}

var remoteDbConnection *exec.Cmd

func getRemoteDb() *MySQLService {
	ssh := initSSHService()

	respContent := getRemoteDBInfo()

	dbConfig := NewDBConfig(respContent)

	ssh.SetForwardRemoteAddr(dbConfig.HOST + ":" + dbConfig.PORT)

	var port int
	remoteDbConnection, port = ssh.MysqlForward()

	return &MySQLService{
		DBHost:          "127.0.0.1",
		DBPort:          IntToString(port),
		DBUser:          dbConfig.USER,
		DBPassword:      dbConfig.PWD,
		DBPrefix:        dbConfig.PREFIX,
		DBName:          dbConfig.NAME,
		LogGetTableList: true,
		ConfigContent:   respContent,
	}
}

func closeRemoteDb() {
	if remoteDbConnection != nil {
		remoteDbConnection.Process.Kill()
	}
}

func getRemoteList() map[string]MySQLTable {
	StepLog("获取远程数据库")
	result := getRemoteDb().GetTableList()
	defer closeRemoteDb()
	return result
}

func (c *WebDeployService) cmdDBCheck() {
	cmd := "db-check"
	c.AddHelp(cmd, `检查数据库表`)

	if argsAct != cmd {
		return
	}
	dest := "server"
	if len(os.Args) >= 3 {
		dest = os.Args[2]
	}

	var leftList map[string]MySQLTable

	if dest == "local" {
		leftList = getLocalList()
	} else {
		leftList = getRemoteList()
	}

	StepLog("检查数据库结构")
	compareService := DBCompareService{
		LeftList:  leftList,
		RightList: leftList,
	}

	compareSQLFile := Pwd() + "/compare.sql"
	content := compareService.Compare()
	FilePutContents(compareSQLFile, content)
	if CommandExist("source-highlight") {
		Execute([]string{"source-highlight", "-f", "esc", "-i", compareSQLFile})
	} else {
		fmt.Println(content)
	}
	RuntimeExit()
}

func (c *WebDeployService) cmdDBCompare() {
	cmd := "db-compare"
	c.AddHelp(cmd, `比较数据库,后可跟参数：server(默认同步远程)；local(同步本地)`)

	if argsAct != cmd {
		return
	}
	var leftList map[string]MySQLTable
	var rightList map[string]MySQLTable

	if hasArg("left-connection") {
		leftList = GetTableList(argsParams["left-connection"], argsParams["left-table-prefix"])
		rightList = GetTableList(argsParams["right-connection"], argsParams["right-table-prefix"])
	} else {
		dest := "server"
		if len(os.Args) >= 3 {
			dest = os.Args[2]
		}

		if dest == "local" {
			leftList = getRemoteList()
			rightList = getLocalList()
		} else {
			leftList = getLocalList()
			rightList = getRemoteList()
		}
	}

	StepLog("比较数据库结构")
	compareService := DBCompareService{
		LeftList:  leftList,
		RightList: rightList,
	}
	compareSQLFile := Pwd() + "/compare.sql"
	content := compareService.Compare()
	FilePutContents(compareSQLFile, content)
	if CommandExist("source-highlight") {
		Execute([]string{"source-highlight", "-f", "esc", "-i", compareSQLFile})
	} else {
		fmt.Println(content)
	}
	RuntimeExit()
}
