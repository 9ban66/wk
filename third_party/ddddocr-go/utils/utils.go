package utils

import (
	"image"

	ort "github.com/yalue/onnxruntime_go"
)

type Detection struct {
	BBox image.Rectangle
	Score float32
	Class int
	Describe string
}

func SemiOCRVerification(img image.Image, shape ort.Shape) string { return "stub" }
func AutoOCRVerification(img image.Image) (string, error) { return "stub", nil }
func AutoDetectionForCalc(img image.Image, resultNum int) ([]Detection, error) { return []Detection{}, nil }
func AutoDetectionForTencent(img image.Image, resultNum int) ([]Detection, error) { return []Detection{}, nil }
func AutoCalc(detections []Detection) (int, error) { return 0, nil }
func DDDDOcrCoreInit() {}
