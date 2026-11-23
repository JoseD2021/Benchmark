package com.example.app.controller;

import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RestController;

import com.example.app.dto.OrdersItemsRequest;
import com.example.app.dto.OrdersRequest;
import com.example.app.dto.OrdersResponse;
import com.example.app.service.OrdersService;

@RestController
public class OrdersController {

    private final OrdersService ordersService;

    public OrdersController(OrdersService ordersService) {
        this.ordersService = ordersService;
    }

    @PostMapping("/api/orders")
    public ResponseEntity<?> createOrder(@RequestBody OrdersRequest request) {
        if (!isValid(request)) {
            return ResponseEntity.badRequest().body(
                    java.util.Map.of("error", "customerId e items (productId, quantity > 0, price >= 0) son requeridos"));
        }
        OrdersResponse response = ordersService.createOrder(request);
        return ResponseEntity.status(HttpStatus.CREATED).body(response);
    }

    private boolean isValid(OrdersRequest request) {
        if (request == null || request.getCustomerId() == null || request.getCustomerId().isBlank()
                || request.getItems() == null || request.getItems().isEmpty()) {
            return false;
        }
        for (OrdersItemsRequest item : request.getItems()) {
            if (item == null || item.getProductId() == null || item.getProductId().isBlank()
                    || item.getQuantity() <= 0 || item.getPrice() == null || item.getPrice().signum() < 0) {
                return false;
            }
        }
        return true;
    }
}
