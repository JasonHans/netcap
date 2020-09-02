/*
 * NETCAP - Traffic Analysis Framework
 * Copyright (c) 2017-2020 Philipp Mieden <dreadl0ck [at] protonmail [dot] ch>
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */

package types

import (
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var fieldsSoftware = []string{
	"Timestamp",
	"Product",
	"Vendor",
	"Version",
	"DeviceProfiles",
	"SourceName",
	"DPIResults",
	"Service",
	"Flows",
	"SourceData",
	"Notes",
}

// CSVHeader returns the CSV header for the audit record.
func (a *Software) CSVHeader() []string {
	return filter(fieldsSoftware)
}

// CSVRecord returns the CSV record for the audit record.
func (a *Software) CSVRecord() []string {
	return filter([]string{
		formatTimestamp(a.Timestamp),
		a.Product,
		a.Vendor,
		a.Version,
		join(a.DeviceProfiles...),
		a.SourceName,
		join(a.DPIResults...),
		a.Service,
		join(a.Flows...),
		a.SourceData,
		a.Notes,
	})
}

// Time returns the timestamp associated with the audit record.
func (a *Software) Time() int64 {
	return a.Timestamp
}

// JSON returns the JSON representation of the audit record.
func (a *Software) JSON() (string, error) {
	// convert unix timestamp from nano to millisecond precision for elastic
	a.Timestamp /= int64(time.Millisecond)

	return jsonMarshaler.MarshalToString(a)
}

var fieldsSoftwareMetric = []string{
	"Product",
	"Vendor",
	"Version",
	"NumDeviceProfiles",
	"SourceName",
	//"NumDPIResults",
	"Service",
	//"Flows",
	//"SourceData",
	//"Notes",
}

var softwareMetric = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: strings.ToLower(Type_NC_Software.String()),
		Help: Type_NC_Software.String() + " audit records",
	},
	fieldsSoftwareMetric,
)

func (a *Software) metricValues() []string {
	return []string{
		a.Product,
		a.Vendor,
		a.Version,
		strconv.Itoa(len(a.DeviceProfiles)),
		a.SourceName,
		a.Service,
	}
}

// Inc increments the metrics for the audit record.
func (a *Software) Inc() {
	softwareMetric.WithLabelValues(a.metricValues()...).Inc()
}

// SetPacketContext sets the associated packet context for the audit record.
func (a *Software) SetPacketContext(*PacketContext) {}

// Src returns the source address of the audit record.
func (a *Software) Src() string {
	return ""
}

// Dst returns the destination address of the audit record.
func (a *Software) Dst() string {
	return ""
}
