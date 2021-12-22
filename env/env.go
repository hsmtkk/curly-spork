package env

import (
	"log"
	"os"
	"strconv"
)

func RequiredString(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("%s environment variable must be defined", key)
	}
	return val
}

func RequiredInt(key string) int {
	val := RequiredString(key)
	n, err := strconv.Atoi(val)
	if err != nil {
		log.Fatalf("failed to parse as int; %s; %v", val, err)
	}
	return n
}
