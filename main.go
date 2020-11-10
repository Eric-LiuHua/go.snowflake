package main

import (
	"fmt"
	"runtime"
	"snowflake/idwork"
)

var quit chan int = make(chan int)

func main() {
	//设置最大开n个原生线程
	runtime.GOMAXPROCS(3)
	fmt.Printf("snowflake.main \n")
	go func() {
		var s idwork.Snowflake = idwork.InitSnowflakeA()
		for i := 0; i < 32; i++ {
			fmt.Printf("quit <- 1,index:%d ,id:%d \n", i, s.NewId(0 != 0))
			runtime.Gosched()
		}
		quit <- 1
	}()

	go func() {
		for i := 0; i < 12; i++ {
			fmt.Printf("quit <- 2,index:%d ,id:%d \n", i, idwork.NewId(0 != 0))
			runtime.Gosched() //需要这样进行并行
		}
		quit <- 2
	}()

	for i := 0; i < 2; i++ {
		cur := <-quit
		fmt.Println("current PROCS:", cur)
	}

	fmt.Println("snowflake.end ")
}
