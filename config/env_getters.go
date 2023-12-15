package config

import (
	"log"
	"os"
	"strconv"
	"strings"
)

const (
	AppEnvBase = "PDW_"
)

func GetIntEnv(env string, defaultValue int) int {
	envValue, found := os.LookupEnv(AppEnvBase + env)
	if !found {
		return defaultValue
	} else {
		value, err := strconv.ParseInt(strings.Trim(envValue, "\r"), 10, 32)
		if err != nil {
			log.Printf("warning:%s not parsed - error: %s\n", env, err)
			return defaultValue
		} else {
			return int(value)
		}
	}
}

func GetBoolEnv(env string, defaultValue bool) bool {
	envValue, found := os.LookupEnv(AppEnvBase + env)
	if !found {
		return defaultValue
	} else {
		value, err := strconv.ParseBool(strings.Trim(envValue, "\r"))
		if err != nil {
			log.Printf("warning:%s not parsed - error: %s\n", env, err)
			return defaultValue
		} else {
			return value
		}
	}
}

func GetStringEnv(env string, defaultValue string) string {
	envValue, found := os.LookupEnv(AppEnvBase + env)
	if !found {
		return defaultValue
	} else {
		return strings.ToLower(strings.Trim(envValue, "\r"))
	}
}

func GetFloatEnv(env string, defaultValue float64) float64 {
	envValue, found := os.LookupEnv(AppEnvBase + env)
	if !found {
		return defaultValue
	} else {
		value, err := strconv.ParseFloat(envValue, 64)
		if err != nil {
			return defaultValue
		}
		return value
	}
}
