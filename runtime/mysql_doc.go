package runtime

import (
	"time"
)

// MySQLDoc 导出数据库字典
type MySQLDoc struct {
	outputFilePath string
}

// SetOutputFilePath 设置输出文件路径
func (c *MySQLDoc) SetOutputFilePath(s string) {
	c.outputFilePath = s
}

// ExportHTML 导出html文档格式
func (c *MySQLDoc) ExportHTML() {
	tableList := getLocalList()

	templatePath := "assets/mysql/doc.html"
	contentBytes, err := Asset(templatePath)
	CheckErr(err)
	content := ParseTemplate(string(contentBytes), map[string]interface{}{
		"tableList": tableList,
		"date":      time.Unix(Time(), 0).Format("2006-01-02"),
	})

	if IsEmpty(c.outputFilePath) {
		c.outputFilePath = "./mysql-doc.html"
	}

	FilePutContents(c.outputFilePath, content)
}

func (c *WebDeployService) cmdDBDoc() {
	cmd := "db-doc"
	c.AddHelp(cmd, `生成文档`)

	if argsAct != cmd {
		return
	}
	doc := &MySQLDoc{}
	outputFilePath := c.webRoot + "/docs/db-doc.html"
	doc.SetOutputFilePath(outputFilePath)
	doc.ExportHTML()

	// 上传文件
	c.zipDist([]string{outputFilePath})
	c.upload()

	RuntimeExit()
}
