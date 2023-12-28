package com.cpp.supplychain.bss.commons.model;

import lombok.Data;

import java.io.Serializable;

/**
 * 分页请求
 *
 * @author lixiaolei
 * @since 1.0
 */
@Data
public class PageParams implements Serializable {

    /**
     * 页码
     */
    private Long pageNum = 1L;

    /**
     * 每页大小
     */
    private Long pageSize = 10L;
}
