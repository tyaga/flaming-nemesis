package main

import "fmt"
import "log"
import "net/http"

func main() {
    http.HandleFunc("/stat", func(w http.ResponseWriter, r *http.Request) {
        log.Printf("[gostat] Got stat: %s", r.URL.RawQuery)
        fmt.Println(r.URL.ParseQuery());

		fmt.Fprint(w, "stat got stat")
    })
    http.ListenAndServe(":8090", nil)
}

