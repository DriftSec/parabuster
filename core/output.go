package core

import (
	"fmt"
)

var DebugOn bool

const (
	Reset        = "\033[0m"
	InfoColor    = "\033[1;34m"
	NoticeColor  = "\033[1;36m"
	WarningColor = "\033[1;33m"
	ErrorColor   = "\033[1;31m"
	DebugColor   = "\033[0;36m"
	SuccessColor = "\033[1;32m"
)

func Dprint(a ...interface{}) {
	if DebugOn {
		aa := make([]interface{}, 0, 2+len(a))
		aa = append(aa, NoticeColor)
		aa = append(append(aa, "[DEBUG]"), a...)
		aa = append(aa, Reset)
		// fmt.Println()
		fmt.Println(aa...)
	}
}

func Eprint(a ...interface{}) {
	aa := make([]interface{}, 0, 2+len(a))
	aa = append(aa, ErrorColor)
	aa = append(append(aa, "[ERROR]"), a...)
	aa = append(aa, Reset)
	// fmt.Println()
	fmt.Println(aa...)
}

func Wprint(a ...interface{}) {
	aa := make([]interface{}, 0, 2+len(a))
	aa = append(aa, WarningColor)
	aa = append(append(aa, "[!]"), a...)
	aa = append(aa, Reset)
	fmt.Println(aa...)
}

func Sprint(a ...interface{}) {
	aa := make([]interface{}, 0, 2+len(a))
	aa = append(aa, SuccessColor)
	aa = append(append(aa, "[+]"), a...)
	aa = append(aa, Reset)
	fmt.Println(aa...)
}

func Nprint(a ...interface{}) {
	aa := make([]interface{}, 0, 2+len(a))
	aa = append(aa, WarningColor)
	aa = append(append(aa, "[!]"), a...)
	aa = append(aa, Reset)
	fmt.Println(aa...)
}

func Fprint(a ...interface{}) {
	aa := make([]interface{}, 0, 2+len(a))
	aa = append(aa, ErrorColor)
	aa = append(append(aa, "[-]"), a...)
	aa = append(aa, Reset)
	fmt.Println(aa...)
}

func Iprint(a ...interface{}) {
	aa := make([]interface{}, 0, 2+len(a))
	aa = append(aa, NoticeColor)
	aa = append(append(aa, "[!]"), a...)
	aa = append(aa, Reset)
	fmt.Println(aa...)
}
