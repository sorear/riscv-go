// Code generated by "stringer -type=RelocType"; DO NOT EDIT

package obj

import "fmt"

const _RelocType_name = "R_ADDRR_ADDRPOWERR_ADDRARM64R_ADDRMIPSR_ADDROFFR_WEAKADDROFFR_SIZER_CALLR_CALLARMR_CALLARM64R_CALLINDR_CALLPOWERR_CALLMIPSR_CALLRISCV1R_CALLRISCV2R_CONSTR_PCRELR_TLS_LER_TLS_IER_GOTOFFR_PLT0R_PLT1R_PLT2R_USEFIELDR_USETYPER_METHODOFFR_POWER_TOCR_GOTPCRELR_JMPMIPSR_DWARFREFR_ARM64_TLS_LER_ARM64_TLS_IER_ARM64_GOTPCRELR_POWER_TLS_LER_POWER_TLS_IER_POWER_TLSR_ADDRPOWER_DSR_ADDRPOWER_GOTR_ADDRPOWER_PCRELR_ADDRPOWER_TOCRELR_ADDRPOWER_TOCREL_DSR_PCRELDBLR_ADDRMIPSUR_ADDRMIPSTLSR_RISCV_PCREL_ITYPER_RISCV_PCREL_STYPE"

var _RelocType_index = [...]uint16{0, 6, 17, 28, 38, 47, 60, 66, 72, 81, 92, 101, 112, 122, 134, 146, 153, 160, 168, 176, 184, 190, 196, 202, 212, 221, 232, 243, 253, 262, 272, 286, 300, 316, 330, 344, 355, 369, 384, 401, 419, 440, 450, 461, 474, 493, 512}

func (i RelocType) String() string {
	i -= 1
	if i < 0 || i >= RelocType(len(_RelocType_index)-1) {
		return fmt.Sprintf("RelocType(%d)", i+1)
	}
	return _RelocType_name[_RelocType_index[i]:_RelocType_index[i+1]]
}
