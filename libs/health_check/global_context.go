package health_check

import (
	"fmt"
	"sync"
)

type HealthContext struct {
	http HTTPServices
	sftp SFTPServices
}

func (context *HealthContext) AddSftp(s SFTPServices) {
	context.sftp = s
}

func (context *HealthContext) AddHttp(h HTTPServices) {
	context.http = h
}

func (context *HealthContext) RunAll() {
	wg := sync.WaitGroup{}

	fmt.Printf("%v", context.http)
	fmt.Printf("%v", context.sftp)

	//wg.Add(1)
	//go run(&wg, context.sftp.RunTesting)

	wg.Add(1)
	go run(&wg, context.http.RunHttpCrons)

	wg.Wait()
}

func run(wg *sync.WaitGroup, fn func()) {
	defer wg.Done()
	fn()
}
