package cfgutils

import (
	"flag"
	"os"
	"strconv"
)

// -------------------------- COMPOSE PARSING WITHOUT VALIDATOR -------------------------- //

func ParseString(
	flagName string,
	envName string,
	description string,
	target *string,
) {
	flag.StringVar(target, flagName, *target, description)
	flag.Parse()
	TryTakeStringFromEnv(envName, target)
}

func ParseInt(
	flagName string,
	envName string,
	description string,
	target *int,
) {
	flag.IntVar(target, flagName, *target, description)
	flag.Parse()
	TryTakeIntFromEnv(envName, target)
}

func ParseBool(
	flagName string,
	envName string,
	description string,
	target *bool,
) {
	flag.BoolVar(target, flagName, *target, description)
	flag.Parse()
	TryGetBoolFromEnv(envName, target)
}

// --------------------------  COMPOSE PARSING WITH VALIDATOR -------------------------- //

func ParseStringWithValidator(
	flagName string,
	envName string,
	description string,
	target *string,
	validator func(v string) error,
) error {
	ParseString(flagName, envName, description, target)
	return validator(*target)
}

func ParseIntWithValidator(
	flagName string,
	envName string,
	description string,
	target *int,
	validator func(v int) error,
) error {
	ParseInt(flagName, envName, description, target)
	return validator(*target)
}

func ParseBoolWithValidator(
	flagName string,
	envName string,
	description string,
	target *bool,
	validator func(v bool) error,
) error {
	ParseBool(flagName, envName, description, target)
	return validator(*target)
}

// -------------------------- FROM ENV -------------------------- //

func TryTakeStringFromEnv(name string, target *string) {
	if fromEnv := os.Getenv(name); fromEnv != "" {
		*target = fromEnv
	}
}

func TryTakeIntFromEnv(name string, target *int) {
	if fromEnv := os.Getenv(name); fromEnv != "" {
		if v, err := strconv.Atoi(fromEnv); err == nil {
			*target = v
		}
	}
}

func TryGetBoolFromEnv(name string, target *bool) {
	if fromEnv := os.Getenv(name); fromEnv != "" {
		if v, err := strconv.ParseBool(fromEnv); err == nil {
			*target = v
		}
	}
}
