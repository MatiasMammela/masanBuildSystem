package main

import (
	"fmt"
	"masanbuildsystem2/src"
	"os"
)

func main(){

	if(len(os.Args)>1){
		err := src.Init_command(os.Args[1:])
		if err != nil {
			fmt.Println(err)
		}
	}else{
		fmt.Println("usage: <command> [options]")
	}
}