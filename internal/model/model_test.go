package model

import "testing"

func TestProductStatusEnum(t *testing.T) {
	for _, s := range []ProductStatus{ProductStatusActive, ProductStatusSold, ProductStatusOffline} {
		if !s.Valid() {
			t.Errorf("status %d should be valid", s)
		}
	}
	for _, s := range []ProductStatus{0, 4, 99, -1} {
		if s.Valid() {
			t.Errorf("status %d should be invalid", s)
		}
	}
	// Establish the int enum contract: 1=active, 2=sold, 3=offline
	if ProductStatusActive != 1 || ProductStatusSold != 2 || ProductStatusOffline != 3 {
		t.Errorf("int enum mismatch: %d %d %d", ProductStatusActive, ProductStatusSold, ProductStatusOffline)
	}
}

func TestRatingSummaryRounding(t *testing.T) {
	cases := []struct {
		in   float64
		want float64
	}{
		{4.0, 4.0},
		{4.44, 4.4},
		{4.45, 4.5},
		{3.14159, 3.1},
	}
	for _, c := range cases {
		rs := &RatingSummary{Count: 1, Average: round1(c.in)}
		if rs.Average != c.want {
			t.Errorf("round1(%v) = %v, want %v", c.in, rs.Average, c.want)
		}
	}
}

func round1(f float64) float64 {
	return float64(int(f*10+0.5)) / 10
}
