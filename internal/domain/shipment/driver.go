package shipment

import "strings"

type Driver struct {
	id   string
	name string
}

func NewDriver(id, name string) (Driver, error) {
	id = strings.TrimSpace(id)
	name = strings.TrimSpace(name)

	if id == "" || name == "" {
		return Driver{}, ErrInvalidDriver
	}

	return Driver{
		id:   id,
		name: name,
	}, nil
}

func (d Driver) ID() string {
	return d.id
}

func (d Driver) Name() string {
	return d.name
}
