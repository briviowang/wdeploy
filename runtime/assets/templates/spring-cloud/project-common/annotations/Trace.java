package com.cpp.supplychain.bss.commons.annotations;


import org.springframework.boot.logging.LogLevel;

import java.lang.annotation.*;

/**
 * 追踪方法入参和出参的标记
 * @author lixiaolei
 */
@Target(ElementType.METHOD)
@Retention(RetentionPolicy.RUNTIME)
@Documented
public @interface Trace {

    /**
     * 日志等级
     *
     * @return 日志等级
     */
    LogLevel level() default LogLevel.INFO;

    /**
     * 标记
     *
     * @return 标记
     */
    String label();

    /**
     * 是否打印入参
     *
     * @return 是否打印入参
     */
    boolean input() default true;

    /**
     * 是否打印返回值
     *
     * @return 是否打印返回值
     */
    boolean output() default true;
}
