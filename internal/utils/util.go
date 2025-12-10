package utils

import (
	"math/rand"
	"time"
)

func RandString(l int) string {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	letras := "abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	resultado := make([]byte, l)

	for i := range resultado {
		resultado[i] = letras[rng.Intn(len(letras))]
	}

	return string(resultado)
}
