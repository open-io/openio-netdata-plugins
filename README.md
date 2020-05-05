OpenIO Plugins for Netdata
===

Description
---

This plugin suite provides different collector to be used with openio products.

Current collectors are:

- openio (SDS conscience and service metrics)
- container (SDS stored data information)
- zookeeper (Zookeeper metrics)
- command (Arbitrary commands, used for version information)
- fs (Filesystem connector metrics)
- s3roundtrip (S3 roundtrip check using AWS SDK)
- redis (redis metrics)
- memcached (memcached metrics)
- beanstalk (beanstalk metrics)

Install
---

#### Prerequisites:
- go 1.8+ (additional testing required for earlier versions)
- netdata 1.7+
- (*optional*) influxdb
- go get github.com/golang/mock/gomock for tests


#### Build:

```
$ cd
$ git clone [this repo] go/src/oionetdata
$ go get github.com/go-redis/redis
$ go get github.com/aws/aws-sdk-go
$ go get gopkg.in/yaml.v2
$ export GOPATH=${GOPATH:-$(go env GOPATH)}:$(pwd)/go/
$ cd $(pwd)/go/src/oionetdata
$ go build ./cmd/openio.plugin/openio.plugin.go;
$ go build ./cmd/zookeeper.plugin/zookeeper.plugin.go;
$ go build ./cmd/container.plugin/container.plugin.go
$ go build ./cmd/command.plugin/command.plugin.go
$ go build ./cmd/oiofs.plugin/oiofs.plugin.go
$ go build ./cmd/s3roundtrip.plugin/s3roundtrip.plugin.go
$ go build ./cmd/redis.plugin/redis.plugin.go
$ go build ./cmd/memcached.plugin/memcached.plugin.go
$ go build ./cmd/beanstalk.plugin/beanstalk.plugin.go
```

Type in `./[name].plugin -h` to get all available options for each plugin

#### Install:

CentOS 7
```sh
$ cp openio.plugin /usr/libexec/netdata/plugins.d/
$ cp zookeeper.plugin /usr/libexec/netdata/plugins.d/
$ cp container.plugin /usr/libexec/netdata/plugins.d/
$ cp command.plugin /usr/libexec/netdata/plugins.d/
$ cp oiofs.plugin /usr/libexec/netdata/plugins.d/
$ cp s3roundtrip.plugin /usr/libexec/netdata/plugins.d/
$ cp redis.plugin /usr/libexec/netdata/plugins.d/
$ cp memcached.plugin /usr/libexec/netdata/plugins.d/
$ cp beanstalk.plugin /usr/libexec/netdata/plugins.d/
```

Ubuntu Xenial
```sh
$ cp openio.plugin /usr/lib/x86_64-linux-gnu/netdata/plugins.d/
$ cp zookeeper.plugin /usr/lib/x86_64-linux-gnu/netdata/plugins.d/
$ cp container.plugin /usr/lib/x86_64-linux-gnu/netdata/plugins.d/
$ cp command.plugin /usr/libexec/netdata/plugins.d/
$ cp oiofs.plugin /usr/libexec/netdata/plugins.d/
$ cp s3roundtrip.plugin /usr/libexec/netdata/plugins.d/
$ cp redis.plugin /usr/libexec/netdata/plugins.d/
$ cp memcached.plugin /usr/libexec/netdata/plugins.d/
$ cp beanstalk.plugin /usr/libexec/netdata/plugins.d/
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

[plugin:command]
    update every = 10
    command options =

[plugin:fs]
    update every = 10

[plugin:s3roundtrip]
    update every = 10

[plugin:redis]
  update every = 10
  command options = --targets 172.30.2.106:6011:redis

[plugin:memcached]
  update every = 10
  command options = --targets 172.30.2.106:6011

[plugin:beanstalk]
  update every = 10
  command options = --targets 172.30.2.106:6014:tube1:tube2:tube3
```

Create and configure plugin config files

```
# /etc/netdata/s3-roundtrip.conf
endpoint=http://localhost:6007
access=
secret=
region=us-east-1
bucket=bucket-roundtrip
object=file-roundtrip
timeout=3
```

Optionally you can add:

```
size=500000
mpu_size=50000000
make_bucket=true
user_agent=OIO-S3RT
```

To control simple object size, MPU size, and whether to roundtrip bucket creation


Since 0.6.0, netdata config files have the following format:
```
# /etc/netdata/commands.yml
config:
  - name: openio_version
    command: "rpm -q --qf '%{VERSION}\n' openio-sds-server"
    interval: 60
    family: version
    value_is_label: true
  - name: swift_version
    command: "rpm -q --qf '%{VERSION}\n' openio-sds-swift
    interval: 60
    family: version
    value_is_label: true
  - name: swift_version
    command: "rpm -q --qf '%{VERSION}\n' openio-sds-swift
    interval: 60
    family: version
    value_is_label: true
[...]
```

For backward-compatibility, .conf files are still accepted:
```
# /etc/netdata/commands.conf
openio_version=rpm -q --qf "%{VERSION}\n" openio-sds-server
swift_version=rpm -q --qf "%{VERSION}\n" openio-sds-swift
s3_version=rpm -q --qf "%{VERSION}\n" openio-sds-swift-plugin-swift3
redis_version=redis-server --version | grep -oP ' v=\K.+? '
zk_version=rpm -q --qf "%{VERSION}\n" zookeeper
zk_lib_version=rpm -q --qf "%{VERSION}\n" zookeeper-lib
beanstalkd_version=beanstalkd -v | awk '{print $2}'
oiofs_version=rpm -q --qf "%{VERSION}\n" oiofs-fuse
keystone_version=keystone-manage --version
```

```
# /etc/netdata/oiofs.conf
/mnt/test=localhost:9000
```


> Replace OPENIO with your namespace name. If you have multiple namespaces on the machine, join the names with ":" (e.g. `command options = --ns OPENIO:OPENIO2`)

> This plugin searches for a valid namespace configuration in `/etc/oio/sds.conf.d`. If your configuration is stored somewhere else, specify the path with `--conf [PATH_TO_DIR]`. For the container plugin, point the option to `/etc/oio/sds/` (directory containing per-namespace configuration)

Restart netdata:
```sh
$ systemctl restart netdata
```

Tests
---

Modules can be tested by running `go test oionetdata/[module]`

Mocks for S3 plugin have been generated as follows:

```sh
mockgen github.com/aws/aws-sdk-go/service/s3/s3iface S3API > s3roundtrip/mocks.go
mockgen github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface UploaderAPI,DownloaderAPI > mock_s3manageriface/mock_s3manageriface.go
```

TODO
---

- Tests for openio/container
- Reload/Update mechanism
- Automatic namespace detection
- Container: cache containers above threshold, separate slow/fast listing
- Container: consider connecting to sentinel via FailoverClient
- improve error handling
- Migrate modules to new API (openio, container)

License
---

[GNU Affero General Public License (AGPL v3)](https://www.gnu.org/licenses/agpl-3.0.html)
