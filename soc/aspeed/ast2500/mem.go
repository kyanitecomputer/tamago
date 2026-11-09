// ASPEED AST2500 SoC support for tamago/arm
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !linkramstart

package ast2500

import _ "unsafe"

// SDRAM base for AST2500. The exception vector table, which VBAR points
// at, and the exception stack live at its start.
//
//go:linkname ramStart runtime/goos.RamStart
var ramStart uint32 = 0x80000000

//go:linkname ramStackOffset runtime/goos.RamStackOffset
var ramStackOffset uint32 = 0x100000 // 1 MB
