#`tail -f` over multiple files

## Build and install

depends on fsnotify:
```console
go get github.com/fsnotify/fsnotify
go get -u golang.org/x/sys/...
```

## usage
```console
  ./supertail file_1 file_2 file_n

  # e.g.
  ./supertail /var/log/postgres.log /var/log/mysql.log
```
