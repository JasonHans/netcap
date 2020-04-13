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

	"github.com/prometheus/client_golang/prometheus"
)

var fieldsIPv6 = []string{
	"Timestamp",
	"Version",        // int32
	"TrafficClass",   // int32
	"FlowLabel",      // uint32
	"Length",         // int32
	"NextHeader",     // int32
	"HopLimit",       // int32
	"SrcIP",          // string
	"DstIP",          // string
	"PayloadEntropy", // float64
	"PayloadSize",    // int32
	"HopByHop",       // *IPv6HopByHop
}

func (i IPv6) CSVHeader() []string {
	return filter(fieldsIPv6)
}

func (i IPv6) CSVRecord() []string {
	var hop string
	if i.HopByHop != nil {
		hop = i.HopByHop.ToString()
	}
	return filter([]string{
		formatTimestamp(i.Timestamp),
		formatInt32(i.Version),      // int32
		formatInt32(i.TrafficClass), // int32
		formatUint32(i.FlowLabel),   // uint32
		formatInt32(i.Length),       // int32
		formatInt32(i.NextHeader),   // int32
		formatInt32(i.HopLimit),     // int32
		i.SrcIP,                     // string
		i.DstIP,                     // string
		strconv.FormatFloat(i.PayloadEntropy, 'f', 6, 64), // float64
		formatInt32(i.PayloadSize),                        // int32
		hop,                                               // *IPv6HopByHop
	})
}

func (i IPv6) Time() string {
	return i.Timestamp
}

func (h IPv6HopByHop) ToString() string {
	var opts []string
	for _, o := range h.Options {
		opts = append(opts, o.ToString())
	}
	return h.Timestamp + Separator + join(opts...)
}

func (a IPv6) JSON() (string, error) {
	return jsonMarshaler.MarshalToString(&a)
}

var (
	ip6Metric = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: strings.ToLower(Type_NC_IPv6.String()),
			Help: Type_NC_IPv6.String() + " audit records",
		},
		fieldsIPv6Metrics,
	)
	ip6PayloadEntropy = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    strings.ToLower(Type_NC_IPv6.String()) + "_entropy",
			Help:    Type_NC_IPv6.String() + " payload entropy",
			Buckets: prometheus.LinearBuckets(20, 5, 5),
		},
		// []string{"SrcIP", "DstIP"},
		[]string{},
	)
	ip6PayloadSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    strings.ToLower(Type_NC_IPv6.String()) + "_size",
			Help:    Type_NC_IPv6.String() + " payload sizes",
			Buckets: prometheus.LinearBuckets(20, 5, 5),
		},
		// []string{"SrcIP", "DstIP"},
		[]string{},
	)
)

var fieldsIPv6Metrics = []string{
	"Version",      // int32
	"TrafficClass", // int32
	"FlowLabel",    // uint32
	"NextHeader",   // int32
	"HopLimit",     // int32
	"SrcIP",        // string
	"DstIP",        // string
}

func (i IPv6) metricValues() []string {
	return []string{
		formatInt32(i.Version),      // int32
		formatInt32(i.TrafficClass), // int32
		formatUint32(i.FlowLabel),   // uint32
		formatInt32(i.NextHeader),   // int32
		formatInt32(i.HopLimit),     // int32
		i.SrcIP,                     // string
		i.DstIP,                     // string
	}
}

func init() {
	prometheus.MustRegister(ip6Metric)
	prometheus.MustRegister(ip6PayloadEntropy)
	prometheus.MustRegister(ip6PayloadSize)
}

func (a IPv6) Inc() {
	ip6Metric.WithLabelValues(a.metricValues()...).Inc()
	ip6PayloadEntropy.WithLabelValues().Observe(a.PayloadEntropy)
	ip6PayloadSize.WithLabelValues().Observe(float64(a.PayloadSize))
}

func (a *IPv6) SetPacketContext(ctx *PacketContext) {

	// create new context and only add information that is
	// not yet present on the audit record type
	a.Context = &PacketContext{
		SrcPort: ctx.SrcPort,
		DstPort: ctx.DstPort,
	}
}

func (a IPv6) Src() string {
	return a.SrcIP
}

func (a IPv6) Dst() string {
	return a.DstIP
}
