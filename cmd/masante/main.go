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
	"github.com/masante/masante/adapter/sms"
	"github.com/masante/masante/adapter/sqlite"
	"github.com/masante/masante/app"
)

// version is set at build time via LDFLAGS.
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

	db, err := sqlite.Open(filepath.Join(*dataDir, "masante.db"))
	if err != nil {
		log.Fatalf("base de donnees: %v", err)
	}
	defer db.Close()

	if er := sqlite.Migrate(db); er != nil {
		log.Fatalf("migration: %v", er)
	}

	// Driven adapters.
	fmt.Println("[masante] Initialisation des adaptateurs...")
	userRepo := sqlite.NewUserRepo(db)
	sessionRepo := sqlite.NewSessionRepo(db)
	centerRepo := sqlite.NewCenterRepo(db)
	smsConfigRepo := sqlite.NewSMSConfigRepo(db)
	auditRepo := sqlite.NewAuditRepo(db)
	patientRepo := sqlite.NewPatientRepo(db)
	appointmentRepo := sqlite.NewAppointmentRepo(db)
	reminderRepo := sqlite.NewReminderRepo(db)
	hasher := adapter.BcryptHasher{}
	fmt.Println("[masante] Adaptateurs OK (sqlite, bcrypt)")

	// Application services.
	fmt.Println("[masante] Demarrage des services applicatifs...")
	authSvc := app.NewAuthService(userRepo, sessionRepo, hasher, auditRepo)
	setupSvc := app.NewSetupService(centerRepo, userRepo, smsConfigRepo, hasher, auditRepo)
	patientSvc := app.NewPatientService(patientRepo, auditRepo)
	appointmentSvc := app.NewAppointmentService(appointmentRepo, patientRepo, auditRepo)
	userSvc := app.NewUserService(userRepo, sessionRepo, hasher, auditRepo)
	reminderSvc := app.NewReminderService(reminderRepo, appointmentRepo, patientRepo, smsConfigRepo, centerRepo)
	fmt.Println("[masante] Services OK (auth, setup, patient, appointment, user, reminder)")

	// Load SMS provider if configured.
	smsCfg, err := smsConfigRepo.Get(context.Background())
	if err == nil && smsCfg.Enabled && smsCfg.Provider != "" {
		provider, err := sms.NewProvider(*smsCfg)
		if err != nil {
			fmt.Printf("[masante] SMS provider %q: %v (rappels desactives)\n", smsCfg.Provider, err)
		} else {
			reminderSvc.SetProvider(provider)
			fmt.Printf("[masante] SMS provider OK (%s)\n", provider.Name())
		}
	} else {
		fmt.Println("[masante] SMS non configure (rappels desactives)")
	}

	// Driving adapter.
	fmt.Println("[masante] Configuration du serveur HTTP...")
	srv := httpAdapter.NewServer(authSvc, setupSvc, patientSvc, appointmentSvc, userSvc, reminderSvc)
	fmt.Println("[masante] Routes enregistrees")

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", *port),
		Handler:      srv,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Session cleanup — every hour.
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_ = authSvc.CleanExpiredSessions(context.Background())
			}
		}
	}()

	// Reminder scheduler — every 5 minutes.
	go func() {
		fmt.Println("[masante] Scheduler de rappels demarre (intervalle: 5 min)")
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := reminderSvc.GenerateReminders(context.Background()); err != nil {
					log.Printf("[masante] generation rappels: %v", err)
				}
				if err := reminderSvc.ProcessQueue(context.Background()); err != nil {
					log.Printf("[masante] envoi rappels: %v", err)
				}
				if err := reminderSvc.RetryFailed(context.Background()); err != nil {
					log.Printf("[masante] retry rappels: %v", err)
				}
			}
		}
	}()

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
