#include "textflag.h"

// func countZeroBytes(buf []byte) uint32
TEXT ·countZeroBytes(SB), NOSPLIT, $0-28
	MOVQ   buf_base+0(FP), AX
	MOVQ   buf_len+8(FP), CX
	XORL   DX, DX
	VXORPS Y0, Y0, Y0
	VXORPS X1, X1, X1

block_loop:
	CMPQ      CX, $0x000000c0
	JL        done
	VMOVDQA   (AX), Y1
	VPCMPEQB  Y1, Y0, Y1
	VPMOVMSKB Y1, BX
	POPCNTL   BX, BX
	ADDL      BX, DX
	VMOVDQA   32(AX), Y1
	VPCMPEQB  Y1, Y0, Y1
	VPMOVMSKB Y1, BX
	POPCNTL   BX, BX
	ADDL      BX, DX
	VMOVDQA   64(AX), Y1
	VPCMPEQB  Y1, Y0, Y1
	VPMOVMSKB Y1, BX
	POPCNTL   BX, BX
	ADDL      BX, DX
	VMOVDQA   96(AX), Y1
	VPCMPEQB  Y1, Y0, Y1
	VPMOVMSKB Y1, BX
	POPCNTL   BX, BX
	ADDL      BX, DX
	VMOVDQA   128(AX), Y1
	VPCMPEQB  Y1, Y0, Y1
	VPMOVMSKB Y1, BX
	POPCNTL   BX, BX
	ADDL      BX, DX
	VMOVDQA   160(AX), Y1
	VPCMPEQB  Y1, Y0, Y1
	VPMOVMSKB Y1, BX
	POPCNTL   BX, BX
	ADDL      BX, DX
	ADDQ      $0x000000c0, AX
	SUBQ      $0x000000c0, CX
	JMP       block_loop

done:
	MOVL DX, ret+24(FP)
	RET