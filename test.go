package logic

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
)

type Cron struct {
	mutex    sync.Mutex
	proccess *os.Process
}

func NewCron(cronStr string) *Cron {
	return &Cron{}
}

func (c *Cron) Start() error {
	idleCronFinished := make(chan struct{})
	stopCron := make(chan struct{})
	go func() {
		defer close(stopCron)
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGINT)

		<-signalChan
		slog.Info("Signal received.")
		c.mutex.Lock()
		if c.proccess != nil {
			c.proccess.Kill()
		}
		c.mutex.Unlock()
	}()

	go func() {
		defer close(idleCronFinished)
		for {
			select {
			case <-stopCron:
				slog.Info("Cron stopped.")
				return
			default:
				slog.Info("Cron loop.")
				cmd := exec.Command("sleep", "60")

				err := cmd.Start()
				if err != nil {
					slog.Error(fmt.Sprintf("Failed to start cron: %v", err))
				}
				c.mutex.Lock()
				c.proccess = cmd.Process
				c.mutex.Unlock()

				err = cmd.Wait()
				if err != nil {
					slog.Error(fmt.Sprintf("Failed to wait cron: %v", err))
				}
				c.mutex.Lock()
				c.proccess = nil
				c.mutex.Unlock()
				return
			}
		}
	}()

	<-idleCronFinished
	return nil
}

func (c *Cron) task() {
	slog.Info("Task Start.")
	defer c.mutex.Unlock()

	c.mutex.Lock()
	if c.proccess != nil {
		slog.Info("Task already running.")
		return
	}
	err := exec.Command("freshclam", "--foreground").Run()
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to run freshclam: %v", err))
		return
	}
	cmd := exec.Command("clamscan", "-r", "/host-fs")
	err = cmd.Start()
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to run cmd: %v", err))
		return
	}
	c.proccess = cmd.Process
	c.mutex.Unlock()

	defer func() {
		c.mutex.Lock()
		c.proccess = nil
		c.mutex.Unlock()
	}()
	err = cmd.Wait()
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to wait cron: %v", err))
		return
	}
	slog.Info("Task Finished.")
}
