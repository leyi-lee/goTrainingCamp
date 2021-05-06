package main

import (
	"context"
	"crypto/md5"
	"fmt"
	"golang.org/x/sync/errgroup"
	"io/fs"
	"io/ioutil"
	"log"
	"path/filepath"
	"time"
)

func main()  {
	m, err := MD5All(context.Background(), ".")
	if err != nil {
		log.Fatal(err)
	}

	for k, sum := range m {
		fmt.Printf("%s:\t%x\n", k, sum)
	}
}

func runTest() error {
	time.Sleep(time.Second)
	fmt.Println("exec runTest")
	return nil
}

type result struct {
	path string
	sum  [md5.Size]byte
}

func MD5All(ctx context.Context, root string) (map[string][md5.Size]byte, error) {
	g,ctx := errgroup.WithContext(ctx)
	paths := make(chan string)

	g.Go(func() error {
		defer close(paths)
		return filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.Mode().IsRegular() {
				return nil
			}

			select {
				case paths <- path:
				case <- ctx.Done():
					return ctx.Err()
			}
			return nil
		})
	})

	// 20个goroutine 计算md5，从paths获取文件路径
	c := make(chan result)
	const numDigesters = 20
	for i := 0; i < numDigesters; i++ {
		g.Go(func() error {
			for path := range paths {
				data, err := ioutil.ReadFile(path)
				if err != nil {
					return err
				}

				select {
					case c <- result{path, md5.Sum(data)}:
					case <- ctx.Done():
						return ctx.Err()
				}
			}
			return nil
		})
	}

	go func() {
		g.Wait()
		close(c)
	}()

	m := make(map[string][md5.Size]byte)
	for r := range c {
		m[r.path] = r.sum
	}

	if err := g.Wait();err != nil {
		return nil,err
	}

	return m, nil
}