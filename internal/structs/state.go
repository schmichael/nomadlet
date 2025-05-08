package structs

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

type State struct {
	NodeID     string `json:"node_id"`
	NodeSecret string `json:"node_secret"`
}

func StateLoad(path string) (*State, error) {
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &State{}, nil
		}
		return nil, err
	}
	defer f.Close()

	s := &State{}
	if err := json.NewDecoder(f).Decode(s); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *State) Store(path string) error {
	buf, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	hash := sha256.Sum256(buf)

	tmpFileName := path + fmt.Sprintf(".%x.tmp", hash)
	of := os.O_RDWR | os.O_CREATE | os.O_TRUNC
	tmpFile, err := os.OpenFile(tmpFileName, of, 0o600)
	if err != nil {
		return err
	}

	defer func() {
		tmpFile.Close()
		os.Remove(tmpFileName)
	}()

	if _, err := tmpFile.Write(buf); err != nil {
		return err
	}

	if err := tmpFile.Sync(); err != nil {
		return err
	}

	if err := tmpFile.Close(); err != nil {
		return err
	}

	return os.Rename(tmpFileName, path)
}
