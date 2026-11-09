// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package exception

import (
	"runtime"
	"unsafe"
)

// Print prints the source location of pc.
func Print(pc uintptr) {
	fn := runtime.FuncForPC(pc)

	if fn == nil {
		print("\tpc=")
		printHex(pc)
		print("\n")
		return
	}

	file, line := fn.FileLine(pc)
	print("\t", file, ":", line, "\n")
}

// Throw prints the source location of pc and panics.
func Throw(pc uintptr) {
	Print(pc)
	panic("unhandled exception")
}

// printHex prints v in hexadecimal without allocating.
func printHex(v uintptr) {
	const digits = "0123456789abcdef"

	print("0x")

	for shift := int(8*unsafe.Sizeof(v)) - 4; shift >= 0; shift -= 4 {
		n := (v >> shift) & 0xf
		print(digits[n : n+1])
	}
}
