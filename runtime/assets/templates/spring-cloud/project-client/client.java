package com.cpp.supplychain.payment.client;


import com.cpp.supplychain.bss.commons.model.PageModel;
import com.cpp.supplychain.bss.commons.model.Result;
import com.cpp.supplychain.payment.client.request.PaymentFlowPageQueryRequest;
import com.cpp.supplychain.payment.client.request.PaymentStateQueryRequest;
import com.cpp.supplychain.payment.client.response.PaymentFlowPageResponse;
import com.cpp.supplychain.payment.client.response.PaymentStateQueryResponse;
import org.springframework.cloud.openfeign.FeignClient;
import org.springframework.cloud.openfeign.SpringQueryMap;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;

import javax.validation.Valid;

/**
 * 支付流水查询Client
 */
@FeignClient(value = "payment", contextId = "PaymentQueryClient")
public interface PaymentQueryClient {

    /**
     * 流水分页查询
     */
    @GetMapping(value = "/payment/v1/payment-flow:page")
    Result<PageModel<PaymentFlowPageResponse>> getPaymentFlowPage(@SpringQueryMap PaymentFlowPageQueryRequest request);

    /**
     * 流水详情
     */
    @GetMapping(value = "/payment/v1/payment-flow/{flowId}")
    Result<PaymentFlowPageResponse> getPaymentFlow(@PathVariable(value = "flowId") Long flowId);

    /**
     * 查询支付状态
     *
     * @param request 查询参数
     * @return 支付状态
     */
    @PostMapping(value = "/payment/v1/getPayState")
    Result<PaymentStateQueryResponse> getPaymentState(@Valid @RequestBody PaymentStateQueryRequest request);
}
