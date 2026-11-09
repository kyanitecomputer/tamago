// Gigabyte MJ11-EC1 BMC support for tamago/arm
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !linkprintk

package mj11ec1bmc

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/soc/aspeed/ast2500"
)

// printk routes runtime console output to UART5, the BMC debug console,
// with a carriage return before each line feed.
//
//go:linkname printk runtime/goos.Printk
func printk(c byte) {
	if c == '\n' {
		ast2500.UART5.Tx('\r')
	}

	ast2500.UART5.Tx(c)
}
