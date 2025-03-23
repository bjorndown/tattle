# Tattle

Performs checks on a Linux machine. Sends alerts to a mattermost channel. Best to schedule it with a systemd timer or similar. 

Currently supported checks
 
* low disk space (via `df`)
* non-active systemd user units (via `systemctl --user`)

## Build

```shell
go build
```

## Run

```shell
./tattle
```

## Help 

```shell
./tattle --help
```

## Config file

Tattle expects a config file of the following shape

```json
{
  "disk": {
    "thresholds": [
      {
        "target": "/",
        "percent": 80
      }
    ]
  },
  "systemd": {
    "activeUnits": ["foo.service"]
  },
  "webhook": "https://mymattermost.com/webhook"
}
```
