// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sys

const (
	ArchFamily    = RISCV
	BigEndian     = 0
	CacheLineSize = 64   // TODO(prattmic)
	PhysPageSize  = 4096 // TODO(prattmic)
	// TODO(sorear) On non-RVC hardware jumping to goexit+2 is likely to crash.  Is that a problem?
	PCQuantum    = 2
	Int64Align   = 8
	HugePageSize = 1 << 21
	MinFrameSize = 8
)

type Uintreg uint64
