package com.cpp.supplychain.payment.client.request;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import io.swagger.annotations.ApiModel;
import io.swagger.annotations.ApiModelProperty;
import lombok.Data;


/**
 * 支付流水Page查询Req
 *
 * @author yuanzhi.wang
 */
@Data
@ApiModel(value = "支付流水Page查询Req")
@JsonIgnoreProperties(ignoreUnknown = true)
public class PaymentFlowPageQueryRequest extends BasePageRequest {

    /**
     * 订单编号
     */
    @ApiModelProperty("订单编号")
    private String orderNo;
    /**
     * 支付时间 左区间
     */
    @ApiModelProperty("支付时间 开始时间")
    private String payTimeBegin;
    /**
     * 支付时间 右区间
     */
    @ApiModelProperty("支付时间 结束时间")
    private String payTimeEnd;
    /**
     * 支付状态
     */
    @ApiModelProperty("支付状态")
    private String payState;
    /**
     * 三方支付编号
     */
    @ApiModelProperty("三方支付编号")
    private String channelSerNo;

    /**
     * 支付渠道Code
     */
    @ApiModelProperty("支付渠道")
    private String channelCode;
    /**
     * 支付方式，关联渠道表的信息
     */
    @ApiModelProperty("支付方式，关联渠道表的信息")
    private String payMethod;

}
