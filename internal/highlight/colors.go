/**********************************************************************}
{                                                                      }
{            L      U   U   DDDD   W      W  IIIII   GGGG              }
{            L      U   U   D   D   W    W     I    G                  }
{            L      U   U   D   D   W ww W     I    G   GG             }
{            L      U   U   D   D    W  W      I    G    G             }
{            LLLLL   UUU    DDDD     W  W    IIIII   GGGG              }
{                                                                      }
{**********************************************************************/

// Name:         COLORS
//
// Description:  This module handles color management for the ncurses version
//               of Ludwig, including mapping syntax groups to ncurses color
//               pairs.

package highlight

import (
	nc "ludwig-go/internal/ncurses"
	"math"
	"strconv"
	"strings"
)

// ColorInit initialises the color system.
func ColorInit(enabled bool) map[string]int {
	// Check if colors are enabled in the config file
	var colors map[string]int = nil
	if enabled {
		colors = initColors()
	}
	return colors
}

// ColorReset shuts down the color system
func ColorReset() {
	// Nothing to do here
}

// sqDist returns the square of the euclidean distance between r1,g1,b1
// and r2,g2,b2
func sqDist(r1, g1, b1, r2, g2, b2 int) int {
	return (r1-r2)*(r1-r2) + (g1-g2)*(g1-g2) + (b1-b2)*(b1-b2)
}

// hexToXterm converts a hex string (#rrggbb) to an xterm 256 color index.
// It uses Euclidean distance to find the closest match.
func hexToXterm(hex string) (int, bool) {
	// Clean and Parse Hex
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return -1, false
	}

	rVal, err := strconv.ParseInt(hex[0:2], 16, 64)
	if err != nil {
		return -1, false
	}
	gVal, err := strconv.ParseInt(hex[2:4], 16, 64)
	if err != nil {
		return -1, false
	}
	bVal, err := strconv.ParseInt(hex[4:6], 16, 64)
	if err != nil {
		return -1, false
	}

	r, g, b := int(rVal), int(gVal), int(bVal)

	// Helper to calculate squared euclidean distance (faster than Sqrt)
	bestIdx := 0
	minDist := math.MaxInt32

	// Handle Grayscale Case (R == G == B)
	// This gives us better grey scales
	if r == g && g == b {
		// The ramp indices 232-255 map to values 8, 18, 28 ... 238.
		// However, pure black (0) is index 16, and pure white (255) is index 231.
		// We check them all to find the absolute best gray match.

		// Check Pure Black (Index 16)
		dBlack := sqDist(r, g, b, 0, 0, 0)
		if dBlack < minDist {
			minDist = dBlack
			bestIdx = 16
		}

		// Check Pure White (Index 231)
		dWhite := sqDist(r, g, b, 255, 255, 255)
		if dWhite < minDist {
			minDist = dWhite
			bestIdx = 231
		}

		// Check Grayscale Ramp (Indices 232-255)
		for i := 232; i <= 255; i++ {
			val := 8 + (i-232)*10 // 8, 18, 28...
			d := sqDist(r, g, b, val, val, val)
			if d < minDist {
				minDist = d
				bestIdx = i
			}
		}
		return bestIdx, true
	}

	// Handle Color Cube (Indices 16-231) via Distance
	// The standard xterm coordinate values.
	// 0x00, 0x5f, 0x87, 0xaf, 0xd7, 0xff
	steps := []int{0, 95, 135, 175, 215, 255}

	// Iterate over the 6x6x6 cube
	for rIdx := range 6 {
		for gIdx := range 6 {
			for bIdx := range 6 {
				// Get actual RGB values for this point in the cube
				cr := steps[rIdx]
				cg := steps[gIdx]
				cb := steps[bIdx]
				d := sqDist(r, g, b, cr, cg, cb)

				if d < minDist {
					minDist = d
					// Calculate actual index
					bestIdx = 16 + (36 * rIdx) + (6 * gIdx) + bIdx
				}
			}
		}
	}

	return bestIdx, true
}

func hexesToXterms(hex string) (int, int) {
	hexes := strings.Split(hex, ",")
	fg, bg := -1, -1
	if len(hexes) > 0 {
		if candidate, ok := hexToXterm(hexes[0]); ok {
			fg = candidate
		}
	}
	if len(hexes) > 1 {
		if candidate, ok := hexToXterm(hexes[1]); ok {
			bg = candidate
		}
	}
	return fg, bg
}

// convertColorSchemeToPairs converts a color scheme to color pair indices
func convertColorSchemeToPairs(scheme map[string]string) map[string]int {
	defaultFg := -1
	defaultBg := -1
	pairIndex := 1
	colorMap := make(map[string]int)
	if defaultColors, ok := scheme["default"]; ok {
		defaultFg, defaultBg = hexesToXterms(defaultColors)
		nc.InitPair(pairIndex, defaultFg, defaultBg)
		nc.BkColor(pairIndex)
		colorMap["default"] = pairIndex
		pairIndex += 1
	}
	for name, colorString := range scheme {
		if name == "default" {
			// Already done
			continue
		}
		var newFg, newBg int
		if colorString == "default" {
			// Could just leave this out
			newFg, newBg = defaultFg, defaultBg
		} else {
			newFg, newBg = hexesToXterms(colorString)
		}
		if newFg < 0 {
			newFg = defaultFg
		}
		if newBg < 0 {
			newBg = defaultBg
		}
		if pairIndex >= nc.ColorPairs() {
			panic("Pairs exhausted")
		}
		nc.InitPair(pairIndex, newFg, newBg)
		colorMap[name] = pairIndex
		pairIndex += 1
	}
	return colorMap
}

var twilightScheme = map[string]string{
	"default":              "#F8F8F8,#141414",
	"color-column":         "#1B1B1B",
	"comment":              "#5F5A60",
	"constant":             "#CF6A4C",
	"constant.specialChar": "#DDF2A4",
	"constant.string":      "#8F9D6A",
	"current-line-number":  "#868686",
	"cursor-line":          "#1B1B1B",
	"divider":              "#1E1E1E",
	"error":                "#D2A8A1",
	"diff-added":           "#00AF00",
	"diff-modified":        "#FFAF00",
	"diff-deleted":         "#D70000",
	"gutter-error":         "#9B859D",
	"gutter-warning":       "#9B859D",
	"hlsearch":             "#141414",
	"identifier":           "#9B703F",
	"identifier.class":     "#DAD085",
	"identifier.var":       "#7587A6",
	"indent-char":          "#515151",
	"line-number":          "#868686",
	"preproc":              "#E0C589",
	"special":              "#E0C589",
	"statement":            "#CDA869",
	"statusline":           "#515151",
	"symbol":               "#AC885B",
	"symbol.brackets":      "#F8F8F8",
	"symbol.operator":      "#CDA869",
	"symbol.tag":           "#AC885B",
	"tabbar":               "#F2F0EC",
	"todo":                 "#8B98AB",
	"type":                 "#F9EE98",
	"type.keyword":         "#CDA869",
	"underlined":           "#8996A8",
	"match-brace":          "#141414",
	"tab-error":            "#D75F5F",
	"trailingws":           "#D75F5F",
}

func basicScheme() map[string]int {
	// Basic color scheme
	nc.InitPair(1, nc.COLOR_YELLOW, -1)
	nc.InitPair(2, nc.COLOR_GREEN, -1)
	nc.InitPair(3, nc.COLOR_CYAN, -1)
	nc.InitPair(4, nc.COLOR_BLUE, -1)
	nc.InitPair(5, nc.COLOR_MAGENTA, -1)
	nc.InitPair(6, nc.COLOR_WHITE, -1)
	nc.InitPair(7, nc.COLOR_WHITE, -1)
	nc.InitPair(8, nc.COLOR_RED, -1)
	colorMap := map[string]int{
		"statement":  1,
		"keyword":    1,
		"type":       4,
		"string":     2,
		"stringx":    2,
		"comment":    3,
		"preproc":    5,
		"constant":   6,
		"special":    7,
		"underlined": 7,
		"todo":       7,
		"error":      8,
	}
	return colorMap
}

// initColors initializes color pairs; returns true if colors are available
func initColors() map[string]int {
	if !nc.HasColors() {
		return nil
	}
	nc.StartColor()
	if nc.Colors() < 8 {
		return nil
	}
	nc.UseDefaultColors()
	// We just wildly guess how many color pairs we need here
	if nc.Colors() < 256 || nc.ColorPairs() < 256 {
		return basicScheme()
	}
	// Here we will assume standard 6x6x6 color cube with 24 gray ramp
	colorMap := convertColorSchemeToPairs(twilightScheme)
	return colorMap
}
