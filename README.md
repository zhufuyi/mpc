## mpc

`mpc`是一个管理prometheus的配置文件自动化运维工具，

特点：
- 支持对prometheus配置文件的job、targets、labels三个对象增删改查。
- 支持使prometheus配置生效。
- 支持在远程服务器安装和启动exporter。

<br>

使用场景：

(1) 如果使用prometheus监控很多目标(例如n个linux、mysql、redis等)，每新增一个监控目标有三个步骤，第一步安装exporter，第二步把exporter的地址和端口添加到prometheus配置文件，第三步使修改后的prometheus配置生效。
同样在prometheus剔除一个监控目标，也需要是三个步骤，第一步在prometheus删除目标地址，第二步使prometheus配置生效，第三步卸载exporter。

(2) 如果监控目标需要打标签，prometheus就用标签来过滤筛选想要的监控数据。

这些场景使用mpc命令来完成，不需要频繁登录不同服务器去操作，如果结合前端，可以实现在界面自动化操作。

<br>

帮助命令：

```bash
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
