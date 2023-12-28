package runtime

import "strings"

/*
根据数据库生成spring cloud微服务基本代码

模板语法：
	https://www.topgoer.com/%E5%B8%B8%E7%94%A8%E6%A0%87%E5%87%86%E5%BA%93/template.html
	https://pkg.go.dev/text/template
*/
var PLATFORM_SPRING_CLOUD string = "spring-cloud"

func (c *APIGenerator) generateSpringCloud() {
	platformName := "spring-cloud"
	c.init(platformName)
	tableList := c.mysql.GetTableList()

	c.sdkOutPath = c.OutPath + "/" + platformName
	projectName := "trade"
	basePackage := "com.cpp.supplychain." + projectName
	javaDir := "/src/main/java/" + strings.ReplaceAll(basePackage, ".", "/")
	resourcesDir := "/src/main/resources"
	data := map[string]interface{}{
		"project_name":   projectName,
		"package_prefix": basePackage,
	}
	c.generateTemplate("/.gitignore", "/gitignore.tpl", data)
	c.generateTemplate("/pom.xml", "/pom.xml", data)
	//生成deploy模块
	c.generateTemplate("/"+projectName+"-deploy"+javaDir+"/Application.java", "/project-deploy/Application.java", data)
	c.generateTemplate("/"+projectName+"-deploy"+resourcesDir+"/application.properties", "/project-deploy/application.properties", data)

	for _, entity := range tableList {
		data["entity"] = entity
		//生成client模块

		//生成dao模块
		c.generateTemplate("/"+projectName+"-dao"+javaDir+"/dao/entity/"+Snake2Camel(entity.TABLE_NAME)+"Entity.java", "/project-dao/entity/entity.java", data)

		c.generateTemplate("/"+projectName+"-dao"+javaDir+"/dao/"+Snake2Camel(entity.TABLE_NAME)+"Dao.java", "/project-dao/dao.java", data)

		c.generateTemplate("/"+projectName+"-dao"+resourcesDir+"/mapper/"+Snake2Camel(entity.TABLE_NAME)+".xml", "/project-dao/mapper/mapper.xml", data)

		//生成service模块
		c.generateTemplate("/"+projectName+"-service"+javaDir+"/service/"+Snake2Camel(entity.TABLE_NAME)+"Service.java", "/project-service/service/service.java", data)
		c.generateTemplate("/"+projectName+"-service"+javaDir+"/service/impl/"+Snake2Camel(entity.TABLE_NAME)+"ServiceImpl.java", "/project-service/service/impl/serviceImpl.java", data)
		c.generateTemplate("/"+projectName+"-service"+javaDir+"/controller/"+Snake2Camel(entity.TABLE_NAME)+"Controller.java", "/project-service/controller/controller.java", data)
	}

}
