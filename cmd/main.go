package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/SKDG042/Zero/devops"
	"github.com/SKDG042/Zero/llm"
	"github.com/SKDG042/Zero/ui"
	"github.com/joho/godotenv"
)

func main() {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		log.Printf("加载环境变量失败: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 初始化EinoDev
	if err := devops.Init(ctx); err != nil {
		log.Printf("初始化EinoDev失败：%v", err)
	} else {
		defer devops.Shutdown(ctx)
	}

	// 初始化 callback handlers
	handlers, err := devops.GetCallbackHandlers(ctx)
	if err != nil {
		log.Printf("初始化callback handlers失败：%v", err)
	}

	// 注册全局handlers
	llm.InitHandlers(handlers)

	// 初始化LLM client
	client, err := llm.NewOpenaiClient(ctx, "openai")
	if err != nil {
		log.Printf("初始化LLM client失败:%v", err)
	}

	log.Println("LLM client和所有链路追踪初始化完毕")

	mainWindow := ui.NewMainWindow(client)

	// 优雅退出
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM) // 将ctrl+c和windows系统发送的终止信号发向signalCHan
	go func() {
		<-signalChan
		log.Println("正在优雅的关闭程序ing")
		cancel()
		mainWindow.App.Quit()
	}()

	mainWindow.Run()
}
