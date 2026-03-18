package shipment

import "strings"

type Unit struct {
	id                 string
	registrationNumber string
}

func NewUnit(id, registrationNumber string) (Unit, error) {
	id = strings.TrimSpace(id)
	registrationNumber = strings.TrimSpace(registrationNumber)

	if id == "" || registrationNumber == "" {
		return Unit{}, ErrInvalidUnit
	}

	return Unit{
		id:                 id,
		registrationNumber: registrationNumber,
	}, nil
}

func (u Unit) ID() string {
	return u.id
}

func (u Unit) RegistrationNumber() string {
	return u.registrationNumber
}
