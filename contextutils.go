package main

import "net/http"

type contextKey string

const (
	Retry    contextKey = "Retry"
	Attempts contextKey = "Attempts"
)

func GetRetryFromContext(r *http.Request) int {
	if retries, ok := r.Context().Value(Retry).(int); ok {
		return retries
	}
	return 0
}

func GetAttemptsFromContext(r *http.Request) int {
	if attempts, ok := r.Context().Value(Attempts).(int); ok {
		return attempts
	}
	return 0
}
