// +build !noprocesses

package collector

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/prometheus/procfs"
)

const (
	processesSubsystem = "process"
)

type processesCollector struct {
	metrics []prometheus.Collector
}

func init() {
	Factories["processes"] = NewProcessesCollector
}

func NewProcessesCollector() (Collector, error) {
	var processesLabelNames = []string{"process"}

	return &processesCollector{
		metrics: []prometheus.Collector{
			prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace: Namespace,
					Subsystem: processesSubsystem,
					Name:      "memory_resident_usage_bytes",
					Help:      "Resident memory size in bytes",
				},
				processesLabelNames,
			),
			prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace: Namespace,
					Subsystem: processesSubsystem,
					Name:      "cpu_usage_total_seconds",
					Help:      "Total CPU user and system time in seconds",
				},
				processesLabelNames,
			),
		},
	}, nil
}

func (c *processesCollector) Update(ch chan<- prometheus.Metric) error {
	processes, err := procfs.AllProcs()
	if err != nil {
		return fmt.Errorf("failed to get processes: %s", err)
	}

	for _, process := range processes {
		cmd, err := process.Comm()
		if err != nil {
			log.Debugf("Failed to get process command: %s", err)
			continue
		}

		stats, err := process.NewStat()
		if err != nil {
			log.Debugf("Failed to get process stats: %s", err)
			continue
		}

		// skip processes with empty stats
		if stats.ResidentMemory() == 0 || stats.CPUTime() == 0 {
			log.Debugf("Skipping process %s due to empty stats", cmd)
			continue
		}

		c.metrics[0].(*prometheus.GaugeVec).WithLabelValues(cmd).Set(float64(stats.ResidentMemory()))
		c.metrics[1].(*prometheus.GaugeVec).WithLabelValues(cmd).Set(stats.CPUTime())
	}

	for _, c := range c.metrics {
		c.Collect(ch)
	}

	return err
}
