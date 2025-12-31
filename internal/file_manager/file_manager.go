package file_manager

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"

	fm "github.com/amauribechtoldjr/msk/internal/file_crypt"
	u "github.com/amauribechtoldjr/msk/utils"
)

var p = u.Panic

const MSK_BASE_PATH = "./assets/msk_base.csv"
const MSK_OUTPUT_DATA = "./bin/msk.txt"

// func read(fileContent string) {
// 	reader := strings.NewReader(fileContent)

// 	r := csv.NewReader(reader)

// 	for {
// 		record, err := r.Read()

// 		if err == io.EOF {
// 			break
// 		}

// 		if err != nil {
// 			log.Fatal(err)
// 		}

// 		fmt.Println(record[1])
// 		fmt.Println(record[0])
// 	}
// }

func saveFile(reader io.Reader) error {
	destinationFile, err := os.Create(MSK_OUTPUT_DATA)
	p(err)
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, reader)
	p(err)

	return nil
}

func SetUpMSK(mk []byte) error {
	if _, err := os.Stat(MSK_OUTPUT_DATA); err == nil {
		return errors.New("MSK is already set up on this machine.")
	}

	fmt.Printf("key to setup: %s", mk)
	
	file, err := os.Open(MSK_BASE_PATH)
	p(err)
	defer file.Close()
	
	fileData, err := io.ReadAll(file)
	p(err)

	cipherFile := fm.Encrypt(fileData, mk)
	reader := bytes.NewReader(cipherFile)

	err = saveFile(reader)
	p(err)

	return nil
}

func DeletePassword(mk []byte, pName string) error {
	return nil
}

func AddPassword(mk []byte, pName, pValue string)  error {
	file, err := os.Open(MSK_OUTPUT_DATA)
	p(err)
	defer file.Close()
	
	fileData, err := io.ReadAll(file)
	p(err)

	csvData := fm.Decrypt(fileData, mk)
	fmt.Println(string(csvData))
	
	return nil
}