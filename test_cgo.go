package main

/*
#include <stdio.h>
void hello() {
    printf("Hello from C!\n");
}
*/
import "C"
import "fmt"

func main() {
	fmt.Println("Hello from Go!")
	C.hello()
}
