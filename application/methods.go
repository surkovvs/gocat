package application

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"sync"

	"github.com/surkovvs/gocat/application/comp"
)

func (a *app) Start(ctx context.Context) {
	a.logger.Debug("App started", "application", a.Meta.Name)
	defer a.logger.Debug("App finished", "application", a.Meta.Name)

	a.mtn.initRunCtx, a.mtn.initRunCancel = context.WithCancel(ctx)

	groups := make(map[sequentialGroup][]comp.Component)
	for comp, group := range a.modules {
		groups[group] = append(groups[group], comp)
	}

	wg := sync.WaitGroup{}
	for group, comps := range groups {
		wg.Add(1)
		go func(comps []comp.Component, gr sequentialGroup) {
			defer wg.Done()

			var groupInitCtx context.Context
			if a.mtn.initTimeout != nil {
				var cancel context.CancelFunc
				groupInitCtx, cancel = context.WithTimeout(a.mtn.initRunCtx, *a.mtn.initTimeout)
				defer cancel()
			} else {
				groupInitCtx = a.mtn.initRunCtx
			}
			for num, comp := range comps {
				if comp.Initializer().IsReady() {
					comp.Initializer().SetInProcess()
					a.logger.Debug("Module initialization",
						"application", a.Meta.Name,
						`group`, gr,
						`module`, num)

					if err := comp.Initializer().Get().Init(groupInitCtx); err != nil {
						a.mtn.errFlow <- fmt.Errorf(
							`initializing module %d, from group %s, failed: %w`,
							num, gr, err)
						comp.Initializer().SetFailed()
						return
					}
					comp.Initializer().SetDone()
				}
			}
			for num, comp := range comps {
				if comp.Runner().IsReady() &&
					(comp.Initializer().IsDone() || !comp.IsInitializer()) {
					comp.Runner().SetInProcess()
					a.logger.Debug("Module running",
						"application", a.Meta.Name,
						`group`, gr,
						`module`, num)

					if err := comp.Runner().Get().Run(a.mtn.initRunCtx); err != nil {
						a.mtn.errFlow <- fmt.Errorf(
							`running module %d, from group %s, failed: %w`,
							num, gr, err)
						comp.Initializer().SetFailed()
						return
					}
					comp.Initializer().SetDone()
				}
			}

			var cancelSD context.CancelFunc
			groupShutdownCtx, cancelSD := context.WithTimeout(a.mtn.shutdownCtx, *a.mtn.shutdownTimeout)
			defer cancelSD()
			for num, comp := range comps {
				if comp.Shutdowner().IsReady() && comp.Runner().IsDone() {
					comp.Shutdowner().SetInProcess()
					a.logger.Debug("Module shutdown",
						"application", a.Meta.Name,
						`group`, gr,
						`module`, num)
					if err := comp.Shutdowner().Get().Shutdown(groupShutdownCtx); err != nil {
						a.mtn.errFlow <- fmt.Errorf(
							`shutdown module %d, from group %s, failed: %w`,
							num, gr, err)
						comp.Shutdowner().SetFailed()
						return
					}
					comp.Shutdowner().SetDone()
				}
			}
		}(comps, group)
	}

	wg.Wait()

	for module := range a.modules {
		if module.Shutdowner().IsReady() || module.Shutdowner().IsInProcess() {
			<-a.mtn.shutdownDone
		}
	}
}

// отменяет контексты всех запущенных компонентов и инициализаторов, запускает все шатдауны
func (a *app) Shutdown(ctx context.Context) error {
	return nil
}

func (a *app) AddModuleToGroup(module any, group string) {
	m := comp.DefineComponent(module)
	if !m.IsValid() {
		a.logger.Error(`module addition, module does not implement valid methods`,
			"application", a.Meta.Name,
			`unapplyed`, reflect.ValueOf(module).Type().Name(),
			`group`, group)
		return
	}
	a.modules[m] = sequentialGroup(group)
}

// component must be a pointer to ctructure?
func (a *app) AddModuleAutoGroup(module any) {
	a.Meta.groupCounter += 1
	group := `Group: ` + strconv.Itoa(a.Meta.groupCounter)
	a.AddModuleToGroup(module, group)
}

// инициализирует все компонены, которые неинициализированны
// Внутри группы запуск выполнен последовательно,
// func (a *app) Init(ctx context.Context) error {
// 	a.logger.Debug("Init started", "application", a.Meta.Name)
// 	defer a.logger.Debug("Init finished", "application", a.Meta.Name)

// 	if a.mtn.initTimeout != nil {
// 		ctx, a.mtn.initRunCancel = context.WithTimeout(ctx, *a.mtn.initTimeout)
// 	} else {
// 		ctx, a.mtn.initRunCancel = context.WithCancel(ctx)
// 	}

// 	var errCounter int64
// 	wg := sync.WaitGroup{}

// 	groups := make(map[sequentialGroup][]comp.Component)
// 	for comp, group := range a.modules {
// 		groups[group] = append(groups[group], comp)
// 	}
// 	for group, comps := range groups {
// 		wg.Add(1)
// 		go func(comps []comp.Component, gr sequentialGroup) {
// 			defer wg.Done()
// 			for num, comp := range comps {
// 				if comp.Initializer().IsReady() {
// 					comp.Initializer().SetInProcess()
// 					if err := comp.Initializer().Get().Init(ctx); err != nil {
// 						a.mtn.errFlow <- fmt.Errorf(
// 							`app %s, initializing module %d, from group %s, failed: %w`,
// 							a.Meta.Name, num, gr, err)
// 						comp.Initializer().SetFailed()
// 						atomic.AddInt64(&errCounter, 1)
// 						break
// 					}
// 					comp.Initializer().SetDone()
// 				}
// 			}
// 		}(comps, group)
// 	}

// 	wg.Wait()

// 	if int(errCounter) == len(groups) {
// 		return fmt.Errorf(`app %s, all groups initializations has been failed`, a.Meta.Name)
// 	}

// 	return nil
// }

// // запускает все компоненты которые не запущены
// func (a *app) Run(ctx context.Context) error {
// 	a.logger.Debug("Run started", "application", a.Meta.Name)
// 	defer a.logger.Debug("Run finished", "application", a.Meta.Name)

// 	if a.mtn.initTimeout != nil {
// 		ctx, a.mtn.initRunCancel = context.WithTimeout(ctx, *a.mtn.initTimeout)
// 	} else {
// 		ctx, a.mtn.initRunCancel = context.WithCancel(ctx)
// 	}

// 	var errCounter int64
// 	wg := sync.WaitGroup{}

// 	groups := make(map[sequentialGroup][]comp.Component)
// 	for comp, group := range a.modules {
// 		groups[group] = append(groups[group], comp)
// 	}
// 	for group, comps := range groups {
// 		wg.Add(1)
// 		go func(comps []comp.Component, gr sequentialGroup) {
// 			defer wg.Done()
// 			for num, comp := range comps {
// 				if comp.IsRunner() {
// 					comp.Runner().SetInProcess()
// 					if err := comp.Runner().Get().Run(ctx); err != nil {
// 						a.mtn.errFlow <- fmt.Errorf(
// 							`app %s, running module %d, from group %s, failed: %w`,
// 							a.Meta.Name, num, gr, err)
// 						comp.Runner().SetFailed()
// 						atomic.AddInt64(&errCounter, 1)
// 						break
// 					}
// 					comp.Runner().SetDone()
// 				}
// 			}
// 		}(comps, group)
// 	}

// 	wg.Wait()

// 	if int(errCounter) == len(groups) {
// 		return fmt.Errorf(`app %s, all groups running has been failed`, a.Meta.Name)
// 	}

// 	return nil
// }

// func (a *app) AddHealthchecker(healthchecker interfaces.Healthchecker) {
// }

// func (a *app) AddInitializer(initializer interfaces.Initializer) {
// }

// func (a *app) AddRunner(healthchecker interfaces.Runner) {
// }

// func (a *app) AddShutdowner(initializer interfaces.Shutdowner) {
// }
