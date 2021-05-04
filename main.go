package main

import (
	"net/http"
	"log"
)

func main() {
	r := masterRouter()

	log.Fatal(http.ListenAndServe(":1337", r))
}

