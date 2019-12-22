package tally_test

import (
	"fmt"
	"log"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/uber-go/tally"
	"github.com/uber-go/tally/m3"
)

func TestScopeCachedFlushTimes(t *testing.T) {
	t.Skip("skipping long-running flush test")

	reporter, err := m3.NewReporter(m3.Options{
		HostPorts:          []string{"127.0.0.1:9052"},
		Service:            "test",
		Env:                "test",
		CommonTags:         nil,
		IncludeHost:        false,
		MaxQueueSize:       8092,
		MaxPacketSizeBytes: 32768,
		Protocol:           m3.Compact,
	})
	require.NoError(t, err)

	scope, closer := tally.NewRootScope(tally.ScopeOptions{
		Prefix:          "",
		Tags:            nil,
		CachedReporter:  reporter,
		SanitizeOptions: &m3.DefaultSanitizerOpts,
	}, 5*time.Second)
	defer func() {
		if err := closer.Close(); err != nil {
			log.Println("failed to close scope:", err)
		}
	}()

	var (
		timer  = time.NewTimer(time.Hour)
		ticker = time.NewTicker(time.Second)
	)

	defer ticker.Stop()

	for {
		select {
		case <-timer.C:
			return
		default:
		}

		start := time.Now().UnixNano()
		for i := 0; i < 1000; i++ {
			ss := scope.Tagged(map[string]string{
				"foo": strconv.Itoa(i),
			})

			for j := 0; j < 5000; j++ {
				ss.Counter(strconv.Itoa(j)).Inc(1)
			}
		}

		fmt.Println(
			">>> incrementing took",
			time.Duration(time.Now().UnixNano()-start).String(),
		)
		<-ticker.C

		break
	}

	<-timer.C
}
