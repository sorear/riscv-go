// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build riscv

// We aim for sequential consistency for all operations, following
// https://github.com/golang/go/issues/5045#issuecomment-252730563

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

TEXT ·Cas(SB), NOSPLIT, $0-17
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

TEXT ·Casp1(SB), NOSPLIT, $0-25
	MOV	ptr+0(FP), A0
	MOV	old+8(FP), A1
	MOV	new+16(FP), A2
cas:
	AMODSC(LR_,13,10,0)	// lr.d.sc a3,(a0)
	BNE	A3, A1, fail
	AMODSC(SC_,10,10,12)	// sc.d.sc a0,a2,(a0)
	// a0 = 0 iff the sc succeeded. Convert that to a boolean.
	SLTIU	$1, A0, A0
	MOV	A0, ret+24(FP)
	RET
fail:
	MOV	$0, A0
	MOV	A0, ret+24(FP)
	RET

TEXT ·Casuintptr(SB),NOSPLIT,$0-25
	JMP ·Casp1(SB)

TEXT ·Storeuintptr(SB),NOSPLIT,$0-16
	MOV	ptr+0(FP), A0
	MOV	new+8(FP), A1
	AMODSC(SWAP_,0,10,11)
	RET

TEXT ·Loaduintptr(SB),NOSPLIT,$0-16
	MOV	ptr+0(FP), A0
	AMODSC(LR_,10,10,0)
	MOV	A0, ret+8(FP)
	RET

TEXT ·Loaduint(SB),NOSPLIT,$0-16
	JMP ·Loaduintptr(SB)

TEXT ·Loadint64(SB),NOSPLIT,$0-16
	JMP ·Loaduintptr(SB)

TEXT ·Xaddint64(SB),NOSPLIT,$0-24
	MOV	ptr+0(FP), A0
	MOV	delta+8(FP), A1
	AMODSC(ADD_,10,10,11)	// amoadd.d.sc a0,a1,(a0)
	ADD	A0, A1, A0
	MOV	A0, ret+16(FP)
	RET
