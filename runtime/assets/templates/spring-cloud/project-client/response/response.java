package com.cpp.supplychain.payment.client.response;

import com.fasterxml.jackson.annotation.JsonFormat;
import com.fasterxml.jackson.annotation.JsonInclude;
import io.swagger.annotations.ApiModel;
import io.swagger.annotations.ApiModelProperty;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.math.BigDecimal;
import java.time.LocalDateTime;


/**
 * 支付流水Resp
 *
 * @author yuanzhi.wang
 */
@Data
@ApiModel(value = "支付流水Resp")
@JsonInclude(JsonInclude.Include.NON_NULL)
public class PaymentFlowPageResponse {

    @ApiModelProperty("支付流水记录主键")
    private Long payFlowId;

    @ApiModelProperty(value = "付款方ID")
    private String payerId;

    @ApiModelProperty(value = "付款方名称")
    private String payerName;

    @ApiModelProperty(value = "商户ID")
    private String merchantId;

    @ApiModelProperty(value = "商户名称")
    private String merchantName;

    @ApiModelProperty(value = "订单编号")
    private String orderNo;

    @ApiModelProperty(value = "订单创建时间")
    @JsonFormat(pattern = "yyyy-MM-dd HH:mm:ss", timezone = "GMT+8")
    private LocalDateTime orderCreateTime;

    @ApiModelProperty(value = "支付渠道")
    private String payChannel;

    @ApiModelProperty(value = "订单金额")
    private BigDecimal orderAmount;

    @ApiModelProperty(value = "支付时间")
    @JsonFormat(pattern = "yyyy-MM-dd HH:mm:ss", timezone = "GMT+8")
    private LocalDateTime payTime;

    @ApiModelProperty(value = "支付金额")
    private BigDecimal payAmount;

    @ApiModelProperty(value = "支付状态")
    private String payState;

    @ApiModelProperty(value = "渠道流水号")
    private String channelSerNo;

    @ApiModelProperty(value = "支付方式，关联渠道表的信息")
    private String payMethod;

    @ApiModelProperty(value = "订单标题")
    private String orderTitle;

    @ApiModelProperty(value = "订单详情")
    private String orderDetail;

    @ApiModelProperty(value = "币种，ISO 4217标准代码，默认 CNY")
    private String currency;

    @ApiModelProperty(value = "调用支付的终端ip")
    private String paymentCreateIp;

    @ApiModelProperty(value = "")
    private String attributes;

    @ApiModelProperty(value = "支付渠道代码")
    private String channelCode;

}
