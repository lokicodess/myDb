package main

import (
	"fmt"
	"os"
)

func solution() {
	err := createFile("../files/createFile.text", []byte("I am writting to the file for the first time and would it work ? "))
	if err != nil {
		fmt.Printf("%#v", err.Error())
	}
}

func createFile(path string, data []byte) error {
	fp, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0664)
	if err != nil {
		return err
	}
	defer fp.Close()
	_, err = fp.Write(data)
	if err != nil {
		return err
	}
	// fsync sys call for Durability
	return fp.Sync()
}
