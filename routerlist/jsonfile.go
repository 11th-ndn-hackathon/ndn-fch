package routerlist

import (
	"encoding/json"
	"errors"
	"io"
	"os"
)

func loadJSONFile(filename string, ptr interface{}) error {
	if filename == "" {
		return errors.New("no filename")
	}

	file, e := os.Open(filename)
	if e != nil {
		return e
	}
	defer file.Close()

	body, e := io.ReadAll(file)
	if e != nil {
		return e
	}

	if e := json.Unmarshal(body, ptr); e != nil {
		return e
	}

	return nil
}

func saveJSONFile(filename string, obj interface{}) error {
	if filename == "" {
		return errors.New("no filename")
	}

	j, e := json.Marshal(obj)
	if e != nil {
		return e
	}

	f, e := os.Create(filename)
	if e != nil {
		return e
	}
	defer f.Close()

	_, e = f.Write(j)
	return e
}
