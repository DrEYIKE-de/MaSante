package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/masante/masante/adapter"
	httpAdapter "github.com/masante/masante/adapter/http"
	"github.com/masante/masante/adapter/sqlite"
	"github.com/masante/masante/app"
)

// version : version of the app
var version = "dev"

func main() {
	port := flag.Int("port", 8080, "Port HTTP")
	dataDir := flag.String("data-dir", "./masante-data", "Repertoire des donnees")
	showVersion := flag.Bool("version", false, "Afficher la version")
	flag.Parse()

	if *showVersion {
		fmt.Printf("MaSante %s\n", version)
		os.Exit(0)
	}

	if err := os.MkdirAll(*dataDir, 0755); err != nil {
		log.Fatalf("impossible de creer %s: %v", *dataDir, err)
	}

	// --- Driven adapters (infrastructure) ---

	db, err := sqlite.Open(filepath.Join(*dataDir, "masante.db"))
	if err != nil {
		log.Fatalf("base de donnees: %v", err)
	}
	defer db.Close()

	if err := sqlite.Migrate(db); err != nil {
		log.Fatalf("migration: %v", err)
	}

	userRepo := sqlite.NewUserRepo(db)
	sessionRepo := sqlite.NewSessionRepo(db)
	centerRepo := sqlite.NewCenterRepo(db)
	smsConfigRepo := sqlite.NewSMSConfigRepo(db)
	auditRepo := sqlite.NewAuditRepo(db)
	hasher := adapter.BcryptHasher{}

	// --- Application services ---

	authService := app.NewAuthService(userRepo, sessionRepo, hasher, auditRepo)
	setupService := app.NewSetupService(centerRepo, userRepo, smsConfigRepo, hasher, auditRepo)

	// --- Driving adapter (HTTP) ---

	srv := httpAdapter.NewServer(authService, setupService)

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", *port),
		Handler:      srv,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// --- Background tasks ---

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Nettoyage des sessions expirees toutes les heures
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_ = authService.CleanExpiredSessions(context.Background())
			}
		}
	}()

	// --- Start ---

	go func() {
		time.Sleep(400 * time.Millisecond)
		openBrowser(fmt.Sprintf("http://localhost:%d", *port))
	}()

	go func() {
		log.Printf("MaSante %s — http://localhost:%d", version, *port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("serveur: %v", err)
		}
	}()

	// --- Graceful shutdown ---

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("arret...")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("arret force: %v", err)
	}
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}
