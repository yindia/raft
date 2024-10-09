package raft

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

const FILE_EXTENSION = "bin"

// opens the file for file operations or creates
// the file if it doesn't exist
func openFile(dirPath string, fileName string) (*os.File, error) {
	// create directory for the file
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return nil, err
	}
	filePath := fmt.Sprintf("%s/%s.%s", dirPath, fileName, FILE_EXTENSION)

	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func readFile(dirPath string, fileName string) ([]byte, error) {
	file, err := openFile(dirPath, fileName)
	if err != nil {
		return []byte(""), err
	}
	// Move the file pointer to the beginning of the file
	if _, err = file.Seek(0, 0); err != nil {
		return []byte(""), err
	}
	// Read the file contents
	data, err := io.ReadAll(file)
	if err != nil {
		return []byte(""), err
	}
	file.Close()
	return data, nil
}

func destructureSnapshot(content []byte) (int, map[string]string) {
	// encodedCommitIndex := bytes.SplitN(content, []byte("\n"), 2)[0]
	encodedKVMap := bytes.SplitN(content, []byte("\n"), 2)[1]
	decodedContent, err := decode(content)
	if err != nil {
		log.Fatal("error while decoding commit index", err)
	}
	decodedCommitIndex := strings.Trim(decodedContent, "\n")
	decodedKVMap, err := decodeMap(encodedKVMap)
	if err != nil {
		log.Fatal("error while deserializing map:", err)
	}
	commitIndex, _ := strconv.Atoi(decodedCommitIndex)
	return commitIndex, decodedKVMap
}

func Encode(data interface{}) ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(data)
	if err != nil {
		return nil, fmt.Errorf("error encoding data: %v", err)
	}
	return buf.Bytes(), nil
}

// Decode bytes back to map[string]string using gob
func decodeMap(data []byte) (map[string]string, error) {
	var result map[string]string
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("error decoding data: %v", err)
	}
	return result, nil
}

func decode(encodedData []byte) (string, error) {
	var result string
	buf := bytes.NewBuffer(encodedData)
	decoder := gob.NewDecoder(buf)
	err := decoder.Decode(&result)
	if err != nil {
		return "", err
	}
	return result, nil
}
