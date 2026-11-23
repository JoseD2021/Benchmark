package com.example.app.dto;
import java.util.List;

public class OrdersRequest {
    private String customerId;
    private List<OrdersItemsRequest> items;

    public String getCustomerId() {
        return this.customerId;
    }

    public void setCustomerId(String customerId) {
        this.customerId = customerId;
    }

    public List<OrdersItemsRequest> getItems() {
        return items;
    }

    public void setItems(List<OrdersItemsRequest> items) {
        this.items = items;
    }
}
