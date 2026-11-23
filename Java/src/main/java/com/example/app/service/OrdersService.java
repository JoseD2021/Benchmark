package com.example.app.service;

import java.math.BigDecimal;
import java.math.RoundingMode;
import java.sql.Timestamp;
import java.time.LocalDateTime;
import java.util.List;
import java.util.UUID;

import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.app.dto.OrdersItemsRequest;
import com.example.app.dto.OrdersRequest;
import com.example.app.dto.OrdersResponse;

@Service
public class OrdersService {

    private static final String INSERT_ORDER =
            "INSERT INTO orders (order_id, customer_id, total_amount, items_count, created_at) VALUES (?, ?, ?, ?, ?)";
    private static final String INSERT_ITEM =
            "INSERT INTO order_items (order_id, product_id, quantity, price) VALUES (?, ?, ?, ?)";

    private final JdbcTemplate jdbc;

    public OrdersService(JdbcTemplate jdbc) {
        this.jdbc = jdbc;
    }

    @Transactional
    public OrdersResponse createOrder(OrdersRequest request) {
        List<OrdersItemsRequest> items = request.getItems();

        BigDecimal totalAmount = BigDecimal.ZERO;
        int itemsCount = 0;
        for (OrdersItemsRequest item : items) {
            totalAmount = totalAmount.add(item.getPrice().multiply(BigDecimal.valueOf(item.getQuantity())));
            itemsCount += item.getQuantity();
        }
        totalAmount = totalAmount.setScale(2, RoundingMode.HALF_UP);

        String orderId = UUID.randomUUID().toString();
        LocalDateTime processedAt = LocalDateTime.now();

        jdbc.update(INSERT_ORDER, orderId, request.getCustomerId(), totalAmount, itemsCount,
                Timestamp.valueOf(processedAt));

        jdbc.batchUpdate(INSERT_ITEM, items, items.size(), (ps, item) -> {
            ps.setString(1, orderId);
            ps.setString(2, item.getProductId());
            ps.setInt(3, item.getQuantity());
            ps.setBigDecimal(4, item.getPrice());
        });

        return new OrdersResponse(orderId, request.getCustomerId(), totalAmount, itemsCount, processedAt);
    }
}
