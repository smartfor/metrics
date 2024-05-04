package utils

import (
	"fmt"
	"log"
	"time"
)

type RetryConfig struct {
	// Количество попыток
	Attempts int
	// Функция увеличения задержки для следующей попытки
	IncrementDelayFn func(prev time.Duration) time.Duration
	// Стартовая задержка
	StartDelay time.Duration
}

type IncrementDelayFn = func(prev time.Duration) time.Duration

var DefaultRetryConfig = RetryConfig{
	Attempts:   3,
	StartDelay: time.Second,
	IncrementDelayFn: func(prev time.Duration) time.Duration {
		return prev + 2
	},
}

func RetryVoid(f func() error, cfg *RetryConfig) error {
	if cfg == nil {
		cfg = &DefaultRetryConfig
	}

	var (
		attempts   = cfg.Attempts
		startDelay = cfg.StartDelay
		incDelayFn = cfg.IncrementDelayFn
	)

	err := f()
	if err == nil {
		return nil
	}

	if attempts--; attempts > 0 {
		log.Printf("Attempt failed with error: %v. Retrying...", err)
		time.Sleep(startDelay)
		return retryVoid(f, attempts, incDelayFn(startDelay), incDelayFn)
	}
	return err
}

func Retry[T any](f func() (T, error), cfg *RetryConfig) (T, error) {
	if cfg == nil {
		cfg = &DefaultRetryConfig
	}

	var (
		attempts   = cfg.Attempts
		startDelay = cfg.StartDelay
		incDelayFn = cfg.IncrementDelayFn
	)

	value, err := f()
	if err == nil {
		return value, err
	}

	if attempts--; attempts > 0 {
		log.Printf("Attempt failed with error: %v. Retrying...", err)
		time.Sleep(startDelay)
		return retry(f, attempts, incDelayFn(startDelay), incDelayFn)
	}
	return value, err
}

func retryVoid(f func() error, attempts int, sleep time.Duration, incDelayFn IncrementDelayFn) error {
	fmt.Printf("Retrying...")

	err := f()
	if err == nil {
		fmt.Printf("Retry error Type :: %T \n", err)
		return nil
	}

	if attempts--; attempts > 0 {
		log.Printf("Attempt failed with error: %v. Retrying...", err)
		time.Sleep(sleep)
		return retryVoid(f, attempts, incDelayFn(sleep), incDelayFn)
	}
	return err
}

func retry[T any](f func() (T, error), attempts int, sleep time.Duration, incDelayFn IncrementDelayFn) (T, error) {
	fmt.Printf("Retrying...")
	value, err := f()
	if err == nil {
		fmt.Printf("Retry error Type :: %T \n", err)
		return value, err
	}

	if attempts--; attempts > 0 {
		log.Printf("Attempt failed with error: %v. Retrying...", err)
		time.Sleep(sleep)
		return retry(f, attempts, incDelayFn(sleep), incDelayFn)
	}
	return value, err
}
