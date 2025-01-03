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
	"compress/gzip"
	"os"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestParseStats(t *testing.T) {
	Convey("Given a parser and reader", t, func() {
		f, err := os.Open("test.stats.gz")
		So(err, ShouldBeNil)

		defer f.Close()

		gr, err := gzip.NewReader(f)
		So(err, ShouldBeNil)

		defer gr.Close()

		p := NewStatsParser(gr, "prefix")
		So(p, ShouldNotBeNil)

		Convey("you can get extract info for all entries", func() {
			i := 0
			for p.Scan() {
				if i == 0 {
					So(string(p.Path), ShouldEqual, "/lustre/scratch122/tol/teams/blaxter/users/am75/assemblies/dataset/ilXesSexs1.2_genomic.fna") //nolint:lll
					So(p.EntryType, ShouldEqual, fileType)
				} else if i == 1 {
					So(string(p.Path), ShouldEqual, "/lustre/scratch122/tol/teams/blaxter/users/am75/assemblies/dataset/ilOpeBrum1.1_genomic.fna.fai") //nolint:lll
				}

				i++
			}
			So(i, ShouldEqual, 18890)

			So(p.Err(), ShouldBeNil)
		})
	})

	Convey("Scan generates Err() when", t, func() {
		prefix := "prefix"

		Convey("first column is not base64 encoded", func() {
			p := NewStatsParser(strings.NewReader("this is invalid since it has spaces\t1\t1\t1\t1\t1\t1\tf\t1\t1\td\n"), prefix)
			So(p.Scan(), ShouldBeFalse)
			So(p.Err(), ShouldEqual, ErrBadPath)
		})

		Convey("there are not enough tab separated columns", func() {
			encodedPath := "L2x1c3RyZS9zY3JhdGNoMTIyL3RvbC90ZWFtcy9ibGF4dGVyL3VzZXJzL2FtNzUvYXNzZW1ibGllcy9kYXRhc2V0L2lsWGVzU2V4czEuMl9nZW5vbWljLmZuYQ==" //nolint:lll

			p := NewStatsParser(strings.NewReader(encodedPath+"\t1\t1\t1\t1\t1\t1\tf\t1\t1\td\n"), prefix)
			So(p.Scan(), ShouldBeTrue)
			So(p.Err(), ShouldBeNil)

			p = NewStatsParser(strings.NewReader(encodedPath+"\t1\t1\t1\t1\t1\n"), prefix)
			So(p.Scan(), ShouldBeFalse)
			So(p.Err(), ShouldEqual, ErrTooFewColumns)

			p = NewStatsParser(strings.NewReader(encodedPath+"\t1\t1\t1\t1\n"), prefix)
			So(p.Scan(), ShouldBeFalse)
			So(p.Err(), ShouldEqual, ErrTooFewColumns)

			p = NewStatsParser(strings.NewReader(encodedPath+"\t1\t1\t1\n"), prefix)
			So(p.Scan(), ShouldBeFalse)
			So(p.Err(), ShouldEqual, ErrTooFewColumns)

			p = NewStatsParser(strings.NewReader(encodedPath+"\t1\t1\n"), prefix)
			So(p.Scan(), ShouldBeFalse)
			So(p.Err(), ShouldEqual, ErrTooFewColumns)

			p = NewStatsParser(strings.NewReader(encodedPath+"\t1\n"), prefix)
			So(p.Scan(), ShouldBeFalse)
			So(p.Err(), ShouldEqual, ErrTooFewColumns)

			p = NewStatsParser(strings.NewReader(encodedPath+"\n"), prefix)
			So(p.Scan(), ShouldBeFalse)
			So(p.Err(), ShouldEqual, ErrTooFewColumns)

			Convey("but not for blank lines", func() {
				p = NewStatsParser(strings.NewReader("\n"), "prefix")
				So(p.Scan(), ShouldBeTrue)
				So(p.Err(), ShouldBeNil)

				p := NewStatsParser(strings.NewReader(""), "prefix")
				So(p.Scan(), ShouldBeFalse)
				So(p.Err(), ShouldBeNil)
			})
		})
	})
}
