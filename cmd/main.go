package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"

	"zerok-injector/internal/config"
	"zerok-injector/pkg/cert"
	"zerok-injector/pkg/server"
	"zerok-injector/pkg/storage"
	"zerok-injector/pkg/utils"
)

//TODO:
// Create local setup and test redis integration.
// Integrate iris into the repo.
// Add a script for stage and prod as well.
// What if the processes change in a pod based on args to a container?
// Add service name from workload to otel agent.

func main() {

	var cfg config.ZkInjectorConfig
	args := ProcessArgs(&cfg)

	// read configuration from the file and environment variables
	log.Println("args.ConfigPath==", args.ConfigPath)

	if err := cleanenv.ReadConfig(args.ConfigPath, &cfg); err != nil {
		log.Println(err)
	}

	runtimeMap := &storage.ImageRuntimeHandler{ImageRuntimeMap: &sync.Map{}}
	runtimeMap.Init(cfg.Redis)

	// start data collector
	go server.StartServer(runtimeMap)

	if cfg.Debug {
		server.StartDebugWebHookServer(cfg.Webhook, runtimeMap)
	} else {
		// initialize certificates
		caPEM, cert, err := cert.InitializeKeysAndCertificates(cfg.Webhook)
		if err != nil {
			fmt.Println(err)
			panic(err)
		}

		// start mutating webhook
		err = utils.CreateOrUpdateMutatingWebhookConfiguration(caPEM, cfg.Webhook)
		if err != nil {
			msg := fmt.Sprintf("Failed to create or update the mutating webhook configuration: %v\n", err)
			fmt.Println(msg)
			panic(msg)
		}

		// start webhook server
		server.StartWebHookServer(cfg.Webhook, cert, runtimeMap)
	}
}

// Args command-line parameters
type Args struct {
	ConfigPath string
}

// ProcessArgs processes and handles CLI arguments
func ProcessArgs(cfg interface{}) Args {
	var a Args

	f := flag.NewFlagSet("Example server", 1)
	f.StringVar(&a.ConfigPath, "c", "config.yaml", "Path to configuration file")

	fu := f.Usage
	f.Usage = func() {
		fu()
		envHelp, _ := cleanenv.GetDescription(cfg, nil)
		fmt.Fprintln(f.Output())
		fmt.Fprintln(f.Output(), envHelp)
	}

	f.Parse(os.Args[1:])
	return a
}
