package runtime

func (c *APIGenerator) generateIOSSDK() {
	c.init("ios")
	c.initDir("/")

	data := map[string]interface{}{
		"prefix": c.prefix,
	}
	c.generateTemplate("/"+c.prefix+"ApiBase.m", "/ApiBase.m", data)
	c.generateTemplate("/"+c.prefix+"ApiBase.h", "/ApiBase.h", data)

	//生成ApiClient
	data = map[string]interface{}{
		"prefix":    c.prefix,
		"apiList":   c.apiList,
		"copyright": c.config.GetString("copyright"),
	}
	c.generateTemplate("/"+c.prefix+"ApiClient.m", "/ApiClient.m", data)
	c.generateTemplate("/"+c.prefix+"ApiClient.h", "/ApiClient.h", data)

	//生成entity
	for _, entity := range c.entityList {
		data["entity"] = entity
		c.generateTemplate("/"+entity.EntityType+"/"+entity.Name+".m", "/Entity.m", data)
		c.generateTemplate("/"+entity.EntityType+"/"+entity.Name+".h", "/Entity.h", data)
	}
}
