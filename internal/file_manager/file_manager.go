package file_manager

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"text/tabwriter"
	"time"

	fm "github.com/amauribechtoldjr/msk/internal/file_crypt"
	u "github.com/amauribechtoldjr/msk/utils"
)

var p = u.Panic

const MSK_BASE_PATH = "./assets/msk_base.csv"
const MSK_OUTPUT_DATA = "./bin/msk.msk"

const CREATED_AT_LAYOUT = "2006-01-02 15:04:05.999999999 -0700 MST m=+0.000000000"

type MskCsv struct {
	Name string
	Password string
	CreatedAt time.Time
}

func mask(s string) string {
	if len(s) <= 4 {
		return "****"
	}
	return "****" + s[len(s)-4:]
}

func read(fileData []byte) map[string]MskCsv {
	reader := bytes.NewReader(fileData)
	r := csv.NewReader(reader)	

	if _, err := r.Read(); err != nil {
		return nil
	}

	content := map[string]MskCsv{}
	
	for {
		record, err := r.Read()

		if err == io.EOF {
			break
		}		
		p(err)

		t, err := time.Parse(CREATED_AT_LAYOUT, record[2])
		p(err)

		content[record[0]] =  MskCsv{
			Name: record[0], 
			Password: mask(record[1]), 
			CreatedAt: t,
		}
	}

	return content;
}

func checkPassword(content map[string]MskCsv, pName string) bool {
	_, ok := content[pName]
	return ok
}

func saveFile(reader io.Reader) error {
	destinationFile, err := os.Create(MSK_OUTPUT_DATA)
	p(err)
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, reader)
	p(err)

	return nil
}

func readMSKFile() []byte {
	file, err := os.Open(MSK_OUTPUT_DATA)
	p(err)
	defer file.Close()
	
	fileData, err := io.ReadAll(file)
	p(err)

	return fileData
}

func readAndDecryptMSKFile(mk []byte) []byte {
	fileData := readMSKFile()
	csvData := fm.Decrypt(fileData, mk)

	return csvData
}

func isMSKFileSetUp() bool {
	_, err := os.Stat(MSK_OUTPUT_DATA);

	return err == nil
}

func ListAll(mk []byte) error {
	if !isMSKFileSetUp() {
		return errors.New("MSK is not set up on this machine.")
	}

	csvData := readAndDecryptMSKFile(mk)
	content := read(csvData)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// Header
	fmt.Fprintln(w, "NAME\tPASSWORD\tCREATED AT")
	fmt.Fprintln(w, "----\t--------\t----------")

	for _, item := range content {
		fmt.Fprintf(
			w, 
			"%s\t%s\t%s\n", 
			item.Name, 
			item.Password, 
			item.CreatedAt.Format("02/01/2006 15:04:05"),
		)
	}

	w.Flush()
	
	return nil
}


func SetUpMSK(mk []byte) error {
	if isMSKFileSetUp() {
		return errors.New("MSK is already set up on this machine.")
	}
	
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
	if !isMSKFileSetUp() {
		return errors.New("MSK is not set up on this machine")
	}

	csvData := readAndDecryptMSKFile(mk)
	content := read(csvData)

	if checkPassword(content, pName) {
		return errors.New("Password already exists")
	}

	newCsvData, err := WriteNewPassword(csvData, pName, pValue)
	p(err)

	cipherFile := fm.Encrypt(newCsvData, mk)
	reader := bytes.NewReader(cipherFile)

	err = saveFile(reader)
	p(err)
	
	return nil
}

func WriteNewPassword(byteData []byte, pName, pValue string) ([]byte, error) {
	buf := bytes.Buffer{}
	buf.Write(byteData)

	writer := csv.NewWriter(&buf)

	err := writer.Write([]string{pName, pValue, time.Now().String()})
	p(err)

	writer.Flush()
	
	return buf.Bytes(), nil
}
