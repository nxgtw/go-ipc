#include "textflag.h"

TEXT runtime·sysenter(SB),NOSPLIT,$0
	POPL	DX
	MOVL	SP, CX
	BYTE $0x0F; BYTE $0x34;  // SYSENTER
	// returns to DX with SP set to CX
	
TEXT runtime·mach_reply_port(SB),NOSPLIT,$0
	MOVL	$-26, AX
	CALL	runtime·sysenter(SB)
	MOVL	AX, ret+0(FP)
	RET
