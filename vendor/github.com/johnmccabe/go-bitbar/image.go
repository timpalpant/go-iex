package bitbar

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/png"
)

func toBase64(img image.Image) string {
	var buf bytes.Buffer
	png.Encode(&buf, img)
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}
