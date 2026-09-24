package cache

import (
	"context"
	"errors"
	"testing"
	"time"

	"terralist/pkg/metrics"

	"github.com/prometheus/client_golang/prometheus/testutil"
	. "github.com/smartystreets/goconvey/convey"
)

// stubCache answers Get with a fixed value and fails every write.
type stubCache struct {
	value []byte
}

func (s *stubCache) Get(_ context.Context, _ string) ([]byte, bool, error) {
	if s.value == nil {
		return nil, false, nil
	}

	return s.value, true, nil
}

func (s *stubCache) Set(_ context.Context, _ string, _ []byte, _ time.Duration) error {
	return errors.New("write failed")
}

func (s *stubCache) Delete(_ context.Context, _ string) error {
	return nil
}

func count(operation, result string) float64 {
	return testutil.ToFloat64(metrics.CacheOperationsTotal.WithLabelValues(operation, "stub", result))
}

func TestMetricsCache(t *testing.T) {
	Convey("Subject: Recording cache metrics", t, func() {
		ctx := context.Background()
		stub := &stubCache{}
		c := &MetricsCache{Cache: stub, Backend: "stub"}

		Convey("When a read misses", func() {
			before := count("get", "miss")
			_, ok, err := c.Get(ctx, "key")

			Convey("Then a miss should be counted", func() {
				So(err, ShouldBeNil)
				So(ok, ShouldBeFalse)
				So(count("get", "miss")-before, ShouldEqual, 1)
			})
		})

		Convey("When a read hits", func() {
			stub.value = []byte("value")
			before := count("get", "hit")
			value, ok, _ := c.Get(ctx, "key")

			Convey("Then a hit should be counted and the value passed through", func() {
				So(ok, ShouldBeTrue)
				So(string(value), ShouldEqual, "value")
				So(count("get", "hit")-before, ShouldEqual, 1)
			})
		})

		Convey("When a write fails", func() {
			before := count("set", "error")
			err := c.Set(ctx, "key", []byte("value"), time.Hour)

			Convey("Then an error should be counted and returned", func() {
				So(err, ShouldNotBeNil)
				So(count("set", "error")-before, ShouldEqual, 1)
			})
		})

		Convey("When a delete succeeds", func() {
			before := count("delete", "success")
			err := c.Delete(ctx, "key")

			Convey("Then a success should be counted", func() {
				So(err, ShouldBeNil)
				So(count("delete", "success")-before, ShouldEqual, 1)
			})
		})
	})
}
