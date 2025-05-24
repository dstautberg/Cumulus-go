package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"

	_ "modernc.org/sqlite"
	"github.com/dustin/go-humanize"
)

var db *sql.DB

func statusHandler(w http.ResponseWriter, r *http.Request) {
	num := 1234567.89
	formatted := humanize.CommafWithDigits(num, 2)
	fmt.Fprintf(w, "Running\nFormatted number: %s", formatted)
}

func getDiskLabels() ([][2]string, error) {
	cmd := exec.Command("powershell", "-Command", "Get-Volume | Select-Object FileSystemLabel,DriveLetter | ConvertTo-Json")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var volumes []struct {
		FileSystemLabel string `json:"FileSystemLabel"`
		DriveLetter    string `json:"DriveLetter"`
	}
	err = json.Unmarshal(output, &volumes)
	if err != nil {
		return nil, err
	}
	var result [][2]string
	for _, v := range volumes {
		label := strings.TrimSpace(v.FileSystemLabel)
		drive := strings.TrimSpace(v.DriveLetter)
		if label != "" || drive != "" {
			result = append(result, [2]string{label, drive})
		}
	}
	return result, nil
}

func saveDiskLabelsToDB(db *sql.DB, labels [][2]string) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS disks (label TEXT, drive TEXT)`)
	if err != nil {
		return err
	}
	_, err = db.Exec(`DELETE FROM disks`)
	if err != nil {
		return err
	}
	for _, pair := range labels {
		_, err := db.Exec(`INSERT INTO disks (label, drive) VALUES (?, ?)`, pair[0], pair[1])
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	var err error
	db, err = sql.Open("sqlite", "file:data.db?_pragma=journal_mode(WAL)")
	if err != nil {
		panic(fmt.Sprintf("failed to open database: %v", err))
	}
	defer db.Close()

	labels, err := getDiskLabels()
	if err != nil {
		fmt.Printf("Failed to get disk labels: %v\n", err)
	} else {
		fmt.Printf("Attached disk labels: %v\n", labels)
		err = saveDiskLabelsToDB(db, labels)
		if err != nil {
			fmt.Printf("Failed to save disk labels to DB: %v\n", err)
		}
	}

	http.HandleFunc("/status", statusHandler)
	fmt.Println("Server listening on port 9999")
	http.ListenAndServe(":9999", nil)
}
