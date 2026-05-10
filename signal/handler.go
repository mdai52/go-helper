package signal

import (
	"os"
	"os/signal"
	"syscall"
)

var quitChan chan os.Signal
var quitFuncs []func()

// OnQuit 注册退出时的回调函数
func OnQuit(fn func()) {
	quitFuncs = append(quitFuncs, fn)

	// 避免重复注册信号通道
	if quitChan != nil {
		return
	}

	// 创建监听中断信号通道
	quitChan = make(chan os.Signal, 1)

	// SIGTERM: `kill`
	// SIGINT : `kill -2` 或 CTRL + C
	// SIGKILL: `kill -9`，无法捕获，故而不做处理
	signal.Notify(quitChan, syscall.SIGTERM, syscall.SIGINT)

	// 等待退出信号
	go func() {
		<-quitChan
		CallQuitFuncs()
		os.Exit(0)
	}()
}

// CallQuitFuncs 调用所有退出函数
func CallQuitFuncs() {
	for _, fn := range quitFuncs {
		fn()
	}
}