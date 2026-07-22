// ASPEED AST2700 DCSCM board support for tamago/arm64
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Board-specific CPU initialization for the AST2750-A1 DCSCM DDR4 board.
//
// The BootMCU releases the CA35 at EL3. This code:
//   1. Parks secondary CPUs (WFE loop).
//   2. Configures EL3 system registers (FP/SIMD, timers, GICv3 SRE).
//   3. Configures EL2 to pass timer/FP access down to EL1.
//   4. Drops CPU0 to Secure EL1h via ERET.
//   5. Sets the stack from runtime symbols, zeroes BSS, and enters
//      _rt0_tamago_start.

//go:build linkcpuinit

#include "textflag.h"

TEXT cpuinit(SB),NOSPLIT|NOFRAME,$0
	// Park secondary CPUs.
	MRS	MPIDR_EL1, R0
	AND	$0xff, R0, R0
	CMP	$0, R0
	BNE	cpu_spin

	// Enable SMP coherency: CPUECTLR_EL1.SMPEN (S3_1_C15_C2_1, bit 6), set
	// before the caches/MMU are enabled (later, in the arm64 runtime InitMMU).
	// The Cortex-A35 does not participate in the Inner Shareable coherency
	// domain until SMPEN is set. The MMU maps DRAM as Inner Shareable Normal
	// memory, so the LL/SC exclusive monitors (LDXR/STXR) used by the Go
	// runtime's atomics require it. Matches ATF cortex_a35.S.
	WORD	$0xD539F220	// mrs x0, S3_1_c15_c2_1 (CPUECTLR_EL1)
	ORR	$1<<6, R0	// set SMPEN
	WORD	$0xD519F220	// msr S3_1_c15_c2_1, x0
	ISB	$0b1111

	// ── EL3 system register setup ─────────────────────────────────────

	// CPTR_EL3: clear TFP (bit 10) to allow FP/SIMD at all ELs.
	MOVD	$0, R0
	WORD	$0xd51e1140	// msr CPTR_EL3, x0

	// CNTFRQ_EL0: counter frequency (1.125 GHz on AST2700).
	MOVD	$1125000000, R0
	MSR	R0, CNTFRQ_EL0

	// Enable the system counter (CNTCR.EN at 0x12C10000).
	MOVD	$0x12c10000, R3
	MOVD	$1, R4
	MOVW	R4, (R3)

	// SCR_EL3: RW=1 (AArch64 at EL2/EL1), HCE=1, RES1 bits 4,5.
	// NS=0: stay secure so EL1 can access IO-die MMIO without TZASC.
	MOVD	$0, R0
	ORR	$1<<10, R0	// RW
	ORR	$1<<8, R0	// HCE
	ORR	$1<<5, R0	// RES1
	ORR	$1<<4, R0	// RES1
	WORD	$0xd51e1100	// msr SCR_EL3, x0

	// ── EL2 passthrough setup ─────────────────────────────────────────

	// HCR_EL2: RW=1 (EL1 is AArch64).
	MOVD	$1<<31, R0
	WORD	$0xd51c1100	// msr HCR_EL2, x0

	// CNTHCTL_EL2: allow EL1/EL0 physical timer and counter access.
	MOVD	$3, R0		// EL1PCEN | EL1PCTEN
	WORD	$0xd51ce100	// msr CNTHCTL_EL2, x0

	// CNTVOFF_EL2: zero virtual offset.
	MOVD	$0, R0
	WORD	$0xd51ce060	// msr CNTVOFF_EL2, x0

	// CPTR_EL2: allow FP/SIMD at EL2.
	MOVD	$0, R0
	WORD	$0xd51c1140	// msr CPTR_EL2, x0

	// ICC_SRE_EL3: enable GICv3 system register interface for all ELs.
	// EN (bit 3) | DIB (bit 2) | DFB (bit 1) | SRE (bit 0)
	MOVD	$0xf, R0
	WORD	$0xd51ecca0	// msr ICC_SRE_EL3, x0
	ISB	$0b1111

	// ── Drop to Secure EL1h ──────────────────────────────────────────

	// SPSR_EL3: all exceptions masked, target EL1h.
	MOVD	$0, R0
	ORR	$0b1111<<6, R0	// DAIF masked
	ORR	$0b0101<<0, R0	// EL1h (M[3:0]=0101)
	WORD	$0xd51e4000	// msr SPSR_EL3, x0

	MOVD	$·cpuinit_el1(SB), R0
	WORD	$0xd51e4020	// msr ELR_EL3, x0
	ISB	$0b1111

	ERET

cpu_spin:
	WFE
	B	cpu_spin

// cpuinit_el1 runs at Secure EL1 after the ERET from EL3.
TEXT ·cpuinit_el1(SB),NOSPLIT|NOFRAME,$0
	// CPACR_EL1: FPEN=0b11 — allow FP/SIMD at EL1/EL0.
	MRS	CPACR_EL1, R0
	ORR	$3<<20, R0
	MSR	R0, CPACR_EL1
	ISB	$0b1111

	// Disable alignment check and MMU (clean state for hwinit0/InitMMU).
	MRS	SCTLR_EL1, R0
	BIC	$1<<1, R0	// A
	BIC	$1<<0, R0	// M
	MSR	R0, SCTLR_EL1
	ISB	$0b1111

	// Set stack pointer from runtime symbols.
	MOVD	runtime∕goos·RamStart(SB), R1
	MOVD	R1, RSP
	MOVD	runtime∕goos·RamSize(SB), R1
	MOVD	runtime∕goos·RamStackOffset(SB), R2
	ADD	R1, RSP
	SUB	R2, RSP

	// Zero BSS: [runtime.bss, runtime.end). tamago's arm64 rt0 does NOT clear
	// BSS — it assumes the loader/DRAM provides zeroed memory. QEMU zero-fills
	// DRAM, so this was masked in emulation, but on real AST2700 silicon the
	// DRAM contains garbage. That garbage leaves runtime.arm64HasATOMICS (in
	// .noptrbss) nonzero, so runtime.check — which runs before internal/cpu
	// feature detection (schedinit) — takes the inlined LSE atomic path and
	// executes a CASAL. FEAT_LSE is not implemented on the ARMv8.0 Cortex-A35,
	// so the CASAL is UNDEFINED and traps synchronously (ESR_EL1.EC=0x00),
	// hanging before VBAR_EL1 is programmed. Clearing BSS here restores the
	// zero state the runtime expects. The range is 8-byte aligned by the linker.
	MOVD	$runtime·bss(SB), R0
	MOVD	$runtime·end(SB), R1
bss_zero:
	CMP	R1, R0
	BHS	bss_done
	MOVD	ZR, (R0)
	ADD	$8, R0, R0
	B	bss_zero
bss_done:

	B	_rt0_tamago_start(SB)
