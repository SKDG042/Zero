// Package devops EinoDev和Coze Loop的trace相关
package devops

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	ccb "github.com/cloudwego/eino-ext/callbacks/cozeloop"
	"github.com/cloudwego/eino-ext/devops"
	"github.com/cloudwego/eino/callbacks"
	"github.com/coze-dev/cozeloop-go"
)

var cozeloopClient cozeloop.Client
var cozeloopOnce sync.Once // 添加sync.Once来避免重复创建client

// Init 用于初始化 EinoDev的链路追踪
func Init(ctx context.Context) error {
	// 首先设置 EinoDev 可视化操作相关的链路追踪
	if os.Getenv("EINO_DEVOPS_ENABLED") == "true" {
		port := os.Getenv("EINO_DEVOPS_PORT")
		if port == "" {
			port = "52538"
		}

		err := devops.Init(ctx, devops.WithDevServerPort(port))
		if err != nil {
			return fmt.Errorf("初始化EinoDev服务失败：%w", err)
		}

		log.Printf("EinoDev 服务已启动: http://127.0.0.1:%s\n", port)
	}
	return nil
}

// GetCallbackHandlers 获取所有的链路追踪handler
// 同时初始化除了EinoDev外的handler
func GetCallbackHandlers(ctx context.Context) ([]callbacks.Handler, error) {
	var handlers []callbacks.Handler

	if os.Getenv("COZELOOP_ENABLED") == "true" {
		// 使用cozeloopOnce.Do 防止重复床啊今client导致coze报错
		cozeloopOnce.Do(func() {
			client, err := cozeloop.NewClient()
			if err != nil {
				log.Printf("cozeloop链路追踪初始化失败：%v", err)
			} else {
				cozeloopClient = client
				log.Println("cozeloop成功启动")
			}
		})

		// 只在客户端成功创建时添加 Handler
		if cozeloopClient != nil {
			handler := ccb.NewLoopHandler(cozeloopClient)
			handlers = append(handlers, handler)
		}
	}

	return handlers, nil
}

func Shutdown(ctx context.Context) {
	cozeloopClient.Close(ctx)
	log.Println("cozeloop成功关闭")
}
