package main

import "log"

func main() {
	cfg := parseFlags()
	s := NewServer(cfg)
	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
