package runtime

import (
	"bytes"
	"fmt"
	"net/http"
	"text/template"
)

var msRemoteDb *MySQLService
var msPort int

func HttpMysqlQuery(w http.ResponseWriter, req *http.Request) {
	if msRemoteDb == nil {
		msRemoteDb = getRemoteDb()
	}
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("content-type", "application/json")
	sql := req.FormValue("sql")
	res := msRemoteDb.Query(sql)
	fmt.Fprintf(w, "%s", JSONEncode(res))
}

func HttpStaticsIndex(w http.ResponseWriter, req *http.Request) {
	indexPath := Pwd() + "/.git/deploy/db-statics.html"

	tplContent := ""
	if FileExists(indexPath) {
		tplContent = FileGetContents(indexPath)
	} else {
		templatePath := "assets/mysql/db-statics.html"
		contentBytes, _ := Asset(templatePath)
		tplContent = string(contentBytes)
	}
	tmpl := template.New("template")

	tmpl.Funcs(template.FuncMap{})

	_, err := tmpl.Parse(tplContent)
	CheckErr(err)

	var buff bytes.Buffer
	pageData := map[string]interface{}{
		"port": msPort,
	}
	err = tmpl.Execute(&buff, pageData)
	CheckErr(err)
	fmt.Fprintf(w, "%s", buff.String())
}
func (c *WebDeployService) cmdDBStatics() {
	cmd := "db-statics"
	c.AddHelp(cmd, `统计`)

	if argsAct != cmd {
		return
	}
	msPort = GetRandomPort()
	http.HandleFunc("/", HttpStaticsIndex)
	http.HandleFunc("/query", HttpMysqlQuery)
	url := fmt.Sprintf("http://localhost:%v", msPort)
	OpenBrowser(url)
	InfoLog(url)

	http.ListenAndServe(fmt.Sprintf(":%v", msPort), nil)
	RuntimeExit()
}
