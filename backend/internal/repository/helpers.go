package repository

import (
	"encoding/base64"
	"unicode/utf8"
)

func base64Std(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	return base64.StdEncoding.EncodeToString(b)
}

func isValidUTF8(b []byte) bool {
	return utf8.Valid(b)
}
