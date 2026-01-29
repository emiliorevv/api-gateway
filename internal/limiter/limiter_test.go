package limiter

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestRateLimiter_Allow(t *testing.T) {

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("error conecting to miniredis: %v", err)
	}

	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	rl := NewRateLimiter(rdb)
	ctx := context.Background()
	key := "user_123"
	limit := 3
	refreshRate := 0.5

	t.Log("Test #1, sending 3 petititons, it should accept them without any problems")
	for i := 0; i < limit; i++ {
		allow, err := rl.Allow(ctx, key, limit, refreshRate)
		if err != nil {
			t.Fatalf("error with Lua : %v", err)
		}

		if !allow {
			t.Errorf("Allow returned blocked: %d", i+1)
		}
	}

	t.Log("Test #2, sending 4 petitions, it should be blocked because of reaching limit")

		allow, err := rl.Allow(ctx, key, limit, refreshRate)
		if err != nil {
			t.Fatalf("error with Lua : %v", err)
		}


		if allow {
			t.Errorf("Allow function wasnt blocked correctly")
		} else {
			t.Logf("Allow returned blocked correctly")
		}

		t.Log("Test #3, waiting 2 seconds to refresh tokens")
		time.Sleep(3 * time.Second)

		allow, err = rl.Allow(ctx, key, limit, refreshRate)
		if err != nil {
			t.Fatalf("error with Lua : %v", err)
		}

		if !allow {
			t.Errorf("Allow function isnt refreshing the tokens")
		} else {
			t.Logf("Allow refreshed the tokens correctly")
		}

}
