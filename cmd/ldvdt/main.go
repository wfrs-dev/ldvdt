// Command ldvdt is the ldvd keyboard layout trainer: it starts a local web
// server (port 10987), serves the SPA, and opens the default browser.
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	hefconfig "github.com/wfrs-dev/hefesto/app/config"
	"github.com/wfrs-dev/hefesto/app/health"
	"github.com/wfrs-dev/hefesto/app/infra/server"

	"github.com/wfrs-dev/ldvdt/internal/config"
	"github.com/wfrs-dev/ldvdt/internal/corpus"
	"github.com/wfrs-dev/ldvdt/internal/domain"
	"github.com/wfrs-dev/ldvdt/internal/infra/browser"
	"github.com/wfrs-dev/ldvdt/internal/infra/storage/sqlite"
	"github.com/wfrs-dev/ldvdt/internal/trainer"
	"github.com/wfrs-dev/ldvdt/internal/web/static"
)

// Version is overridable at build time via -ldflags "-X main.Version=...".
var Version = "0.1.0-dev"

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg := config.Default()

	if err := cfg.EnsureDBDir(); err != nil {
		return fmt.Errorf("prepare db dir: %w", err)
	}

	store, err := sqlite.Open(cfg.DBPath)
	if err != nil {
		return fmt.Errorf("open storage: %w", err)
	}
	defer store.Close()

	wordCorpus, err := corpus.Load()
	if err != nil {
		return fmt.Errorf("load corpus: %w", err)
	}

	kb := domain.DefaultKeyboard()
	levels := domain.BuildLevels(kb)

	hefCfg := &hefconfig.Config{
		Server: hefconfig.Server{
			Port:     cfg.Port,
			LogLevel: cfg.LogLevel,
		},
	}

	srv, err := server.New(hefCfg)
	if err != nil {
		return fmt.Errorf("create server: %w", err)
	}

	srv.AddModule(health.New(Version, time.Now().Unix()))
	srv.EchoInstance().StaticFS("/static", static.FS)

	svc := trainer.NewService(store, wordCorpus, levels)
	srv.AddModule(trainer.NewModule(svc, kb))

	addr := fmt.Sprintf("http://127.0.0.1:%d", cfg.Port)
	if cfg.OpenBrowser {
		go func() {
			waitForServer(addr)
			if err := browser.Open(addr); err != nil {
				log.Printf("could not open browser: %v", err)
			}
		}()
	}

	log.Printf("ldvdt %s listening on %s (db: %s)", Version, addr, cfg.DBPath)
	return srv.Start()
}

// waitForServer blocks until the server responds or a timeout elapses.
func waitForServer(addr string) {
	client := &http.Client{Timeout: time.Second}
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := client.Get(addr + "/status/health")
		if err == nil {
			resp.Body.Close()
			return
		}
		time.Sleep(150 * time.Millisecond)
	}
	// Fall through: give up quietly; the user can open the browser manually.
	_, _ = fmt.Fprintf(os.Stderr, "server did not answer at %s; open it manually\n", addr)
}
