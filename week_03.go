package main

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/sync/errgroup"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	g, ctx := errgroup.WithContext(context.Background())

	closeChan := make(chan string)
	srv := &http.Server{Addr: ":9090"}

	http.HandleFunc("/closeServer", func(writer http.ResponseWriter, request *http.Request) {
		closeChan <- "close"
	})

	// 关闭
	g.Go(func() error {
		fmt.Println("closeServer")
		// 不管是哪种，都需要关闭服务
		select { // 监控阻塞等待
			case <-ctx.Done():
				 fmt.Println("收到context", ctx.Err())
			case <-closeChan:
				 fmt.Println("手动关闭")
		}
		return srv.Shutdown(ctx)
	})

	g.Go(func() error {
		fmt.Println("signal")
		c := make(chan os.Signal, 0)
		signal.Notify(c, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
		select {
			case <-ctx.Done():
				return ctx.Err()
			case <-c:
				return errors.New("进程信号退出")
		}
	})

	g.Go(func() error {
		fmt.Println("start")
		return srv.ListenAndServe()
		//return errors.New("启动服务失败")
	})

	if err := g.Wait(); err != nil {
		fmt.Println("错误退出:", err)
	}
}
