// Gigabyte MJ11-EC1 BMC support for tamago/arm
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !linkramsize

package mj11ec1bmc

import _ "unsafe"

// The BMC has 512 MiB of DDR4 at 0x80000000. With ECC on and the 32 MiB VGA
// carve-out at its top, the loader leaves the first 424 MiB usable.
//
//go:linkname ramSize runtime/goos.RamSize
var ramSize uint32 = 0x1a800000 // 424 MiB
