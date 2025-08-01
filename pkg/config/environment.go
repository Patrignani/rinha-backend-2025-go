package config

import (
	"log"

	"github.com/Netflix/go-env"
)

var Env Environment

func init() {
	_, err := env.UnmarshalFromEnviron(&Env)
	if err != nil {
		log.Fatal(err)
	}
}
