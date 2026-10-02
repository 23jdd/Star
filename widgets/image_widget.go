package widgets

import (
	"image"
	"math"

	"github.com/gdamore/tcell/v2"
)

// ImageFit controls how an Image is scaled into its arranged bounds.
type ImageFit uint8

const (
	// ImageFitContain preserves the aspect ratio and shows the complete image.
	ImageFitContain ImageFit = iota
	// ImageFitCover preserves the aspect ratio and crops overflowing pixels.
	ImageFitCover
	// ImageFitFill stretches the image to exactly fill the available area.
	ImageFitFill
	// ImageFitNone renders one source pixel per terminal half-cell.
	ImageFitNone
)

// ImageVerticalAlign controls vertical placement when the fitted image is
// smaller than its available area.
type ImageVerticalAlign uint8

const (
	// ImageAlignTop places an image at the top edge.
	ImageAlignTop ImageVerticalAlign = iota
	// ImageAlignMiddle centres an image vertically.
	ImageAlignMiddle
	// ImageAlignBottom places an image at the bottom edge.
	ImageAlignBottom
)

// Image renders a Go image.Image with true-colour terminal cells. Each cell
// uses an upper/lower half-block pair, providing two vertical image pixels.
// Fully transparent pixels preserve content already drawn by the parent.
type Image struct {
	WidgetBase
	Source             image.Image
	Fit                ImageFit
	Align              TextAlign
	VerticalAlign      ImageVerticalAlign
	FallbackBackground tcell.Color
}

// NewImage creates a centred, aspect-preserving image widget.
func NewImage(source image.Image) *Image {
	return &Image{
		Source:        source,
		Fit:           ImageFitContain,
		Align:         AlignCenter,
		VerticalAlign: ImageAlignMiddle,
	}
}

// SetImage replaces the rendered source. A nil source renders nothing.
func (i *Image) SetImage(source image.Image) { i.Source = source }

func (i *Image) Measure(constraints Constraints) Size {
	if i.Source == nil {
		return constraints.Constrain(Size{})
	}
	bounds := i.Source.Bounds()
	return constraints.Constrain(Size{W: bounds.Dx(), H: (bounds.Dy() + 1) / 2})
}

func (i *Image) Arrange(bounds Rect) { i.SetBounds(bounds) }
func (i *Image) HandlerEvent(*Event) {}

func (i *Image) Render(screen tcell.Screen) {
	if i.Hidden() || i.Source == nil || i.Bounds().Empty() {
		return
	}
	sourceBounds := i.Source.Bounds()
	if sourceBounds.Empty() {
		return
	}

	targetWidth, targetHeight := i.Bounds().W, i.Bounds().H*2
	drawWidth, drawHeight := imageFitSize(i.Fit, sourceBounds.Dx(), sourceBounds.Dy(), targetWidth, targetHeight)
	if drawWidth <= 0 || drawHeight <= 0 {
		return
	}
	drawX := imageAlignedOffset(i.Align, targetWidth, drawWidth)
	drawY := imageVerticalOffset(i.VerticalAlign, targetHeight, drawHeight)
	screenWidth, screenHeight := screen.Size()

	for cellY := 0; cellY < i.Bounds().H; cellY++ {
		y := i.Bounds().Y + cellY
		if y < 0 || y >= screenHeight {
			continue
		}
		for cellX := 0; cellX < i.Bounds().W; cellX++ {
			x := i.Bounds().X + cellX
			if x < 0 || x >= screenWidth {
				continue
			}

			underTop, underBottom := imageUnderlyingColors(screen, x, y)
			top, hasTop := i.sample(sourceBounds, cellX, cellY*2, drawX, drawY, drawWidth, drawHeight, underTop)
			bottom, hasBottom := i.sample(sourceBounds, cellX, cellY*2+1, drawX, drawY, drawWidth, drawHeight, underBottom)
			switch {
			case hasTop && hasBottom:
				screen.SetContent(x, y, '▀', nil, tcell.StyleDefault.Foreground(top).Background(bottom))
			case hasTop:
				screen.SetContent(x, y, '▀', nil, tcell.StyleDefault.Foreground(top).Background(underBottom))
			case hasBottom:
				screen.SetContent(x, y, '▄', nil, tcell.StyleDefault.Foreground(bottom).Background(underTop))
			}
		}
	}
}

func (i *Image) sample(sourceBounds image.Rectangle, x, y, drawX, drawY, drawWidth, drawHeight int, under tcell.Color) (tcell.Color, bool) {
	if x < drawX || x >= drawX+drawWidth || y < drawY || y >= drawY+drawHeight {
		return tcell.ColorDefault, false
	}
	sourceX := sourceBounds.Min.X + min(sourceBounds.Dx()-1, (x-drawX)*sourceBounds.Dx()/drawWidth)
	sourceY := sourceBounds.Min.Y + min(sourceBounds.Dy()-1, (y-drawY)*sourceBounds.Dy()/drawHeight)
	r, g, b, a := i.Source.At(sourceX, sourceY).RGBA()
	if a == 0 {
		return tcell.ColorDefault, false
	}
	if a == 0xffff {
		return tcell.NewRGBColor(int32(r>>8), int32(g>>8), int32(b>>8)), true
	}

	background := under
	if !background.Valid() {
		background = i.FallbackBackground
	}
	backgroundR, backgroundG, backgroundB := background.RGB()
	if backgroundR >= 0 {
		alpha := int64(a >> 8)
		blend := func(premultiplied uint32, base int32) int32 {
			return int32(min(int64(255), int64(premultiplied>>8)+int64(base)*(255-alpha)/255))
		}
		return tcell.NewRGBColor(blend(r, backgroundR), blend(g, backgroundG), blend(b, backgroundB)), true
	}

	// The terminal default colour cannot be queried. Preserve the source hue
	// by removing premultiplication and treat the resulting pixel as opaque.
	unpremultiply := func(value uint32) int32 {
		return int32(min(uint64(255), uint64(value)*255/uint64(a)))
	}
	return tcell.NewRGBColor(unpremultiply(r), unpremultiply(g), unpremultiply(b)), true
}

func imageFitSize(fit ImageFit, sourceWidth, sourceHeight, targetWidth, targetHeight int) (int, int) {
	if sourceWidth <= 0 || sourceHeight <= 0 || targetWidth <= 0 || targetHeight <= 0 {
		return 0, 0
	}
	switch fit {
	case ImageFitNone:
		return sourceWidth, sourceHeight
	case ImageFitFill:
		return targetWidth, targetHeight
	}

	widthScale := float64(targetWidth) / float64(sourceWidth)
	heightScale := float64(targetHeight) / float64(sourceHeight)
	scale := math.Min(widthScale, heightScale)
	if fit == ImageFitCover {
		scale = math.Max(widthScale, heightScale)
	}
	return max(1, int(math.Round(float64(sourceWidth)*scale))), max(1, int(math.Round(float64(sourceHeight)*scale)))
}

func imageAlignedOffset(align TextAlign, available, used int) int {
	switch align {
	case AlignCenter:
		return (available - used) / 2
	case AlignRight:
		return available - used
	default:
		return 0
	}
}

func imageVerticalOffset(align ImageVerticalAlign, available, used int) int {
	switch align {
	case ImageAlignMiddle:
		return (available - used) / 2
	case ImageAlignBottom:
		return available - used
	default:
		return 0
	}
}

func imageUnderlyingColors(screen tcell.Screen, x, y int) (tcell.Color, tcell.Color) {
	main, _, style, _ := screen.GetContent(x, y)
	foreground, background, attributes := style.Decompose()
	if attributes&tcell.AttrReverse != 0 {
		foreground, background = background, foreground
	}
	switch main {
	case '▀':
		return foreground, background
	case '▄':
		return background, foreground
	case '█':
		return foreground, foreground
	default:
		return background, background
	}
}
