package com.cpp.supplychain.bss.commons.enums;

import lombok.AllArgsConstructor;

/**
 * 数据中心
 */
@AllArgsConstructor
public enum DataCenter {
    /**
     * 中国-上海
     */
    CN_SHANGHAI(1, "cn_shanghai", "中国-上海"),
    /**
     * 中国-北京
     */
    CN_BEIJING(2, "cn_beijing", "中国-北京"),
    ;

    /**
     * long型的标志
     */
    public final long id;
    /**
     * 区域id
     */
    public final String regionId;
    /**
     * 名称
     */
    public final String name;
}
