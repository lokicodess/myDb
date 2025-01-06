package main

import (
	"fmt"
	"os"
)

func main() {
	err := createFile("../files/createFile.text", []byte("I am writting to the file for the first time"))
	if err != nil {
		fmt.Printf("%#v", err.Error())
	}
}

func createFile(path string, data []byte) error {
	fp, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0664)
	if err != nil {
		return err
	}
	_, err = fp.Write(data)
	if err != nil {
		return err
	}
	// fsync sys call for Durability
	return fp.Sync()
}
