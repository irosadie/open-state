package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/irosadie/open-state/api/internal/e2efixtures"
	"github.com/irosadie/open-state/api/internal/infrastructure/config"
)

func main() {
	mode := flag.String("mode", "seed", "seed, verify, verify-builder-draft, verify-builder-published, verify-runtime, verify-runtime-suspend, verify-runtime-resume, verify-runtime-retry, or bump-builder-version")
	flag.Parse()

	if os.Getenv("E2E_FIXTURES") != "1" {
		log.Fatal("refusing fixture command without E2E_FIXTURES=1")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	ctx := context.Background()
	pool, err := config.NewPool(ctx, databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	password := os.Getenv("E2E_FIXTURE_PASSWORD")
	switch *mode {
	case "seed", "reset":
		if err := e2efixtures.Seed(ctx, pool, password); err != nil {
			log.Fatal(err)
		}
		printJSON(e2efixtures.Verification{Mode: "seed", Checks: []string{"reset", "seed"}})
	case "bump-builder-version":
		if err := e2efixtures.BumpBuilderVersion(ctx, pool); err != nil {
			log.Fatal(err)
		}
		printJSON(e2efixtures.Verification{Mode: *mode, Checks: []string{"builder version bumped"}})
	default:
		result, err := e2efixtures.Verify(ctx, pool, *mode)
		if err != nil {
			log.Fatal(err)
		}
		printJSON(*result)
	}
}

func printJSON(value e2efixtures.Verification) {
	encoded, err := json.Marshal(value)
	if err != nil {
		log.Fatal(fmt.Sprintf("encode fixture result: %v", err))
	}
	fmt.Println(string(encoded))
}
