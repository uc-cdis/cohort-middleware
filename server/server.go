package server

import (
	"log"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func Init() {
	tracer.Start()
	defer tracer.Stop()

	r := NewRouter()
	if err := r.Run(); err != nil {
		log.Printf("unhandled server error:\n%s", err.Error())
	}
}
