package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"testing"
)

func BenchmarkGetFriends(b *testing.B) {
	// 模拟设置包含token的Cookie
	cookie := &http.Cookie{
		Name:  "token",
		Value: "Bearer valid_token",
	}

	var wg sync.WaitGroup
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req, err := http.NewRequest("GET", "http://localhost:8080/api/user/friends", nil)
			if err != nil {
				fmt.Printf("创建请求失败: %v\n", err)
				return
			}
			// 添加Cookie到请求中
			req.AddCookie(cookie)

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				fmt.Printf("发送请求失败: %v\n", err)
				return
			}
			defer resp.Body.Close()

			_, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("读取响应失败: %v\n", err)
				return
			}
		}()
	}
	wg.Wait()
}
