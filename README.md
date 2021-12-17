dd-trace-shell
---------------
First attempt to trace a shell script with Datadog's go tracer.

WARNING
-------
Avoid passing credentials as command line arguments in your shell script. Otherwise, they will be sent to Datadog.

Example
--------
```
go build .
./dd-trace-shell brew update
```

Result
------
![Screenshot 2021-12-17 at 13 30 04](https://user-images.githubusercontent.com/1836721/146544964-7827f5b6-5901-4bf2-af82-389f2579c06c.png)

TODO
----
- [ ] Remove Datadog agent and dd-go-tracer from the dependency
- [ ] Write own `ps` library in order to reduce `ps` calls
- [ ] Make `interactive` (currently, `stdin` and `stdout` are not working)
