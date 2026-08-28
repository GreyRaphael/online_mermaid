package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"golang.org/x/crypto/bcrypt"

	"online_mermaid/internal/auth"
	"online_mermaid/internal/config"
	"online_mermaid/internal/db"
	"online_mermaid/internal/server"
	"online_mermaid/internal/webui"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "hash-password" {
		fmt.Fprint(os.Stderr, "Password: ")
		if err := hashPassword(os.Stdin, os.Stdout); err != nil {
			fmt.Fprintln(os.Stderr, "hash password:", err)
			os.Exit(1)
		}
		return
	}

	if err := runServer(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runServer(args []string) error {
	cfg, err := config.Parse(args)
	if err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	database, err := db.Open(cfg.DBPath)
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}
	defer database.Close()

	if err := database.EnsureDefaultDiagram(context.Background()); err != nil {
		slog.Warn("ensure default diagram error", "error", err)
	}

	assets, err := webui.Dist()
	if err != nil {
		return fmt.Errorf("frontend assets error: %w", err)
	}

	sessions := auth.NewStore(cfg.SessionTTL, cfg.SecureCookie)
	authHandler := auth.NewHandler(cfg.Username, cfg.PasswordHash, sessions, auth.NewLoginLimiter(10, time.Minute))

	app := server.New(server.Config{
		AppConfig: cfg,
		Auth:      authHandler,
		Sessions:  sessions,
		DB:        database,
		Assets:    assets,
	})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go sessions.RunCleanup(ctx)

	httpServer := app.HTTPServer()
	serveErr := make(chan error, 1)
	go func() {
		slog.Info("online mermaid starting", "addr", cfg.Addr, "db", cfg.DBPath, "admin_user", cfg.Username)
		serveErr <- httpServer.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		stop()
	case err := <-serveErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("online mermaid stopped unexpectedly", "error", err)
			return err
		}
		slog.Info("online mermaid stopped")
		return nil
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
	}
	<-serveErr
	slog.Info("online mermaid stopped")
	return nil
}

func hashPassword(input io.Reader, output io.Writer) error {
	scanner := bufio.NewScanner(input)
	scanner.Buffer(make([]byte, 1024), 64<<10)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return err
		}
		return errors.New("no password provided")
	}
	password := strings.TrimSuffix(scanner.Text(), "\r")
	if password == "" {
		return errors.New("password cannot be empty")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(output, string(hash))
	return err
}
