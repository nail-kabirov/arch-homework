package app

import (
	"math"

	"github.com/pkg/errors"
)

var ErrNegativePrice = errors.New("price should be positive value")
var ErrNotRoundedPrice = errors.New("price should be multiple of 0.01")

const priceMultiplier = 100
const float64ComparisonThreshold = 1e-9

type Price interface {
	Value() float64
	RawValue() uint64
}

func PriceFromRawValue(value uint64) Price {
	return &price{value: value}
}

func PriceFromFloat(value float64) (Price, error) {
	val := math.Round(value * priceMultiplier)
	if val <= 0 {
		return nil, errors.WithStack(ErrNegativePrice)
	}
	diff := math.Abs(val - value*priceMultiplier)
	if diff > float64ComparisonThreshold {
		return nil, errors.WithStack(ErrNotRoundedPrice)
	}
	return &price{value: uint64(val)}, nil
}

type price struct {
	value uint64
}

func (p *price) Value() float64 {
	return float64(p.value) / priceMultiplier
}

func (p *price) RawValue() uint64 {
	return p.value
}
