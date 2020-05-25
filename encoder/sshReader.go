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

package encoder

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/dreadl0ck/netcap/reassembly"
	"github.com/dreadl0ck/netcap/sshx"
	"github.com/dreadl0ck/netcap/types"
	"github.com/sasha-s/go-deadlock"
)

/*
 * SSH - The Secure Shell Protocol
 */

type sshReader struct {
	parent        *tcpConnection
	clientIdent   string
	serverIdent   string
	clientKexInit *sshx.KexInitMsg
	serverKexInit *sshx.KexInitMsg
	software      []*types.Software
}

func (h *sshReader) Decode(s2c Stream, c2s Stream) {

	// parse conversation
	var (
		buf         bytes.Buffer
		previousDir reassembly.TCPFlowDirection
	)
	if len(h.parent.merged) > 0 {
		previousDir = h.parent.merged[0].dir
	}

	for _, d := range h.parent.merged {

		if d.dir == previousDir {
			buf.Write(d.raw)
		} else {
			h.searchKexInit(bufio.NewReader(&buf))

			buf.Reset()

			previousDir = d.dir
			buf.Write(d.raw)
			continue
		}
	}
	h.searchKexInit(bufio.NewReader(&buf))
	if len(h.software) == 0 {
		return
	}

	// add new audit records or update existing
	SoftwareStore.Lock()
	for _, s := range h.software {
		if _, ok := SoftwareStore.Items[s.Product+"/"+s.Version]; ok {
			// TODO updateSoftwareAuditRecord(dp, p, i)
		} else {
			SoftwareStore.Items[s.Product+"/"+s.Version] = &Software{
				s,
				deadlock.Mutex{},
			}
			statsMutex.Lock()
			reassemblyStats.numSoftware++
			statsMutex.Unlock()
		}
	}
	SoftwareStore.Unlock()
}

func (h *sshReader) searchKexInit(r *bufio.Reader) {

	if h.serverKexInit != nil && h.clientKexInit != nil {
		return
	}

	data, err := ioutil.ReadAll(r)
	if err != nil {
		fmt.Println(err)
		return
	}

	if h.clientIdent == "" {
		h.clientIdent = string(data)
		return
	} else if h.serverIdent == "" {
		h.serverIdent = string(data)
		return
	}

	for i, b := range data {

		if b == 0x14 { // Marks the beginning of the KexInitMsg // TODO: stop checking after X bytes, and after we already have server and client hashes

			if i == 0 {
				break
			}

			if len(data[:i-1]) != 4 {
				break
			}

			length := int(binary.BigEndian.Uint32(data[:i-1]))
			padding := int(data[i-1])
			if len(data) < i+length-padding-1 {
				break
			}

			//fmt.Println("padding", padding, "length", length)
			//fmt.Println(hex.Dump(data[i:i+length-padding-1]))

			var init sshx.KexInitMsg
			err := sshx.Unmarshal(data[i:i+length-padding-1], &init)
			if err != nil {
				fmt.Println(err)
			}

			//spew.Dump("found SSH KexInit", h.parent.ident, init)
			hash, raw := computeHASSH(init)
			if h.clientKexInit == nil {
				sshEncoder.write(&types.SSH{
					Timestamp:  h.parent.client.FirstPacket().String(),
					HASSH:      hash,
					Flow:       h.parent.ident,
					Ident:      h.clientIdent,
					Algorithms: raw,
					IsClient:   true,
				})
				h.clientKexInit = &init
			} else {
				sshEncoder.write(&types.SSH{
					Timestamp:  h.parent.client.FirstPacket().String(),
					HASSH:      hash,
					Flow:       reverseIdent(h.parent.ident),
					Ident:      h.serverIdent,
					Algorithms: raw,
					IsClient:   false,
				})
				h.serverKexInit = &init
			}

			// TODO fetch device profile
			for _, soft := range hashDBMap[hash] {
				vendor, product, version, os := parseSSH(soft.Version)
				h.software = append(h.software, &types.Software{
					Timestamp: h.parent.client.FirstPacket().String(),
					Product:   product, // Name of the server (Apache, Nginx, ...)
					Vendor:    vendor,  // Unfitting name, but operating system
					Version:   version, // Version as found after the '/'
					//DeviceProfiles: []string{dpIdent},
					SourceName: "HASSH",
					SourceData: hash,
					Service:    "SSH",
					//DPIResults:     protos,
					Flows: []string{h.parent.ident},
					Notes: "Likelyhood: " + soft.Likelyhood + " Possible OS: " + os,
				})
			}
			break
		}
	}
}

func parseSSH(soft string) (string, string, string, string) {
	firstSplit := strings.Split(soft, " ? ")
	sshVersionTmp := firstSplit[0]
	sshVersion := strings.Split(sshVersionTmp, " | ")
	product := sshVersion[0]
	vendorVersion := strings.Split(sshVersion[1], " ")
	var os string
	if len(firstSplit) > 1 {
		os = firstSplit[len(firstSplit)-1]
		return product, vendorVersion[0], vendorVersion[1], os
	} else {
		os = ""
	}
	return product, vendorVersion[0], vendorVersion[1], os
}

// HASSH SSH Fingerprint
func computeHASSH(init sshx.KexInitMsg) (hash string, raw string) {

	var b strings.Builder
	b.WriteString(strings.Join(init.KexAlgos, ","))
	b.WriteString(";")
	b.WriteString(strings.Join(init.CiphersClientServer, ","))
	b.WriteString(";")
	b.WriteString(strings.Join(init.MACsClientServer, ","))
	b.WriteString(";")
	b.WriteString(strings.Join(init.CompressionClientServer, ","))

	return fmt.Sprintf("%x", md5.Sum([]byte(b.String()))), b.String()
}
