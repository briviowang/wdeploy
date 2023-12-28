package runtime

import (
	"bytes"
	"io/ioutil"
	"strings"
	"text/template"

	"github.com/buger/jsonparser"
)

var (
	currentGeneratorPlatform string
)

// ParseTemplate 解析模板
func ParseTemplate(tplContent string, pageData interface{}) string {
	tmpl := template.New("template")

	tmpl.Funcs(template.FuncMap{
		"isEmpty":         IsEmpty,
		"isNotEmpty":      IsNotEmpty,
		"trim":            Trim,
		"getField":        sdkGetField,
		"getFieldType":    sdkGetFieldType,
		"getFieldKey":     sdkGetFieldJSON,
		"isList":          sdkIsListType,
		"isString":        sdkIsStringType,
		"ucfirst":         Ucfirst,
		"camel":           Snake2Camel,
		"camelVar":        Snake2CamelVar,
		"snake":           Camel2Snake,
		"springFieldType": springFieldType,
		"upper":           StrToUpper,
	})

	_, err := tmpl.Parse(tplContent)
	CheckErr(err)

	var buff bytes.Buffer
	err = tmpl.Execute(&buff, pageData)
	CheckErr(err)

	return buff.String()
}

func _getPlatformSDK() string {
	return currentGeneratorPlatform
}
func sdkGetField(s string) string {
	config := map[string][]string{
		"ios": {"id", "auto", "break", "case", "char", "const", "continue", "default", "do", "double", "else", "enum", "extern", "float", "for", "goto", "if", "implementation", "int", "interface", "long", "property", "protocol", "register", "return", "short", "signed", "sizeof", "static", "struct", "switch", "typedef", "union", "unsigned", "void", "volatile", "while", "description"},

		"android": {"abstract", "assert", "boolean", "break", "byte", "case", "catch", "char", "class", "const", "continue", "default", "do", "double", "else", "enum", "extends", "final", "finally", "float", "for", "goto", "if", "implements", "import", "instanceof", "int", "interface", "long", "native", "new", "package", "private", "protected", "public", "return", "short", "static", "strictfp", "super", "switch", "synchronized", "this", "throw", "throws", "transient", "try", "void", "volatile", "while"},
		"delphi":  {"and", "array", "as", "asm", "begin", "case", "class", "const", "constructor", "destructor", "dispinterface", "div", "do", "downto", "else", "end", "except", "exports", "file", "finalization", "finally", "for", "function", "goto", "if", "implementation", "in", "inherited", "initialization", "inline", "interface", "is", "label", "library", "mod", "name", "nil", "not", "object", "of", "or", "out", "packed", "procedure", "program", "property", "raise", "record", "repeat", "resourcestring", "result", "set", "shl", "shr", "string", "then", "threadvar", "to", "try", "type", "unit", "until", "uses", "var", "while", "with", "xor"},
		"dart": {
			"as", "assert", "async", "async*", "await", "break", "case", "catch", "class", "const", "default", "deferred", "do", "dynamic", "else", "enum", "export", "extends", "external", "factory", "final", "finally", "for", "get", "if", "implements", "import", "in", "is", "library", "null", "operator", "part", "rethrow", "return", "set", "static", "super", "switch", "sync", "throw", "true", "try", "typedef", "var", "void", "while", "with", "yield",
		},
	}

	if s[len(s)-1:] == "?" {
		s = s[:len(s)-1]
	}
	platformSdk := _getPlatformSDK()
	if InStringSlice(s, config[platformSdk]) {
		if platformSdk == "delphi" {
			return "f" + Ucfirst(s)
		}
		return Ucfirst(s)
	}
	return s
}
func sdkGetFieldJSON(s string) string {
	if s[len(s)-1:] == "?" {
		return s[:len(s)-1]
	}
	return s
}
func sdkGetFieldType(val string) string {
	if val[len(val)-1:] == "?" {
		return val[:len(val)-1]
	} else if StrToLower(val) == "string" {
		platformSdk := _getPlatformSDK()

		if platformSdk == "ios" {
			return "NSString"
		}
		return "String"
	} else {
		return val
	}
}
func sdkIsListType(s string) bool {
	if len(s) == 0 {
		return false
	}
	return s[len(s)-1:] == "?"
}

func sdkIsStringType(s string) bool {
	return StrToLower(s) == "string"
}

//APIEntity 表示一个entity
type APIEntity struct {
	URL              string
	Name             string
	ShortName        string
	OriginalName     string
	Comment          string
	ImportList       map[string]string
	FieldList        map[string]string
	FieldCommentList map[string]string
	EntityType       string
	FieldMaxLength   int
}

func (c APIEntity) init(prefix, name, entityType string) APIEntity {
	c.Name = Ucfirst(prefix) + Ucfirst(name) + Ucfirst(entityType)
	c.EntityType = StrToLower(entityType)
	if c.EntityType == "table" {
		c.ShortName = StrToLower(name)
	}
	c.ImportList = map[string]string{}
	c.FieldList = map[string]string{}
	c.FieldCommentList = map[string]string{}
	return c
}

type typeAPIList map[string]map[string]string

// APIGenerator api解析
type APIGenerator struct {
	ConfigPHPPath string
	OutPath       string

	ConfigFilePath     string
	config             JSONService
	entityList         []APIEntity
	mysql              MySQLService
	apiList            typeAPIList
	apiComments        map[string]map[string]string
	sdkOutPath         string
	usedEntityNameList []string

	tableNameIndexMap map[string]int
	moduleNameList    []string

	prefix string

	errors []string
}

func (c *APIGenerator) init(platform string) {
	currentGeneratorPlatform = platform

	c.errors = []string{}

	c.prefix = Ucfirst(c.config.GetString("platform", currentGeneratorPlatform, "prefix"))

	dbConfig := NewLocalDBConfig()

	c.mysql = MySQLService{
		DBHost:     dbConfig.HOST,
		DBPort:     dbConfig.PORT,
		DBUser:     dbConfig.USER,
		DBPassword: dbConfig.PWD,
		DBPrefix:   dbConfig.PREFIX,
		DBName:     dbConfig.NAME,
	}

	if currentGeneratorPlatform == PLATFORM_SPRING_CLOUD {
		return
	}

	//避免重复解析注释
	if c.apiList == nil {
		c.apiList = typeAPIList{}
		dir := Pwd() + "/app/Lib/Action/api"
		if !FileExists(dir) {
			dir = Pwd() + "/app/api/controller"
		}
		c.apiComments = c.parsePHPComments(dir)
	}

	c.parseAPIEntityList()
}

func (c *APIGenerator) parsePHPComments(d string) map[string]map[string]string {
	result := map[string]map[string]string{}

	res, err := ioutil.ReadDir(d)
	if err != nil {
		return result
	}

	for _, file := range res {
		fPath := d + "/" + file.Name()
		if GetFileExtension(fPath) == "php" {
			phpClass := ParsePHP(fPath)
			temp := map[string]string{}
			for _, phpFunction := range phpClass.FunctionList {
				temp[phpFunction.FunctionName] = phpFunction.Comment
			}
			result[phpClass.ClassName] = temp
		}
	}

	return result
}

func (c *APIGenerator) getAPIComment(apiName string) string {
	res := strings.Split(apiName, "/")
	result := ""
	if functions, classExist := c.apiComments[res[0]]; classExist {
		if comment, functionExist := functions[res[1]]; functionExist {
			return comment
		}
	}
	c.errors = append(c.errors, "接口定义："+apiName+"不存在")
	return result
}

func (c *APIGenerator) initDir(p string) {
	res := strings.Split(BaseName(c.ConfigFilePath), ".")
	c.sdkOutPath = c.OutPath + "/" + res[0] + "-" + _getPlatformSDK()
	dirs := []string{
		"table", "data", "request", "response",
	}
	for _, v := range dirs {
		Mkdir(c.sdkOutPath + p + "/" + v)
	}
}
func (c *APIGenerator) generateTemplate(fileOutPath string, templatePath string, data interface{}) {
	if data == nil {
		data = map[string]interface{}{}
	}
	templatePath = "assets/templates/" + _getPlatformSDK() + templatePath
	contentBytes, err := Asset(templatePath)
	CheckErr(err)
	content := ParseTemplate(string(contentBytes), data)

	content = strings.Replace(content, "\n\n\n", "\n", -1)
	content = strings.Replace(content, "\n\n", "\n", -1)
	content = strings.Replace(content, "\\n", "\n", -1)
	content = strings.Replace(content, "【", "{", -1)
	content = strings.Replace(content, "】", "}", -1)
	fileOutPath = c.sdkOutPath + fileOutPath
	FilePutContents(fileOutPath, content)
}

func (c *APIGenerator) parseAPIEntityList() {
	tableList := c.mysql.GetTableList()
	c.entityList = []APIEntity{}

	//数据库中的table
	c.tableNameIndexMap = map[string]int{}
	index := 0
	for tableName, tableItem := range tableList {
		tableShortName := tableName[len(c.mysql.DBPrefix):]
		c.tableNameIndexMap[tableShortName] = index

		entity := APIEntity{}.init(c.prefix, tableShortName, "table")
		entity.OriginalName = tableName

		for _, column := range tableItem.COLUMN_LIST {
			if currentGeneratorPlatform == PLATFORM_SPRING_CLOUD {
				entity.FieldList[column.COLUMN_NAME] = column.DATA_TYPE
			} else {
				entity.FieldList[column.COLUMN_NAME] = "STRING"
			}

			if len(column.COLUMN_COMMENT) > 0 {
				entity.FieldCommentList[column.COLUMN_NAME] = column.COLUMN_COMMENT
			}
		}
		c.entityList = append(c.entityList, entity)
		index++
	}
	if currentGeneratorPlatform == PLATFORM_SPRING_CLOUD {
		return
	}

	c.config.Each(func(key string, value string, dataType jsonparser.ValueType) {
		c.moduleNameList = append(c.moduleNameList, key)
	}, "modules")

	//配置文件中的table
	c.config.Each(func(key string, value string, dataType jsonparser.ValueType) {

		tableName := key

		entity := c.entityList[c.tableNameIndexMap[tableName]]
		c.config.Each(func(key string, value string, dataType jsonparser.ValueType) {
			entity = c.setEntity(entity, key, value)
		}, "tables", tableName)

		c.entityList[c.tableNameIndexMap[tableName]] = entity

	}, "tables")

	//配置文件中的modules
	c.config.Each(func(key string, value string, dataType jsonparser.ValueType) {
		entityName := key
		entity := APIEntity{}.init(c.prefix, entityName, "data")

		c.config.Each(func(key string, value string, dataType jsonparser.ValueType) {
			entity = c.setEntity(entity, key, value)
		}, "modules", entityName)

		c.entityList = append(c.entityList, entity)
	}, "modules")

	//配置文件中的api
	c.config.Each(func(key string, value string, dataType jsonparser.ValueType) {
		apiName := key
		c.apiList[c.getAPIName(apiName)] = map[string]string{
			"url":     apiName,
			"comment": c.getAPIComment(apiName),
		}

		//request
		requestEntity := APIEntity{}.init(c.prefix, c.getAPIName(apiName), "request")

		c.config.Each(func(key string, value string, dataType jsonparser.ValueType) {
			requestEntity = c.setEntity(requestEntity, key, value)
		}, "api", apiName, "request")
		c.entityList = append(c.entityList, requestEntity)

		//response
		responseEntity := APIEntity{}.init(c.prefix, c.getAPIName(apiName), "response")

		keys := []string{"api", apiName, "response"}

		responseEntity.FieldList = map[string]string{
			"status": "STRING",
			"result": "STRING",
		}

		responseConfigType := c.config.GetType(keys...)
		if responseConfigType == jsonparser.String {
			responseEntity = c.setEntity(responseEntity, "data", c.config.GetString(keys...))
		} else if responseConfigType == jsonparser.Object || responseConfigType == jsonparser.Array {

			responseDataEntity := APIEntity{}.init(c.prefix, c.getAPIName(apiName), "data")

			c.config.Each(func(key string, value string, dataType jsonparser.ValueType) {
				responseDataEntity = c.setEntity(responseDataEntity, key, value)
			}, keys...)

			if len(responseDataEntity.FieldList) > 0 {
				c.entityList = append(c.entityList, responseDataEntity)
				responseEntity.FieldList["data"] = responseDataEntity.Name
				responseEntity.ImportList[responseDataEntity.Name] = "data"
			} else {
				responseEntity.FieldList["data"] = "STRING"
			}

			c.addUsedEntityNameList(responseDataEntity.Name)
		}

		c.entityList = append(c.entityList, responseEntity)
	}, "api")

	var tempList []APIEntity
	for _, entity := range c.entityList {
		if InStringSlice(entity.EntityType, []string{"table", "data"}) {
			if InStringSlice(entity.Name, c.usedEntityNameList) {
				tempList = append(tempList, entity)
			}
		} else {
			tempList = append(tempList, entity)
		}
	}
	//fmt.Println(JSONEncode(c.usedEntityNameList))
	c.entityList = tempList
	if len(c.errors) > 0 {
		ColorErrorText(Implode("\n\t", c.errors) + "\n")
		RuntimeExit()
	}
}

func (c *APIGenerator) setEntity(entity APIEntity, key string, value string) APIEntity {
	fieldGroup, fieldName, fieldType := c.getInstanceInfo(key, value)
	entity.FieldList[fieldName] = fieldType
	if len(fieldGroup) > 0 {
		entity.ImportList[fieldType] = fieldGroup
	}
	return entity
}

func (c *APIGenerator) addUsedEntityNameList(name string) {
	if !InStringSlice(name, c.usedEntityNameList) {
		c.usedEntityNameList = append(c.usedEntityNameList, name)
	}
}

func (c *APIGenerator) getInstanceInfo(key string, value string) (string, string, string) {
	firstCharacter := value[0:1]

	var fieldGroup, fieldName, fieldType string

	fieldGroup = ""
	fieldType = "STRING"

	if !IsNumber(key) {
		fieldName = key
	}

	if firstCharacter == "@" {
		fieldGroup = "data"
		fieldType = c.getDataClass(value)
		if len(fieldName) == 0 {
			fieldName = value[1:]
		}

		c.addUsedEntityNameList(fieldType)
	} else if firstCharacter == "/" {
		fieldGroup = "table"
		fieldType = c.getTableClass(value)
		if len(fieldName) == 0 {
			fieldName = value[1:]
		}

		c.addUsedEntityNameList(fieldType)
	} else if StrToLower(value) == "string" {
		if len(fieldName) == 0 {
			fieldName = key
		}
	} else {
		if len(fieldName) == 0 {
			fieldName = value
		}
	}
	return fieldGroup, fieldName, fieldType
}

func (c *APIGenerator) getAPIName(s string) string {
	res := strings.Split(s, "/")
	result := ""
	for _, v := range res {
		result += Ucfirst(v)
	}
	return result
}

func (c *APIGenerator) getTableClass(s string) string {
	s = s[1:]
	if _, exist := c.tableNameIndexMap[s]; !exist {
		c.errors = append(c.errors, "表"+s+"不存在")
	}

	return c.prefix + Ucfirst(s) + "Table"
}

func (c *APIGenerator) getDataClass(s string) string {
	s = s[1:]
	if !InStringSlice(s, c.moduleNameList) {
		c.errors = append(c.errors, "module "+s+"不存在")
	}

	return c.prefix + Ucfirst(s) + "Data"
}

//indentApiComment 对注释进行缩进
//blank 空格字符串
func (c *APIGenerator) indentApiComment(blank string) typeAPIList {
	apiList := c.apiList
	for _, entity := range apiList {
		comment := entity["comment"]
		res := Explode("\n", comment)
		for k, l := range res {
			res[k] = "    " + l
		}
		entity["comment"] = Implode("\n", res)
	}
	return apiList
}

// Generate 生成sdk
func (c *APIGenerator) Generate() {
	if IsEmpty(c.ConfigPHPPath) {
		c.ConfigPHPPath = Pwd() + "/api.config.php"
	}

	c.ConfigFilePath = Pwd() + "/.git/deploy/api.config.json"

	if FileExists(c.ConfigPHPPath) {
		ExecutePHP("convert_api_config", map[string]interface{}{
			"json_path":       c.ConfigFilePath,
			"api_config_path": c.ConfigPHPPath,
		})
	}

	if !FileExists(c.ConfigFilePath) {
		ErrorLog(c.ConfigFilePath + "配置文件不存在!")
	}
	c.config = JSONService{File: c.ConfigFilePath}
	c.config.init()

	platformList := []string{}

	jsonparser.ObjectEach(c.config.Data, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
		platformList = append(platformList, string(key))
		return nil
	}, "platform")

	StepLog("生成接口SDK")

	for _, platform := range platformList {
		println("*生成" + platform)
		if platform == "android" {
			c.generateAndroidSDK()
		} else if platform == "ios" {
			c.generateIOSSDK()
		} else if platform == "react-native" {
			c.generateReactSDK()
		} else if platform == "flutter" {
			c.generateFlutterSDK()
		} else if platform == "spring-cloud" {
			c.generateSpringCloud()
		}
	}
}

func (c *APIGenerator) generateReactSDK() {
	c.init("react-native")
	c.initDir("/")

	//生成ApiClient
	data := map[string]interface{}{
		"prefix":     c.prefix,
		"apiList":    c.apiList,
		"entityList": c.entityList,
		"copyright":  c.config.GetString("copyright"),
	}
	c.generateTemplate("/"+c.prefix+"ApiClient.js", "/ApiClient.js", data)
}

func (c *APIGenerator) generateFlutterSDK() {
	c.init("flutter")
	c.initDir("/lib")

	packagePrefix := c.config.GetString("platform", "flutter", "package")
	//生成ApiClient.java
	data := map[string]interface{}{
		"prefix":         c.prefix,
		"apiList":        c.apiList,
		"copyright":      c.config.GetString("copyright"),
		"package_prefix": packagePrefix,
	}
	c.generateTemplate("/"+c.prefix+"pubspec.yaml", "/pubspec.tpl", data)
	c.generateTemplate("/lib/"+c.prefix+"ApiClient.dart", "/ApiClient.tpl", data)
	c.generateTemplate("/lib/"+c.prefix+"BaseEntity.dart", "/BaseEntity.tpl", data)

	//生成entity
	for _, entity := range c.entityList {
		data["entity"] = entity
		c.generateTemplate("/lib/"+entity.EntityType+"/"+entity.Name+".dart", "/Entity.tpl", data)
	}
}

func springFieldType(s string) string {
	if s == "datetime" {
		return "LocalDateTime"
	}

	if s == "bigint" {
		return "Long"
	}

	if s == "int" {
		return "Integer"
	}

	if s == "decimal" {
		return "BigDecimal"
	}
	if s == "smallint" || s == "tinyint" {
		return "Boolean"
	}
	return "String"
}

func (c *WebDeployService) cmdAPI() {
	cmd := "api"
	c.AddHelp(cmd, `生成api；参数module：区分不同分组`)

	if argsAct != cmd {
		return
	}

	distPath := c.webRoot + "/tools/sdk"
	Rm(distPath)

	moduleName := "api"
	if !IsEmpty(argsParams["module"]) {
		moduleName = argsParams["module"]
	}

	generator := APIGenerator{
		OutPath:       distPath,
		ConfigPHPPath: Pwd() + "/" + moduleName + ".config.php",
	}

	generator.Generate()

	var zipFiles []string

	dirs, _ := ioutil.ReadDir(distPath)
	for _, dir := range dirs {
		f := distPath + "/" + dir.Name() + ".zip"
		zipFiles = append(zipFiles, f)
		ZipDirectory(f, distPath+"/"+dir.Name())
	}
	if hasArg("test") {
		RuntimeExit()
	}

	c.zipDist(zipFiles)
	c.upload()

	RuntimeExit()
}
