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
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

var fieldsICMPv6NeighborAdvertisement = []string{
	"Timestamp",
	"Flags",         // int32
	"TargetAddress", // string
	"Options",       // []*ICMPv6Option
	"SrcIP",
	"DstIP",
}

func (i ICMPv6NeighborAdvertisement) CSVHeader() []string {
	return filter(fieldsICMPv6NeighborAdvertisement)
}

func (i ICMPv6NeighborAdvertisement) CSVRecord() []string {
	var opts []string
	for _, o := range i.Options {
		opts = append(opts, o.ToString())
	}
	// prevent accessing nil pointer
	if i.Context == nil {
		i.Context = &PacketContext{}
	}
	return filter([]string{
		formatTimestamp(i.Timestamp),
		formatInt32(i.Flags),
		i.TargetAddress,
		strings.Join(opts, ""),
		i.Context.SrcIP,
		i.Context.DstIP,
	})
}

func (i ICMPv6NeighborAdvertisement) Time() string {
	return i.Timestamp
}

func (a ICMPv6NeighborAdvertisement) JSON() (string, error) {
	return jsonMarshaler.MarshalToString(&a)
}

var icmp6naMetric = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: strings.ToLower(Type_NC_ICMPv6NeighborAdvertisement.String()),
		Help: Type_NC_ICMPv6NeighborAdvertisement.String() + " audit records",
	},
	fieldsICMPv6NeighborAdvertisement[1:],
)

func init() {
	prometheus.MustRegister(icmp6naMetric)
}

func (a ICMPv6NeighborAdvertisement) Inc() {
	icmp6naMetric.WithLabelValues(a.CSVRecord()[1:]...).Inc()
}

func (a *ICMPv6NeighborAdvertisement) SetPacketContext(ctx *PacketContext) {
	a.Context = ctx
}

func (a ICMPv6NeighborAdvertisement) Src() string {
	if a.Context != nil {
		return a.Context.SrcIP
	}
	return ""
}

func (a ICMPv6NeighborAdvertisement) Dst() string {
	if a.Context != nil {
		return a.Context.DstIP
	}
	return ""
}
