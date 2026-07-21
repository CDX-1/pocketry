package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/CDX-1/pocketry/internal/auth"
	"github.com/CDX-1/pocketry/internal/db"
	"github.com/CDX-1/pocketry/internal/handlers"
	"github.com/CDX-1/pocketry/internal/instance"
	"github.com/CDX-1/pocketry/internal/middleware"
)

func Run(ctx context.Context, output io.Writer, inst *instance.Instance) error {
	accessTokenTTL, err := time.ParseDuration(inst.Config.Security.AccessTokenTTL)
	if err != nil {
		return fmt.Errorf("parse access token TTL: %w", err)
	}

	accessTokenSecret, err := inst.LoadAccessTokenSecret()
	if err != nil {
		return fmt.Errorf("load access token secret: %w", err)
	}

	if err := auth.SetAccessTokenSecret(accessTokenSecret); err != nil {
		return fmt.Errorf("configure access token secret: %w", err)
	}

	opaqueServerKeyMaterial, err := inst.LoadOpaqueServerKeyMaterial()
	if err != nil {
		return fmt.Errorf("load OPAQUE server key material: %w", err)
	}

	opaqueServer, err := auth.NewBytemareOpaqueServer(
		auth.DefaultOpaqueServerIdentity,
		opaqueServerKeyMaterial,
	)
	if err != nil {
		return fmt.Errorf("initialize OPAQUE server: %w", err)
	}

	store, err := db.Open(inst.DatabasePath())
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer store.Close()

	mux := handlers.RegisterRoutes(
		store.Q,
		opaqueServer,
		accessTokenTTL,
	)

	cors := middleware.NewCORS(
		inst.Config.CORS.AllowedOrigins,
		inst.Config.CORS.AllowFirefoxExtensions,
	)

	address := net.JoinHostPort(
		inst.Config.Server.Host,
		strconv.Itoa(inst.Config.Server.Port),
	)

	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", address, err)
	}
	defer listener.Close()

	httpServer := &http.Server{
		Handler: cors.Handler(mux),
	}

	return serve(ctx, output, httpServer, listener)
}

func serve(
	ctx context.Context,
	output io.Writer,
	httpServer *http.Server,
	listener net.Listener,
) error {
	serverErr := make(chan error, 1)

	go func() {
		err := httpServer.Serve(listener)

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
			return
		}

		serverErr <- nil
	}()

	fmt.Fprintf(output, "Pocketry server is running at http://%s\n", listener.Addr())

	select {
	case err := <-serverErr:
		if err != nil {
			return fmt.Errorf("serve HTTP: %w", err)
		}

		return nil

	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			10*time.Second,
		)
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown HTTP server: %w", err)
		}

		if err := <-serverErr; err != nil {
			return fmt.Errorf("wait for server shutdown: %w", err)
		}

		return nil
	}
}
