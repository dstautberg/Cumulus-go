// Here is how to create a basic HTTP server in Go:
// Import the net/http package.
// This package provides the necessary functions for creating HTTP servers and handling requests.
// Create a handler function.
// This function will be called for each incoming request. It takes an http.ResponseWriter and an http.Request as arguments. The http.ResponseWriter is used to write the response, and the http.Request contains information about the request.
// Register the handler function with a route.
// Use the http.HandleFunc function to register the handler function for a specific route. The first argument is the route path, and the second argument is the handler function.
// Start the server.
// Use the http.ListenAndServe function to start the server. The first argument is the address to listen on (e.g., ":8080"), and the second argument is the handler to use (usually nil to use the default handler).
// Below is an example:

// package main

// import (
// 	"fmt"
// 	"net/http"
// )

// func helloHandler(w http.ResponseWriter, r *http.Request) {
// 	fmt.Fprintf(w, "Hello, World!")
// }

// func main() {
// 	http.HandleFunc("/hello", helloHandler)
// 	fmt.Println("Server listening on port 8080")
// 	http.ListenAndServe(":8080", nil)
// }

package server

import (
	"fmt"
	"net/http"
)

func statusHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Running")
}

func run() {
	http.HandleFunc("/status", statusHandler)
	fmt.Println("Server listening on port 9999")
	http.ListenAndServe(":9999", nil)
}
