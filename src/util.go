package src

import "os"

func GetClientName() string {
	val, exists := os.LookupEnv("NAME")
	if !exists {
		val = "default"
	}

	return val
}
