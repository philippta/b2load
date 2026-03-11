<h1 align="center">
    <br>
    b(uild h)2load
    <br>
    <br>
</h1>

b2load builds h2load commands. that's about it.

<img src=".github/demo.gif">

<br />

### install
```
go install github.com/philippta/b2load@latest
```


### usage

1. run b2load
2. fill in request parameters
3. hit enter
4. enjoy the h2load command

```
$ b2load
h2load https://localhost:8443
--h1 (use http/1.1)
  -D 10s (duration)
  -c 100 (clients)
  -t 1 (threads)
  -m 100 (max concurrent streams)
  -H Accept-Encoding: gzip

<ctrl-h> add header | <ctrl-x> remove header | <enter> build

$ h2load -D 10s -c 100 -t 1 -m 100 -H 'Accept-Encoding: gzip' https://localhost:8443
```

### license
[MIT](/LICENSE)
