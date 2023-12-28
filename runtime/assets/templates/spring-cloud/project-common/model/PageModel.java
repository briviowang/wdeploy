package com.cpp.supplychain.bss.commons.model;

import lombok.Data;
import lombok.NoArgsConstructor;
import lombok.experimental.Accessors;

import java.io.Serializable;
import java.util.Collections;
import java.util.List;

/**
 * 页数据
 *
 * @author lixiaolei
 * @since 1.0
 */
@Data
@Accessors(chain = true)
@NoArgsConstructor
public class PageModel<T> implements Serializable {

    /**
     * 页码
     */
    private Long pageNum;

    /**
     * 每页大小
     */
    private Long pageSize;

    /**
     * 总行数
     */
    private Long total;

    /**
     * 数据
     */
    private List<T> rows;

    /**
     * 构造函数工厂
     *
     * @param pageNum  页码
     * @param pageSize 每页大小
     * @param total    总行数
     * @param rows     数据
     */
    public PageModel(Long pageNum, Long pageSize, Long total, List<T> rows) {
        this.pageNum = pageNum;
        this.pageSize = pageSize;
        this.total = total;
        this.rows = rows;
    }

    /**
     * 构建一个空的分页
     *
     * @param pageNum  页码
     * @param pageSize 每页大小
     * @param <T>      数据类型
     * @return 一页数据
     */
    public static <T> PageModel<T> empty(Long pageNum, Long pageSize) {
        return new PageModel<T>()
                .setRows(Collections.emptyList())
                .setTotal(0L)
                .setPageSize(pageSize)
                .setPageNum(pageNum);
    }
}
