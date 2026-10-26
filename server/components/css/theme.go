// Package css has the site's stylesheets, and gives Go the theme's colours for what CSS does not draw
package css

import (
	_ "embed"
	"fmt"
	"math"
	"regexp"
	"strings"
)

//go:embed theme.css
var theme string

var themeColour = regexp.MustCompile(`--color-([a-z0-9-]+):\s*([^;]+);`)
var oklch = regexp.MustCompile(`^oklch\(([\d.]+)%\s+([\d.]+)\s+([\d.]+)\)$`)

// Colour is the theme's colour of that name, e.g. "red-600", as hex for what cannot use CSS, such as
// ECharts. It panics for a colour the theme does not define.
func Colour(name string) string {
	for _, match := range themeColour.FindAllStringSubmatch(theme, -1) {
		if match[1] != name {
			continue
		}
		value := strings.TrimSpace(match[2])
		if strings.HasPrefix(value, "#") {
			return value
		}
		if lch := oklch.FindStringSubmatch(value); lch != nil {
			var l, c, h float64
			fmt.Sscan(lch[1], &l)
			fmt.Sscan(lch[2], &c)
			fmt.Sscan(lch[3], &h)
			return oklchToHex(l/100, c, h)
		}
		panic(fmt.Sprintf("theme colour %s is %q, which is neither hex nor oklch", name, value))
	}
	panic("theme.css has no colour " + name)
}

// oklchToHex converts through OKLab to sRGB, clamping a colour outside sRGB, as Björn Ottosson defines OKLab
func oklchToHex(l, c, h float64) string {
	a, b := c*math.Cos(h*math.Pi/180), c*math.Sin(h*math.Pi/180)
	lc := math.Pow(l+0.3963377774*a+0.2158037573*b, 3)
	mc := math.Pow(l-0.1055613458*a-0.0638541728*b, 3)
	sc := math.Pow(l-0.0894841775*a-1.2914855480*b, 3)
	channel := func(linear float64) int {
		if linear > 0.0031308 {
			linear = 1.055*math.Pow(linear, 1/2.4) - 0.055
		} else {
			linear *= 12.92
		}
		return int(math.Round(math.Max(0, math.Min(1, linear)) * 255))
	}
	return fmt.Sprintf("#%02x%02x%02x",
		channel(4.0767416621*lc-3.3077115913*mc+0.2309699292*sc),
		channel(-1.2684380046*lc+2.6097574011*mc-0.3413193965*sc),
		channel(-0.0041960863*lc-0.7034186147*mc+1.7076147010*sc))
}
