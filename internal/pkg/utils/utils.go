package utils

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func HandleSignals(cancel context.CancelFunc) {
	signalChanel := make(chan os.Signal, 1)
	signal.Notify(signalChanel, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	go func() {
		select {
		case <-signalChanel:
			fmt.Printf("Shutting down\n")
			cancel()
			return
		}
	}()

}

func WaitCancel(ctx context.Context) {
	select {
	case <-ctx.Done():
		fmt.Printf("Waiting for background tasks to finish")
		return
	}
}
