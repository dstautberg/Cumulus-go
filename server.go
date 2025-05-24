package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"

	_ "modernc.org/sqlite"
	"github.com/dustin/go-humanize"
	"github.com/joho/godotenv"
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

// Add this function to parse BACKUP env variables
func parseBackupEnvVars() [][2]string {
	var result [][2]string
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "BACKUP") {
			fmt.Printf("Found BACKUP env: %s\n", env) // Print out each env found
			parts := strings.SplitN(env, "=", 2)
			if len(parts) != 2 {
				continue
			}
			value := parts[1]
			vparts := strings.SplitN(value, "/", 2)
			label := strings.TrimSpace(vparts[0])
			path := ""
			if len(vparts) > 1 {
				path = strings.TrimSpace(vparts[1])
			}
			result = append(result, [2]string{label, path})
		}
	}
	return result
}

func ensureBackupPathsExist(diskInfo [][2]string, backupInfo [][2]string) {
	fmt.Printf("ensureBackupPathsExist called with diskInfo: %v, backupInfo: %v\n", diskInfo, backupInfo) // Print method name and parameters at line 90
	for _, backup := range backupInfo {
		fmt.Printf("Processing backup: %v\n", backup) // Print out backup at line 91
		backupLabel := backup[0]
		backupPath := backup[1]
		for _, disk := range diskInfo {
			fmt.Printf("Checking disk: %v\n", disk) // Print out disk at line 95
			diskLabel := disk[0]
			driveLetter := disk[1]
			if diskLabel == backupLabel && driveLetter != "" && backupPath != "" {
				fullPath := driveLetter + ":\\" + backupPath
				if _, err := os.Stat(fullPath); os.IsNotExist(err) {
					err := os.MkdirAll(fullPath, 0755)
					if err != nil {
						fmt.Printf("Failed to create path %s: %v\n", fullPath, err)
					} else {
						fmt.Printf("Created path: %s\n", fullPath)
					}
				} else {
					fmt.Printf("Path exists: %s\n", fullPath)
					// Print out each path that was verified to exist
					fmt.Println(fullPath)
				}
			}
		}
	}
}

func main() {
	_ = godotenv.Load() // Load environment variables from .env if present
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

	backupInfo := parseBackupEnvVars()
	ensureBackupPathsExist(labels, backupInfo)

	http.HandleFunc("/status", statusHandler)
	fmt.Println("Server listening on port 9999")
	http.ListenAndServe(":9999", nil)
}
