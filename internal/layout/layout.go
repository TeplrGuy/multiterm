package layout

import (
	"fmt"
	"math"
	"strings"
)

// Layout type constants.
const (
	Grid       = "grid"
	Vertical   = "vertical"
	Horizontal = "horizontal"
	MainSide   = "main-side"
)

// SplitDirection represents the axis along which a tmux pane is split.
type SplitDirection string

const (
	SplitV SplitDirection = "vertical"   // -v flag in tmux (top/bottom)
	SplitH SplitDirection = "horizontal" // -h flag in tmux (left/right)
)

// Split describes a single tmux split operation.
type Split struct {
	Direction  SplitDirection
	TargetPane int
	Percentage int
}

// LayoutPlan holds the full set of splits needed to produce the desired layout.
type LayoutPlan struct {
	Name           string
	Splits         []Split
	Count          int
	TmuxLayoutName string // tmux built-in layout applied via select-layout
}

// AvailableLayouts returns the list of recognised layout names.
func AvailableLayouts() []string {
	return []string{Grid, Vertical, Horizontal, MainSide}
}

// ParseLayout validates a layout name string and returns the corresponding
// constant.  It is case-insensitive.
func ParseLayout(name string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case Grid:
		return Grid, nil
	case Vertical:
		return Vertical, nil
	case Horizontal:
		return Horizontal, nil
	case MainSide, "mainside", "main_side":
		return MainSide, nil
	default:
		return "", fmt.Errorf("unknown layout %q: valid layouts are %s",
			name, strings.Join(AvailableLayouts(), ", "))
	}
}

// Calculate produces a LayoutPlan for the given layout type and pane count.
//
// The key insight is that tmux's built-in layouts (tiled, even-vertical, etc.)
// handle the actual geometry.  We only need to create the right number of panes
// via (count-1) splits, then apply select-layout with the appropriate name.
func Calculate(layoutType string, count int) (*LayoutPlan, error) {
	if count < 1 {
		return nil, fmt.Errorf("pane count must be at least 1, got %d", count)
	}

	plan := &LayoutPlan{
		Name:  layoutType,
		Count: count,
	}

	if count == 1 {
		plan.TmuxLayoutName = tmuxLayoutName(layoutType)
		return plan, nil
	}

	switch layoutType {
	case Grid:
		plan.Splits = gridSplits(count)
		plan.TmuxLayoutName = "tiled"
	case Vertical:
		plan.Splits = verticalSplits(count)
		plan.TmuxLayoutName = "even-vertical"
	case Horizontal:
		plan.Splits = horizontalSplits(count)
		plan.TmuxLayoutName = "even-horizontal"
	case MainSide:
		plan.Splits = mainSideSplits(count)
		plan.TmuxLayoutName = "main-vertical"
	default:
		return nil, fmt.Errorf("unsupported layout type %q", layoutType)
	}

	return plan, nil
}

// tmuxLayoutName maps a layout constant to the corresponding tmux built-in
// layout name.
func tmuxLayoutName(layout string) string {
	switch layout {
	case Grid:
		return "tiled"
	case Vertical:
		return "even-vertical"
	case Horizontal:
		return "even-horizontal"
	case MainSide:
		return "main-vertical"
	default:
		return "tiled"
	}
}

// gridSplits creates (count-1) splits for a grid layout.
//
// Algorithm:
//   - cols = ceil(sqrt(N))
//   - rows = ceil(N / cols)
//   - First split the window into `rows` horizontal strips.
//   - Then split each strip into the needed columns.
//
// After all panes exist tmux's "tiled" layout rearranges them.
func gridSplits(count int) []Split {
	cols := int(math.Ceil(math.Sqrt(float64(count))))
	rows := int(math.Ceil(float64(count) / float64(cols)))

	splits := make([]Split, 0, count-1)

	// Create horizontal strips (rows).
	for i := 1; i < rows; i++ {
		remaining := rows - i + 1
		pct := 100 / remaining
		splits = append(splits, Split{
			Direction:  SplitV,
			TargetPane: 0,
			Percentage: pct,
		})
	}

	// Split each row into columns.
	pane := 0
	for r := 0; r < rows; r++ {
		// The last row may have fewer columns.
		colsInRow := cols
		if r == rows-1 && count%cols != 0 {
			colsInRow = count % cols
		}
		for c := 1; c < colsInRow; c++ {
			remaining := colsInRow - c + 1
			pct := 100 / remaining
			splits = append(splits, Split{
				Direction:  SplitH,
				TargetPane: pane,
				Percentage: pct,
			})
		}
		pane += colsInRow
	}

	return splits
}

// verticalSplits creates (count-1) horizontal splits so panes stack top-to-bottom.
func verticalSplits(count int) []Split {
	splits := make([]Split, 0, count-1)
	for i := 1; i < count; i++ {
		remaining := count - i + 1
		pct := 100 / remaining
		splits = append(splits, Split{
			Direction:  SplitV,
			TargetPane: i - 1,
			Percentage: pct,
		})
	}
	return splits
}

// horizontalSplits creates (count-1) vertical splits so panes sit side-by-side.
func horizontalSplits(count int) []Split {
	splits := make([]Split, 0, count-1)
	for i := 1; i < count; i++ {
		remaining := count - i + 1
		pct := 100 / remaining
		splits = append(splits, Split{
			Direction:  SplitH,
			TargetPane: i - 1,
			Percentage: pct,
		})
	}
	return splits
}

// mainSideSplits creates a 60/40 main-side layout.
//
//   - First split vertically at 60 % for the main pane.
//   - Then split the right pane horizontally (count-2) times.
func mainSideSplits(count int) []Split {
	splits := make([]Split, 0, count-1)

	// Initial vertical split: main (60 %) | side (40 %).
	splits = append(splits, Split{
		Direction:  SplitH,
		TargetPane: 0,
		Percentage: 60,
	})

	// Stack remaining panes on the right side.
	sidePanes := count - 1
	for i := 1; i < sidePanes; i++ {
		remaining := sidePanes - i + 1
		pct := 100 / remaining
		splits = append(splits, Split{
			Direction:  SplitV,
			TargetPane: 1,
			Percentage: pct,
		})
	}

	return splits
}
