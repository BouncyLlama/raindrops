# Cloud Monitor

Tool scrapes the status pages of AWS, Azure, GCP and reports the status of
each service. It always reports in text format to stdout and will dump to
influxdb if the 'influx' subcommand is used.

## Usage
```bash
usage: monitor <Command> [-h|--help] -r|--report (down|up|all) -c|--cloud
               (google|amazon|azure|all)

               monitors cloud host status

Commands:

  influx  Dump statii to influxdb

Arguments:

  -h  --help    Print help information
  -r  --report  report if services are down, up, or all. For the Google
                platform, we only retrieve if the service is down.
  -c  --cloud   which platforms to report on

```

### Influx subcommand
```bash
usage: monitor influx -i|--influx-url "<value>" -u|--username "<value>"
               -p|--pasword "<value>" -d|--influxdb "<value>" [-h|--help]
               -r|--report (down|up|all) -c|--cloud (google|amazon|azure|all)

               Dump statii to influxdb

Arguments:

  -i  --influx-url  The url to influxdb
  -u  --username    influxdb username
  -p  --pasword     influxdb password
  -d  --influxdb    db name to use
  -h  --help        Print help information
  -r  --report      report if services are down, up, or all. For the Google
                    platform, we only retrieve if the service is down.
  -c  --cloud       which platforms to report on

```