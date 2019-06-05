package polls

import (
	"context"
	"log"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	poll_db "lightning-poll/polls/internal/db/polls"
	"lightning-poll/polls/internal/types"
)

var pollCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: "main",
	Subsystem: "polls",
	Name:      "count",
	Help:      "Count of polls by status.",
}, []string{"status"})

func init() {
	prometheus.MustRegister(pollCount)
}

func updateMetricsForever(b Backends) {
	for {
		ctx := context.Background()
		if err := updatePollMetrics(ctx, b); err != nil {
			log.Printf("updateMetricsForever: error %v", err)
		}

		time.Sleep(time.Minute * 30)
	}
}

func updatePollMetrics(ctx context.Context, b Backends) error {
	//TODO(carla): Add DB function that returns map status -> poll rather than dodgy loop
	i := types.PollStatus(1)
	for {
		if !i.Valid() {
			break
		}

		p, err := poll_db.ListByStatus(ctx, b.GetDB(), types.PollStatusCreated)
		if err != nil {
			return err
		}

		pollCount.WithLabelValues(types.PollStatusCreated.String()).Sub(float64(len(p)))
		i++

	}

	return nil
}
