# Cloud Monitor

Tool scrapes the status pages of AWS, Azure, GCP and reports the status of
each service. It always reports in text format to stdout and can also dump to Mongo.

## Usage - status command
```bash
usage: raindrops status -r|--report (down|up|all) [-h|--help] -c|--cloud
                 (google|amazon|azure|all) [-m|--mongo-url "<value>"]
                 [-u|--username "<value>"] [-p|--password "<value>"]
                 [-d|--mongo-dbname "<value>"]

                 scrape current statii

Arguments:

  -r  --report        report if services are types.Down, types.Up, or
                      types.All. For the Google platform, we only retrieve if
                      the service is types.Down.
  -h  --help          Print help information
  -c  --cloud         which platforms to report on
  -m  --mongo-url     The url to mongo
  -u  --username      mongo username
  -p  --password      mongo password
  -d  --mongo-dbname  db name to use


```

### Usage - incidents command
```bash
usage: raindrops incidents [-h|--help] -c|--cloud (google|amazon|azure|all)
                 [-m|--mongo-url "<value>"] [-u|--username "<value>"]
                 [-p|--password "<value>"] [-d|--mongo-dbname "<value>"]

                 scrape incident descriptions

Arguments:

  -h  --help          Print help information
  -c  --cloud         which platforms to report on
  -m  --mongo-url     The url to mongo
  -u  --username      mongo username
  -p  --password      mongo password
  -d  --mongo-dbname  db name to use

```