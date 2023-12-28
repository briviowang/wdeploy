package com.cpp.supplychain.bss.commons.constants;

/**
 * 供应链请求头 常量类
 *
 * @author : zongri (｡￫‿￩｡)
 * @link : zhongri.ye@henhenchina.com
 * @since : 2021/12/4 10:12
 */
public interface ScHeadConstants {

    /**
     * 认证请求头
     */
    String AUTH_HEADER = "SC-Authorization";

    /**
     * 生成认证 token 的 用户id 字段
     */
    String AUTH_USER_ID_FIELD = "userId";

    /**
     * 用户id 请求头
     */
    String USER_ID_HEADER = "SC-User-Id";

    /**
     * trace id 请求头
     */
    String TRACE_ID_HEADER = "SC-Trace-Id";

}
