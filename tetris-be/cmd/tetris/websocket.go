package main

import "net/http"

const (
	// Maximum message size allowed from peer.
	readLimit = 1024 * 10
)

func serveWs(w http.ResponseWriter, r *http.Request) {

}
