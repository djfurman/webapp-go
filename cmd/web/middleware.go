package main

import "net/http"

// SessionLoad is a middleware that loads the session and moves along in the stack
func SessionLoad(next http.Handler) http.Handler {
	return session.LoadAndSave(next)
}
