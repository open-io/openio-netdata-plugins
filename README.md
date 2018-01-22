OpenIO Plugin for Netdata
===

Description
---

This plugin collects metrics from OpenIO services. Currently reported metrics are (more on their way):

- Rawx: Request/Response info, connexion info, volume info (via statfs)
- Metax: Request/Response info, connexion info, volume info (via statfs)
- Score (for all scored services)
- Zookeeper metrics for local Zookeeper instances

Suggestions are welcome!

Install
---

#### Prerequisites:
- go 1.8+ (additional testing required for earlier versions)
- netdata 1.7+
- (*optional*) influxdb


#### Build:

```
$ cd
$ git clone [this repo] go/src/oionetdata
$ export GOPATH=${GOPATH:-$(go env GOPATH)}:$(pwd)/go/
$ cd $(pwd)/go/src/oionetdata
$ go build openio.plugin.go; go build zookeeper.plugin.go
$ chmod +x openio.plugin zookeeper.plugin
```

Test-run the plugins (Abort with Ctrl+C):

> As metrics are gathered for __local services__, there might not be any output from those plugins on the test machine (e.g. if it isn't an OpenIO node). Also make sure you have a valid OPENIO config file in `/etc/oio/sds.conf.d/OPENIO`

```sh
$ ./openio.plugin 1 --ns OPENIO
$ ./zookeeper.plugin 1 --ns OPENIO
```

#### Install:
```sh
$ sudo cp openio.plugin /usr/lib/netdata/plugins.d/
$ sudo cp zookeeper.plugin /usr/lib/netdata/plugins.d/
```

Add the following /etc/netdata/netdata.conf:
```ini
[plugin:openio]
    update every = 1
    command options = --ns OPENIO

[plugin:zookeeper]
    update every = 1
    command options = --ns OPENIO
```

> Replace OPENIO with your namespace name. If you have multiple namespaces on the machine, join the names with ":" (e.g. `command options = --ns OPENIO:OPENIO2`)

> This plugin searches for a valid namespace configuration in `/etc/oio/sds.conf.d`. If your configuration is stored somewhere else, specify the path with `--conf [PATH_TO_DIR]`.

Restart netdata:
```sh
$ systemctl restart netdata
```

Head to the dashboard at http://[IP]:19999, and look for an __openio__ section.

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

TODO
---

- Tests
- ~~Tag services with volume information~~
- ~~Make it work with InfluxDB~~
- ~~More collectors: ZK~~
- Reload/Update mechanism
- Automatic namespace detection
