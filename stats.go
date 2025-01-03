// Copyright © 2024 Genome Research Limited
// Authors:
//  Sendu Bala <sb10@sanger.ac.uk>.
//  Dan Elia <de7@sanger.ac.uk>.
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
	"bufio"
	"encoding/base64"
	"io"
)

// Error is the type of the constant Err* variables.
type Error string

// Error returns a string version of the error.
func (e Error) Error() string { return string(e) }

const (
	fileType                   = byte('f')
	maxLineLength              = 64 * 1024
	maxBase64EncodedPathLength = 1024

	ErrBadPath       = Error("invalid file format: path is not base64 encoded")
	ErrTooFewColumns = Error("invalid file format: too few tab separated columns")
)

// StatsParser is used to parse wrstat stats files.
type StatsParser struct {
	scanner    *bufio.Scanner
	pathBuffer []byte
	lineBytes  []byte
	lineLength int
	lineIndex  int
	Path       []byte
	EntryType  byte
	error      error
}

// NewStatsParser is used to create a new StatsParser, given uncompressed wrstat
// stats data.
func NewStatsParser(r io.Reader) *StatsParser {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, maxLineLength), maxLineLength)

	return &StatsParser{
		scanner:    scanner,
		pathBuffer: make([]byte, base64.StdEncoding.DecodedLen(maxBase64EncodedPathLength)),
	}
}

// Scan is used to read the next line of stats data, which will then be
// available through the Path, Size, GID, MTime, CTime and EntryType properties.
//
// It returns false when the scan stops, either by reaching the end of the input
// or an error. After Scan returns false, the Err method will return any error
// that occurred during scanning, except that if it was io.EOF, Err will return
// nil.
func (p *StatsParser) Scan() bool {
	keepGoing := p.scanner.Scan()
	if !keepGoing {
		return false
	}

	return p.parseLine()
}

func (p *StatsParser) parseLine() bool {
	p.lineBytes = p.scanner.Bytes()
	p.lineLength = len(p.lineBytes)

	if p.lineLength <= 1 {
		return true
	}

	p.lineIndex = 0

	encodedPath, ok := p.parseNextColumn()
	if !ok {
		return false
	}

	if !p.skipColumns2to7() {
		return false
	}

	entryTypeCol, ok := p.parseNextColumn()
	if !ok {
		return false
	}

	p.EntryType = entryTypeCol[0]

	return p.decodePath(encodedPath)
}

func (p *StatsParser) skipColumns2to7() bool {
	for i := 0; i < 6; i++ {
		if _, ok := p.parseNextColumn(); !ok {
			return false
		}
	}

	return true
}

func (p *StatsParser) parseNextColumn() ([]byte, bool) {
	start := p.lineIndex

	for p.lineBytes[p.lineIndex] != '\t' {
		p.lineIndex++

		if p.lineIndex >= p.lineLength {
			p.error = ErrTooFewColumns

			return nil, false
		}
	}

	end := p.lineIndex
	p.lineIndex++

	return p.lineBytes[start:end], true
}

func (p *StatsParser) decodePath(encodedPath []byte) bool {
	l, err := base64.StdEncoding.Decode(p.pathBuffer, encodedPath)
	if err != nil {
		p.error = ErrBadPath

		return false
	}

	p.Path = p.pathBuffer[:l]

	return true
}

// Err returns the first non-EOF error that was encountered, available after
// Scan() returns false.
func (p *StatsParser) Err() error {
	return p.error
}
