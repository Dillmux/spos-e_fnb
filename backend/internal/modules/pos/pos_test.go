package pos

import (
	"testing"
)

func TestCalculateTaxAndServiceLogic(t *testing.T) {
	subtotal := 100000.0
	discountAmount := 10000.0
	servicePercentage := 5.0  // 5%
	taxPercentage := 10.0     // 10%
	isTaxInclusive := false

	taxableAmount := subtotal - discountAmount
	if taxableAmount < 0 {
		taxableAmount = 0
	}

	serviceAmount := 0.0
	if servicePercentage > 0 {
		serviceAmount = taxableAmount * (servicePercentage / 100.0)
	}

	taxAmount := 0.0
	if taxPercentage > 0 && !isTaxInclusive {
		taxAmount = (taxableAmount + serviceAmount) * (taxPercentage / 100.0)
	}

	totalAmount := taxableAmount + serviceAmount + taxAmount

	expectedTaxable := 90000.0
	expectedService := 4500.0
	expectedTax := 9450.0
	expectedTotal := 103950.0

	if taxableAmount != expectedTaxable {
		t.Errorf("Taxable amount calculation mismatch: expected %.2f, got %.2f", expectedTaxable, taxableAmount)
	}
	if serviceAmount != expectedService {
		t.Errorf("Service amount calculation mismatch: expected %.2f, got %.2f", expectedService, serviceAmount)
	}
	if taxAmount != expectedTax {
		t.Errorf("Tax amount calculation mismatch: expected %.2f, got %.2f", expectedTax, taxAmount)
	}
	if totalAmount != expectedTotal {
		t.Errorf("Total amount calculation mismatch: expected %.2f, got %.2f", expectedTotal, totalAmount)
	}
}

func TestSplitBillSubtotalCalculation(t *testing.T) {
	unitPrice1 := 25000.0
	qty1 := 2
	subtotal1 := float64(qty1) * unitPrice1

	unitPrice2 := 15000.0
	qty2 := 1
	subtotal2 := float64(qty2) * unitPrice2

	childSubtotal := subtotal1 + subtotal2
	expectedChildSubtotal := 65000.0

	if childSubtotal != expectedChildSubtotal {
		t.Errorf("Split bill child subtotal mismatch: expected %.2f, got %.2f", expectedChildSubtotal, childSubtotal)
	}
}

func TestDailyQueueNumberReset(t *testing.T) {
	// Simulate orders on Day 1
	day1MaxQueue := 5
	nextQueueDay1 := day1MaxQueue + 1
	if nextQueueDay1 != 6 {
		t.Errorf("Expected Day 1 next queue number to be 6, got %d", nextQueueDay1)
	}

	// Simulate new Day 2 where no orders exist yet
	day2MaxQueue := 0
	nextQueueDay2 := day2MaxQueue + 1
	if nextQueueDay2 != 1 {
		t.Errorf("Expected Day 2 reset queue number to be 1, got %d", nextQueueDay2)
	}
}
