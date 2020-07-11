package models

const THIS_MODULE  = 1

type DomainError struct {
	s string
}

func (de *DomainError) Error() string {
	return de.s
}

