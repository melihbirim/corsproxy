package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true,"items":[1,2,3,4,5,6,7,8,9,10],"msg":"hello from upstream"}`))
	})
	http.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"method": r.Method, "headers": r.Header})
	})
	http.HandleFunc("/redir", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "http://169.254.169.254/latest/meta-data/", http.StatusFound)
	})
	http.HandleFunc("/big", func(w http.ResponseWriter, r *http.Request) {
		mb, _ := strconv.Atoi(r.URL.Query().Get("mb"))
		chunk := []byte(strings.Repeat("x", 1024*1024))
		for i := 0; i < mb; i++ {
			w.Write(chunk)
		}
	})
	http.ListenAndServe(":9000", nil)
}
