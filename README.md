OpenIO Plugin for Netdata
===

Description
---

This plugin collects metrics from OpenIO services. Currently reported metrics are (more on their way):

- Rawx: Request/Response info, connexion info, volume info (via statfs)
- Metax: Request/Response info, connexion info, volume info (via statfs)
- Score (for all scored services)

Suggestions are welcome!

Install
---

#### Prerequisites:
- go 1.8+
- netdata 1.7+
- *optional* influxdb
- *optional* gccgo


#### Build (static):

> There is a way to build the plugin dynamically, see **Compile with shared libraries**

```
$ cd
$ git clone [this repo] go/src/oionetdata
$ export GOPATH=${GOPATH:-$(go env GOPATH)}:$(pwd)/go/
$ cd $(pwd)/go/src/oionetdata
$ go build openio.plugin.go
$ chmod +x openio.plugin
```

Test-run the plugin (Abort with Ctrl+C):
```sh
$ ./openio.plugin 1 --ns OPENIO
```

#### Install:
```sh
$ sudo cp openio.plugin /usr/lib/netdata/plugins.d/
```

Add the following /etc/netdata/netdata.conf:
```ini
[plugin:openio]
    update every = 1
    command options = --ns OPENIO
```

> Replace OPENIO with your namespace name. If you have multiple namespaces on the machine, join the names with ":" (e.g. `command options = --ns OPENIO:OPENIO2`)

> This plugin searches for a valid namespace configuration in `/etc/oio/sds.conf.d`. If your configuration is stored somewhere else, specify the path with `--conf [PATH_TO_DIR]`.

Restart netdata:
```sh
$ systemctl restart netdata
```

Head to the dashboard at http://[IP]:19999, and look for an openio section.

InfluxDB
---

> We suppose that an InfluxDB is installed on the same machine

To integrate with InfluxDB, first enable the graphite backend in `/etc/netdata/netdata.conf`:


```ini
[backend]
     enabled = yes
     type = graphite
     destination = localhost
     prefix = netdata
     send charts matching = openio.*
```

Then in `/etc/influxdb/influxdb.conf`, add the following to graphite > templates:

```ini
"netdata.*.openio.*.*.*.*.host.measurement.measurement.ns.service.volume",
"netdata.*.openio.*.*.*.host.measurement.measurement.ns.service",
```

Restart both netdata and influxdb:

```sh
$ systemctl restart netdata influxdb
```

Query InfluxDB for the newly stored metrics:

```sh
$ curl -G 'http://localhost:8086/query?pretty=true' --data-urlencode "db=graphite" --data-urlencode "q=SELECT * from openio_byte_used limit 3"
{
    "results": [
        {
            "statement_id": 0,
            "series": [
                {
                    "name": "openio_byte_used",
                    "columns": [
                        "time",
                        "host",
                        "ns",
                        "service",
                        "value",
                        "volume"
                    ],
                    "values": [
                        [
                            "2017-12-04T21:38:32Z",
                            "myhost",
                            "OPENIO",
                            "192_168_50_2_6001",
                            0,
                            "_var_lib_oio_sds_OPENIO_meta0_0"
                        ],
                        [
                            "2017-12-04T21:38:32Z",
                            "myhost",
                            "OPENIO",
                            "192_168_50_2_6004",
                            0,
                            "_var_lib_oio_sds_OPENIO_rawx_0"
                        ],
                        [
                            "2017-12-04T21:38:32Z",
                            "myhost",
                            "OPENIO",
                            "192_168_50_2_6002",
                            0,
                            "_var_lib_oio_sds_OPENIO_meta1_0"
                        ]
                    ]
                }
            ]
        }
    ]
}
```

Compile with shared libraries
---

It is possible to compile the binary to make use of shared libraries. However, version requirements for both GCC and go
are not compatible with every distro. (GCC 7 + Go 1.8 required).

Supported distros:
- Fedora 26/27
- OpenSuse Thumbleweed
- Arch Linux

Below is an example of a build under Fedora 26 (done via docker):

```sh
$ docker run -ti fedora:26 /bin/bash
$ echo "fastestmirror=true" >> /etc/dnf/dnf.conf
$ dnf -y update && dnf -y install golang git gcc-go
$ cd
$ git clone [this repo] go/src/oionetdata
$ export GOPATH=${GOPATH:-$(go env GOPATH)}:$(pwd)/go/
$ cd $(pwd)/go/src/oionetdata
$ go build openio.plugin.go
```

File details:

```sh
$ du -sh openio.plugin
172K	openio.plugin

$ ldd openio.plugin
linux-vdso.so.1 (0x00007ffc46ef0000)
libgo.so.11 => /lib64/libgo.so.11 (0x00007f1711432000)
libm.so.6 => /lib64/libm.so.6 (0x00007f171111c000)
libgcc_s.so.1 => /lib64/libgcc_s.so.1 (0x00007f1710f05000)
libc.so.6 => /lib64/libc.so.6 (0x00007f1710b30000)
/lib64/ld-linux-x86-64.so.2 (0x00007f1712f5d000)
libpthread.so.0 => /lib64/libpthread.so.0 (0x00007f1710911000)
```


TODO
---

- Tests
- ~~Tag services with volume information~~
- ~~Make it work with InfluxDB~~
- More collectors: ZK
