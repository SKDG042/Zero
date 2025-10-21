// Package ui 使用Fyne来写GUI界面
package ui

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// MainWindow 主界面
type MainWindow struct {
	App       fyne.App
	Window    fyne.Window
	statusBar *widget.Label // 状态栏
}

func NewMainWindow() *MainWindow {

	// 用fyne创建一个新的窗口应用
	zeroApp := app.NewWithID("zero")
	zeroWindow := zeroApp.NewWindow("Zero Agent")

	mainWindow := &MainWindow{
		App:       zeroApp,
		Window:    zeroWindow,
		statusBar: widget.NewLabel("Zero: Waiting for u..."), // 状态栏
	}

	// 绑定icon
	icon, err := fyne.LoadResourceFromPath("assets/icon.svg")
	if err != nil {
		log.Fatal("绑定Icon失败：", err)
	}
	zeroWindow.SetIcon(icon)

	// 创建左上角的那种菜单栏
	fileMenu := fyne.NewMenu("打开文件",
		fyne.NewMenuItem("新对话", func() {
			content := widget.NewLabel("还没有写好捏~")
			content.Alignment = fyne.TextAlignCenter

			dialogBox := dialog.NewCustom(
				"新的对话",
				"好的",
				content,
				zeroWindow,
			)

			dialogBox.Show()
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

	// 在创建一个设置相关菜单
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

	// 把两个菜单栏全部加到窗口上面
	mainMenu := fyne.NewMainMenu(fileMenu, settingsMenu)
	zeroWindow.SetMainMenu(mainMenu)

	// 完善窗口菜单栏下的内容
	// 创建欢迎文本
	welcomeText := canvas.NewText("欢迎使用 Zero", nil) // nil 表示使用默认颜色
	welcomeText.Alignment = fyne.TextAlignCenter    // 居中
	welcomeText.TextSize = 32
	welcomeText.TextStyle = fyne.TextStyle{
		Bold: true, // 加粗
	}

	centerContent := container.NewCenter(welcomeText)

	// 超级拼装(
	// 把菜单栏，状态栏还有窗口的内容合并
	content := container.NewBorder(
		nil, // 菜单栏自动绑定置顶
		mainWindow.statusBar,
		nil,
		nil,
		centerContent,
	)

	// 设置窗口内容
	zeroWindow.SetContent(content)
	zeroWindow.Resize(fyne.NewSize(800, 600))
	zeroWindow.CenterOnScreen() // 窗口居中

	return mainWindow
}

func (mainWindow *MainWindow) Run() {
	mainWindow.Window.ShowAndRun()
}
