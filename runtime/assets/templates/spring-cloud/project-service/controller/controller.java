package {{.package_prefix}}.service.controller;\n

{{$entity_name := camel .entity.TABLE_NAME}}
{{$primary_key := camelVar .entity.PRIMARY_KEY}}

import com.cpp.supplychain.bss.commons.model.PageModel;
import com.cpp.supplychain.bss.commons.model.Result;
import com.cpp.supplychain.payment.acquirerorder.service.PaymentFlowService;
import com.cpp.supplychain.payment.client.request.PaymentFlowPageQueryRequest;
import com.cpp.supplychain.payment.client.response.PaymentFlowPageResponse;
import io.swagger.annotations.Api;
import lombok.extern.slf4j.Slf4j;
import org.springframework.cloud.openfeign.SpringQueryMap;
import org.springframework.web.bind.annotation.*;

import javax.annotation.Resource;\n

@Slf4j
@Api(tags = "{{$entity_name}}接口")
@RequestMapping(value = "/v1")
@RestController
public class {{$entity_name}}Controller {
    @Resource
    private {{$entity_name}}Service service;\n

    @Trace(label = "详情", level = LogLevel.INFO)
    @ApiOperation("详情")
    @GetMapping(value = "/{{snake $entity_name}}/【{{$primary_key}}】")
    ProductDetailDto get{{$entity_name}}Detail(Long {{$primary_key}}){
        return service.get{{$entity_name}}Detail({{$primary_key}});
    }\n

    @Trace(label = "列表", level = LogLevel.INFO)
    @ApiOperation("列表")
    @GetMapping(value = "/{{snake $entity_name}}:page")
    PageModel<{{$entity_name}}ListResponse> get{{$entity_name}}ListByPage({{$entity_name}}ListRequest request){
        return service.get{{$entity_name}}ListByPage(request);
    }\n

    @Trace(label = "添加", level = LogLevel.INFO)
    @ApiOperation("新增")
    @PostMapping(value = "/{{snake $entity_name}}")
    {{$entity_name}}Response add{{$entity_name}}({{$entity_name}}AddRequest request){
        return service.add{{$entity_name}}(request);
    }\n

    @Trace(label = "更新", level = LogLevel.INFO)
    @ApiOperation("更新")
    @PutMapping(value = "/{{snake $entity_name}}")
    void update{{$entity_name}}({{$entity_name}}UpdateRequest request){
        return service.update{{$entity_name}}(request);
    }\n

    @Trace(label = "删除", level = LogLevel.INFO)
    @ApiOperation("删除")
    @DeleteMapping(value = "/{{snake $entity_name}}/{id}")
    boolean delete{{$entity_name}}(Long {{$primary_key}}){
        return service.delete{{$entity_name}}({{$primary_key}});
    }\n

    @Trace(label = "批量获取商品基本信息", level = LogLevel.INFO)
    @ApiOperation("批量获取")
    @GetMapping(value = "/{{snake $entity_name}}/list")
    List<{{$entity_name}}DetailDto> get{{$entity_name}}ListByIds({{$entity_name}}ListByIdsRequest request){
        return service.get{{$entity_name}}ListByIds(request);
    }\n
}
