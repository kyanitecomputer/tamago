// Gigabyte MJ11-EC1 BMC support for tamago/arm
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package mj11ec1bmc provides hardware initialization for the ASPEED AST2500
// baseboard management controller (ARM1176JZ-S) of the Gigabyte MJ11-EC1.
//
// A loader, such as fstart, initializes the DRAM, configures the UART5
// console (115200 8N1) and enters the ELF in System or Supervisor mode with
// the MMU off. Build with:
//
//	GOOS=tamago GOOSPKG=github.com/usbarmory/tamago GOARCH=arm GOARM=5 \
//	  ${TAMAGO} build -ldflags "-T 0x80010000 -R 0x1000" main.go
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package mj11ec1bmc

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/soc/aspeed/ast2500"
)

// Peripheral instances re-exported for application convenience.
var (
	ARM   = ast2500.ARM
	UART5 = ast2500.UART5
	VIC   = ast2500.VIC
)

// Init takes care of the lower level SoC initialization triggered early in
// runtime setup (post Go-World start). The console keeps the loader's
// configuration.
//
//go:linkname Init runtime/goos.Hwinit1
func Init() {
	ast2500.Init()
}
