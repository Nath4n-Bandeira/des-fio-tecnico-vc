package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestAnalyzeFileBusinessRulesAndOrdering(t *testing.T) {
	path := writeTestSalesFile(t, []testSale{
		{product: 1, timestamp: "2024-01-10T00:00:00Z", quantity: 10},
		{product: 1, timestamp: "2024-02-10T00:00:00Z", quantity: 20},
		{product: 2, timestamp: "2024-01-10T00:00:00Z", quantity: 9},
		{product: 2, timestamp: "2024-02-10T00:00:00Z", quantity: 19},
		{product: 3, timestamp: "2024-01-10T00:00:00Z", quantity: 5},
		{product: 4, timestamp: "2024-02-10T00:00:00Z", quantity: 99},
		{product: 5, timestamp: "2024-01-10T00:00:00Z", quantity: 3},
		{product: 5, timestamp: "2024-02-10T00:00:00Z", quantity: 4},
		{product: 6, timestamp: "2024-01-10T00:00:00Z", quantity: 5},
		{product: 6, timestamp: "2024-02-10T00:00:00Z", quantity: 5},
	})

	rows, got, err := analyzeFile(path)
	if err != nil {
		t.Fatalf("analyzeFile() error = %v", err)
	}
	if rows != 10 {
		t.Fatalf("rows = %d, want 10", rows)
	}

	want := []productGrowth{
		{productID: "produto02", previous: 9, current: 19, growth: 10, percent: 111.12},
		{productID: "produto01", previous: 10, current: 20, growth: 10, percent: 100.00},
		{productID: "produto05", previous: 3, current: 4, growth: 1, percent: 33.34},
		{productID: "produto06", previous: 5, current: 5, growth: 0, percent: 0.00},
		{productID: "produto03", previous: 5, current: 0, growth: -5, percent: -100.00},
	}

	if len(got) != len(want) {
		t.Fatalf("len(got) = %d, want %d; got = %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got[%d] = %#v, want %#v", i, got[i], want[i])
		}
	}
}

func TestFormatMetrics(t *testing.T) {
	if got := formatDuration(4240_000_000); got != "4,24 s" {
		t.Fatalf("formatDuration() = %q, want %q", got, "4,24 s")
	}

	got := formatCPU(metrics{cpuMeasured: true, cpuPercent: 12.4, logicalCPUs: 1})
	if got != "12% (1 núcleo)" {
		t.Fatalf("formatCPU() = %q, want %q", got, "12% (1 núcleo)")
	}
}

type testSale struct {
	product   int
	timestamp string
	quantity  int
}

func writeTestSalesFile(t *testing.T, sales []testSale) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "sales.dat")
	content := "sale_id:0,11\nstore_id:11,7\nproduct_id:18,7\ntimestamp:25,20\nquantity:45,2\nunit_price:47,7\n-------\n"
	for i, sale := range sales {
		content += fmt.Sprintf("S%010dST00001PR%05d%s%02d%7.2f\n", i+1, sale.product, sale.timestamp, sale.quantity, 9.99)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	return path
}
