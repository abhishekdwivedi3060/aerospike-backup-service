package shared

/*
#include <stdbool.h>
#include <stdint.h>
*/
import "C"

func setCString(cchar **C.char, str *string) {
	if str != nil {
		*cchar = C.CString(*str)
	}
}

func setCInt(cint *C.int, i *int) {
	if i != nil {
		*cint = C.int(*i)
	}
}

func setCLong(clong *C.int64_t, l *int64) {
	if l != nil {
		*clong = C.int64_t(*l)
	}
}

func setCBool(cbool *C.bool, b *bool) {
	if b != nil {
		*cbool = C.bool(*b)
	}
}
