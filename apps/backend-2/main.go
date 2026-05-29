package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total HTTP requests",
	}, []string{"method", "route", "status"})

	httpDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "http_request_duration_ms",
		Help: "HTTP request duration in ms",
	}, []string{"method", "route"})
)

func init() {
	prometheus.MustRegister(httpRequests, httpDuration)
}

func basefunc(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello world")
}

func connectb1(w http.ResponseWriter, r *http.Request) {
	back1_url := os.Getenv("BACKEND1_URL")
	
	resp, err := http.Get(back1_url)
	
	if err != nil {
		fmt.Fprintf(w, "Error connecting to backend 1: %v", err)
		return
	}
	defer resp.Body.Close() // Good practice to close the body
	
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(w, "Error reading response from backend 1: %v", err)
		return
	}

	fmt.Printf("Data from backend 1: %s\n", string(data))
	fmt.Fprintf(w, "Data from backend 1: %s", string(data))
}

func initMsg() {
	fmt.Println("Backend 2 is starting...")
	fmt.Println("Backend 2 is ready to accept requests.")
}

// 1. Updated logger to act as a Middleware wrapper
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: 200}

		next.ServeHTTP(rec, r)

		duration := float64(time.Since(start).Milliseconds())
		log.Printf("%s %s %d took %s", r.Method, r.URL.Path, rec.status, time.Since(start))
		httpRequests.WithLabelValues(r.Method, r.URL.Path, fmt.Sprintf("%d", rec.status)).Inc()
		httpDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
	})
}

func main() {
	initMsg()
	mux := http.NewServeMux()

	mux.HandleFunc("/", basefunc)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Backend 2 is healthy!!")
	})
	mux.HandleFunc("/test", connectb1)
	mux.HandleFunc("/backend1", connectb1)
	mux.Handle("/metrics", promhttp.Handler())

	fmt.Println("Server listening on :8080...")
	http.ListenAndServe(":8080", logger(mux))
}
