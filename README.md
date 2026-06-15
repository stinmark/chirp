## running application

1. build compile

- go build -o sigcat .

1. starting daemon

- nohup ./sigcat --start-watch > daemon.log 2>&1 &

1. stop daemon

- ./sigcat --stop-watch
