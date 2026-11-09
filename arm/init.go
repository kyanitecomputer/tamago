// ARM processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build arm.6

package arm

import (
	_ "unsafe"
)

// goarmsoftfp is set by the linker when GOARM includes the softfloat option.
//
//go:linkname goarmsoftfp runtime.goarmsoftfp
var goarmsoftfp uint8

// Init takes care of the lower level initialization triggered before runtime
// setup (pre World start).
//
//go:nosplit
//go:linkname Init runtime/goos.Hwinit0
func Init() {
	if int(read_cpsr()&0x1f) != SYS_MODE {
		// initialization required only when in PL1
		return
	}

	// softfloat builds issue no VFP instructions, the VFP is left disabled
	// as it might not be present (e.g. ARM1176JZ-S).
	if goarmsoftfp != 0 {
		return
	}

	vfp_enable()
}
