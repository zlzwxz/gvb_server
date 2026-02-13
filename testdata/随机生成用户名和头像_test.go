package testdata

import (
	"fmt"
	"github.com/DanPlayer/randomname"
	"github.com/disintegration/letteravatar"
	"github.com/golang/freetype"
	"image/png"
	"os"
	"path"
	"testing"
	"unicode/utf8"
)

func Test_Name_image(t *testing.T) {
	GenerateNameAvatar()
}

// 随机生成头像
func GenerateNameAvatar() {
	dir := "uploads/chat_avatar"
	for _, s := range randomname.AdjectiveSlice {
		DrawImage(string([]rune(s)[0]), dir)
	}
	for _, s := range randomname.PersonSlice {
		DrawImage(string([]rune(s)[0]), dir)
	}
}

// 随机生成图片
func DrawImage(name string, dir string) {
	// 检查并创建目录
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Printf("创建目录失败: %v\n", err)
		return
	}

	// 使用绝对路径或确保相对路径正确
	fontPath := "uploads/system/NotoSansSC-Regular.ttf"

	// 检查字体文件是否存在
	if _, err := os.Stat(fontPath); os.IsNotExist(err) {
		fmt.Printf("字体文件不存在: %s\n", fontPath)
		return
	}

	// 读取字体文件
	fontFile, err := os.ReadFile(fontPath)
	if err != nil {
		fmt.Printf("读取字体文件失败: %v\n", err)
		return
	}

	// 检查文件大小
	if len(fontFile) < 100 {
		fmt.Printf("字体文件太小: %d 字节\n", len(fontFile))
		return
	}

	// 解析字体
	font, err := freetype.ParseFont(fontFile)
	if err != nil {
		fmt.Printf("解析字体失败: %v\n", err)
		return
	}

	options := &letteravatar.Options{
		Font: font,
	}

	// 绘制文字
	firstLetter, _ := utf8.DecodeRuneInString(name)
	img, err := letteravatar.Draw(140, firstLetter, options)
	if err != nil {
		fmt.Printf("绘制头像失败: %v\n", err)
		return
	}

	// 保存
	filePath := path.Join(dir, name+".png")
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Printf("创建文件失败: %v\n", err)
		return
	}
	defer file.Close()

	err = png.Encode(file, img)
	if err != nil {
		fmt.Printf("保存图片失败: %v\n", err)
		return
	}

	fmt.Printf("成功生成头像: %s\n", filePath)
}
