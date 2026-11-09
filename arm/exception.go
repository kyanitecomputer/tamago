// ARM processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package arm

import (
	"runtime"
	"runtime/goos"
	"unsafe"

	"github.com/usbarmory/tamago/internal/exception"
	"github.com/usbarmory/tamago/internal/reg"
)

// ARM exception vector offsets
// (Table 11-1, ARM® Cortex™ -A Series Programmer’s Guide).
const (
	RESET          = 0x00
	UNDEFINED      = 0x04
	SUPERVISOR     = 0x08
	PREFETCH_ABORT = 0x0c
	DATA_ABORT     = 0x10
	IRQ            = 0x18
	FIQ            = 0x1c
)

// linknamed by application as needed
var vecTableStart uint32

// excStack is the exception stack address set by initVectorTable,
// read by exception_v5.s which cannot use the ARMv6+ banked SP MRS.
var excStack uint32

const (
	vecTableJump   = 0xe59ff018 // ldr pc, [pc, #24]
	excStackOffset = 0x8000     // 32 kB
	excStackSize   = 0x4000     // 16 kB
)

// defined in exception.s or exception_v5.s (see build constraints)
func read_spsr() uint32
func set_exc_stack(addr uint32)
func set_vbar(addr uint32)
func set_mvbar(addr uint32)
func resetHandler()
func undefinedHandler()
func supervisorHandler()
func prefetchAbortHandler()
func dataAbortHandler()
func irqHandler()
func fiqHandler()
func nullHandler()

var (
	offset     int
	eip        uintptr
	isThrowing bool

	// excG is the g register (R10) of the interrupted context, zero
	// before the runtime installs one.
	excG uint32
	// excMode is the processor mode the exception was taken to.
	excMode int
	// inject has the exception vector resume the interrupted goroutine
	// in exceptionPanic once the system exception handler returns.
	inject bool
	// injectLR, when not zero, replaces the interrupted link register
	// for exceptionPanic.
	injectLR uintptr
)

// CPSR/SPSR mode and Thumb state bits
const (
	psrModeMask = 0x1f
	psrThumb    = 1 << 5
)

const (
	// Go ARM prologue saving the link register: str lr, [sp, #-n]!
	pushLR     = 0xe52de000
	pushLRMask = 0xfffff000
	// prologue instructions searched for pushLR
	prologueLen = 8
)

// linkSaved reports whether the function executing pc pushed its link
// register, and with it moved the stack pointer, before reaching pc.
func linkSaved(pc uintptr) bool {
	fn := runtime.FuncForPC(pc)

	if fn == nil {
		return false
	}

	for p := fn.Entry(); p < pc && p < fn.Entry()+4*prologueLen; p += 4 {
		if reg.Read(uint32(p))&pushLRMask == pushLR {
			return true
		}
	}

	return false
}

type ExceptionHandler func()

func vector(fn ExceptionHandler) uint32 {
	return **((**uint32)(unsafe.Pointer(&fn)))
}

type VectorTable struct {
	Reset         ExceptionHandler
	Undefined     ExceptionHandler
	Supervisor    ExceptionHandler
	PrefetchAbort ExceptionHandler
	DataAbort     ExceptionHandler
	IRQ           ExceptionHandler
	FIQ           ExceptionHandler
}

// DefaultExceptionHandler handles an exception by panicking on the
// interrupted goroutine, as if its faulting instruction had called a function
// that panics: the panic then unwinds that goroutine's stack, running its
// deferred calls, and recover works as for any panic.
//
// An exception taken outside of a goroutine (in another exception mode,
// before the runtime starts or in Thumb state) has no stack to unwind: its
// vector, processor mode and location are printed before halting.
func DefaultExceptionHandler() {
	if isThrowing {
		goos.Exit(1)
	}

	excMode = int(read_cpsr() & psrModeMask)
	spsr := read_spsr()

	if spsr&psrModeMask == SYS_MODE && spsr&psrThumb == 0 && excG != 0 {
		// Once the faulting function saved its return address,
		// exceptionPanic appears called by the faulting instruction.
		// Before, or in a leaf function, its caller's frame is still
		// current: exceptionPanic appears called from there.
		injectLR = 0

		if linkSaved(eip) {
			injectLR = eip
		}

		inject = true
		return
	}

	isThrowing = true

	print("exception: vector ", offset, " mode ", excMode, " from mode ", int(spsr&psrModeMask), "\n")
	exception.Print(eip)
	goos.Exit(1)
}

// exceptionPanic runs on the interrupted goroutine, entered from the exception
// vector with a return address that continues the goroutine's traceback:
// the faulting instruction, or its caller's call site (see
// DefaultExceptionHandler).
//
//go:noinline
func exceptionPanic() {
	print("exception: vector ", offset, " mode ", excMode, "\n")
	exception.Print(eip)
	panic("unhandled exception: " + VectorName(offset))
}

// SystemExceptionHandler allows to override the default exception handler
// executed at any exception by the table returned by SystemVectorTable(),
// which is used by default when initializing the CPU instance (e.g.
// CPU.Init()).
var SystemExceptionHandler = DefaultExceptionHandler

// SystemVectorTable returns a vector table that, for all exceptions, switches
// to system mode and calls the SystemExceptionHandler on the Go runtime stack
// within goroutine g0.
func SystemVectorTable() VectorTable {
	return VectorTable{
		Reset:         resetHandler,
		Undefined:     undefinedHandler,
		Supervisor:    supervisorHandler,
		PrefetchAbort: prefetchAbortHandler,
		DataAbort:     dataAbortHandler,
		IRQ:           irqHandler,
		FIQ:           fiqHandler,
	}
}

// VectorName returns the exception vector offset name.
func VectorName(off int) string {
	switch off {
	case RESET:
		return "RESET"
	case UNDEFINED:
		return "UNDEFINED"
	case SUPERVISOR:
		return "SUPERVISOR"
	case PREFETCH_ABORT:
		return "PREFETCH_ABORT"
	case DATA_ABORT:
		return "DATA_ABORT"
	case IRQ:
		return "IRQ"
	case FIQ:
		return "FIQ"
	}

	return "Unknown"
}

// SetVectorTable updates the CPU exception handling vector table with the
// addresses of the functions defined in the passed structure.
func (cpu *CPU) SetVectorTable(t VectorTable) {
	vecTable := cpu.vbar + 8*4

	// set handler pointers
	// Table 11-1 ARM® Cortex™ -A Series Programmer’s Guide

	reg.Write(vecTable+RESET, vector(t.Reset))
	reg.Write(vecTable+UNDEFINED, vector(t.Undefined))
	reg.Write(vecTable+SUPERVISOR, vector(t.Supervisor))
	reg.Write(vecTable+PREFETCH_ABORT, vector(t.PrefetchAbort))
	reg.Write(vecTable+DATA_ABORT, vector(t.DataAbort))
	reg.Write(vecTable+IRQ, vector(t.IRQ))
	reg.Write(vecTable+FIQ, vector(t.FIQ))
}

//go:nosplit
func (cpu *CPU) initVectorTable() {
	// 32-bytes alignment is required
	cpu.vbar = uint32(goos.RamStart)

	// the application is allowed to override the reserved area
	if vecTableStart != 0 {
		cpu.vbar = vecTableStart
	}

	// initialize jump table
	// Table 11-1 ARM® Cortex™ -A Series Programmer’s Guide
	for i := range uint32(8) {
		reg.Write(cpu.vbar+4*i, vecTableJump)
	}

	// set exception handlers
	cpu.SetVectorTable(SystemVectorTable())

	// Do not set VBAR on cores with fixed exception vectors (e.g. ARMv5)
	// which have no VBAR/MVBAR registers.
	if cpu.vbar != 0 {
		set_vbar(cpu.vbar)

		if cpu.Secure() {
			// set monitor vector base address register
			set_mvbar(cpu.vbar)
		}
	}

	// Set the stack pointer for exception modes to provide a stack when
	// summoned by exception vectors.
	excStack = cpu.vbar + excStackOffset + excStackSize
	set_exc_stack(excStack)
}
