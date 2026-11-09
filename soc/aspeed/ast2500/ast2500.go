// ASPEED AST2500 SoC support for tamago/arm
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package ast2500 provides support to Go bare metal unikernels written using
// the TamaGo framework on the ASPEED AST2500 SoC (ARM1176JZS, ARMv6).
//
// The package implements initialization and drivers for the AST2500 SoC,
// adopting the following reference specifications:
//   - AST2500/AST2520 A2 Datasheet v1.8 (Jul 2020)
//
// A loader initializes SDRAM, which starts at goos.RamStart = 0x80000000,
// and enters the image there; the linker text base is 0x80010000. The
// ARM1176JZS implements the Security Extensions and with them VBAR, which
// arm.CPU.Init points at RamStart: address 0 stays the SPI flash.
//
// Build with:
//
//	GOOS=tamago GOARCH=arm GOARM=5
//
// With the MMU off, ARMv6 LDREX/STREX need a global exclusive monitor, which
// the AST2500 lacks: a GOARM=6 image hangs in its first atomic operation.
// GOARM=5 selects single-processor atomics and the soft-float ABI; the
// ARM1176JZS has no VFP unit either.
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package ast2500

import (
	"github.com/usbarmory/tamago/arm"
	"github.com/usbarmory/tamago/soc/aspeed/intc"
	"github.com/usbarmory/tamago/soc/aspeed/uart"
)

// SiliconRevision holds the hardware revision read from SCU7C bits[23:16].
// Populated during init().
var SiliconRevision uint32

// ARM processor instance.
var ARM = &arm.CPU{}

// VIC is the Vectored Interrupt Controller.
var VIC = &intc.INTC{
	Base: VIC_BASE,
}

// SCU is the System Control Unit peripheral accessor (chiptool-generated type).
var SCU = &Scu{Base: uintptr(SCU_BASE)}

// UART5 is the primary console UART (16550-compatible, 115200-8N1).
var UART5 = &uart.UART{
	Index: 5,
	Base:  UART5_BASE,
}

// TIMER is the Timer Controller peripheral accessor (chiptool-generated type).
// Uses the chiptool-generated Timer struct (Base uintptr) for nanotime since
// nanotime is called before Go world start and must avoid calling-convention
// issues with imported-struct methods.
var TIMER = &Timer{Base: uintptr(TIMER_BASE)}
