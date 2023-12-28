package com.cpp.supplychain.bss.commons.exception;

import lombok.Getter;

/**
 * 自定义的异常
 *
 * @author yuanzhi.wang
 */
public class CustomException extends RuntimeException{

    @Getter
    private final String code;

    @Getter
    private final String message;

    public CustomException(ErrorCode errorCode){
        this.code = errorCode.getErrorCode();
        this.message = errorCode.getErrorMessage();
    }

    public interface ErrorCode{

        /**
         * 获得错误码
         *
         * @return 错误码
         *  */
        String getErrorCode();

        /**
         * 获得错误信息
         *
         * @return 错误信息
         * */
        String getErrorMessage();
    }

}
