// ASPEED AST2500EVB board support for tamago/arm
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Board-specific CPU initialization for the AST2500EVB (ARM1176JZS, ARMv6).
//
// Linker text base: 0x80010000 (SDRAM, goos.RamStart = 0x80000000).
//
// ARM1176JZS specifics:
//   - VBAR (Security Extensions) holds the exception vectors at RamStart;
//     address 0 is the SPI flash and must not be written
//   - GOARM=5: no exclusive monitor with the MMU off, and no VFP unit

//go:build linkcpuinit

#include "textflag.h"

TEXT cpuinit(SB),NOSPLIT|NOFRAME,$0
	// Step 1: SoC early init — unlocks SCU, enables UART5 clock.
	// No stack used.
	BL	github·com∕usbarmory∕tamago∕soc∕aspeed∕ast2500·EarlyInit(SB)

	// Step 2: Compute stack pointer from goos.RamStart.
	MOVW	runtime∕goos·RamStart(SB), R13
	MOVW	runtime∕goos·RamSize(SB), R1
	MOVW	runtime∕goos·RamStackOffset(SB), R2
	ADD	R1, R13
	SUB	R2, R13
	MOVW	R13, R3		// save for after mode switch

	// Step 3: Detect current mode.
	WORD	$0xe10f0000	// mrs r0, CPSR
	AND	$0x1f, R0, R0

	// USR mode → jump directly (rare path).
	CMP	$0x10, R0
	BL.EQ	_rt0_tamago_start(SB)

	// No HYP mode on ARM1176; skip HYP handling.
	// Switch to System mode (0xdf = SYS + I+F masked).
	WORD	$0xe321f0df		// msr CPSR_c, #0xdf

	MOVW	R3, R13			// restore stack pointer

	// Step 4: Zero BSS.
	BL	clearBSS(SB)

	B	_rt0_tamago_start(SB)

// clearBSS zeroes .bss and .noptrbss sections.
TEXT clearBSS(SB),NOSPLIT|NOFRAME,$0
	MOVW	$runtime·bss(SB), R0
	MOVW	$runtime·enoptrbss(SB), R1
	MOVW	$0, R2
bss_loop:
	CMP	R0, R1
	B.EQ	bss_done
	MOVW	R2, (R0)
	ADD	$4, R0
	B	bss_loop
bss_done:
	RET
