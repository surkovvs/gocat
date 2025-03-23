package application

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/surkovvs/gocat/application/comp"
)

func (a *app) gracefulShutdown() {
	wg := sync.WaitGroup{}

	gsDone := make(chan struct{})

	ctx, cancel := context.WithTimeout(a.mtn.shutdownCtx, *a.mtn.shutdownTimeout)
	defer cancel()
	for module, group := range a.modules {
		wg.Add(1)
		go func(module comp.Component, gr sequentialGroup) {
			defer wg.Done()

			if module.Shutdowner().IsReady() {
				module.Shutdowner().SetInProcess()
				if err := module.Shutdowner().Get().Shutdown(ctx); err != nil {
					a.mtn.errFlow <- fmt.Errorf(
						`shutdown module, from group %s, failed: %w`,
						gr, err)
					module.Shutdowner().SetFailed()
					return
				}
				module.Shutdowner().SetDone()
			}
		}(module, group)
	}

	go func() {
		wg.Wait()
		close(gsDone)
	}()

SelLabel:
	select {
	case <-gsDone:
		a.logger.Info(`graceful shutdown finished for all bacground runners`,
			"application", a.Meta.Name)
		for module, group := range a.modules {
			if module.Shutdowner().IsInProcess() {
				a.logger.Info(`graceful shutdown for cron module still in process, app shutdown time goes to max limit`,
					"application", a.Meta.Name,
					`group`, group)
				<-ctx.Done()
				break SelLabel
			}
		}

	case <-ctx.Done():
		a.logger.Error(`graceful shutdown timeout exeeded`,
			"application", a.Meta.Name)
	}

	os.Exit(a.mtn.exitCode)
}
