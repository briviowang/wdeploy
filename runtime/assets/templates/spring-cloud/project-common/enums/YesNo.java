package com.cpp.supplychain.bss.commons.enums;

import lombok.AllArgsConstructor;

import java.util.Arrays;

/**
 * yes or no
 *
 * @author lixiaolei
 */
@AllArgsConstructor
public enum YesNo {
    /**
     * yes
     */
    Y(1, true, "yes"),

    /**
     * no
     */
    N(0, false, "no");
    public final int code;
    public final boolean logic;
    public final String desc;

    public static YesNo of(int code) {
        return Arrays.stream(YesNo.values())
                .filter(i -> i.code == code).findFirst()
                .orElseThrow(() -> new IllegalArgumentException("错误的枚举值"));
    }
}
