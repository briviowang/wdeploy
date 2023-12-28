package {{.package_prefix}}.service;\n

{{$entity_name := camel .entity.TABLE_NAME}}
{{$primary_key := camelVar .entity.PRIMARY_KEY}}

import com.baomidou.mybatisplus.extension.service.IService;
import com.cpp.supplychain.bss.commons.model.PageModel;
import com.cpp.supplychain.payment.acquirerorder.entity.PaymentFlow;
import com.cpp.supplychain.payment.client.request.PaymentFlowPageQueryRequest;
import com.cpp.supplychain.payment.client.response.PaymentFlowPageResponse;\n

{{if .entity.TABLE_COMMENT}}
/**
 * {{.entity.TABLE_COMMENT}}
 */
{{end}}

public interface {{$entity_name}}Service {\n

    /**
     * 详情
     *
     * @param {{$primary_key}} id
     * @return 结果
     */
    @Override
    public {{$entity_name}}Response get{{$entity_name}}Detail(Long {{$primary_key}});\n

    /**
     * 商品列表
     * @param request
     * @return
     */
    @Override
    public PageModel<{{$entity_name}}ListResponse> get{{$entity_name}}ListByPage({{$entity_name}}ListRequest request);\n

    /**
     * 更新
     *
     * @param request 参数
     * @return 商品id
     */
    @Override
    public {{$entity_name}}Response add{{$entity_name}}({{$entity_name}}AddRequest request);\n

    /**
     * 更新
     *
     * @param request 参数
     * @return 商品id
     */
    @Override
    public void update{{$entity_name}}({{$entity_name}}UpdateRequest request);\n

    /**
     * 删除
     *
     * @param {{$primary_key}} id
     * @return 是否删除成功
     */
    @Override
    public boolean delete{{$entity_name}}(Long {{$primary_key}});\n

    /**
     * 批量获取商品基本信息
     * @param request 商品主键列表参数
     * @return 商品基本信息
     */
    @Override
    public List<{{$entity_name}}DetailDto> get{{$entity_name}}ListByIds({{$entity_name}}ListByIdsRequest request);\n
}
