# Fptn HTTP DoS attack utility

## Build

Fptn is using packages from standard library only. Build is easy like this:

```
go build fptn.go
```

## Usage

```
Usage of ./fptn:
  -delay int
    	Sleep time in milliseconds between each request per worker. Can be increased for keep-alive attacks similar to slowloris
  -keep-alive
    	Whether to use keep-alive connections (true), or initiate new TCP connection on each request (false) (default true)
  -method string
    	HTTP method to use. Can be HEAD for low bandwidth attacks. POST payloads are not implemented (default "GET")
  -site string
    	Site URL to attack (default "https://kremlin.ru")
  -sites-file string
    	Path to file with URLs, each on a new line (default "./sites.txt")
  -workers int
    	Number of workers per URL (default 20)
```
