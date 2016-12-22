#include "textflag.h"

TEXT mach_task_self(SB),NOSPLIT,$0
	MOVL	$(0x1000000+28), AX	// task_self_trap
	SYSCALL
	MOVL	AX, ret+0(FP)
	RET
