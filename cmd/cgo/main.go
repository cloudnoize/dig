package main

// #include <stdio.h>
// #include <stdlib.h>
//
// static void myprint(char* s) {
//   printf("%s\n", s);
// }
import "C"
import (
	"log"
	"net/http"
	"os"
	"reflect"
	"unsafe"
)

func main() {
	cs := C.CString("Hello from stdio")
	C.myprint(cs)
	log.Println(reflect.TypeOf(cs))
	C.free(unsafe.Pointer(cs))

	cd := "momo"
	log.Println(reflect.TypeOf(cd))

	if v := os.Getenv("CONNECTION_STRING"); v != "" {

	}
}

func GetHandler(pmian *CTEngine) http.Handler {

}
