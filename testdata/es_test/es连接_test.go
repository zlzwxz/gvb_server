package elasticsearch

/*func EsConnect() (*elastic.Client, error) {
	sniffOpt := elastic.SetSniff(false)
	host := "http://127.0.0.1:9200"
	c, err := elastic.NewClient(
		elastic.SetURL(host),
		sniffOpt,
		elastic.SetBasicAuth("", ""),
	)
	if err != nil {
		return nil, err // 返回错误而不是致命退出
	}
	return c, nil
}

func TestEs(t *testing.T) {
	client, err := EsConnect()
	if err != nil {
		t.Fatalf("ES连接失败: %v", err)
	}
	defer func() {
		if client != nil {
			// 关闭客户端资源
		}
	}()
}*/
