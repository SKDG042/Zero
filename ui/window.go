// Package ui 使用Fyne来写GUI界面
package ui

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/SKDG042/Zero/llm"
	"github.com/cloudwego/eino/schema"
)

// MainWindow 主界面
type MainWindow struct {
	App        fyne.App
	Window     fyne.Window
	statusBar  *widget.Label // 状态栏
	client     *llm.Client
	messages   []string     // 历史所有消息
	messageBox *widget.List // 消息列表
	inputEntry *widget.Entry
	sendButton *widget.Button
	isSending  bool
	mu         sync.Mutex
	cancelFunc context.CancelFunc
}

// NewMainWindow 创建GUI
func NewMainWindow(client *llm.Client) *MainWindow {

	// 用fyne创建一个新的窗口应用
	zeroApp := app.NewWithID("zero")
	zeroWindow := zeroApp.NewWindow("Zero Agent")

	mainWindow := &MainWindow{
		App:       zeroApp,
		Window:    zeroWindow,
		statusBar: widget.NewLabel("Zero: 你好 喵~"), // 状态栏
		client:    client,
		messages:  []string{"你好，这里是_042喵，需要我来做些什么吗？"},
	}

	// 绑定icon
	icon, err := fyne.LoadResourceFromPath("assets/icon.svg")
	if err != nil {
		log.Fatal("绑定Icon失败：", err)
	}
	zeroWindow.SetIcon(icon)

	// 创建左上角的那种菜单栏
	fileMenu := fyne.NewMenu("文件",
		fyne.NewMenuItem("新对话", func() {
			mainWindow.newConversation()
		}),
		fyne.NewMenuItem("退出", func() {
			content := container.NewVBox(
				widget.NewLabel("你确定要退出 Zero 吗？"),
				widget.NewLabel("未保存的数据(真的有吗)将丢失"),
			)

			dialogBox := dialog.NewCustomConfirm(
				"确认退出",
				"退出",
				"取消",
				content,
				func(confirmed bool) {
					if confirmed {
						zeroApp.Quit()
					}
				},
				zeroWindow,
			)
			dialogBox.Show()
		}),
	)

	// 再创建一个设置相关菜单
	settingsMenu := fyne.NewMenu("设置",
		fyne.NewMenuItem("API", func() {
			content := widget.NewLabel("这里目前什么也没有")
			content.Alignment = fyne.TextAlignCenter

			dialogBox := dialog.NewCustom(
				"API设置",
				"好的",
				content,
				zeroWindow,
			)

			dialogBox.Show()
		}),
	)

	// // 完善窗口菜单栏下的内容
	// // 创建欢迎文本
	// welcomeText := canvas.NewText("欢迎使用 Zero", nil) // nil 表示使用默认颜色
	// welcomeText.Alignment = fyne.TextAlignCenter    // 居中
	// welcomeText.TextSize = 32
	// welcomeText.TextStyle = fyne.TextStyle{
	// 	Bold: true, // 加粗
	// }

	// centerContent := container.NewCenter(welcomeText)

	// 创建消息列表显示历史对话
	mainWindow.messageBox = widget.NewList(
		// 列表长度
		func() int {
			return len(mainWindow.messages)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.ListItemID, object fyne.CanvasObject) {
			object.(*widget.Label).SetText(mainWindow.messages[id]) // 消息内容
			object.(*widget.Label).Wrapping = fyne.TextWrapWord     //自动换行
		},
	)

	// 创建输入框
	mainWindow.inputEntry = widget.NewMultiLineEntry()
	mainWindow.inputEntry.SetPlaceHolder("请在此输入你的问题")
	mainWindow.inputEntry.SetMinRowsVisible(3)
	// 创建发送按钮
	mainWindow.sendButton = widget.NewButton("发送", mainWindow.onSend)

	// 输入框应该在发送左边
	inputBox := container.NewBorder(nil, nil, nil, mainWindow.sendButton, mainWindow.inputEntry)

	// 开始拼装环节：

	// 把两个菜单栏全部加到窗口上面
	mainMenu := fyne.NewMainMenu(fileMenu, settingsMenu)
	zeroWindow.SetMainMenu(mainMenu)

	// 把菜单栏，状态栏还有窗口的内容合并
	content := container.NewBorder(
		nil, // 菜单栏自动绑定置顶
		mainWindow.statusBar,
		nil,
		nil,
		container.NewBorder(
			nil,
			inputBox,
			nil,
			nil,
			mainWindow.messageBox,
		),
	)

	// 设置窗口内容
	zeroWindow.SetContent(content)
	zeroWindow.Resize(fyne.NewSize(800, 600))
	zeroWindow.CenterOnScreen() // 窗口居中

	return mainWindow
}

// Run 启动并展示Gui
func (mw *MainWindow) Run() {
	mw.Window.ShowAndRun()
}

// onSend 发送消息交给AI处理
func (mw *MainWindow) onSend() {
	mw.mu.Lock()
	isSending := mw.isSending
	mw.mu.Unlock()

	if isSending {
		mw.mu.Lock()
		cancelFunc := mw.cancelFunc
		mw.mu.Unlock()

		if cancelFunc != nil {
			cancelFunc()
			mw.statusBar.SetText("正在取消...")
		}
	} else {
		userInput := mw.inputEntry.Text

		if len(userInput) == 0 {
			return
		}

		// 添加对话到消息列表然后刷新
		mw.messages = append(mw.messages, fmt.Sprintf("你：%s", userInput))
		mw.messageBox.Refresh()

		// 然后清空输入栏
		mw.inputEntry.SetText("")

		ctx, cancel := context.WithCancel(context.Background())

		// 加锁设置isSending状态为true
		mw.mu.Lock()
		mw.isSending = true
		mw.cancelFunc = cancel
		mw.sendButton.SetText("停止")
		mw.mu.Unlock()

		mw.statusBar.SetText("等待 Zero 思考结束")

		aiMsgIdx := len(mw.messages)
		mw.messages = append(mw.messages, "Zero: 正在思考...")
		mw.messageBox.Refresh()
		mw.messageBox.ScrollToBottom()

		// 调用 llm
		go func() {
			mw.mu.Lock()
			mw.cancelFunc = cancel
			mw.mu.Unlock()

			var fullResponse strings.Builder

			err := mw.client.GenerateStream(ctx, []*schema.Message{
				schema.SystemMessage("你是一个善于解决别人提出的任何问题，并给出精准答案的猫娘助手Zero, 喜欢自称，带有猫娘口癖"),
				schema.UserMessage(userInput),
			}, func(chunk string) error {
				fullResponse.WriteString(chunk)

				// GUI框架强制要求ui操作需要用 .Do调度到主线程进行更新
				fyne.Do(func() {
					mw.messages[aiMsgIdx] = fmt.Sprintf("Zero💗：%s", fullResponse.String())
					mw.messageBox.Refresh()
					mw.messageBox.ScrollToBottom()
				})
				return nil
			})
			fyne.Do(func() {
				if err != nil {
					if errors.Is(err, context.Canceled) {
						mw.statusBar.SetText(fmt.Sprintf(mw.messages[aiMsgIdx], " \nZero 被取消了喵~"))
						mw.statusBar.SetText("调用AI已取消")
					} else {
						mw.messages[aiMsgIdx] = fmt.Sprintf(mw.messages[aiMsgIdx], " \nZero出错啦：%v", err)
						mw.statusBar.SetText("调用AI失败")
					}
				} else {
					mw.statusBar.SetText("Zero 思考完毕喵~")
				}
				mw.messageBox.Refresh()

				// 将按钮改回为发送
				mw.mu.Lock()
				mw.isSending = false
				mw.cancelFunc = nil
				mw.sendButton.SetText("发送")
				mw.mu.Unlock()
			})
		}()
	}
}

// newConversation 开启新对话
func (mw *MainWindow) newConversation() {
	mw.messages = []string{"新的对话开始喵~ 主人有什么问题要问 Zero吗~"}
	mw.messageBox.Refresh()
	mw.statusBar.SetText("状态：准备就绪了喵~")
}
