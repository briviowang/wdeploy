package {{.package_prefix}}.dao;

{{$entity_name := camel .entity.TABLE_NAME}}

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.baomidou.mybatisplus.core.metadata.IPage;
import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
import {{.package_prefix}}.dao.entity.{{$entity_name}}Entity;
import org.apache.ibatis.annotations.Param;

import java.time.LocalDateTime;
import java.util.List;


/**
 * @author brivio
 */
public interface {{$entity_name}}Dao extends BaseMapper<{{$entity_name}}Entity> {
    /**
     * 列表搜索
     *
     * @param keywords        类目id
     * @param sortedBy        排序字段
     * @param sortedType      排序类型
     * @param page            分页参数
     * @return 商品列表
     */
    IPage<{{$entity_name}}Entity> getListByPage(
        @Param(value = "keywords") String keywords,
        @Param(value = "sortedBy") String sortedBy,
        @Param(value = "sortedType") String sortedType,
        Page<{{$entity_name}}Entity> page);
}