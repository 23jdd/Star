package widgets

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
)

// NewImageFromPath decodes a PNG, JPEG, or GIF file and creates an Image.
// GIF files render their first frame. The file is closed before this function
// returns, so the widget does not retain file-system resources.
func NewImageFromPath(path string) (*Image, error) {
	source, err := decodeImagePath(path)
	if err != nil {
		return nil, err
	}
	return NewImage(source), nil
}

// SetImagePath decodes a PNG, JPEG, or GIF file and replaces the current
// source. If decoding fails, the existing source remains unchanged.
func (i *Image) SetImagePath(path string) error {
	source, err := decodeImagePath(path)
	if err != nil {
		return err
	}
	i.SetImage(source)
	return nil
}

func decodeImagePath(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("widgets: open image %q: %w", path, err)
	}
	defer file.Close()

	source, format, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("widgets: decode image %q: %w", path, err)
	}
	if format == "" {
		return nil, fmt.Errorf("widgets: decode image %q: unknown format", path)
	}
	return source, nil
}
