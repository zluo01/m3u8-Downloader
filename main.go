package main

import (
	"bytes"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"github.com/cheggaaa/pb/v3"
	"github.com/grafov/m3u8"
	"io"
	"io/ioutil"
	"log"
	"m3u8-Downloader-Go/decrypter"
	"m3u8-Downloader-Go/hackpool"
	"m3u8-Downloader-Go/request"
	"m3u8-Downloader-Go/sort"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	Client       *request.ReqClient
	keyCache     = map[string][]byte{}
	keyCacheLock sync.Mutex
	headers      map[string]string
	directory    string
)

var (
	URL       = flag.String("u", "", "url of m3u8 file")
	File      = flag.String("f", "", "local m3u8 file")
	ThreadNum = flag.Int("n", 10, "thread number")
	OutFile   = flag.String("o", "", "out file")
	Retry     = flag.Int("r", 3, "number of retries")
	Timeout   = flag.Duration("t", time.Second*30, "timeout")
	Proxy     = flag.String("p", "", "proxy. Example: http://127.0.0.1:8080")
	Headers   = flag.String("H", "", "http header. Example: Referer:https://www.example.com")
)

func init() {
	flag.Parse()

	if *URL == "" && *File == "" {
		fmt.Println("You must set the -u or -f parameter")
		flag.Usage()
	}

	if *ThreadNum <= 0 {
		*ThreadNum = 10
	}

	if *Retry <= 0 {
		*Retry = 1
	}

	if *Timeout <= 0 {
		*Timeout = time.Second * 30
	}

	if len(*Headers) > 0 {
		headers = map[string]string{}
		for _, header := range strings.Split(*Headers, ";") {
			s := strings.SplitN(header, ":", 2)
			key := strings.TrimRight(s[0], " ")
			if len(s) == 2 {
				headers[key] = strings.TrimLeft(s[1], " ")
			} else {
				headers[key] = ""
			}
		}
	}
}

func start(mpl *m3u8.MediaPlaylist) {
	pool := hackpool.New(*ThreadNum, download)

	var count = int(mpl.Count())
	go func(count int) {
		for i := 0; i < count; i++ {
			pool.Push(i, mpl.Segments[i], mpl.Key)
		}
		pool.CloseQueue()
	}(count)

	pool.Run(count)
}

func getKey(url string) ([]byte, error) {
	keyCacheLock.Lock()
	defer keyCacheLock.Unlock()

	key := keyCache[url]
	if key != nil {
		return key, nil
	}

	key, err := Client.Get(url, headers, *Retry)
	if err != nil {
		return nil, err
	}

	keyCache[url] = key

	return key, nil
}

func download(args ...interface{}) {
	id := args[0].(int)
	segment := args[1].(*m3u8.MediaSegment)
	globalKey := args[2].(*m3u8.Key)

	data, err := Client.Get(segment.URI, headers, *Retry)
	if err != nil {
		log.Fatalln("[-] Download failed:", id, err)
	}

	var keyURL, ivStr string
	if segment.Key != nil && segment.Key.URI != "" {
		keyURL = segment.Key.URI
		ivStr = segment.Key.IV
	} else if globalKey != nil && globalKey.URI != "" {
		keyURL = globalKey.URI
		ivStr = globalKey.IV
	}

	if keyURL != "" {
		var key, iv []byte
		key, err = getKey(keyURL)
		if err != nil {
			log.Fatalln("[-] Download key failed:", keyURL, err)
		}

		if ivStr != "" {
			iv, err = hex.DecodeString(strings.TrimPrefix(ivStr, "0x"))
			if err != nil {
				log.Fatalln("[-] Decode iv failed:", err)
			}
		} else {
			iv = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, byte(id)}
		}

		data, err = decrypter.Decrypt(data, key, iv)
		if err != nil {
			log.Fatalln("[-] Decrypt failed:", err)
		}
	}

	if err := ioutil.WriteFile(path.Join(directory, filename(segment.URI)), data, 0755); err != nil {
		log.Fatal(err)
	}
}

func filename(u string) string {
	obj, _ := url.Parse(u)
	_, filename := filepath.Split(obj.Path)
	return filename
}

func DownloadM3u8(m3u8URL string) ([]byte, error) {
	data, err := Client.Get(m3u8URL, headers, *Retry)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func ParseM3u8(data []byte) (*m3u8.MediaPlaylist, error) {
	playlist, listType, err := m3u8.Decode(*bytes.NewBuffer(data), true)
	if err != nil {
		return nil, err
	}

	if listType == m3u8.MEDIA {
		var obj *url.URL
		if *URL != "" {
			obj, err = url.Parse(*URL)
			if err != nil {
				return nil, errors.New("parse m3u8 url failed: " + err.Error())
			}
		}

		mpl := playlist.(*m3u8.MediaPlaylist)

		if mpl.Key != nil && mpl.Key.URI != "" {
			uri, err := formatURI(obj, mpl.Key.URI)
			if err != nil {
				return nil, err
			}
			mpl.Key.URI = uri
		}

		count := int(mpl.Count())
		for i := 0; i < count; i++ {
			segment := mpl.Segments[i]

			uri, err := formatURI(obj, segment.URI)
			if err != nil {
				return nil, err
			}
			segment.URI = uri

			if segment.Key != nil && segment.Key.URI != "" {
				uri, err := formatURI(obj, segment.Key.URI)
				if err != nil {
					return nil, err
				}
				segment.Key.URI = uri
			}

			mpl.Segments[i] = segment
		}

		return mpl, nil
	}

	return nil, errors.New("unsupported m3u8 type")
}

func formatURI(base *url.URL, u string) (string, error) {
	if strings.HasPrefix(u, "http") {
		return u, nil
	}

	if base == nil {
		return "", errors.New("you must set m3u8 url for " + *File + " to download")
	}

	obj, err := base.Parse(u)
	if err != nil {
		return "", err
	}

	return obj.String(), nil
}

func combinedFile() {
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		log.Fatal(err)
	}

	output, err := os.OpenFile(*OutFile, os.O_CREATE|os.O_TRUNC|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(output)

	sort.Compare(sort.CompareStringNumber).Sort(files)
	bar := pb.StartNew(len(files))
	for _, i := range files {
		input, err := os.Open(path.Join(directory, i.Name()))
		if err != nil {
			log.Fatal(err)
		}
		if _, err := io.Copy(output, input); err != nil {
			log.Fatal(err)
		}
		bar.Increment()
	}
	bar.Finish()
}

func main() {
	var err error
	Client, err = request.New(*Timeout, *Proxy)
	if err != nil {
		log.Fatalln("[-] Init failed:", err)
	}

	t := time.Now()

	var data []byte
	if *File != "" {
		data, err = ioutil.ReadFile(*File)
		if err != nil {
			log.Fatalln("[-] Load m3u8 file failed:", err)
		}
	} else {
		data, err = DownloadM3u8(*URL)
		if err != nil {
			log.Fatalln("[-] Download m3u8 file failed:", err)
		}
	}

	mpl, err := ParseM3u8(data)
	if err != nil {
		log.Fatalln("[-] Parse m3u8 file failed:", err)
	} else {
		log.Println("[+] Parse m3u8 file succed")
	}

	if mpl.Count() > 0 {
		log.Println("[+] Total", mpl.Count(), "files to download")

		if *OutFile == "" {
			*OutFile = "total_" + filename(mpl.Segments[0].URI)
			directory = strings.Split(*OutFile, ".")[0]
		}

		if _, err := os.Stat(directory); os.IsNotExist(err) {
			if err := os.Mkdir(directory, 0755); err != nil {
				log.Fatal(err)
			}
		}

		start(mpl)

		log.Println("[+] Download succeed, saved to", directory, "cost:", time.Now().Sub(t))
		combinedFile()
	}
}
