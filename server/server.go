package server

import "log"

func Init() {
	r := NewRouter()
	if err := r.Run(); err != nil {
		log.Printf("unhandled server error:\n%s", err.Error())
	}
}
