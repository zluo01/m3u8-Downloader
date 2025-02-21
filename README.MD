# m3u8-Downloader

Download m3u8 media with multithreading, support decryption.

## Build

```
go build
```

## How to use

`./m3u8-Downloader -u "http://wwww.example.com/example.m3u8" -o video.ts`

`./m3u8-Downloader -f example.m3u8 -H Referer:http://www.example.com -H 'User-Agent:Chrome/83.0.4103.61 Safari/537.36'`

Note: When using the -f parameter, if the m3u8 file does not contain a specific link to the media, but only the media
name, you must specify the -u parameter

```
./m3u8-Downloader -h

Usage:
    ./m3u8-Downloader [Options]

Options:
  -H string
        http header. Example: Referer:https://www.example.com
  -f string
        local m3u8 file
  -n int
        thread number (default 10)
  -o string
        custom output file name. Example: <EXAMPLE>.ts
  -p string
        proxy. Example: http://127.0.0.1:8080
  -r int
        number of retries (default 3)
  -s int
        Start chunk number
  -t duration
        timeout (default 30s)
  -u string
        url of m3u8 file
```
