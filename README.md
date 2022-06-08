## mpc

`mpc` is a configuration file automation ops tool that manages prometheus

### Features

- Support for adding, deleting, and checking the jobs, targets, and labels objects of the prometheus configuration file.
- Support for making prometheus configuration effective.
- Support for installing and starting exporter on remote servers.

<br>

### Usage scenarios

- If you use prometheus to monitor many targets (e.g. linux, mysql, redis, etc.), there are three steps for each new monitoring target, the first step is to install the exporter, the second step is to add the exporter's address and port to the prometheus configuration file, and the third step is to make the modified prometheus configuration take effect. The steps for deleting a monitoring target are similar.
- If a monitoring target needs to be tagged, prometheus uses tags to filter the desired monitoring data.

These scenarios are done using the `mpc` command and do not require frequent logins to different servers to operate, and if combined with the front-end, can be automated in the interface.

You can of course configure consul in prometheus so that consul automatically registers to discover monitoring targets

<br>

### Usage

**Install mpc**

> go install github.com/zhufuyi/mpc@latest

**Get  job targets**

> mpc get targets -f prometheus.yaml -n node_exporter

**Append new value to job targets**

> mpc add targets -f prometheus.yaml -n node_exporter -v 127.0.0.1:9100

**Delete address in job targets**

> mpc delete targets -f prometheus.yaml -n node_exporter -v 127.0.0.1:9100

**Replace job targets**

> mpc replace targets -f prometheus.yaml -n node_exporter -v 127.0.0.1:9100

**Install exporter on a remote server**

> mpc exec -u root -p 123456 -H 192.168.1.10 -P 22 -e node_exporter_install.sh -f node_exporter-1.3.1.linux-amd64.tar.gz

<br>

For more information on using the command, see the help.

```bash
$ mpc -h
manage prometheus configuration, add,delete,update job

Usage:
  mpc [command]

Available Commands:
  add         Add job,targets,labels to prometheus configuration file
  completion  Generate the autocompletion script for the specified shell
  delete      Delete job,targets,labels in prometheus configuration file
  exec        Install and run service to one remote server
  execs       Install and run service to multiple remote servers
  get         Show job,targets,labels from prometheus configuration file
  help        Help about any command
  reload      Make the prometheus configuration effective
  replace     Replace job,targets,labels to prometheus configuration file
  resources   List of supported resources

Flags:
  -h, --help      help for mpc
  -v, --version   version for mpc

Use "mpc [command] --help" for more information about a command.
```
