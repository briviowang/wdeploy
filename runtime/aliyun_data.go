package runtime

import (
	"bytes"
	"net/http"
	"text/template"
)

func (c *AliyunHelper) GetAliyunData(w http.ResponseWriter, req *http.Request) {
	tplContent := ""

	templatePath := "assets/aliyun/data.json"
	contentBytes, _ := Asset(templatePath)
	tplContent = string(contentBytes)

	tmpl := template.New("template")
	tmpl.Delims("${{", "}}")
	tmpl.Funcs(template.FuncMap{})

	_, err := tmpl.Parse(tplContent)
	CheckErr(err)

	var buff bytes.Buffer
	pageData := map[string]interface{}{
		"port": msPort,
	}
	err = tmpl.Execute(&buff, pageData)
	CheckErr(err)
	c.AjaxResult(w, buff.String())
}
