package screenshot

import (
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"nuther/internal/config"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

var ansiRE = regexp.MustCompile(`\x1b\[[0-9;?]*[ -/]*[@-~]`)

const overviewLogo = `███╗   ██╗██╗   ██╗████████╗██╗  ██╗███████╗██████╗
████╗  ██║██║   ██║╚══██╔══╝██║  ██║██╔════╝██╔══██╗
██╔██╗ ██║██║   ██║   ██║   ███████║█████╗  ██████╔╝
██║╚██╗██║██║   ██║   ██║   ██╔══██║██╔══╝  ██╔══██╗
██║ ╚████║╚██████╔╝   ██║   ██║  ██║███████╗██║  ██║
╚═╝  ╚═══╝ ╚═════╝    ╚═╝   ╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝`

type imagePalette struct {
	bg, text, dim, border, accent, accent2, success color.Color
}

func RenderTextImage(text, outputPath string) error {
	return RenderOverviewImage(text, outputPath, config.DefaultConfig())
}

func RenderOverviewImage(text, outputPath string, cfg *config.Config) error {
	bodyFace, err := loadMonoFaceSize(15)
	if err != nil {
		return err
	}
	defer bodyFace.Close()

	logoFace, err := loadMonoFaceSize(6)
	if err != nil {
		return err
	}
	defer logoFace.Close()

	p := paletteFromConfig(cfg)
	logoLines := strings.Split(overviewLogo, "\n")
	bodyLines := overviewImageLines(text)
	padding := 26
	logoLineHeight := logoFace.Metrics().Height.Ceil() + 1
	bodyLineHeight := bodyFace.Metrics().Height.Ceil() + 4
	logoAscent := logoFace.Metrics().Ascent.Ceil()
	bodyAscent := bodyFace.Metrics().Ascent.Ceil()

	maxWidth := 1
	for _, line := range logoLines {
		if w := font.MeasureString(logoFace, line).Ceil(); w > maxWidth {
			maxWidth = w
		}
	}
	for _, line := range bodyLines {
		if w := font.MeasureString(bodyFace, line).Ceil(); w > maxWidth {
			maxWidth = w
		}
	}

	logoHeight := len(logoLines) * logoLineHeight
	gap := 18
	img := image.NewRGBA(image.Rect(0, 0, maxWidth+padding*2, logoHeight+gap+len(bodyLines)*bodyLineHeight+padding*2))
	draw.Draw(img, img.Bounds(), image.NewUniform(p.bg), image.Point{}, draw.Src)

	d := &font.Drawer{Dst: img}
	for i, line := range logoLines {
		d.Face = logoFace
		d.Src = image.NewUniform(p.accent)
		d.Dot = fixed.P(padding, padding+logoAscent+i*logoLineHeight)
		d.DrawString(line)
	}

	bodyY := padding + logoHeight + gap
	d.Face = bodyFace
	for i, line := range bodyLines {
		y := bodyY + bodyAscent + i*bodyLineHeight
		drawStyledLine(d, padding, y, line, p)
	}

	return encodeImage(img, outputPath)
}

func overviewImageLines(text string) []string {
	body := strings.TrimRight(ansiRE.ReplaceAllString(text, ""), "\n")
	body = strings.ReplaceAll(body, "s  Screenshot overview", "")
	bodyLines := strings.Split(body, "\n")
	for len(bodyLines) > 0 && strings.TrimSpace(bodyLines[0]) == "" {
		bodyLines = bodyLines[1:]
	}
	return bodyLines
}

func drawStyledLine(d *font.Drawer, x, y int, line string, p imagePalette) {
	trimmed := strings.TrimSpace(line)
	if isCardValueLine(line) {
		drawBoxValueLine(d, x, y, line, p)
		return
	}

	col := p.text
	if strings.HasPrefix(trimmed, "≡") {
		col = p.accent
	} else if strings.Contains(line, "✓") || strings.Contains(line, "GOOD") || strings.Contains(line, "%") {
		col = p.success
	} else if isDimLine(trimmed) {
		col = p.dim
	} else if strings.ContainsAny(line, "╭╮╰╯─│") {
		col = p.border
	}

	d.Src = image.NewUniform(col)
	d.Dot = fixed.P(x, y)
	d.DrawString(line)
}

func drawBoxValueLine(d *font.Drawer, x, y int, line string, p imagePalette) {
	currentX := x
	for _, r := range line {
		col := p.accent
		if r == '│' || r == ' ' {
			col = p.border
		}
		d.Src = image.NewUniform(col)
		d.Dot = fixed.P(currentX, y)
		d.DrawString(string(r))
		currentX += font.MeasureString(d.Face, string(r)).Ceil()
	}
}

func isCardValueLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	if !strings.Contains(trimmed, "│") || !strings.HasPrefix(trimmed, "│") {
		return false
	}
	upper := strings.ToUpper(trimmed)
	if strings.Contains(upper, "TEMPERATURE") || strings.Contains(upper, "POWER ON") || strings.Contains(upper, "PWR CYCLES") || strings.Contains(upper, "DATA WRITTEN") {
		return false
	}
	return strings.Count(trimmed, "│") >= 4
}

func isDimLine(line string) bool {
	return strings.HasPrefix(line, "ID") || strings.Contains(line, "PROTOCOL") || strings.Contains(line, "TEMPERATURE") || strings.HasPrefix(line, "·")
}

func paletteFromConfig(cfg *config.Config) imagePalette {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}
	return imagePalette{
		bg:      hexColor(cfg.Colors.Background, color.RGBA{12, 14, 18, 255}),
		text:    hexColor(cfg.Colors.Text, color.RGBA{232, 238, 246, 255}),
		dim:     hexColor(cfg.Colors.TextDim, color.RGBA{126, 139, 160, 255}),
		border:  hexColor(cfg.Colors.Border, color.RGBA{92, 75, 138, 255}),
		accent:  hexColor(cfg.Colors.AccentPrimary, color.RGBA{164, 145, 255, 255}),
		accent2: hexColor(cfg.Colors.AccentSecondary, color.RGBA{127, 184, 255, 255}),
		success: hexColor(cfg.Colors.Success, color.RGBA{92, 255, 129, 255}),
	}
}

func hexColor(s string, fallback color.RGBA) color.Color {
	s = strings.TrimPrefix(strings.TrimSpace(s), "#")
	if len(s) != 6 {
		return fallback
	}
	r, err := strconv.ParseUint(s[0:2], 16, 8)
	if err != nil {
		return fallback
	}
	g, err := strconv.ParseUint(s[2:4], 16, 8)
	if err != nil {
		return fallback
	}
	b, err := strconv.ParseUint(s[4:6], 16, 8)
	if err != nil {
		return fallback
	}
	return color.RGBA{uint8(r), uint8(g), uint8(b), 255}
}

func encodeImage(img image.Image, outputPath string) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return err
	}
	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	switch strings.ToLower(filepath.Ext(outputPath)) {
	case ".jpg", ".jpeg":
		return jpeg.Encode(f, img, &jpeg.Options{Quality: 92})
	default:
		return png.Encode(f, img)
	}
}

func loadMonoFace() (font.Face, error) {
	return loadMonoFaceSize(15)
}

func loadMonoFaceSize(size float64) (font.Face, error) {
	candidates := []string{
		"/usr/share/fonts/truetype/dejavu/DejaVuSansMono.ttf",
		"/usr/share/fonts/dejavu/DejaVuSansMono.ttf",
		"/Library/Fonts/Menlo.ttc",
		"C:/Windows/Fonts/consola.ttf",
	}
	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		ft, err := opentype.Parse(data)
		if err != nil {
			continue
		}
		return opentype.NewFace(ft, &opentype.FaceOptions{Size: size, DPI: 96, Hinting: font.HintingFull})
	}
	return nil, os.ErrNotExist
}
