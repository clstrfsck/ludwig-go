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

package ludwig

import (
	nc "ludwig-go/internal/ncurses"
	"os"
	"strconv"
)

type ncursesColor struct {
	r int
	g int
	b int
}

// Color pair assignments (indices into ncurses color table)
const (
	ColorPairKeyword  = 1
	ColorPairString   = 2
	ColorPairComment  = 3
	ColorPairType     = 4
	ColorPairPreproc  = 5
	ColorPairConstant = 6
	ColorPairSpecial  = 7
	ColorPairError    = 8
)

// Global variable to track if colors are enabled
var (
	colorsEnabled = false
)

func ColorInit() map[string]int {
	// Check if colors are enabled in the config file
	var colors map[string]int = nil
	if FileData.Highlighting {
		colors, colorsEnabled = initColors()
	}
	return colors
}

func ColorReset() {
	if colorsEnabled {
		// Reset xterm colors to default: OSC 104 BEL
		// A mystery why ncurses does not do this.
		os.Stdout.WriteString("\033]104\007")
		os.Stdout.Sync()
	}
}

func hexToRGB(hex string) (int, int, int) {
	if len(hex) != 7 || hex[0] != '#' {
		return 0, 0, 0 // Default to black on invalid input
	}
	r := hexToInt(hex[1:3])
	g := hexToInt(hex[3:5])
	b := hexToInt(hex[5:7])
	return r, g, b
}

func hexToInt(s string) int {
	value, err := strconv.ParseInt(s, 16, 0)
	if err != nil {
		return 0
	}
	// Scale to ncurses range (0-1000)
	return int(value * 1000 / 255)
}

// convertColorSchemeToPairs converts a color scheme to color pair indices
func convertColorSchemeToPairs(scheme map[string]string) map[string]int {
	colorMap := make(map[string]int)
	colorIndexes := make(map[ncursesColor]int)
	colorIndex := 17 // Start after the basic ANSI colors + bright variants
	pairIndex := 1
	for name, hex := range scheme {
		var newColor ncursesColor
		newColor.r, newColor.g, newColor.b = hexToRGB(hex)
		index, exists := colorIndexes[newColor]
		if !exists {
			if colorIndex >= nc.Colors() || pairIndex >= nc.ColorPairs() {
				// No more colors/pairs available, break out of the loop
				break
			}
			// Create a new color and color pair for this color
			thisColorIndex := colorIndex
			colorIndex += 1
			index = pairIndex
			pairIndex += 1
			colorIndexes[newColor] = index
			nc.InitColor(thisColorIndex, newColor.r, newColor.g, newColor.b)
			nc.InitPair(index, thisColorIndex, -1)
		}
		colorMap[name] = index
	}
	return colorMap
}

// initColors initializes color pairs; returns true if colors are available
func initColors() (map[string]int, bool) {
	if !nc.HasColors() {
		return nil, false
	}
	nc.StartColor()
	if nc.Colors() < 8 {
		return nil, false
	}
	nc.UseDefaultColors()
	if nc.Colors() < 256 || nc.ColorPairs() < 256 || !nc.CanChangeColor() {
		// Basic color scheme
		nc.InitPair(ColorPairKeyword, nc.COLOR_YELLOW, -1)
		nc.InitPair(ColorPairString, nc.COLOR_GREEN, -1)
		nc.InitPair(ColorPairComment, nc.COLOR_CYAN, -1)
		nc.InitPair(ColorPairType, nc.COLOR_BLUE, -1)
		nc.InitPair(ColorPairPreproc, nc.COLOR_MAGENTA, -1)
		nc.InitPair(ColorPairConstant, nc.COLOR_WHITE, -1)
		nc.InitPair(ColorPairSpecial, nc.COLOR_WHITE, -1)
		nc.InitPair(ColorPairError, nc.COLOR_RED, -1)
		colorMap := map[string]int{
			"statement":  ColorPairKeyword,
			"keyword":    ColorPairKeyword,
			"type":       ColorPairType,
			"string":     ColorPairString,
			"stringx":    ColorPairString,
			"comment":    ColorPairComment,
			"preproc":    ColorPairPreproc,
			"constant":   ColorPairConstant,
			"special":    ColorPairSpecial,
			"underlined": ColorPairSpecial,
			"todo":       ColorPairSpecial,
			"error":      ColorPairError,
		}
		return colorMap, true
	}

	colorScheme := map[string]string{
		"default":              "#F8F8F8",
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

	colorMap := convertColorSchemeToPairs(colorScheme)
	return colorMap, true
}

// ColorOn turns on the specified color pair
func ColorOn(pair int) {
	stdscr.ColorOn(pair)
}

// ColorOff turns off the specified color pair
func ColorOff(pair int) {
	stdscr.ColorOff(pair)
}
