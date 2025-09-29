package internal

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"gitlab.mvk.com/go/vkgo/projects/donut/shooter/internal/action"
)

type Shooter struct {
	ShooterName string
	Parameters  Parameters
	Action      action.Action
}

func (s *Shooter) Launch() {
	var (
		wg              sync.WaitGroup
		terminationChan = make(chan struct{})
	)

	wg.Add(s.Parameters.Workers)
	fmt.Printf("[%v] shooting started\n", time.Now().Format("2006-01-02 15:04:05"))

	tStart := time.Now()

	for i := 0; i < s.Parameters.Workers; i++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-terminationChan:
					return
				default:
					action.DoWrapped(s.Action, s.Parameters.AvgTiming)
				}
			}
		}()
	}

	go func() {
		ticker := time.NewTicker(s.Parameters.StatsPeriod)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.Action.PrintCurrentAndFlush(s.ShooterName, s.Parameters.StatsPeriod)
				s.Action.PrintTotal(s.ShooterName, time.Since(tStart))
				fmt.Printf("----------------------------------------------------------------\n")
			case <-terminationChan:
				return
			}
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT, os.Interrupt)

		sig := <-sigChan

		log.Printf("Получен сигнал: %v. Начинаем graceful shutdown...", sig)

		cancel()
	}()

	go s.GracefulShutdown(ctx)

	time.Sleep(s.Parameters.ShootingDuration)
	close(terminationChan)

	wg.Wait()
	cancel()

	s.Action.Close()
}

func (s *Shooter) GracefulShutdown(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("[%v] shooting finished\n", time.Now().Format("2006-01-02 15:04:05"))
			s.Action.PrintTotal(s.ShooterName, s.Parameters.ShootingDuration)

			os.Exit(0)
		}
	}
}
