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

package dbs

import (
	"encoding/hex"
	"fmt"
	"github.com/dreadl0ck/cryptoutils"
	"github.com/dreadl0ck/netcap/resolvers"
	"github.com/dreadl0ck/netcap/utils"
	"github.com/evilsocket/islazy/zip"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func UpdateDBs() {

	// check if database root path exists already
	if _, err := os.Stat(resolvers.ConfigRootPath); err != nil {
		log.Fatal("database root path: ", resolvers.ConfigRootPath, " does not exist")
	}

	var (
		pathASN = filepath.Join(resolvers.DataBaseFolderPath, "GeoLite2-ASN.mmdb")
		pathCity = filepath.Join(resolvers.DataBaseFolderPath, "GeoLite2-City.mmdb")
	)

	// backup the recent versions of the GeoLite databases
	// so they wont get overwritten by the outdated ones from upstream
	asnDB, err := ioutil.ReadFile(pathASN)
	if err != nil {
		log.Fatal(err)
	}

	cityDB, err := ioutil.ReadFile(pathCity)
	if err != nil {
		log.Fatal(err)
	}

	var (
		asnHash = hex.EncodeToString(cryptoutils.MD5Data(asnDB))
		cityHash = hex.EncodeToString(cryptoutils.MD5Data(cityDB))
	)

	if asnHash != "17eea01c955ada90ad922c2c95455515" {
		utils.CopyFile(pathASN, "/tmp")
	}
	if cityHash != "10b66842fd51336ae7c4f34c058deb46" {
		utils.CopyFile(pathCity, "/tmp")
	}

	cmd := exec.Command("git", "pull")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	if err != nil {
		log.Fatal("failed to pull netcap-dbs repo: ", err)
	}

	fmt.Println("pulled netcap-dbs repository")

	// restore geolite dbs
	if asnHash != "17eea01c955ada90ad922c2c95455515" {
		utils.CopyFile(filepath.Join("/tmp", filepath.Base(pathASN)), pathASN)
	}
	if cityHash != "10b66842fd51336ae7c4f34c058deb46" {
		utils.CopyFile(filepath.Join("/tmp", filepath.Base(pathCity)), pathCity)
	}

	// decompress bleve stores
	files, err := ioutil.ReadDir(resolvers.DataBaseFolderPath)
	if err != nil {
		log.Fatal("failed to read dir: ", resolvers.DataBaseFolderPath, err)
	}

	for _, f := range files {
		if filepath.Ext(f.Name()) == ".zip" {
			fmt.Println("decompressing", f.Name())
			_, err = zip.Unzip(
				filepath.Join(resolvers.DataBaseFolderPath, f.Name()),
				resolvers.DataBaseFolderPath,
			)
			if err != nil {
				log.Fatal("failed to unzip: ", f.Name(), " error: ", err)
			}
		}
	}

	fmt.Println("done! Updated databases to", resolvers.ConfigRootPath)
}