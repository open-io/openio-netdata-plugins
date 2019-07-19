// OpenIO netdata collectors
// Copyright (C) 2019 OpenIO SAS
//
// This library is free software; you can redistribute it and/or
// modify it under the terms of the GNU Lesser General Public
// License as published by the Free Software Foundation; either
// version 3.0 of the License, or (at your option) any later version.
//
// This library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// Lesser General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public
// License along with this program. If not, see <http://www.gnu.org/licenses/>.

package openio

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"oionetdata/netdata"
	"testing"
	"time"
)

var testAddr = "127.0.0.1:6006"

type testServer struct{}

func newTestServer() *testServer {
	return &testServer{}
}

func (s *testServer) Run() {
	http.HandleFunc("/stat", func(w http.ResponseWriter, r *http.Request) {
		b, err := ioutil.ReadFile("./testdata/stat_rawx")
		if err != nil {
			fmt.Print(err)
		}
		fmt.Fprintf(w, string(b))
	})

	http.HandleFunc("/v3.0/OPENIO/conscience/info", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w,
			"[\"account\",\"beanstalkd\",\"meta0\",\"meta1\",\"meta2\",\"oioproxy\",\"rawx\",\"rdir\",\"redis\",\"sqlx\"]")
	})

	http.HandleFunc("/v3.0/OPENIO/conscience/list", func(w http.ResponseWriter, r *http.Request) {
		b, err := ioutil.ReadFile(fmt.Sprintf("./testdata/types_%s.json", r.URL.Query().Get("type")))
		if err != nil {
			fmt.Println("Warning, type not implemented", r.URL.Query().Get("type"))
			fmt.Fprintf(w, "[]")
			return
		}
		fmt.Fprintf(w, string(b))
	})

	http.HandleFunc("/v3.0/OPENIO/forward/stats", func(w http.ResponseWriter, r *http.Request) {
		// Not implemented
		// b, err := ioutil.ReadFile("./testdata/stat_metax")
		// if err != nil {
		// 	fmt.Print(err)
		// }
		// fmt.Fprintf(w, string(b))
	})

	http.HandleFunc("/v3.0/OPENIO/forward/info", func(w http.ResponseWriter, r *http.Request) {
		// Not implemented
		// b, err := ioutil.ReadFile(s.specFile)
		// if err != nil {
		// 	fmt.Print(err)
		// }
		// fmt.Fprintf(w, string(b))
	})
	err := http.ListenAndServe(testAddr, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func TestOpenIOCollector(t *testing.T) {
	srv := newTestServer()
	go srv.Run()

	c := make(chan netdata.Metric, 1e5)
	go Collect(testAddr, "OPENIO", c)

	time.Sleep(time.Duration(2) * time.Second)
	close(c)
	// for m := range c {
	//     // fmt.Println(m)
	//     // buf[m.Chart] = append(buf[m.Chart], fmt.Sprintf("SET %s %s\n", m.Dim, m.Value)...)
	// }
}
