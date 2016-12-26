#include "textflag.h"

TEXT sysenter(SB),NOSPLIT,$0
	POPL	DX
	MOVL	SP, CX
	BYTE $0x0F; BYTE $0x34;  // SYSENTER
	// returns to DX with SP set to CX
	
TEXT ·mach_thread_self(SB),NOSPLIT,$0
	MOVL	$-27, AX
	CALL	sysenter(SB)
	MOVL	AX, ret+0(FP)
	RET
