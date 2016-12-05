package main

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"os"
	"sync"
	"time"
)

var addr = flag.String("addr", "127.0.0.1:3306", "MySQL ip:port")
var user = flag.String("user", "root", "MySQL user")
var password = flag.String("password", "password", "MySQL password")
var count = flag.Int("count", 1000, "connection count")

type result struct {
	id  int
	err error
}

/*
	Make _count_ connnections to MySQL concurrently
*/
func main() {
	flag.Parse()

	startBarrier := make(chan struct{})
	results := make(chan result, *count)
	startWg := sync.WaitGroup{}
	startWg.Add(*count)

	for i := 0; i < *count; i++ {
		go func(id int) {
			startWg.Done()
			select {
			case <-startBarrier:
			}

			sqlDb, err := sql.Open("mysql", fmt.Sprintf("%s:%s@(%s)/?timeout=5s", *user, *password, *addr))
			if nil == err {
				err = sqlDb.Ping()
			}

			results <- result{
				id:  id,
				err: err,
			}
		}(i)
	}

	startWg.Wait()
	fmt.Println("all makers are ready, start connecting")
	close(startBarrier)

	startTime := time.Now()
	hasErr := false
	for i := 0; i < *count; i++ {
		select {
		case result := <-results:
			if nil != result.err {
				hasErr = true
				i_ := i
				fmt.Printf("maker %v error: %v\n", i_, result.err)
			}
		}
	}
	endTime := time.Now()
	fmt.Printf("Elapsed time: %.6fs\n", endTime.Sub(startTime).Seconds())

	if hasErr {
		os.Exit(1)
	}
}
