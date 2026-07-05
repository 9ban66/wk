package onnxruntime_go

import (
	"fmt"
	"reflect"
)

type Shape []int

type InputOutputInfo struct {
	Name       string
	Dimensions []int64
	OrtValueType struct{ name string }
	DataType struct{ name string }
}

func (i InputOutputInfo) String() string { return i.Name }
func (i InputOutputInfo) GetName() string { return i.Name }

func (s Shape) String() string { return fmt.Sprintf("%v", []int(s)) }
func (s Shape) MarshalText() ([]byte, error) { return []byte(s.String()), nil }
func (s Shape) Get() []int { return []int(s) }
func (s Shape) Len() int { return len(s) }
func (s Shape) Index(i int) int { return []int(s)[i] }
func (s Shape) LenInt() int { return len(s) }

type Tensor[T any] struct {
	data []T
	shape Shape
}

func NewShape(dim ...int) Shape { return Shape(dim) }
func NewShapeInt64(dim ...int64) Shape { out := make(Shape, len(dim)); for i, d := range dim { out[i] = int(d) }; return out }

func IsInitialized() bool { return true }
func InitializeEnvironment() error { return nil }
func DestroyEnvironment() {}

func NewTensor[T any](shape Shape, data []T) (*Tensor[T], error) {
	return &Tensor[T]{data: data, shape: shape}, nil
}

func NewEmptyTensor[T any](shape Shape) (*Tensor[T], error) {
	return &Tensor[T]{shape: shape}, nil
}

func GetInputOutputInfo(modelPath string) ([]InputOutputInfo, []InputOutputInfo, error) {
	return []InputOutputInfo{{Name: "input"}}, []InputOutputInfo{{Name: "output", Dimensions: []int64{1}}}, nil
}

type Value interface{}

type AdvancedSession struct {
	modelPath string
}

func NewAdvancedSession(modelPath string, inputNames, outputNames []string, inputs, outputs []Value, _ interface{}) (*AdvancedSession, error) {
	_ = inputNames
	_ = outputNames
	_ = inputs
	_ = outputs
	return &AdvancedSession{modelPath: modelPath}, nil
}

func (s *AdvancedSession) Run() error { return nil }
func (s *AdvancedSession) Destroy() {}

type SessionOptions struct{}

type Session struct{}

func NewSession(modelPath string, options *SessionOptions) (*Session, error) { return &Session{}, fmt.Errorf("onnxruntime stub not implemented") }
func (s *Session) Run(inputs map[string][]float32) ([]float32, error) { return nil, fmt.Errorf("onnxruntime stub not implemented") }

func (t *Tensor[T]) Destroy() {}
func (t *Tensor[T]) GetData() []T { return t.data }
func (t *Tensor[T]) GetShape() Shape { return t.shape }

func (t *Tensor[T]) GetDataAsFloat32() []float32 { return reflect.ValueOf(t.data).Convert(reflect.TypeOf([]float32{})).Interface().([]float32) }
