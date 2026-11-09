// ASPEED AST2500EVB board support for tamago/arm
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !linkramsize

package ast2500evb

import _ "unsafe"

// The AST2500EVB has 512 MB DDR3 SDRAM at 0x80000000.

//go:linkname ramSize runtime/goos.RamSize
var ramSize uint32 = 0x20000000 // 512 MB
