package transform

import (
	"errors"
	"fmt"
	"github.com/dreadl0ck/netcap/types"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dreadl0ck/netcap/defaults"
	"github.com/dreadl0ck/netcap/maltego"
)

// dieIfExecutable can be used to check if the file at the passed in location is executable
// if so, the function will exit the program cleanly and print an error for maltego.
func dieIfExecutable(trx *maltego.Transform, loc string) {
	// the open tool uses the file extension to decide which program to pass the file to
	// if there is an extension for known executable formats - abort directly
	ext := filepath.Ext(loc)
	log.Println("file extension", ext)

	if ext == ".exe" || ext == ".bin" {
		log.Println("detected known executable file extension - aborting to prevent accidental execution!")
		trx.AddUIMessage("completed!", maltego.UIMessageInform)
		fmt.Println(trx.ReturnOutput())
		os.Exit(0)
	}

	// if there is no extension, use content type detection to determine if its an executable
	log.Println("open path for determining content type:", loc)

	f, err := os.OpenFile(loc, os.O_RDONLY, defaults.DirectoryPermission)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		errClose := f.Close()
		if errClose != nil && !errors.Is(errClose, io.EOF) {
			fmt.Println("failed to close:", errClose)
		}
	}()

	// read file banner to determine content type
	buf := make([]byte, 512)
	_, err = io.ReadFull(f, buf)
	if err != nil && !errors.Is(err, io.EOF) {
		log.Fatal(err)
	}

	// check if file is executable to prevent accidental execution
	cType := http.DetectContentType(buf)
	log.Println("cType:", cType)

	// get file stats to determine if executable bit is set
	stat, err := os.Stat(loc)
	if err != nil {
		log.Fatal(err)
	}

	// if content type is octet or the file has the executable bit set, abort
	if cType == octetStream || isExecAny(stat.Mode()) {
		log.Println("detected executable file - aborting to prevent accidental execution!")
		trx.AddUIMessage("completed!", maltego.UIMessageInform)
		fmt.Println(trx.ReturnOutput())
		os.Exit(0)
	}
}

func die(err string, msg string) {
	trx := maltego.Transform{}
	// add error message for the user
	trx.AddUIMessage(msg+": "+err, maltego.UIMessageFatal)
	fmt.Println(trx.ReturnOutput())
	log.Println(msg, err)
	os.Exit(0) // don't signal an error for the transform invocation
}

func createJa3TableHTML(m map[string]string) string {

	var out = []string{"<table style='width:100%'>"}

	out = append(out, `<tr>
    <th>Ja3</th>
    <th>Lookup Result</th>
  </tr>`)

	for k, v := range m {
		if v != "" {
			out = append(out, "<tr><td style='color:red'>"+k+"</td><td>"+v+"</td></tr>")
		} else {
			out = append(out, "<tr><td>"+k+"</td><td>"+v+"</td></tr>")
		}
	}

	out = append(out, "</table>")

	return strings.Join(out, "")
}

func createSNITableHTML(m map[string]int64) string {

	var out = []string{"<table style='width:100%'>"}

	out = append(out, `<tr>
    <th>SNI</th>
    <th>Number of Packets</th>
  </tr>`)

	for k, v := range m {
		out = append(out, "<tr><td>"+k+"</td><td>"+strconv.FormatInt(v, 10)+"</td></tr>")
	}

	out = append(out, "</table>")

	return strings.Join(out, "")
}

func createProtocolsTableHTML(m map[string]*types.Protocol) string {

	var out = []string{"<table style='width:100%'>"}

	out = append(out, `<tr>
    <th>Application</th>
	<th>Category</th>
    <th>Number of Packets</th>
  </tr>`)

	for k, v := range m {
		out = append(out, "<tr><td>"+k+"</td><td>"+v.Category+"</td><td>"+strconv.FormatUint(v.Packets, 10)+"</td></tr>")
	}

	out = append(out, "</table>")

	return strings.Join(out, "")
}
