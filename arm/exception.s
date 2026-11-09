// ARM processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "go_asm.h"
#include "textflag.h"

// func set_exc_stack(addr uint32)
TEXT ·set_exc_stack(SB),NOSPLIT,$0-4
	MOVW addr+0(FP), R0

	// set FIQ mode SP
	WORD	$0xe321f0d1	// msr CPSR_c, 0xd1
	MOVW R0, R13

	// set IRQ mode SP
	WORD	$0xe321f0d2	// msr CPSR_c, 0xd2
	MOVW R0, R13

	// set Supervisor mode SP
	WORD	$0xe321f0d3	// msr CPSR_c, 0xd3
	MOVW R0, R13
#ifndef GOARM_5
	// set Monitor mode SP
	WORD	$0xe321f0d6	// msr CPSR_c, 0xd6
	MOVW R0, R13
#endif
	// set Abort mode SP
	WORD	$0xe321f0d7	// msr CPSR_c, 0xd7
	MOVW R0, R13

	// set Undefined mode SP
	WORD	$0xe321f0db	// msr CPSR_c, 0xdb
	MOVW R0, R13

	// return to System mode
	WORD	$0xe321f0df	// msr CPSR_c, 0xdf

	RET

// func read_spsr() uint32
TEXT ·read_spsr(SB),NOSPLIT,$0-4
	WORD	$0xe14f0000 // mrs r0, SPSR
	MOVW	R0, ret+0(FP)
	RET

// func set_vbar(addr uint32)
TEXT ·set_vbar(SB),NOSPLIT,$0-4
	MOVW	addr+0(FP), R0
	MCR	15, 0, R0, C12, C0, 0
	RET

// func set_mvbar(addr uint32)
TEXT ·set_mvbar(SB),NOSPLIT,$0-4
	MOVW	addr+0(FP), R0
	MCR	15, 0, R0, C12, C0, 1
	RET

#define EXCEPTION(OFFSET, FN, LROFFSET)				\
	/* save exception vector */				\
	MOVW	$OFFSET, R0					\
	MOVW	R0, ·offset(SB)					\
	MOVW	R14, R0						\
								\
	/* remove exception specific LR offset */		\
	SUB	$LROFFSET, R14, R14				\
	MOVW	R14, ·eip(SB)					\
								\
	/* save the interrupted goroutine */			\
	MOVW	g, ·excG(SB)					\
								\
	/* call exception handler on system stack (g0) */	\
	MOVW	FN(SB), R0					\
	MOVW	R0, 4(R13)					\
	CALL	runtime·CallOnG0(SB)				\
	B	·resumeException(SB)

// resumeException runs when the system exception handler returns. When it
// asked for it, the interrupted System mode goroutine resumes in
// exceptionPanic, its link register set to injectLR unless that is zero, as
// if exceptionPanic had been called there: the goroutine's own stack then
// unwinds the panic. Otherwise the CPU halts.
TEXT ·resumeException(SB),NOSPLIT|NOFRAME,$0
	MOVB	·inject(SB), R0
	CMP	$0, R0
	B.EQ	halt
	MOVW	$0, R0
	MOVB	R0, ·inject(SB)

	MOVW	·injectLR(SB), R0
	CMP	$0, R0
	B.EQ	resume
	WORD	$0xe10f1000	// mrs r1, CPSR
	WORD	$0xe321f0df	// msr CPSR_c, 0xdf (System mode, IRQ/FIQ masked)
	MOVW	R0, R14
	WORD	$0xe121f001	// msr CPSR_c, r1 (back to the exception mode)
resume:

	MOVW	$·exceptionPanic(SB), R14
	MOVW.S	R14, R15	// return to System mode, restoring CPSR from SPSR
halt:
	B	·exit(SB)

TEXT ·resetHandler(SB),NOSPLIT|NOFRAME,$0
	EXCEPTION(const_RESET, ·SystemExceptionHandler, 0)

TEXT ·undefinedHandler(SB),NOSPLIT|NOFRAME,$0
	EXCEPTION(const_UNDEFINED, ·SystemExceptionHandler, 4)

TEXT ·supervisorHandler(SB),NOSPLIT|NOFRAME,$0
	EXCEPTION(const_SUPERVISOR, ·SystemExceptionHandler, 0)

TEXT ·prefetchAbortHandler(SB),NOSPLIT|NOFRAME,$0
	EXCEPTION(const_PREFETCH_ABORT, ·SystemExceptionHandler, 4)

TEXT ·dataAbortHandler(SB),NOSPLIT|NOFRAME,$0
	EXCEPTION(const_DATA_ABORT, ·SystemExceptionHandler, 8)

TEXT ·fiqHandler(SB),NOSPLIT|NOFRAME,$0
	EXCEPTION(const_FIQ, ·SystemExceptionHandler, 4)

TEXT ·nullHandler(SB),NOSPLIT|NOFRAME,$0
	MOVW.S	R14, R15
