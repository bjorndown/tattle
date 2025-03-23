# Tattle

Sends low disk space warnings to a mattermost channel. Best to schedule it with a systemd timer or similar. 

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
  "webhook": "https://mymattermost.com/webhook"
}
```
