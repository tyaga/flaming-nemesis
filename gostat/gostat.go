package main

import "fmt"
import "net/http"

func main() {
    http.HandleFunc("/stat", func(w http.ResponseWriter, r *http.Request) {
		//fmt.Fprintln(w, "Hello, world.")
    })
    http.ListenAndServe(":8090", nil)
}

