// Package llm 大模型调用相关
package llm

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	openai "github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/schema"
)

var InitHandlersOnce sync.Once

// Client LLM调用的客户端
type Client struct {
	provider string
	model    *openai.ChatModel
}

func InitHandlers(handlers []callbacks.Handler) {
	// 如果创建的Client还有handler, 则自动将其添加到全局
	InitHandlersOnce.Do(func() {
		if len(handlers) > 0 {
			callbacks.AppendGlobalHandlers(handlers...)
		}
	})

}

// NewOpenaiClient 创建Openai接口格式的客户端
func NewOpenaiClient(ctx context.Context, provider string) (*Client, error) {
	model, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: os.Getenv("OPENAI_BASE_URL"),
		APIKey:  os.Getenv("OPENAI_API_KEY"),
		Model:   os.Getenv("OPENAI_MODEL"),
		Timeout: 30 * time.Second,
	})
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("链接provider超时：%w", err)
		} else {
			return nil, fmt.Errorf("创建model失败: %w", err)
		}
	}

	Client := &Client{
		provider: provider,
		model:    model,
	}

	return Client, nil
}

// Generate 调用client的model生成消息
func (c *Client) Generate(ctx context.Context, message []*schema.Message) (*schema.Message, error) {
	response, err := c.model.Generate(ctx, message)
	if err != nil {
		return nil, fmt.Errorf("生成消息失败: %w", err)
	}

	return response, nil
}
