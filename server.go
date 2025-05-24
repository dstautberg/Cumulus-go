package main

import (
	"database/sql"
	"fmt"
	"net/http"

	_ "modernc.org/sqlite"
	"github.com/dustin/go-humanize"
)

var db *sql.DB

func statusHandler(w http.ResponseWriter, r *http.Request) {
	num := 1234567.89
	formatted := humanize.CommafWithDigits(num, 2)
	fmt.Fprintf(w, "Running\nFormatted number: %s", formatted)
}

func main() {
	var err error
	db, err = sql.Open("sqlite", "file:data.db?_pragma=journal_mode(WAL)")
	if err != nil {
		panic(fmt.Sprintf("failed to open database: %v", err))
	}
	defer db.Close()

	http.HandleFunc("/status", statusHandler)
	fmt.Println("Server listening on port 9999")
	http.ListenAndServe(":9999", nil)
}
