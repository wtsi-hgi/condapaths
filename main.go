// Copyright © 2025 Genome Research Limited
// Authors:
//  Sendu Bala <sb10@sanger.ac.uk>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const statsFileSuffix = ".stats.gz"

const helpText = `condapaths parses wrstat stats.gz files quickly, in low mem.

Provide a directory as an argument and it will parse the most recent stats.gz
file inside.

It outputs files with one quoted path per line:
* <date>.condarc: paths where the file basename was ".condarc"
* <date>.conda-meta: paths where the file basename was "history", in a directory
                     named "conda-meta"
* <date>.singularity: paths where the file basename suffix was one of ".sif",
                      ".simg", and ".img"

Usage: condapaths /wrstat/output/dir
Options:
  -h          this help text
`

var l = log.New(os.Stderr, "", 0) //nolint:gochecknoglobals

func main() {
	var help = flag.Bool("h", false, "print help text")
	flag.Parse()

	if *help {
		exitHelp("")
	}

	if flag.Arg(0) == "" {
		exitHelp("ERROR: you must provide a wrstat output directory")
	}

	statsPath, date, err := getLatestStatFile(flag.Arg(0))
	if err != nil {
		die(err)
	}

	input, cleanup, err := decompress(statsPath)
	if err != nil {
		die(err)
	}

	defer cleanup()

	parseStats(input, date)
}

// exitHelp prints help text and exits 0, unless a message is passed in which
// case it also prints that and exits 1.
func exitHelp(msg string) {
	print(helpText) //nolint:forbidigo

	if msg != "" {
		fmt.Printf("\n%s\n", msg) //nolint:forbidigo
		os.Exit(1)
	}

	os.Exit(0)
}

func getLatestStatFile(dir string) (string, string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return "", "", err
	}

	var latestFile os.DirEntry
	var latestModTime time.Time

	for _, file := range files {
		if file.IsDir() || !file.Type().IsRegular() || !strings.HasSuffix(file.Name(), statsFileSuffix) {
			continue
		}

		info, err := file.Info()
		if err != nil {
			return "", "", err
		}

		if info.ModTime().After(latestModTime) {
			latestModTime = info.ModTime()
			latestFile = file
		}
	}

	if latestFile == nil {
		return "", "", fmt.Errorf("no stats file found in directory: %s", dir)
	}

	fileName := latestFile.Name()

	return filepath.Join(dir, fileName), fileName[:8], nil
}

func decompress(path string) (io.ReadCloser, func() error, error) {
	cmd := exec.Command("pigz", "-d", "-c", path)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, nil, err
	}

	cleanup := func() error {
		return cmd.Wait()
	}

	return stdout, cleanup, nil
}

func parseStats(in io.ReadCloser, date string) {
	p := NewStatsParser(in)

	lines := 0

	for p.Scan() {
		lines++

		if lines == 1 {
			l.Printf("got path %s\n", p.Path)
		}
	}

	l.Printf("Parsed %d lines from %s file\n", lines, date)

	if p.Err() != nil {
		die(p.Err())
	}
}

func die(err error) {
	l.Printf("ERROR: %s", err.Error())
	os.Exit(1)
}
