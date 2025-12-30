package file_manager

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	fm "github.com/amauribechtoldjr/msk/internal/file_crypt"
	u "github.com/amauribechtoldjr/msk/utils"
)

var c = u.Check

const MSK_BASE_PATH = "./assets/msk_base.csv"
const MSK_OUTPUT_DATA = "./bin/msk.txt"

func read(fileContent string) {
	reader := strings.NewReader(fileContent)

	r := csv.NewReader(reader)

	for {
		record, err := r.Read()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(record[1])
		fmt.Println(record[0])
	}
}

func saveFile(reader io.Reader) error {
	if _, err := os.Stat(MSK_OUTPUT_DATA); err == nil {
		fmt.Printf("You already have configured MSK on this machine.")
		return nil
	}

	destinationFile, err := os.Create(MSK_OUTPUT_DATA)
	c(err)
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, reader)
	c(err)

	return nil
}

func InitMSKConfig(masterKey []byte) error {
	file, err := os.Open(MSK_BASE_PATH)
	c(err)
	
	fileData, err := io.ReadAll(file)
	c(err)

	cipherFile := fm.Encrypt(fileData, masterKey)
	reader := bytes.NewReader(cipherFile)

	err = saveFile(reader)
	c(err)

	fmt.Println("MSK initialized successfully.")

	return nil
}