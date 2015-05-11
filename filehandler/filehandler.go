package filehandler

import (
	"github.com/henkburgstra/spoor"
	"os"
)

type FileHandler struct {
	spoor.StreamHandler
	filename string
	mode     string
}

func NewFileHandler(filename string, mode string) *FileHandler {
	fileHandler := new(FileHandler)
	fileHandler.filename = filename
	fileHandler.mode = mode

	logfile, _ := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.FileMode(0666))
	fileHandler.StreamHandler = *spoor.NewStreamHandler(logfile)

	return fileHandler
}
