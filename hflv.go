package main

import (
	"fmt"
	"net"
	"net/http"
    "time"
    "io"
    "crypto/tls"
    "errors"
)

    const (
            AUDIO_TAG       = byte(0x08)
            VIDEO_TAG       = byte(0x09)
            SCRIPT_DATA_TAG = byte(0x12)
            DURATION_OFFSET = 53
            HEADER_LEN      = 13
            KEY_FRAME       = byte(0x17)
          )

type TagHeader struct {
    TagType   byte
    DataSize  uint32
    Timestamp uint32
}

func ReadTag(reader io.ReadCloser) (header *TagHeader, data []byte, err error) {
    tmpBuf := make([]byte, 4)
    header = &TagHeader{}
    // Read tag header
    if _, err = io.ReadFull(reader, tmpBuf[3:]); err != nil {
        return
    }
    header.TagType = tmpBuf[3]

    // Read tag size
    if _, err = io.ReadFull(reader, tmpBuf[1:]); err != nil {
        return
    }
    header.DataSize = uint32(tmpBuf[1])<<16 | uint32(tmpBuf[2])<<8 | uint32(tmpBuf[3])

    // Read timestamp
    if _, err = io.ReadFull(reader, tmpBuf); err != nil {
        return
    }
    header.Timestamp = uint32(tmpBuf[3])<<32 + uint32(tmpBuf[0])<<16 + uint32(tmpBuf[1])<<8 + uint32(tmpBuf[2])

    // Read stream ID
    if _, err = io.ReadFull(reader, tmpBuf[1:]); err != nil {
        return
    }

    // Read data
    data = make([]byte, header.DataSize)
    if _, err = io.ReadFull(reader, data); err != nil {
        return
    }

    // Read previous tag size
    if _, err = io.ReadFull(reader, tmpBuf); err != nil {
        return
    }

    return
}


func httpClient(reqUrl string, timeout int) (client *http.Client, err error) {

	client = &http.Client{
        CheckRedirect: func (req *http.Request, via []*http.Request) error {
            return errors.New("")
        },
		Transport: &http.Transport{
            TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
			Dial: (&net.Dialer{
                Timeout:   time.Duration(timeout) * time.Second,
            }).Dial,
			ResponseHeaderTimeout: time.Second * time.Duration(timeout),
            DisableKeepAlives: true,
		},
	}

	return
}


func main() {
	reqUrl := "http://httpflv.fastweb.com.cn.cloudcdn.net/live_fw/fastweb-stream-test.flv"
    
    client , err2:= httpClient(reqUrl, 10)
    if err2 != nil {
        fmt.Println("client not sure")
    }
	request, _ := http.NewRequest("GET", reqUrl, nil)
    request.Header.Add("User-Agent", "curl/7.19.7 (x86_64-redhat-linux-gnu) libcurl/7.19.7 NSS/3.13.1.0 zlib/1.2.3 libidn/1.18 libssh2/1.2.2")
    request.Header.Add("Accept" , "*/*")
	
    response, err := client.Do(request)
    if response != nil {
        fmt.Println(response.StatusCode, err)
    } else {
        return
    }
    defer response.Body.Close()

    flvHeader := make([]byte, HEADER_LEN)
    if _, err := io.ReadFull(response.Body, flvHeader); err != nil {
        return
    }
    if flvHeader[0] != 'F' || flvHeader[1] != 'L' || flvHeader[2] != 'V' {
        return
    }
    fmt.Println(string(flvHeader))
    
    for {
        header, data, err := ReadTag(response.Body)
        if err != nil {
            fmt.Println("ERRROR TAG\n")
            return
        }

        fmt.Println(header.TagType, data[0])
        if header.TagType == VIDEO_TAG && data[0] == KEY_FRAME {
            fmt.Println("FOUND KEY FRAME\n")
            break
        }
    }
}
