package catapp

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"

	"github.com/surkovvs/gocat/catapp/component"
	"github.com/surkovvs/gocat/catapp/compstor"
)

func (a *app) Start(ctx context.Context) {
	a.logger.Debug(`App started`, `application`, a.name)
	defer a.logger.Debug(`App finished`, `application`, a.name)

	var initRunCtx context.Context
	initRunCtx, a.execution.initRunCancel = context.WithCancel(ctx)

	var initCtx context.Context
	if a.execution.initTimeout != nil {
		var cancel context.CancelFunc
		initCtx, cancel = context.WithTimeout(initRunCtx, *a.execution.initTimeout)
		defer cancel()
	} else {
		initCtx = initRunCtx
	}

	group, err := a.storage.GetGroupByName(privelegedGroupName)
	if err != nil {
		if errors.Is(err, compstor.ErrGroupNotFound) {
			a.logger.Info(`priveleged group not found`,
				"application", a.name)
		} else {
			a.logger.Error(`unexpected error`,
				"application", a.name,
				`group`, group.GetName(),
				"error", err)
		}
	} else {
		a.processInitializers(initCtx, group)
		a.processRunners(initCtx, group)
		a.processShutdowners(initCtx, group)
	}

	wg := sync.WaitGroup{}
	for _, group := range a.storage.GetOrderedGroupList() {
		wg.Add(1)
		go func(group compstor.SequentialGroup) {
			defer wg.Done()
			a.processInitializers(initCtx, group)
			a.processRunners(ctx, group)
			groupShutdownCtx, cancelSD := context.WithTimeout(a.shutdown.ctx, *a.shutdown.timeout)
			defer cancelSD()
			a.processShutdowners(groupShutdownCtx, group)
		}(group)
	}

	wg.Wait()

	close(a.execution.done)
	<-a.shutdown.shutdownDone
}

func (a *app) processInitializers(ctx context.Context, group compstor.SequentialGroup) {
	for _, module := range group.GetComponents() {
		if module.Initializer().TrySetInProcess() {
			a.logger.Debug(`Module initialization`,
				`application`, a.name,
				`group`, group.GetName(),
				`module`, module.Name())

			if err := module.Initializer().Get().Init(ctx); err != nil {
				a.execution.errFlow <- fmt.Errorf(
					`initializing module %s, from group %s, failed: %w`,
					module.Name(), group.GetName(), err)

				module.Initializer().SetFailed()
				return
			}
			module.Initializer().SetDone()
		}
	}
}

func (a *app) processRunners(ctx context.Context, group compstor.SequentialGroup) {
	for _, module := range group.GetComponents() {
		if (module.Initializer().IsDone() || !module.IsInitializer()) &&
			module.Runner().TrySetInProcess() {
			a.logger.Debug(`Module running`,
				`application`, a.name,
				`group`, group.GetName(),
				`module`, module.Name())

			if err := module.Runner().Get().Run(ctx); err != nil {
				a.execution.errFlow <- fmt.Errorf(
					`running module %s, from group %s, failed: %w`,
					module.Name(), group.GetName(), err)

				module.Runner().SetFailed()
				return
			}
			module.Runner().SetDone()
		}
	}
}

func (a *app) processShutdowners(ctx context.Context, group compstor.SequentialGroup) {
	for _, module := range group.GetComponents() {
		if module.Runner().IsDone() && module.Shutdowner().TrySetInProcess() {
			a.logger.Debug(`Module shutdown`,
				`application`, a.name,
				`group`, group.GetName(),
				`module`, module.Name())

			if err := module.Shutdowner().Get().Shutdown(ctx); err != nil {
				a.execution.errFlow <- fmt.Errorf(
					`shutdown module %s, from group %s, failed: %w`,
					module.Name(), group.GetName(), err)

				module.Shutdowner().SetFailed()
				return
			}
			module.Shutdowner().SetDone()
		}
	}
}

func (a *app) AddModuleToGroup(groupName, moduleName string, module any) {
	comp := component.DefineComponent(moduleName, module)
	if !comp.IsValid() {
		a.logger.Error(`module addition`,
			"application", a.name,
			`group`, groupName,
			`module`, moduleName,
			`unapplyed`, reflect.ValueOf(module).Type().Name(),
			`error`, "module does not implement valid methods")
		return
	}
	if err := a.storage.AddComponent(groupName, moduleName, comp); err != nil {
		a.logger.Error(`module addition`,
			"application", a.name,
			`group`, groupName,
			`module`, moduleName,
			`unapplyed`, reflect.ValueOf(module).Type().Name(),
			`error`, err)
	}
}
