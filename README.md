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
./dd-trace-shell brew
```

Result
------
![Screenshot 2021-12-17 at 12 31 08](https://user-images.githubusercontent.com/1836721/146539179-186416fe-729f-41fd-92ba-fbec9599ea94.png)
