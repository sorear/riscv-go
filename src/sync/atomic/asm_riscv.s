// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build riscv

// RISC-V's atomic operations have two bits, aq ("acquire") and rl ("release"),
// which may be toggled on and off. Their precise semantics are defined in
// section 6.3 of the specification, but the basic idea is as follows:
//
//   - If neither aq nor rl is set, the CPU may reorder the atomic arbitrarily.
//     It guarantees only that it will execute atomically.
//
//   - If aq is set, the CPU may move the instruction backward, but not forward.
//
//   - If rl is set, the CPU may move the instruction forward, but not backward.
//
//   - If both are set, the CPU may not reorder the instruction at all.
//
// These four modes correspond to other well-known memory models on other CPUs.
// On ARM, aq corresponds to a dmb ishst, aq+rl corresponds to a dmb ish. On
// Intel, aq corresponds to an lfence, rl to an sfence, and aq+rl to an mfence
// (or a lock prefix).
//
// Go's memory model does not provide semantics for sync/atomic, but proper
// operation of sync.Mutex requires that atomic operations serve as memory
// barriers in both Lock and Unlock.  We employ sequential consistency, per
// https://github.com/golang/go/issues/5045#issuecomment-252730563 .

#include "textflag.h"

// for A0-A7 add 10 to get the register number
#define AMOWSC(op,rd,rs1,rs2) WORD $0x0600202f+rd<<7+rs1<<15+rs2<<20+op<<27
#define AMODSC(op,rd,rs1,rs2) WORD $0x0600302f+rd<<7+rs1<<15+rs2<<20+op<<27
#define ADD_ 0
#define SWAP_ 1
#define LR_ 2
#define SC_ 3
#define OR_ 8
#define AND_ 12
#define FENCE WORD $0x0ff0000f

TEXT ·SwapInt32(SB),NOSPLIT,$0-20
	JMP	·SwapUint32(SB)

TEXT ·SwapInt64(SB),NOSPLIT,$0-24
	JMP	·SwapUint64(SB)

TEXT ·SwapUint32(SB),NOSPLIT,$0-20
	MOV	ptr+0(FP), A0
	MOVW	new+8(FP), A1
	AMOWSC(SWAP_,11,10,11)
	MOVW	A1, ret+16(FP)
	RET

TEXT ·SwapUint64(SB),NOSPLIT,$0-24
	MOV	ptr+0(FP), A0
	MOV	new+8(FP), A1
	AMODSC(SWAP_,11,10,11)
	MOV	A1, ret+16(FP)
	RET

TEXT ·SwapUintptr(SB),NOSPLIT,$0-24
	JMP	·SwapUint64(SB)

TEXT ·CompareAndSwapInt32(SB),NOSPLIT,$0-17
	JMP	·CompareAndSwapUint32(SB)

TEXT ·CompareAndSwapInt64(SB),NOSPLIT,$0-25
	JMP	·CompareAndSwapUint64(SB)

TEXT ·CompareAndSwapUint32(SB),NOSPLIT,$0-17
	MOV	ptr+0(FP), A0
	MOVW	old+8(FP), A1
	MOVW	new+12(FP), A2
again:
	AMOWSC(LR_,13,10,0)	// lr.w.sc a3,(a0)
	BNE	A3, A1, fail
	AMOWSC(SC_,14,10,12)	// sc.w.sc a0,a2,(a0)
	BNE	A4, ZERO, again // a4=0 if sc succeeded
	MOV	$1, A0
	MOVB	A0, ret+16(FP)
	RET
fail:
	MOVB	ZERO, ret+16(FP)
	RET

TEXT ·CompareAndSwapUint64(SB),NOSPLIT,$0-25
	MOV	ptr+0(FP), A0
	MOV	old+8(FP), A1
	MOV	new+16(FP), A2
again:
	AMODSC(LR_,13,10,0)
	BNE	A3, A1, fail
	AMODSC(SC_,14,10,12)
	BNE	A4, ZERO, again
	MOV	$1, A0
	MOVB	A0, ret+24(FP)
	RET
fail:
	MOVB	ZERO, ret+24(FP)
	RET

TEXT ·CompareAndSwapUintptr(SB),NOSPLIT,$0-25
	JMP	·CompareAndSwapUint64(SB)

TEXT ·AddInt32(SB),NOSPLIT,$0-20
	JMP	·AddUint32(SB)

TEXT ·AddUint32(SB),NOSPLIT,$0-20
	MOV	ptr+0(FP), A0
	MOVW	delta+8(FP), A1
	AMOWSC(ADD_,12,10,11)
	ADD	A2,A1,A0
	MOVW	A0, ret+16(FP)
	RET

TEXT ·AddInt64(SB),NOSPLIT,$0-24
	JMP	·AddUint64(SB)

TEXT ·AddUint64(SB),NOSPLIT,$0-24
	MOV	ptr+0(FP), A0
	MOV	delta+8(FP), A1
	AMODSC(ADD_,12,10,11)
	ADD	A2,A1,A0
	MOV	A0, ret+16(FP)
	RET

TEXT ·AddUintptr(SB),NOSPLIT,$0-24
	JMP	·AddUint64(SB)

TEXT ·LoadInt32(SB),NOSPLIT,$0-12
	JMP	·LoadUint32(SB)

TEXT ·LoadInt64(SB),NOSPLIT,$0-16
	JMP	·LoadUint64(SB)

TEXT ·LoadUint32(SB),NOSPLIT,$0-12
	MOV	ptr+0(FP), A0
	AMOWSC(LR_,10,10,0)
	MOVW	A0, ret+8(FP)
	RET

TEXT ·LoadUint64(SB),NOSPLIT,$0-16
	MOV	ptr+0(FP), A0
	AMODSC(LR_,10,10,0)
	MOV	A0, ret+8(FP)
	RET

TEXT ·LoadUintptr(SB),NOSPLIT,$0-16
	JMP	·LoadUint64(SB)

TEXT ·LoadPointer(SB),NOSPLIT,$0-16
	JMP	·LoadUint64(SB)

TEXT ·StoreInt32(SB),NOSPLIT,$0-12
	JMP	·StoreUint32(SB)

TEXT ·StoreInt64(SB),NOSPLIT,$0-16
	JMP	·StoreUint64(SB)

TEXT ·StoreUint32(SB),NOSPLIT,$0-12
	MOV	ptr+0(FP), A0
	MOVW	val+8(FP), A1
	AMOWSC(SWAP_,0,10,11)
	RET

TEXT ·StoreUint64(SB),NOSPLIT,$0-16
	MOV	ptr+0(FP), A0
	MOV	val+8(FP), A1
	AMODSC(SWAP_,0,10,11)
	RET

TEXT ·StoreUintptr(SB),NOSPLIT,$0-16
	JMP	·StoreUint64(SB)
