package models

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/gotd/td/tg"
	"reflect"
	"strconv"
)

type File struct {
	Location *tg.InputDocumentFileLocation
	FileSize int64
	FileName string
	MimeType string
	ID       int64
}

type HashFileStruct struct {
	FileName string
	FileSize int64
	MimeType string
	FileID   int64
}

func (f *HashFileStruct) Pack() string {
	hash := md5.New()
	val := reflect.ValueOf(*f)
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)

		var fieldValue []byte
		switch field.Kind() {
		case reflect.String:
			fieldValue = []byte(field.String())
		case reflect.Int64:
			fieldValue = []byte(strconv.FormatInt(field.Int(), 10))
		default:
			panic("unhandled default case")
		}

		hash.Write(fieldValue)
	}
	return hex.EncodeToString(hash.Sum(nil))
}
