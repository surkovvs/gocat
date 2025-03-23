package main

import (
	"context"
	"log"
	"time"

	"github.com/surkovvs/gocat/application"
	"github.com/surkovvs/gocat/configuration"
	"github.com/surkovvs/gocat/logging"
	"github.com/surkovvs/gocat/logging/logdeafult"
)

func main() {
	cfg, err := configuration.ParseFile(`config.yml`)
	if err != nil {
		log.Fatal(err)
	}
	logger := logdeafult.NewZapDefault(cfg.Logger)
	app := application.New(
		application.WithName("Main app"),
		application.WithLogger(logging.NewZapAdapter(logger)),
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
	app.AddModuleAutoGroup(module1)

	module2 := &moduleInitRunSd{
		cfg: moduleCfg{
			Name: "module2",
			init: elemCfg{
				totalDur: time.Second * 2,
				wantFail: false,
			},
			run: elemCfg{
				totalDur: 0,
				wantFail: false,
			},
			shutdown: elemCfg{
				totalDur: time.Second,
				wantFail: false,
			},
		},
	}
	app.AddModuleAutoGroup(module2)

	module3 := &moduleSd{
		cfg: moduleCfg{
			Name: "module3",
			shutdown: elemCfg{
				totalDur: time.Second / 2,
				wantFail: false,
			},
		},
	}
	app.AddModuleAutoGroup(module3)

	app.Start(context.Background())
}

// func Grpc() {
// 	// Init
// 	listener, _ := net.Listen("tcp", "127.0.0.1")
// 	server := grpc.NewServer()
// 	server.RegisterService()

// 	// Run
// 	server.Serve(listener)

// 	// Shutdown
// 	server.GracefulStop()

// 	server.GetServiceInfo()[""]
// }
