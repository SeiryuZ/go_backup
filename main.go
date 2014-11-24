package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"
)

const KEEP_FILE_COUNT = 2
const DB_NAME = "momo_cuppy"

func main() {
	now := time.Now().Format("02-01-2006_15:04:05")
	finalName := DB_NAME + now

	// Execute the dump utility
	backupCommand := fmt.Sprintf("pg_dump -d %s -f %s.sql", DB_NAME, finalName)
	fmt.Println(backupCommand)

	err := exec.Command("sh", "-c", backupCommand).Run()
	if err != nil {
		fmt.Println(err.Error())
		log.Fatal(err)
	}

	// execute the compress
	compressCommand := fmt.Sprintf("7z a -t7z -m0=lzma -mx=9 -mfb=64 -md=32m -ms=on %s.7z %s.sql", finalName, finalName)
	err = exec.Command("sh", "-c", compressCommand).Run()
	if err != nil {
		log.Fatal(err)
	}

	// Upload the dump file
	uploadCommand := fmt.Sprintf("dropbox_uploader.sh upload %s.7z /", finalName)
	err = exec.Command("sh", "-c", uploadCommand).Run()
	if err != nil {
		log.Fatal(err)
	}

	// Delete the unused file
	deleteCommand := fmt.Sprintf("rm %s.sql %s.7z", finalName, finalName)
	err = exec.Command("sh", "-c", deleteCommand).Run()
	if err != nil {
		log.Fatal(err)
	}

	cleanupOldFiles()
}

func cleanupOldFiles() {
	out, err := exec.Command("sh", "-c", "dropbox_uploader.sh list").Output()
	if err != nil {
		log.Fatal(err)
	}

	listings := strings.Split(string(out[:]), "\n")
	// trim unnecessary listings
	listings = listings[1 : len(listings)-1]

	listingCount := len(listings)
	// Delete all files that are excessive
	if listingCount > KEEP_FILE_COUNT {

		for _, element := range listings[:listingCount-KEEP_FILE_COUNT] {
			// Result from listing dropbox_uploader is like this
			// [F] 10324 momo_cuppy24-11-2014_22:35:48.7z
			deletedFilename := strings.Split(element, " ")[3]
			deleteCommand := fmt.Sprintf("dropbox_uploader.sh delete %s", deletedFilename)

			err := exec.Command("sh", "-c", deleteCommand).Run()
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
