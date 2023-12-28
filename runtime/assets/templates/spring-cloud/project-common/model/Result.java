package com.cpp.supplychain.bss.commons.model;

import io.swagger.annotations.ApiModel;
import io.swagger.annotations.ApiModelProperty;
import lombok.Data;
import lombok.experimental.Accessors;
import org.slf4j.MDC;
import org.springframework.util.Assert;

import java.io.Serializable;
import java.util.function.Supplier;

/**
 * 结果封装
 *
 * @author lixiaolei
 * @since 1.0
 */
@Data
@ApiModel("结果")
@Accessors(chain = true)
public class Result<T> implements Serializable {

    public static final String SUCCESS = "success";
    public static final String SUCCESS_MESSAGE = "ok";

    /**
     * 应答编码
     */
    @ApiModelProperty("应答编码")
    private String code = SUCCESS;

    /**
     * 应答消息
     */
    @ApiModelProperty("应答消息")
    private String message = SUCCESS_MESSAGE;

    /**
     * 时间戳
     */
    @ApiModelProperty("时间戳")
    private Long ts = System.currentTimeMillis();


    /**
     * 应答数据
     */
    @ApiModelProperty("数据")
    private T data;

    /**
     * 成功应答
     *
     * @return 成功应答
     */
    public static <T> Result<T> success() {
        Result<T> response = new Result<>();
        response.setCode(SUCCESS);
        return response;
    }

    /**
     * 成功应答
     *
     * @param data 应答数据
     * @return 成功应答
     */
    public static <T> Result<T> success(T data) {
        Result<T> response = new Result<>();
        response.setCode(SUCCESS);
        response.setData(data);
        return response;
    }

    /**
     * 成功应答
     *
     * @param data    应答数据
     * @param message 应答消息
     * @return 成功应答
     */
    public static <T> Result<T> success(T data, String message) {
        Result<T> response = new Result<>();
        response.setCode(SUCCESS);
        response.setData(data);
        response.setMessage(message);
        return response;
    }

    /**
     * 失败应答
     *
     * @param code    应答编码
     * @param message 应答消息
     * @return 失败应答
     */
    public static <T> Result<T> failure(String code, String message) {
        Result<T> response = new Result<>();
        response.setCode(code);
        response.setMessage(message);
        return response;
    }

    /**
     * 是否成功
     *
     * @return 是否成功
     */
    @ApiModelProperty("是否成功")
    public boolean isSuccess() {
        return SUCCESS.equals(code);
    }

    /**
     * 断言
     */
    public Result<T> assertTrue() {
        Assert.isTrue(this.isSuccess(), this.message);
        return this;
    }

    /**
     * 如果不是成功,则由调用者注入异常提供类,抛出异常
     *
     * @param supplier 异常处理列
     * @return 数据
     */
    public <E extends Throwable> T orElseThrow(Supplier<? extends E> supplier) throws Throwable {
        if (!this.isSuccess()) {
            throw supplier.get();
        } else {
            return this.getData();
        }
    }
}
