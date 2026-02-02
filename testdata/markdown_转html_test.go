package testdata

import (
	"fmt"
	"strings"
	"testing"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"github.com/russross/blackfriday"
)

func TestMarkdown(t *testing.T) {

	// markdown转html
	unsafe := blackfriday.MarkdownCommon([]byte("### 你好\n ```" +
		"python\nprint('你好')\n```\n - 123 \n \n" +
		"<script>alert(123)</script>\n\n ![图片](http://xxx.com)"))
	fmt.Println(string(unsafe))
	// html获取文本内容，xss过滤
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(string(unsafe)))
	//fmt.Println(doc.Text())
	doc.Find("script").Remove()
	fmt.Println(doc.Text())
	// 创建转换器 html转markdown
	converter := md.NewConverter("", true, nil)
	html, _ := doc.Html()
	markdown, err := converter.ConvertString(html)
	fmt.Println(markdown, err)
}
