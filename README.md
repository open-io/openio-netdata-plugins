OpenIO Plugin for Netdata
===

Description
---

This plugin collects metrics from OpenIO services. Currently reported metrics are (more on their way):

- Rawx: Request/Response info, connexion info, volume info (via statfs)
- Metax: Request/Response info, connexion info, volume info (via statfs)
- Score (for all scored services)
- Zookeeper metrics for local Zookeeper instances
- Account container listing (account container count, size and object count for containers above threshold)

#### Zookeeper

This plugin provides statistics from zookeeper server. Statistics are collected using `mntr` command.

Charts:

1. **latency** - latency in milliseconds
 * min
 * max
 * avg

2. **packets** - packets stats in packets/s
 * received
 * sent

3. **connections** - connections count
 * alive

4. **requests** - requests count
 * outstanding

5. **nodes** - nodes count
 * znode
 * watch
 * ephemeral

6. **data** - data size
 * size

7. **fds** - file descriptors count
 * open
 * max

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
$ go get github.com/go-redis/redis
$ export GOPATH=${GOPATH:-$(go env GOPATH)}:$(pwd)/go/
$ cd $(pwd)/go/src/oionetdata
$ go build ./cmd/openio.plugin/openio.plugin.go; go build ./cmd/zookeeper.plugin/zookeeper.plugin.go; go build ./cmd/container.plugin/container.plugin.go
```

Test-run the plugins (Abort with Ctrl+C):

> As metrics are gathered for __local services__, there might not be any output from those plugins on the test machine (e.g. if it isn't an OpenIO node). Also make sure you have a valid OPENIO config file in `/etc/oio/sds.conf.d/OPENIO`. The only exception is the __container__ plugin, which requires a local redis and a redis configuration file in `/etc/oio/sds/OPENIO/redis-X/redis.conf`.

```sh
$ ./openio.plugin 1 --ns OPENIO
$ ./zookeeper.plugin 1 --ns OPENIO
$ ./container.plugin 10 --ns OPENIO
```

Type in `./[name].plugin -h` to get all available options for each plugin

#### Install:

CentOS 7
```sh
$ cp openio.plugin /usr/libexec/netdata/plugins.d/
$ cp zookeeper.plugin /usr/libexec/netdata/plugins.d/
$ cp container.plugin /usr/libexec/netdata/plugins.d/
```

Ubuntu Xenial
```sh
$ cp openio.plugin /usr/lib/x86_64-linux-gnu/netdata/plugins.d/
$ cp zookeeper.plugin /usr/lib/x86_64-linux-gnu/netdata/plugins.d/
$ cp container.plugin /usr/lib/x86_64-linux-gnu/netdata/plugins.d/
```

Add the following /etc/netdata/netdata.conf:
```ini
[plugin:openio]
    update every = 1
    command options = --ns OPENIO

[plugin:zookeeper]
    update every = 1
    command options = --ns OPENIO

[plugin:container]
    update every = 60
    command options = --ns OPENIO --threshold 0 --limit 1000
```

> Replace OPENIO with your namespace name. If you have multiple namespaces on the machine, join the names with ":" (e.g. `command options = --ns OPENIO:OPENIO2`)

> This plugin searches for a valid namespace configuration in `/etc/oio/sds.conf.d`. If your configuration is stored somewhere else, specify the path with `--conf [PATH_TO_DIR]`. For the container plugin, point the option to `/etc/oio/sds/` (directory containing per-namespace configuration)

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
"netdata.*.openio.container_bytes.*.*.* .host.measurement.measurement.ns.account.container",
"netdata.*.openio.container_objects.*.*.* .host.measurement.measurement.ns.account.container",
"netdata.*.openio.container_count.*.* .host.measurement.measurement.ns.account",
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
- ~~More collectors: container~~
- Reload/Update mechanism
- Automatic namespace detection
- Container: cache containers above threshold, separate slow/fast listing
- Container: consider connecting to sentinel via FailoverClient
- OpenIO: implement error handling
- Zookeeper: implement error handling
- Container: improve error handling

License
---

[GNU Affero General Public License (AGPL v3)](https://www.gnu.org/licenses/agpl-3.0.html)
