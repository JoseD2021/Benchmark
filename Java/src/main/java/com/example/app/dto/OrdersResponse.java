package com.example.app.dto;

import java.math.BigDecimal;
import java.time.LocalDateTime;

public record OrdersResponse(
        String orderId,
        String customerId,
        BigDecimal totalAmount,
        int itemsCount,
        LocalDateTime processedAt) {
}
