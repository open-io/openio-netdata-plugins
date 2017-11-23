OpenIO Plugin for Netdata
===

Description
---

This plugin collects metrics from OpenIO services. Currently reported metrics are (more to come):

- Rawx: req_hits/req_time
- Metax: req_hits/req_time
- Score

Issues/enhancements are welcome!

Install
---

> Prerequisites:
- go 1.8+
- netdata 1.7+

From the cloned project:

```
$ go build openio.plugin.go
$ chmod +x openio.plugin
```

Test-run the plugin (Abort with Ctrl+C):
```
$ ./openio.plugin 1 --ns OPENIO
```

Install it:
```sh
$ sudo cp openio.plugin /usr/lib/netdata/plugins.d/
```

Add the following /etc/netdata/netdata.conf:
```
[plugin:openio]
    update every = 1
    command options = --ns OPENIO
```

> Replace OPENIO with your namespace name. If you have multiple namespaces on the machine, join the names with ":" (e.g. `command options = --ns OPENIO:OPENIO2`)

> This plugin searches for a valid namespace configuration in `/etc/oio/sds.conf.d`. If your configuration is stored somewhere else, specify the path with `--conf [PATH_TO_DIR]`.

Restart netdata:
```
$ systemctl restart netdata
```

Head to the dashboard at http://[IP]:19999, and look for an openio section.
