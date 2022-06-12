// Copyright © 2022 Attestant Limited.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package standard

import (
	"context"

	"github.com/attestantio/vouch/services/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var blockRootToSlotProcessed *prometheus.CounterVec
var blockRootToSlotEntries prometheus.Gauge

func registerMetrics(ctx context.Context, monitor metrics.Service) error {
	if blockRootToSlotProcessed != nil {
		// Already registered.
		return nil
	}
	if monitor == nil {
		// No monitor.
		return nil
	}
	if monitor.Presenter() == "prometheus" {
		return registerPrometheusMetrics(ctx)
	}
	return nil
}

func registerPrometheusMetrics(_ context.Context) error {
	blockRootToSlotProcessed = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "vouch",
		Subsystem: "cache",
		Name:      "blockroottoslot_lookups",
		Help:      "The number of lookups for block root to slot.",
	}, []string{"result"})
	if err := prometheus.Register(blockRootToSlotProcessed); err != nil {
		return err
	}

	blockRootToSlotEntries = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "vouch",
		Subsystem: "cache",
		Name:      "blockroottoslot_entries",
		Help:      "The number of entries in the block root to slot cache.",
	})
	return prometheus.Register(blockRootToSlotEntries)
}

func monitorBlockRootToSlotEntriesUpdated(entries int) {
	if blockRootToSlotEntries == nil {
		return
	}
	blockRootToSlotEntries.Set(float64(entries))
}

func monitorBlockRootToSlot(source string) {
	if blockRootToSlotProcessed == nil {
		return
	}
	blockRootToSlotProcessed.WithLabelValues(source).Inc()
}
