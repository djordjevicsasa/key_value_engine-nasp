package model

import "errors"

var (
	ErrCorruptedRecord = errors.New("zapis je ostecen")
	ErrCRCMismatch     = errors.New("CRC provera nije uspela — podaci su mozda osteceni")
	ErrKeyNotFound     = errors.New("kljuc nije pronadjen")
	ErrRateLimited     = errors.New("prekoracen broj dozvoljenih zahteva")
	ErrDeleted         = errors.New("kljuc je obrisan")
)
