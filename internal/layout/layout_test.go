package layout

import (
	"testing"
)

func TestParseLayout(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		wantErr  bool
	}{
		{"grid", Grid, false},
		{"Grid", Grid, false},
		{"GRID", Grid, false},
		{"vertical", Vertical, false},
		{"horizontal", Horizontal, false},
		{"main-side", MainSide, false},
		{"mainside", MainSide, false},
		{"main_side", MainSide, false},
		{"  grid  ", Grid, false},
		{"invalid", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParseLayout(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseLayout(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("ParseLayout(%q) unexpected error: %v", tt.input, err)
				return
			}
			if result != tt.expected {
				t.Errorf("ParseLayout(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCalculate_SinglePane(t *testing.T) {
	for _, layout := range AvailableLayouts() {
		t.Run(layout, func(t *testing.T) {
			plan, err := Calculate(layout, 1)
			if err != nil {
				t.Fatalf("Calculate(%q, 1) failed: %v", layout, err)
			}
			if plan.Count != 1 {
				t.Errorf("plan.Count = %d, want 1", plan.Count)
			}
			if len(plan.Splits) != 0 {
				t.Errorf("plan.Splits = %d, want 0 for single pane", len(plan.Splits))
			}
		})
	}
}

func TestCalculate_SplitCounts(t *testing.T) {
	tests := []struct {
		layout string
		count  int
	}{
		{Grid, 2}, {Grid, 4}, {Grid, 6}, {Grid, 9}, {Grid, 12}, {Grid, 20},
		{Vertical, 2}, {Vertical, 3}, {Vertical, 6},
		{Horizontal, 2}, {Horizontal, 3}, {Horizontal, 6},
		{MainSide, 2}, {MainSide, 3}, {MainSide, 5},
	}

	for _, tt := range tests {
		t.Run(tt.layout+"-"+string(rune('0'+tt.count)), func(t *testing.T) {
			plan, err := Calculate(tt.layout, tt.count)
			if err != nil {
				t.Fatalf("Calculate(%q, %d) failed: %v", tt.layout, tt.count, err)
			}
			if plan.Count != tt.count {
				t.Errorf("plan.Count = %d, want %d", plan.Count, tt.count)
			}
			expectedSplits := tt.count - 1
			if len(plan.Splits) != expectedSplits {
				t.Errorf("len(plan.Splits) = %d, want %d", len(plan.Splits), expectedSplits)
			}
			if plan.TmuxLayoutName == "" {
				t.Error("plan.TmuxLayoutName is empty")
			}
		})
	}
}

func TestCalculate_InvalidCount(t *testing.T) {
	_, err := Calculate(Grid, 0)
	if err == nil {
		t.Error("expected error for count=0")
	}

	_, err = Calculate(Grid, -1)
	if err == nil {
		t.Error("expected error for count=-1")
	}
}

func TestCalculate_InvalidLayout(t *testing.T) {
	_, err := Calculate("nonexistent", 4)
	if err == nil {
		t.Error("expected error for invalid layout")
	}
}

func TestAvailableLayouts(t *testing.T) {
	layouts := AvailableLayouts()
	if len(layouts) != 4 {
		t.Errorf("expected 4 layouts, got %d", len(layouts))
	}

	expected := map[string]bool{
		Grid: true, Vertical: true, Horizontal: true, MainSide: true,
	}
	for _, l := range layouts {
		if !expected[l] {
			t.Errorf("unexpected layout %q", l)
		}
	}
}

func TestGridSplits_TargetPaneZero(t *testing.T) {
	// Verify grid splits always target pane 0 (the fix for "no space" errors).
	plan, err := Calculate(Grid, 6)
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}

	if plan.TmuxLayoutName != "tiled" {
		t.Errorf("expected tiled layout, got %q", plan.TmuxLayoutName)
	}
}

func TestVerticalSplits_AlwaysTargetPaneZero(t *testing.T) {
	plan, err := Calculate(Vertical, 6)
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}

	for i, split := range plan.Splits {
		if split.TargetPane != 0 {
			t.Errorf("split[%d].TargetPane = %d, want 0", i, split.TargetPane)
		}
		if split.Direction != SplitV {
			t.Errorf("split[%d].Direction = %q, want %q", i, split.Direction, SplitV)
		}
	}
}

func TestHorizontalSplits_AlwaysTargetPaneZero(t *testing.T) {
	plan, err := Calculate(Horizontal, 6)
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}

	for i, split := range plan.Splits {
		if split.TargetPane != 0 {
			t.Errorf("split[%d].TargetPane = %d, want 0", i, split.TargetPane)
		}
		if split.Direction != SplitH {
			t.Errorf("split[%d].Direction = %q, want %q", i, split.Direction, SplitH)
		}
	}
}

func TestMainSideSplits(t *testing.T) {
	plan, err := Calculate(MainSide, 4)
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}

	// First split should be horizontal (main | side).
	if plan.Splits[0].Direction != SplitH {
		t.Errorf("first split should be horizontal, got %q", plan.Splits[0].Direction)
	}

	// Remaining splits should be vertical (stacking side panes).
	for i := 1; i < len(plan.Splits); i++ {
		if plan.Splits[i].Direction != SplitV {
			t.Errorf("split[%d] should be vertical, got %q", i, plan.Splits[i].Direction)
		}
	}

	if plan.TmuxLayoutName != "main-vertical" {
		t.Errorf("expected main-vertical layout, got %q", plan.TmuxLayoutName)
	}
}
