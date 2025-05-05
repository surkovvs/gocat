package main

import (
	"context"
	"log"
	"time"

	"github.com/surkovvs/gocat/catapp"
	"github.com/surkovvs/gocat/catcfg"
	catdefzap "github.com/surkovvs/gocat/catdef/loggers/catdef_zap"
	"github.com/surkovvs/gocat/catlog"
)

func main() {
	cfg, err := catcfg.ParseFile(`config.yml`)
	if err != nil {
		log.Fatal(err)
	}

	logger := catdefzap.NewZapDefault(cfg)
	app := catapp.New(
		catapp.WithName("Main app"),
		catapp.WithLogger(catlog.NewZapAdapter(logger)),
		catapp.WithInitTimeout(time.Second),
	)

	module1 := &moduleInitRun{
		cfg: moduleCfg{
			Name: "module1",
			init: elemCfg{
				totalDur: time.Second,
				wantFail: false,
			},
			run: elemCfg{
				totalDur: 0,
				wantFail: false,
			},
		},
	}
	app.AddModuleToGroup("group1", "moduleInitRun", module1)

	module2_1 := &moduleInitRunSd{
		cfg: moduleCfg{
			Name: "module2_1",
			init: elemCfg{
				totalDur: time.Second * 2,
				wantFail: false,
			},
			run: elemCfg{
				totalDur: time.Second / 2,
				wantFail: false,
			},
			shutdown: elemCfg{
				totalDur: time.Second,
				wantFail: false,
			},
		},
	}
	app.AddModuleToGroup("Ordercheck", "moduleInitRunSd", module2_1)

	module2_2 := &moduleInitRunSd{
		cfg: moduleCfg{
			Name: "module2_2",
			init: elemCfg{
				totalDur: time.Second / 2,
				wantFail: false,
			},
			run: elemCfg{
				totalDur: time.Second * 2,
				wantFail: false,
			},
			shutdown: elemCfg{
				totalDur: time.Second,
				wantFail: false,
			},
		},
	}
	app.AddModuleToGroup("Ordercheck", "moduleInitRunSd", module2_2)

	module3 := &moduleSd{
		cfg: moduleCfg{
			Name: "module3",
			shutdown: elemCfg{
				totalDur: time.Second / 2,
				wantFail: false,
			},
		},
	}
	app.AddModuleToGroup("group3", "moduleSd", module3)

	module4 := &moduleInitRunSd{
		cfg: moduleCfg{
			Name: "module4",
			init: elemCfg{
				totalDur: time.Second / 2,
				wantFail: false,
			},
			run: elemCfg{
				totalDur: time.Second * 2,
				wantFail: false,
			},
			shutdown: elemCfg{
				totalDur: time.Second,
				wantFail: false,
			},
		},
	}
	app.AddModuleToGroup("global", "moduleInitRunSd", module4)

	app.Start(context.Background())
}
