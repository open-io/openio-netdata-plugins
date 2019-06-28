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

package util

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestVolumeInfo(t *testing.T) {
	info, id, err := VolumeInfo("/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Logf("info: %v, id: %v", info, id)
	// TODO verify output
}

func TestReadConf(t *testing.T) {
	tests := []struct {
		name         string
		conf         string
		expectedConf map[string]string
		separator    string
	}{
		{
			name:         "empty conf",
			expectedConf: map[string]string{},
		},
		{
			name: "sds conf",
			conf: `
[OPENIO]
conscience=10.240.0.13:6000
zookeeper=10.240.0.11:6005,10.240.0.12:6005,10.240.0.13:6005
proxy=10.240.0.13:6006
`,
			expectedConf: map[string]string{
				"conscience": "10.240.0.13:6000",
				"zookeeper":  "10.240.0.11:6005,10.240.0.12:6005,10.240.0.13:6005",
				"proxy":      "10.240.0.13:6006",
			},
		},
		{
			name: "redis conf",
			conf: `
daemonize no
pidfile "/var/lib/oio/sds/OPENIO/redis-1/redis-1.pid"
port 6011
bind 10.240.0.13
`,
			expectedConf: map[string]string{
				"daemonize": "no",
				"pidfile":   "\"/var/lib/oio/sds/OPENIO/redis-1/redis-1.pid\"",
				"port":      "6011",
				"bind":      "10.240.0.13",
			},
			separator: " ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := ioutil.TempDir("", "test_readconf_")
			if err != nil {
				t.Fatalf("failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(path)

			confPath := filepath.Join(path, "config")
			f, err := os.OpenFile(confPath, os.O_CREATE|os.O_RDWR, 0644)
			if err != nil {
				t.Fatalf("failed to create file: %v", err)
			}

			_, err = f.WriteString(tt.conf)
			if err != nil {
				t.Fatal(err)
			}
			f.Close()

			separator := "="
			if len(tt.separator) != 0 {
				separator = tt.separator
			}
			conf, err := ReadConf(confPath, separator)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(conf, tt.expectedConf) {
				t.Fatalf("unexpected conf got\n%v\n expected\n%v", conf, tt.expectedConf)
			}
		})
	}

}
