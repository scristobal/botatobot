package botatobot

import (
	"fmt"
	"net/http"
	"time"
)

func Http(handler func(w http.ResponseWriter, r *http.Request)) {

	s := http.Server{
		Addr:         fmt.Sprintf(":%s", PORT),
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

func Start_health() {
	Http(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
}
