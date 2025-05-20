package config

import (
	"fmt"
	"path/filepath"
)

type NoFileToReadError struct {
	CandidateConfigPaths []string
}

func (e *NoFileToReadError) Error() string {
	return fmt.Sprintf("could not found config file to read from candidates: %s", e.CandidateConfigPaths)
}

func (e *NoFileToReadError) Unwrap() error {
	return nil
}

type NoFileToWriteError struct {
}

func (e *NoFileToWriteError) Error() string {
	return "could not found config file to write"
}

func (e *NoFileToWriteError) Unwrap() error {
	return nil
}

type UnsupportedConfigFileError struct {
	Path string
}

func (e *UnsupportedConfigFileError) Error() string {
	return fmt.Sprintf("config file path %s has unsupported extension: %s", e.Path, filepath.Ext(e.Path))
}

func (e *UnsupportedConfigFileError) Unwrap() error {
	return nil
}
