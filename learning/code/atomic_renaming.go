package main

import (
	"fmt"
	"os"
)

func append(path string, data []byte) error {
	tmp := fmt.Sprintf("%s_%d", path, 1)
	fp, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0664)
	if err != nil {
		return err
	}
	defer func() {
		fp.Close()
		if err != nil {
			os.Remove(tmp)
		}
	}()

	_, err = fp.Write(data)
	if err != nil {
		return err
	}

	err = fp.Sync()
	if err != nil {
		return err
	}

	os.Rename(tmp, path)

	return nil
}

func main() {
	err := append("../files/createFile.text", []byte("everything changed to the new content"))
	if err != nil {
		fmt.Println(err.Error())
	}
}
