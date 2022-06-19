package utils

import (
	"context"
	"fmt"
	"github.com/coreos/go-systemd/daemon"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

func GoId() int {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		panic(fmt.Sprintf("cannot get goroutine id: %v", err))
	}
	return id
}

func HandleSignals(cancel context.CancelFunc) {
	signalChanel := make(chan os.Signal, 1)
	signal.Notify(signalChanel, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	go func() {
		select {
		case <-signalChanel:
			log.Printf("Shutting down\n")
			cancel()
			return
		}
	}()

}

func Watchdog(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	interval, err := daemon.SdWatchdogEnabled(false)

	if err != nil || interval == 0 {
		return
	}
	log.Printf("starting systemd watchdog with interval %d", interval)
	for {
		_, err := daemon.SdNotify(false, daemon.SdNotifyWatchdog)
		if err != nil {
			log.Panicf("unable to notify systemd")
		}
		time.Sleep(interval / 3)
		select {
		case <-ctx.Done():
			return
		default:
		}
	}
}

func WatchdogAndWait(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	go Watchdog(ctx, wg)
	wg.Wait()
	log.Printf("exiting")
}

func WaitCancel(ctx context.Context) {
	select {
	case <-ctx.Done():
		log.Printf("Waiting for background tasks to finish")
		return
	}
}
