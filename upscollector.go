package apcupsdexporter

import (
	"log/slog"
	"strings"
	"time"

	"github.com/mdlayher/apcupsd"
	"github.com/prometheus/client_golang/prometheus"
)

var _ StatusSource = &apcupsd.Client{}

// A StatusSource is a type which can retrieve UPS status information from
// apcupsd.  It is implemented by *apcupsd.Client.
type StatusSource interface {
	Status() (*apcupsd.Status, error)
}

// A UPSCollector is a Prometheus collector for metrics regarding an APC UPS.
type UPSCollector struct {
	Info *prometheus.Desc

	UPSLoadPercent                      *prometheus.Desc
	BatteryChargePercent                *prometheus.Desc
	LineVolts                           *prometheus.Desc
	LineNominalVolts                    *prometheus.Desc
	OutputVolts                         *prometheus.Desc
	OutputAmps                          *prometheus.Desc
	BatteryVolts                        *prometheus.Desc
	BatteryNominalVolts                 *prometheus.Desc
	BatteryNumberTransfersTotal         *prometheus.Desc
	BatteryTimeLeftSeconds              *prometheus.Desc
	BatteryTimeOnSeconds                *prometheus.Desc
	BatteryCumulativeTimeOnSecondsTotal *prometheus.Desc
	LastTransferOnBatteryTimeSeconds    *prometheus.Desc
	LastTransferOffBatteryTimeSeconds   *prometheus.Desc
	LastSelftestTimeSeconds             *prometheus.Desc
	NominalPowerWatts                   *prometheus.Desc
	InternalTemperatureCelsius          *prometheus.Desc

	ss     StatusSource
	logger *slog.Logger
}

var _ prometheus.Collector = &UPSCollector{}

// NewUPSCollectorWithLogger creates a new UPSCollector with an explicit slog logger.
func NewUPSCollectorWithLogger(ss StatusSource, logger *slog.Logger) *UPSCollector {
	labels := []string{"ups"}

	return &UPSCollector{
		Info: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "info"),
			"Metadata about a given UPS.",
			[]string{"ups", "hostname", "model", "status"},
			nil,
		),

		UPSLoadPercent: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "ups_load_percent"),
			"Current UPS load percentage.",
			labels,
			nil,
		),

		BatteryChargePercent: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "battery_charge_percent"),
			"Current UPS battery charge percentage.",
			labels,
			nil,
		),

		LineVolts: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "line_volts"),
			"Current AC input line voltage.",
			labels,
			nil,
		),

		LineNominalVolts: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "line_nominal_volts"),
			"Nominal AC input line voltage.",
			labels,
			nil,
		),

		OutputVolts: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "output_volts"),
			"Current AC output voltage.",
			labels,
			nil,
		),

		OutputAmps: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "output_amps"),
			"Current AC output amperage.",
			labels,
			nil,
		),

		BatteryVolts: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "battery_volts"),
			"Current UPS battery voltage.",
			labels,
			nil,
		),

		BatteryNominalVolts: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "battery_nominal_volts"),
			"Nominal UPS battery voltage.",
			labels,
			nil,
		),

		BatteryNumberTransfersTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "battery_number_transfers_total"),
			"Total number of transfers to UPS battery power.",
			labels,
			nil,
		),

		BatteryTimeLeftSeconds: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "battery_time_left_seconds"),
			"Number of seconds remaining of UPS battery power.",
			labels,
			nil,
		),

		BatteryTimeOnSeconds: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "battery_time_on_seconds"),
			"Number of seconds the UPS has been providing battery power due to an AC input line outage.",
			labels,
			nil,
		),

		BatteryCumulativeTimeOnSecondsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "battery_cumulative_time_on_seconds_total"),
			"Total number of seconds the UPS has provided battery power due to AC input line outages.",
			labels,
			nil,
		),

		LastTransferOnBatteryTimeSeconds: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "last_transfer_on_battery_time_seconds"),
			"UNIX timestamp of last transfer to battery since apcupsd startup.",
			labels,
			nil,
		),

		LastTransferOffBatteryTimeSeconds: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "last_transfer_off_battery_time_seconds"),
			"UNIX timestamp of last transfer from battery since apcupsd startup.",
			labels,
			nil,
		),

		LastSelftestTimeSeconds: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "last_selftest_time_seconds"),
			"UNIX timestamp of last selftest since apcupsd startup.",
			labels,
			nil,
		),

		NominalPowerWatts: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "nominal_power_watts"),
			"Nominal power output in watts.",
			labels,
			nil,
		),

		InternalTemperatureCelsius: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "internal_temperature_celsius"),
			"Internal temperature in °C.",
			labels,
			nil,
		),

		ss:     ss,
		logger: logger,
	}
}

// NewUPSCollector creates a new UPSCollector with a default logger.
// This is a backward-compatible constructor for tests and callers that don't
// need to pass an explicit logger.
func NewUPSCollector(ss StatusSource) *UPSCollector {
	return NewUPSCollectorWithLogger(ss, slog.Default())
}

// Describe sends the descriptors of each metric over to the provided channel.
// The corresponding metric values are sent separately.
func (c *UPSCollector) Describe(ch chan<- *prometheus.Desc) {
	ds := []*prometheus.Desc{
		c.Info,
		c.UPSLoadPercent,
		c.BatteryChargePercent,
		c.LineVolts,
		c.LineNominalVolts,
		c.OutputVolts,
		c.OutputAmps,
		c.BatteryVolts,
		c.BatteryNominalVolts,
		c.BatteryNumberTransfersTotal,
		c.BatteryTimeLeftSeconds,
		c.BatteryTimeOnSeconds,
		c.BatteryCumulativeTimeOnSecondsTotal,
		c.LastTransferOnBatteryTimeSeconds,
		c.LastTransferOffBatteryTimeSeconds,
		c.LastSelftestTimeSeconds,
		c.NominalPowerWatts,
		c.InternalTemperatureCelsius,
	}

	for _, d := range ds {
		ch <- d
	}
}

// Collect sends the metric values for each metric created by the UPSCollector
// to the provided prometheus Metric channel.
func (c *UPSCollector) Collect(ch chan<- prometheus.Metric) {
	s, err := c.ss.Status()
	if err != nil {
		if c.logger != nil {
			c.logger.Error("failed collecting UPS metrics", "err", err)
		}
		ch <- prometheus.NewInvalidMetric(c.Info, err)
		return
	}

	ups := sanitizeLabel(s.UPSName)
	host := sanitizeLabel(s.Hostname)
	model := sanitizeLabel(s.Model)
	status := sanitizeLabel(s.Status)

	ch <- prometheus.MustNewConstMetric(
		c.Info,
		prometheus.GaugeValue,
		1,
		ups, host, model, status,
	)

	ch <- prometheus.MustNewConstMetric(
		c.UPSLoadPercent,
		prometheus.GaugeValue,
		s.LoadPercent,
		ups,
	)

	ch <- prometheus.MustNewConstMetric(
		c.BatteryChargePercent,
		prometheus.GaugeValue,
		s.BatteryChargePercent,
		ups,
	)

	ch <- prometheus.MustNewConstMetric(
		c.LineVolts,
		prometheus.GaugeValue,
		s.LineVoltage,
		ups,
	)

	ch <- prometheus.MustNewConstMetric(
		c.LineNominalVolts,
		prometheus.GaugeValue,
		s.NominalInputVoltage,
		ups,
	)

	ch <- prometheus.MustNewConstMetric(
		c.OutputVolts,
		prometheus.GaugeValue,
		s.OutputVoltage,
		ups,
	)

	ch <- prometheus.MustNewConstMetric(
		c.OutputAmps,
		prometheus.GaugeValue,
		s.OutputAmps,
		ups,
	)

	ch <- prometheus.MustNewConstMetric(
		c.BatteryVolts,
		prometheus.GaugeValue,
		s.BatteryVoltage,
		ups,
	)

	ch <- prometheus.MustNewConstMetric(
		c.BatteryNominalVolts,
		prometheus.GaugeValue,
		s.NominalBatteryVoltage,
		ups,
	)

	ch <- prometheus.MustNewConstMetric(
		c.BatteryNumberTransfersTotal,
		prometheus.CounterValue,
		float64(s.NumberTransfers),
		ups,
	)

	ch <- prometheus.MustNewConstMetric(
		c.BatteryTimeLeftSeconds,
		prometheus.GaugeValue,
		s.TimeLeft.Seconds(),
		ups,
	)

	ch <- prometheus.MustNewConstMetric(
		c.BatteryTimeOnSeconds,
		prometheus.GaugeValue,
		s.TimeOnBattery.Seconds(),
		ups,
	)

	ch <- prometheus.MustNewConstMetric(
		c.BatteryCumulativeTimeOnSecondsTotal,
		prometheus.CounterValue,
		s.CumulativeTimeOnBattery.Seconds(),
		ups,
	)

	ch <- prometheus.MustNewConstMetric(
		c.LastTransferOnBatteryTimeSeconds,
		prometheus.GaugeValue,
		timestamp(s.XOnBattery),
		ups,
	)

	ch <- prometheus.MustNewConstMetric(
		c.LastTransferOffBatteryTimeSeconds,
		prometheus.GaugeValue,
		timestamp(s.XOffBattery),
		ups,
	)

	ch <- prometheus.MustNewConstMetric(
		c.LastSelftestTimeSeconds,
		prometheus.GaugeValue,
		timestamp(s.LastSelftest),
		ups,
	)

	ch <- prometheus.MustNewConstMetric(
		c.NominalPowerWatts,
		prometheus.GaugeValue,
		float64(s.NominalPower),
		ups,
	)

	ch <- prometheus.MustNewConstMetric(
		c.InternalTemperatureCelsius,
		prometheus.GaugeValue,
		s.InternalTemp,
		ups,
	)
}

func timestamp(t time.Time) float64 {
	if t.IsZero() {
		return 0
	}

	return float64(t.Unix())
}

// sanitizeLabel removes control characters and truncates label values to a
// reasonable length to avoid excessive cardinality or invalid label values.
func sanitizeLabel(v string) string {
	if v == "" {
		return "unknown"
	}
	// Remove control chars.
	v = strings.Map(func(r rune) rune {
		if r < 32 || r == 127 {
			return -1
		}
		return r
	}, v)
	v = strings.TrimSpace(v)
	const maxLabelLen = 128
	if len(v) > maxLabelLen {
		return v[:maxLabelLen]
	}
	return v
}
