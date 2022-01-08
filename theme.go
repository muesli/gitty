package main

import (
	"github.com/muesli/termenv"
)

// Theme defines the colors used by gitty.
type Theme struct {
	colorBlack    string
	colorRed      string
	colorYellow   string
	colorGreen    string
	colorBlue     string
	colorTooltip  string
	colorDarkGray string
	colorGray     string
	colorMagenta  string
	colorCyan     string
}

func defaultThemeName() string {
	if !termenv.HasDarkBackground() {
		return "light"
	}
	return "dark"
}

func initTheme() {
	themes := make(map[string]Theme)

	themes["dark"] = Theme{
		colorBlack:    "#222222",
		colorRed:      "#E88388",
		colorYellow:   "#DBAB79",
		colorGreen:    "#A8CC8C",
		colorBlue:     "#71BEF2",
		colorDarkGray: "#888888",
		colorTooltip:  "#555555",
		colorGray:     "#B9BFCA",
		colorMagenta:  "#D290E4",
		colorCyan:     "#66C2CD",
	}

	themes["light"] = Theme{
		colorBlack:    "#eeeeee",
		colorRed:      "#D70000",
		colorYellow:   "#FFAF00",
		colorGreen:    "#005F00",
		colorBlue:     "#000087",
		colorDarkGray: "#303030",
		colorTooltip:  "#303030",
		colorGray:     "#303030",
		colorMagenta:  "#AF00FF",
		colorCyan:     "#0087FF",
	}
	theme = themes[defaultThemeName()]
}
