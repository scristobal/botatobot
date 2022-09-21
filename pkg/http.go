package botatobot

import (
	"net/http"
	"time"
)

func Http(handler func(w http.ResponseWriter, r *http.Request)) {

	s := http.Server{
		Addr:         ":8080",
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 90 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", handler)

	s.Handler = mux

	err := s.ListenAndServe()

	if err != nil {
		if err != http.ErrServerClosed {
			panic(err)
		}
	}
}
