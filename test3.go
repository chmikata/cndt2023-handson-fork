package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-co-op/gocron/v2"
)

var wg sync.WaitGroup

func main() {
	// go func() {
	// 	sig := make(chan os.Signal, 1)
	// 	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	// 	s := <-sig
	// 	fmt.Printf("Signal received: %s \n", s.String())
	// }()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	ns, err := gocron.NewScheduler()
	if err != nil {
		fmt.Println(err)
		return
	}

	ns.Start()
	job, err := ns.NewJob(
		gocron.CronJob("*/1 * * * *", false),
		gocron.NewTask(task, ctx),
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("job ID: %v\n", job.ID)

	<-ctx.Done()
	ns.Shutdown()
	wg.Wait()
}

func task(sigCtx context.Context) {
	defer wg.Done()
	wg.Add(1)
	ctx, cancel := context.WithTimeout(sigCtx, 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "sleep", "60")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Cancel = func() error {
		fmt.Println("cancel")
		return nil
	}
	cmd.WaitDelay = 10 * time.Second

	fmt.Println("start")
	cmd.Run()
	fmt.Println(cmd.Stdout)
}
