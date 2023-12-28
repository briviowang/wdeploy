package com.cpp.supplychain.bss.commons.aop;

import com.cpp.supplychain.bss.commons.annotations.Trace;
import lombok.extern.slf4j.Slf4j;
import org.apache.commons.lang3.StringUtils;
import org.aspectj.lang.ProceedingJoinPoint;
import org.aspectj.lang.annotation.Around;
import org.aspectj.lang.annotation.Aspect;
import org.aspectj.lang.reflect.MethodSignature;
import org.springframework.boot.logging.LogLevel;
import org.springframework.context.annotation.Configuration;
import org.springframework.stereotype.Component;

import java.util.Objects;

/**
 * 切面处理 - 跟踪日志
 *
 * @author lixiaolei
 * @since 1.0
 */
@Component
@Aspect
@Slf4j
public class TraceAspect {

    public final static LogLevel DEFAULT_LOG_LEVEL = LogLevel.DEBUG;

    @Around("@annotation(com.cpp.supplychain.bss.commons.annotations.Trace)")
    public Object trace(ProceedingJoinPoint jointPoint) throws Throwable {
        Object[] request = jointPoint.getArgs();
        if (jointPoint.getSignature() instanceof MethodSignature) {
            MethodSignature methodSignature = (MethodSignature) jointPoint.getSignature();
            Trace traceAnnotation = methodSignature.getMethod().getDeclaredAnnotation(Trace.class);
            LogLevel logLevel = traceAnnotation.level();
            if (traceAnnotation.input()) {
                log(logLevel, "{}, input -> {}", buildKey(jointPoint), StringUtils.join(request, ","));
            }
            Object response = jointPoint.proceed();
            if (Objects.nonNull(response) && traceAnnotation.output()) {
                log(logLevel, "{}, output -> {}", buildKey(jointPoint), response);
            }
            return response;
        } else {
            return jointPoint.proceed();
        }
    }

    /**
     * 记录日志
     *
     * @param level     日志级别
     * @param format    格式化字符
     * @param arguments 占位符数据
     */
    void log(LogLevel level, String format, Object... arguments) {
        switch (level) {
            case TRACE:
                log.trace(format, arguments);
                break;
            case DEBUG:
                log.debug(format, arguments);
                break;
            case INFO:
                log.info(format, arguments);
                break;
            case WARN:
                log.warn(format, arguments);
                break;
            case ERROR:
            case OFF:
            case FATAL:
                log.error(format, arguments);
                break;
            default:
                log.info(format, arguments);
                break;
        }
    }

    /**
     * 获取日志级别
     *
     * @param jointPoint 切点
     * @return 日志级别
     */
    static LogLevel getLogLevel(ProceedingJoinPoint jointPoint) {
        if (jointPoint.getSignature() instanceof MethodSignature) {
            MethodSignature methodSignature = (MethodSignature) jointPoint.getSignature();
            Trace declaredAnnotation = methodSignature.getMethod().getDeclaredAnnotation(Trace.class);
            return declaredAnnotation.level();
        } else {
            return DEFAULT_LOG_LEVEL;
        }
    }


    /**
     * 获取关键词
     *
     * @param jointPoint 切点
     * @return 管检测
     */
    static String buildKey(ProceedingJoinPoint jointPoint) {
        if (jointPoint.getSignature() instanceof MethodSignature) {
            MethodSignature methodSignature = (MethodSignature) jointPoint.getSignature();
            Trace declaredAnnotation = methodSignature.getMethod().getDeclaredAnnotation(Trace.class);
            if (StringUtils.isNotBlank(declaredAnnotation.label())) {
                return declaredAnnotation.label();
            } else {
                return String.format("%s#%s", methodSignature.getDeclaringTypeName(), methodSignature);
            }
        } else {
            return "";
        }
    }
}
