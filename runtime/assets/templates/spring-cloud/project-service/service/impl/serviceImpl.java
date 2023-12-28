package {{.package_prefix}}.service.impl;\n

{{$entity_name := camel .entity.TABLE_NAME}}
{{$primary_key := camelVar .entity.PRIMARY_KEY}}

import com.baomidou.mybatisplus.core.metadata.IPage;
import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import lombok.RequiredArgsConstructor;
import org.springframework.beans.BeanUtils;
import org.springframework.stereotype.Service;
import org.springframework.util.Assert;

import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import java.util.ArrayList;
import java.util.List;
import java.util.Optional;\n

/**
 * @author brivio
 */
@RequiredArgsConstructor
@Service("{{$entity_name}}Service")
public class {{$entity_name}}ServiceImpl implements {{$entity_name}}Service {
    private final {{$entity_name}}Dao dao;\n

    /**
     * 详情
     *
     * @param {{$primary_key}} 参数
     * @return 结果
     */
    @Override
    public {{$entity_name}}Response get{{$entity_name}}Detail(Long {{$primary_key}}){
        checkExist({{$primary_key}});
        {{$entity_name}}Entity entity = dao.selectById({{$primary_key}});
        {{$entity_name}}Response response = new {{$entity_name}}Response();
        BeanUtil.copyProperties(entity, response);
        return response;
    }\n

    /**
     * 商品列表
     * @param request
     * @return
     */
    @Override
    public PageModel<{{$entity_name}}ListResponse> get{{$entity_name}}ListByPage({{$entity_name}}ListRequest request){
        IPage<{{$entity_name}}Entity> page = dao.getListByPage(
                new Page<>(request.getPageNum(), request.getPageSize(), true)
        );
        List<{{$entity_name}}ListResponse> list = new ArrayList<>();
        page.getRecords().forEach(item -> {
            {{$entity_name}}ListResponse response = new {{$entity_name}}ListResponse();
            BeanUtils.copyProperties(item, response);
            list.add(response);
        });
        return new PageModel<>(page.getCurrent(),page.getSize(),page.getTotal(),list);
    }\n

    /**
     * 更新
     *
     * @param request 参数
     * @return 商品id
     */
    @Override
    public {{$entity_name}}Response add{{$entity_name}}({{$entity_name}}AddRequest request){
        {{$entity_name}}Entity entity = new {{$entity_name}}Entity();
        BeanUtil.copyProperties(request, entity);
        entity.set{{camel .entity.PRIMARY_KEY}}(idGeneratorClient.nextId());

        dao.insert(entity);
        {{$entity_name}}Response response = new CategoryResponse();
        BeanUtil.copyProperties(entity, response);
        return response;
    }\n

    /**
     * 更新
     *
     * @param request 参数
     * @return 商品id
     */
    @Override
    public void update{{$entity_name}}({{$entity_name}}UpdateRequest request){
        {{$entity_name}}Entity entity = new {{$entity_name}}Entity();
        BeanUtil.copyProperties(request, entity);

        categoryDao.updateById(entity);
    }\n

    /**
     * 删除
     *
     * @param request 参数
     * @return 是否删除成功
     */
    @Override
    public boolean delete{{$entity_name}}(Long {{$primary_key}}){
        checkExist({{$primary_key}});

        {{if .entity.DELETE_FLAG_COLUMN}}
        {{$entity_name}}Entity entity = {{$entity_name}}Entity.builder()
                .set{{camelVar .entity.PRIMARY_KEY}}({{$primary_key}})
                .{{camelVar .entity.DELETE_FLAG_COLUMN}}(true)
                .build();
        categoryDao.updateById(entity);
        {{else}}
        dao.deleteById(request.{{$primary_key}});
        {{end}}
    }\n

    /**
     * 批量获取商品基本信息
     * @param request 商品主键列表参数
     * @return 商品基本信息
     */
    @Override
    public List<{{$entity_name}}DetailDto> get{{$entity_name}}ListByIds({{$entity_name}}ListByIdsRequest request){

    }\n

    private void checkExist(Long id) {
        LambdaQueryWrapper<{{$entity_name}}Entity> wrapper = new LambdaQueryWrapper<>();
        wrapper.eq({{$entity_name}}Entity::getCategoryId, id);
        {{if .entity.DELETE_FLAG_COLUMN}}
        wrapper.eq({{$entity_name}}Entity::getDelFlag, false);
        {{end}}
        if (dao.selectOne(wrapper) == null) {
            throw new CustomException(ProductErrorCode.ERROR_CATEGORY_NOT_EXIST);
        }
    }\n

}