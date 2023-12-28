package {{.package_prefix}};

import org.mybatis.spring.annotation.MapperScan;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.boot.context.properties.ConfigurationPropertiesScan;
import org.springframework.cloud.client.discovery.EnableDiscoveryClient;
import org.springframework.cloud.openfeign.EnableFeignClients;
import org.springframework.web.servlet.config.annotation.EnableWebMvc;

@EnableDiscoveryClient
@EnableWebMvc
@SpringBootApplication
@EnableFeignClients(basePackages = {

})
@MapperScan(basePackages= {
        "com.cpp.supplychain.trade.order.mapper",
})
public class {{ucfirst .project_name}}Application {

    public static void main(String[] args) {
        SpringApplication.run({{ucfirst .project_name}}Application.class, args);
    }

}
