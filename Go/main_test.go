package main

import "testing"

func TestTotals(t *testing.T) {
	total, count := totals([]Item{{"P1", 2, 10.5}, {"P2", 1, 5.0}})
	if total != 26.0 || count != 3 {
		t.Fatalf("got total=%v count=%v, want 26 and 3", total, count)
	}
}

func TestValidate(t *testing.T) {
	if validate(OrderRequest{CustomerID: "C1", Items: []Item{{"P1", 1, 1}}}) != "" {
		t.Fatal("valid order rejected")
	}
	for _, bad := range []OrderRequest{
		{Items: []Item{{"P1", 1, 1}}},
		{CustomerID: "C1"},
		{CustomerID: "C1", Items: []Item{{"P1", 0, 1}}},
		{CustomerID: "C1", Items: []Item{{"P1", 1, -1}}},
	} {
		if validate(bad) == "" {
			t.Fatalf("invalid order accepted: %+v", bad)
		}
	}
}
