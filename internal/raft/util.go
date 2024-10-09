package raft

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"strconv"
	"strings"
)

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
