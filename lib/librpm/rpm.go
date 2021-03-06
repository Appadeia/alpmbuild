package librpm

// #cgo pkg-config: rpm
// #include <rpm/rpmmacro.h>
//
// static char* expandy(char* in) {
//	char* result = rpmExpand(in, NULL);
//	return result;
// }
// static int definemacro(char* macro, int level) {
//	return rpmDefineMacro(NULL, macro, level);
// }
// static int loadfromfile(char* path) {
//	return rpmLoadMacroFile(NULL, path);
// }
// static void dumpmacros() {
//	FILE* file = fopen("/tmp/alpmbuild", "w+");
//	rpmDumpMacroTable(NULL, file);
//	fclose(file);
// }
// static void delmacro(char* macro) {
//	delMacro(NULL, macro);
// }
//
import "C"
import (
	"io/ioutil"
	"strings"
	"unsafe"
)

type Macro struct {
	Macro    string
	ExpandTo string
}

func ExpandMacro(macro string) string {
	cs := C.CString(macro)
	defer C.free(unsafe.Pointer(cs))

	expanded := C.expandy(cs)
	defer C.free(unsafe.Pointer(expanded))

	goStr := C.GoString(expanded)
	return goStr
}

func DeleteMacro(macro string) {
	cs := C.CString(macro)
	defer C.free(unsafe.Pointer(cs))

	C.delmacro(cs)
}

func DefineMacro(macro string, level int) int {
	cs := C.CString(macro)
	defer C.free(unsafe.Pointer(cs))

	result := C.definemacro(cs, C.int(level))

	return int(result)
}

func LoadFromFile(path string) int {
	cs := C.CString(path)
	defer C.free(unsafe.Pointer(cs))

	result := C.loadfromfile(cs)

	return int(result)
}

func DumpMacros() []Macro {
	C.dumpmacros()

	bytes, err := ioutil.ReadFile("/tmp/alpmbuild")
	if err != nil {
		return []Macro{}
	}
	str := string(bytes)

	macros := []Macro{}

	for _, line := range strings.Split(str, "\n") {
		if strings.Contains(line, ":") {
			fields := strings.Fields(line)
			if len(fields) < 3 {
				continue
			}
			macros = append(macros, Macro{Macro: fields[1], ExpandTo: fields[2]})
		}
	}

	return macros
}

func DumpMacroNamesAsString() []string {
	C.dumpmacros()

	bytes, err := ioutil.ReadFile("/tmp/alpmbuild")
	if err != nil {
		return []string{}
	}
	str := string(bytes)

	macros := []string{}

	for _, line := range strings.Split(str, "\n") {
		if strings.Contains(line, ":") {
			fields := strings.Fields(line)
			if len(fields) < 3 {
				continue
			}
			macros = append(macros, fields[1])
		}
	}

	return macros
}
