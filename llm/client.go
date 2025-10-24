// Package llm 大模型调用相关
package llm

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	openai "github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
)

// Client LLM调用的客户端
type Client struct {
	provider string
	model    *openai.ChatModel
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

// GenerateStream 调用流式输出
func (c *Client) GenerateStream(ctx context.Context, message []*schema.Message, onChunk func(string) error) error {
	streamReader, err := c.model.Stream(ctx, message)
	if err != nil {
		return fmt.Errorf("流式生成消息失败: %w", err)
	}
	defer streamReader.Close()

	// 循环接收消息
	for {
		chunk, err := streamReader.Recv()
		// 遇到EOF截止信号则退出循环
		if err == io.EOF {
			break
		}

		if err != nil {
			return fmt.Errorf("接收消息失败：%w", err)
		}

		// 将消息交给回调函数
		if err := onChunk(chunk.Content); err != nil {
			return fmt.Errorf("处理消息块失败：%w", err)
		}
	}

	return nil
}
