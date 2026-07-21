package admin

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/CDX-1/pocketry/internal/auth"
	"github.com/CDX-1/pocketry/internal/db"
	"github.com/CDX-1/pocketry/internal/instance"
)

type CheckResult struct {
	Name string
	Err  error
}

func Doctor(ctx context.Context, inst *instance.Instance) []CheckResult {
	results := make([]CheckResult, 0, 6)

	results = append(results, CheckResult{
		Name: "Access-token secret",
		Err:  checkAccessTokenSecret(inst),
	})

	results = append(results, CheckResult{
		Name: "Opaque key material",
		Err:  checkOpaqueKeyMaterial(inst),
	})

	results = append(results, CheckResult{
		Name: "Database",
		Err:  checkDatabase(ctx, inst),
	})

	results = append(results, CheckResult{
		Name: "Listen address",
		Err:  checkListenAddress(inst),
	})

	return results
}

func checkAccessTokenSecret(inst *instance.Instance) error {
	secret, err := inst.LoadAccessTokenSecret()
	if err != nil {
		return err
	}

	if err := auth.SetAccessTokenSecret(secret); err != nil {
		return fmt.Errorf("configure access-token secret: %w", err)
	}

	return nil
}

func checkOpaqueKeyMaterial(inst *instance.Instance) error {
	keyMaterial, err := inst.LoadOpaqueServerKeyMaterial()
	if err != nil {
		return err
	}

	_, err = auth.NewBytemareOpaqueServer(
		auth.DefaultOpaqueServerIdentity,
		keyMaterial,
	)
	if err != nil {
		return fmt.Errorf("initialize OPAQUE server: %w", err)
	}

	return nil
}

func checkDatabase(ctx context.Context, inst *instance.Instance) error {
	store, err := db.Open(inst.DatabasePath())
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer func() {
		_ = store.Close()
	}()

	if err := store.DB.PingContext(ctx); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}

	if store.Q == nil {
		return fmt.Errorf("database queries are unavailable")
	}

	return nil
}

func checkListenAddress(inst *instance.Instance) error {
	address := net.JoinHostPort(
		inst.Config.Server.Host,
		strconv.Itoa(inst.Config.Server.Port),
	)

	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", address, err)
	}

	return listener.Close()
}