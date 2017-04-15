//	Copyright © 1994-1999 Lucent Technologies Inc.  All rights reserved.
//	Portions Copyright © 1995-1997 C H Forsyth (forsyth@terzarima.net)
//	Portions Copyright © 1997-1999 Vita Nuova Limited
//	Portions Copyright © 2000-2007 Vita Nuova Holdings Limited (www.vitanuova.com)
//	Portions Copyright © 2004,2006 Bruce Ellis
//	Portions Copyright © 2005-2007 C H Forsyth (forsyth@terzarima.net)
//	Revisions Copyright © 2000-2007 Lucent Technologies Inc. and others
//	Portions Copyright © 2009 The Go Authors.  All rights reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.  IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package riscv

import (
	"cmd/internal/obj"
	"cmd/internal/obj/riscv"
	"cmd/link/internal/ld"
	"fmt"
	"log"
)

func gentext(ctxt *ld.Link) {
}

func adddynrela(ctxt *ld.Link, rel *ld.Symbol, s *ld.Symbol, r *ld.Reloc) {
	log.Fatalf("adddynrela not implemented")
}

func adddynrel(ctxt *ld.Link, s *ld.Symbol, r *ld.Reloc) bool {
	log.Fatalf("adddynrel not implemented")
	return false
}

func elfreloc1(ctxt *ld.Link, r *ld.Reloc, sectoff int64) int {
	log.Printf("elfreloc1")
	return -1
}

func elfsetupplt(ctxt *ld.Link) {
	log.Printf("elfsetuplt")
	return
}

func machoreloc1(s *ld.Symbol, r *ld.Reloc, sectoff int64) int {
	log.Fatalf("machoreloc1 not implemented")
	return -1
}

func jumpInRange(s *ld.Symbol, r *ld.Reloc, rsym *ld.Symbol, radd int64) bool {
	pc := s.Value + int64(r.Off)
	off := ld.Symaddr(rsym) + radd - pc
	return off >= -(1<<20) && off < (1<<20)
}

// Convert the direct jump relocation r to refer to a trampoline if the target is too far
func trampoline(ctxt *ld.Link, r *ld.Reloc, s *ld.Symbol) {
	switch r.Type {
	case obj.R_CALLRISCV1:
		if jumpInRange(s, r, r.Sym, r.Add) {
			return
		}

		// direct call too far, need to insert trampoline.
		// look up existing trampolines first. if we found one within the range
		// of direct call, we can reuse it. otherwise create a new one.
		if r.Sym.Attr&ld.AttrTrampoline != 0 {
			// we had reused a trampoline, but now it's too far
			// start fresh
			r.Add = r.Sym.R[0].Add
			r.Sym = r.Sym.R[0].Sym
		}

		var tramp *ld.Symbol
		for i := 0; ; i++ {
			name := r.Sym.Name + fmt.Sprintf("%+d-tramp%d", r.Add, i)
			tramp = ctxt.Syms.Lookup(name, int(r.Sym.Version))
			if tramp.Type == obj.SDYNIMPORT {
				// don't reuse trampoline defined in other module
				continue
			}
			if tramp.Value == 0 {
				// either the trampoline does not exist -- we need to create one,
				// or found one the address which is not assigned -- this will be
				// laid down immediately after the current function. use this one.
				break
			}

			if jumpInRange(s, r, tramp, 0) {
				// found an existing trampoline that is not too far
				// we can just use it
				break
			}
		}
		if tramp.Type == 0 {
			// trampoline does not exist, create one
			ctxt.AddTramp(tramp)
			if ctxt.DynlinkingGo() {
				ld.Errorf(s, "dynamic linking not implemented")
			} else {
				gentramp(tramp, r.Sym, r.Add)
			}
		}
		// modify reloc to point to tramp, which will be resolved later
		r.Sym = tramp
		r.Add = 0
		r.Done = 0
	default:
		ld.Errorf(s, "trampoline called with non-jump reloc: %v", r.Type)
	}
}

// generate a trampoline to target+offset without PLT ind
func gentramp(tramp, target *ld.Symbol, offset int64) {
	tramp.Size = 8 // 2 instructions
	tramp.P = make([]byte, tramp.Size)
	o1 := uint32(0x00000f97) // AUIPC T6, 0
	o2 := uint32(0x000f8067) // JR    T6
	ld.SysArch.ByteOrder.PutUint32(tramp.P, o1)
	ld.SysArch.ByteOrder.PutUint32(tramp.P[4:], o2)

	r := ld.Addrel(tramp)
	r.Off = 0
	r.Type = obj.R_CALLRISCV2
	r.Siz = 8
	r.Sym = target
	r.Add = offset
}

func archreloc(ctxt *ld.Link, r *ld.Reloc, s *ld.Symbol, val *int64) int {
	switch r.Type {
	case obj.R_RISCV_PCREL_ITYPE, obj.R_RISCV_PCREL_STYPE, obj.R_CALLRISCV2:
		pc := s.Value + int64(r.Off)
		off := ld.Symaddr(r.Sym) + r.Add - pc

		// Generate AUIPC and second instruction immediates.
		low, high, err := riscv.Split32BitImmediate(off)
		if err != nil {
			ld.Errorf(s, "R_RISCV_PCREL_ relocation does not fit in 32-bits: %d", off)
			return 0
		}

		auipcImm, err := riscv.EncodeUImmediate(high)
		if err != nil {
			ld.Errorf(s, "cannot encode R_RISCV_PCREL_ AUIPC relocation offset for %s: %v", r.Sym.Name, err)
			return 0
		}

		var secondImm, secondImmMask int64
		switch r.Type {
		case obj.R_RISCV_PCREL_ITYPE, obj.R_CALLRISCV2:
			secondImmMask = riscv.ITypeImmMask
			secondImm, err = riscv.EncodeIImmediate(low)
			if err != nil {
				ld.Errorf(s, "cannot encode R_RISCV_PCREL_ITYPE I-type instruction relocation offset for %s: %v", r.Sym.Name, err)
				return 0
			}
		case obj.R_RISCV_PCREL_STYPE:
			secondImmMask = riscv.STypeImmMask
			secondImm, err = riscv.EncodeSImmediate(low)
			if err != nil {
				ld.Errorf(s, "cannot encode R_RISCV_PCREL_STYPE S-type instruction relocation offset for %s: %v", r.Sym.Name, err)
				return 0
			}
		default:
			panic(fmt.Sprintf("Unknown relocation type: %v", r.Type))
		}

		auipc := int64(uint32(*val))
		second := int64(uint32(*val >> 32))

		auipc = (auipc &^ riscv.UTypeImmMask) | int64(uint32(auipcImm))
		second = (second &^ secondImmMask) | int64(uint32(secondImm))

		*val = second<<32 | auipc
	case obj.R_CALLRISCV1:
		pc := s.Value + int64(r.Off)
		off := ld.Symaddr(r.Sym) + r.Add - pc

		// This is always a JAL instruction, we just need
		// to replace the immediate.
		//
		// TODO(prattmic): sanity check the opcode.
		if off&1 != 0 {
			ld.Errorf(s, "R_CALLRISCV relocation for %s is not aligned: %#x", r.Sym.Name, off)
			return 0
		}

		imm, err := riscv.EncodeUJImmediate(off)
		if err != nil {
			// Anything larger should have resulted in a trampoline
			ld.Errorf(s, "cannot encode R_CALLRISCV1 relocation offset for %s: %v", r.Sym.Name, err)
			return 0
		}

		// The assembler is encoding a normal JAL instruction with the
		// immediate as whatever p.To.Offset. We need to replace that
		// immediate with the relocated value.
		*val = (*val &^ riscv.UJTypeImmMask) | int64(imm)
	default:
		return -1
	}

	return 0
}

func archrelocvariant(ctxt *ld.Link, r *ld.Reloc, s *ld.Symbol, t int64) int64 {
	log.Printf("archrelocvariant")
	return -1
}

// TODO(mpratt): Refactor asmb for other archs. This worked literally copied
// and pasted from mips64.
func asmb(ctxt *ld.Link) {
	if ctxt.Debugvlog != 0 {
		fmt.Fprintf(ctxt.Bso, "%5.2f asmb\n", obj.Cputime())
	}
	ctxt.Bso.Flush()

	if ld.Iself {
		ld.Asmbelfsetup()
	}

	sect := ld.Segtext.Sect
	ld.Cseek(int64(sect.Vaddr - ld.Segtext.Vaddr + ld.Segtext.Fileoff))
	ld.Codeblk(ctxt, int64(sect.Vaddr), int64(sect.Length))
	for sect = sect.Next; sect != nil; sect = sect.Next {
		ld.Cseek(int64(sect.Vaddr - ld.Segtext.Vaddr + ld.Segtext.Fileoff))
		ld.Datblk(ctxt, int64(sect.Vaddr), int64(sect.Length))
	}

	if ld.Segrodata.Filelen > 0 {
		if ctxt.Debugvlog != 0 {
			fmt.Fprintf(ctxt.Bso, "%5.2f rodatblk\n", obj.Cputime())
		}
		ctxt.Bso.Flush()

		ld.Cseek(int64(ld.Segrodata.Fileoff))
		ld.Datblk(ctxt, int64(ld.Segrodata.Vaddr), int64(ld.Segrodata.Filelen))
	}

	if ctxt.Debugvlog != 0 {
		fmt.Fprintf(ctxt.Bso, "%5.2f datblk\n", obj.Cputime())
	}
	ctxt.Bso.Flush()

	ld.Cseek(int64(ld.Segdata.Fileoff))
	ld.Datblk(ctxt, int64(ld.Segdata.Vaddr), int64(ld.Segdata.Filelen))

	ld.Cseek(int64(ld.Segdwarf.Fileoff))
	ld.Dwarfblk(ctxt, int64(ld.Segdwarf.Vaddr), int64(ld.Segdwarf.Filelen))

	/* output symbol table */
	ld.Symsize = 0

	ld.Lcsize = 0
	symo := uint32(0)
	if !*ld.FlagS {
		if ctxt.Debugvlog != 0 {
			fmt.Fprintf(ctxt.Bso, "%5.2f sym\n", obj.Cputime())
		}
		ctxt.Bso.Flush()
		switch ld.Headtype {
		default:
			if ld.Iself {
				symo = uint32(ld.Segdwarf.Fileoff + ld.Segdwarf.Filelen)
				symo = uint32(ld.Rnd(int64(symo), int64(*ld.FlagRound)))
			}

		case obj.Hplan9:
			log.Fatalf("Plan 9 unsupported")
		}

		ld.Cseek(int64(symo))
		switch ld.Headtype {
		default:
			if ld.Iself {
				if ctxt.Debugvlog != 0 {
					fmt.Fprintf(ctxt.Bso, "%5.2f elfsym\n", obj.Cputime())
				}
				ld.Asmelfsym(ctxt)
				ld.Cflush()
				ld.Cwrite(ld.Elfstrdat)

				if ld.Linkmode == ld.LinkExternal {
					ld.Elfemitreloc(ctxt)
				}
			}

		case obj.Hplan9:
			log.Fatalf("Plan 9 unsupported")
		}
	}

	if ctxt.Debugvlog != 0 {
		fmt.Fprintf(ctxt.Bso, "%5.2f header\n", obj.Cputime())
	}
	ctxt.Bso.Flush()
	ld.Cseek(0)
	switch ld.Headtype {
	default:
		log.Fatalf("Unsupported OS: %v", ld.Headtype)

	case obj.Hlinux,
		obj.Hfreebsd,
		obj.Hnetbsd,
		obj.Hopenbsd,
		obj.Hnacl:
		ld.Asmbelf(ctxt, int64(symo))
	}

	ld.Cflush()
	if *ld.FlagC {
		fmt.Printf("textsize=%d\n", ld.Segtext.Filelen)
		fmt.Printf("datsize=%d\n", ld.Segdata.Filelen)
		fmt.Printf("bsssize=%d\n", ld.Segdata.Length-ld.Segdata.Filelen)
		fmt.Printf("symsize=%d\n", ld.Symsize)
		fmt.Printf("lcsize=%d\n", ld.Lcsize)
		fmt.Printf("total=%d\n", ld.Segtext.Filelen+ld.Segdata.Length+uint64(ld.Symsize)+uint64(ld.Lcsize))
	}
}
