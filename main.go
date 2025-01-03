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
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	statsFileSuffix         = ".stats.gz"
	condarcSuffix           = ".condarc"
	condaMetaOutputSuffix   = ".conda-meta"
	condaMetaSuffix         = "/conda-meta/history"
	singularityOutputSuffix = ".singularity"
)

var singularitySuffixes = []string{".sif", ".simg", ".img"}

const helpText = `condapaths parses wrstat stats.gz files quickly, in low mem.

Provide one or more stats.gz files output by wrstat.

It outputs files with one path per line:
* <input prefix>.condarc: paths where the file basename was ".condarc"
* <input prefix>.conda-meta: paths where the file basename was "history", in a
                             directory named "conda-meta"
* <input prefix>.singularity: paths where the file basename suffix was one of
                              ".sif",  ".simg", and ".img"

Usage: condapaths 20241222_mount.unique.stats.gz
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
		exitHelp("ERROR: you must provide at least 1 wrstat stats file")
	}

	for _, statsPath := range flag.Args() {
		prefix, err := getPathPrefix(statsPath)
		if err != nil {
			die(err)
		}

		input, cleanup, err := decompress(statsPath)
		if err != nil {
			die(err)
		}

		err = parseStats(input, prefix)

		cleanup()

		if err != nil {
			die(err)
		}
	}
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

func getPathPrefix(path string) (string, error) {
	if !strings.HasSuffix(path, statsFileSuffix) {
		return "", fmt.Errorf("path must end with %s", statsFileSuffix)
	}

	base := filepath.Base(path)
	idx := strings.Index(base, ".")

	return base[:idx], nil
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

func parseStats(in io.ReadCloser, prefix string) error {
	defer in.Close()

	rcOut, err := os.Create(prefix + condarcSuffix)
	if err != nil {
		return err
	}

	defer rcOut.Close()

	cmOut, err := os.Create(prefix + condaMetaOutputSuffix)
	if err != nil {
		return err
	}

	defer cmOut.Close()

	smOut, err := os.Create(prefix + singularityOutputSuffix)
	if err != nil {
		return err
	}

	defer smOut.Close()

	p := NewStatsParser(in)

	condarcSuffixBytes := []byte(condarcSuffix)
	condaMetaSuffixBytes := []byte(condaMetaSuffix)
	singularitySuffixesBytes := make([][]byte, len(singularitySuffixes))
	for i, suffix := range singularitySuffixes {
		singularitySuffixesBytes[i] = []byte(suffix)
	}

	for p.Scan() {
		if p.EntryType != fileType {
			continue
		}

		switch {
		case bytes.HasSuffix(p.Path, condarcSuffixBytes):
			if _, err := rcOut.Write(append(p.Path, '\n')); err != nil {
				return err
			}
		case bytes.HasSuffix(p.Path, condaMetaSuffixBytes):
			if _, err := cmOut.Write(append(p.Path, '\n')); err != nil {
				return err
			}
		default:
			for _, suffix := range singularitySuffixesBytes {
				if bytes.HasSuffix(p.Path, suffix) {
					if _, err := smOut.Write(append(p.Path, '\n')); err != nil {
						return err
					}
					break
				}
			}
		}
	}

	return p.Err()
}

func die(err error) {
	l.Printf("ERROR: %s", err.Error())
	os.Exit(1)
}
