package utils

import "image"

func ResizeImage(inputImage image.Image, newWidth, newHeight int) image.Image {
	bounds := inputImage.Bounds()
	inputWidth := bounds.Dx()
	inputHeight := bounds.Dy()

	outputImage := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			sx := x * inputWidth / newWidth
			sy := y * inputHeight / newHeight
			outputImage.Set(x, y, inputImage.At(sx, sy))
		}
	}

	return outputImage
}
