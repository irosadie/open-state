package ratelimit

import (
	"context"
	"testing"
)

func TestTokenBucketBurstThenLimit(t *testing.T) {
	ctx := context.Background()
	rl := NewTokenBucket(Config{Rate: 1000, Burst: 3})

	// Burst of 3 allowed.
	for i := 0; i < 3; i++ {
		ok, err := rl.Allow(ctx, "k")
		if err != nil || !ok {
			t.Fatalf("expected burst request %d allowed, got ok=%v err=%v", i, ok, err)
		}
	}
	// 4th immediately denied (burst consumed).
	ok, err := rl.Allow(ctx, "k")
	if err != nil || ok {
		t.Fatalf("expected 4th request denied, got ok=%v err=%v", ok, err)
	}
}

func TestTokenBucketIndependentKeys(t *testing.T) {
	ctx := context.Background()
	rl := NewTokenBucket(Config{Rate: 1000, Burst: 1})

	if ok, _ := rl.Allow(ctx, "a"); !ok {
		t.Fatal("expected key a allowed")
	}
	if ok, _ := rl.Allow(ctx, "a"); ok {
		t.Fatal("expected key a denied after burst")
	}
	// Key b unaffected.
	if ok, _ := rl.Allow(ctx, "b"); !ok {
		t.Fatal("expected key b allowed independently")
	}
}

func TestKeyHelpers(t *testing.T) {
	if TenantKey("t1") != "tenant:t1" {
		t.Errorf("TenantKey = %q", TenantKey("t1"))
	}
	if UserKey("u1") != "user:u1" {
		t.Errorf("UserKey = %q", UserKey("u1"))
	}
	if APIKey("k") != "apikey:k" {
		t.Errorf("APIKey = %q", APIKey("k"))
	}
	if RouteUserKey("login", "u1") != "route:login:user:u1" {
		t.Errorf("RouteUserKey = %q", RouteUserKey("login", "u1"))
	}
	if RouteIPKey("register", "1.2.3.4") != "route:register:ip:1.2.3.4" {
		t.Errorf("RouteIPKey = %q", RouteIPKey("register", "1.2.3.4"))
	}
	if TenantCapabilityKey("t1", "c1") != "tenant:t1:capability:c1" {
		t.Errorf("TenantCapabilityKey = %q", TenantCapabilityKey("t1", "c1"))
	}
}
