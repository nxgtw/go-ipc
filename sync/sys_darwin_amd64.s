#include "textflag.h"

TEXT ·mach_thread_self(SB),NOSPLIT,$0
	MOVL	$(0x1000000+27), AX	// mach_thread_self
	SYSCALL
	MOVL	AX, ret+0(FP)
	RET
