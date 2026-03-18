package gui

import (
	_ "embed"
)

//go:embed qzone.png
var qzone []byte

func QzoneICON() []byte {
	return qzone
}

//go:embed qmail.png
var mail []byte

func MAILICON() []byte {
	return mail
}

//go:embed qun.png
var qun []byte

func QUNICON() []byte {
	return qun
}

//go:embed logo.png
var logo []byte

func LOGO() []byte {
	return logo
}
