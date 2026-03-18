package shipment

type Money struct {
	minorUnits int64
	valid      bool
}

func NewMoney(minorUnits int64) (Money, error) {
	if minorUnits < 0 {
		return Money{}, ErrInvalidMoney
	}

	return Money{
		minorUnits: minorUnits,
		valid:      true,
	}, nil
}

func (m Money) MinorUnits() int64 {
	return m.minorUnits
}

func (m Money) IsValid() bool {
	return m.valid
}

func (m Money) GreaterThan(other Money) bool {
	return m.minorUnits > other.minorUnits
}
