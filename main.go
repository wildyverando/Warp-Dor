/*

	This code used to auto referrer for cloudflare warp
	
	Author: Wildy Sheverando <wildy@wildyverando.com>

	This Project Licensed under The MIT License.

	Repo: https://github.com/wildyverando/Warp-Dor

*/

package main

import (
	"fmt"
	// "bytes"
	// "compress/gzip"
	// "io/ioutil"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"os/exec"
	"runtime"
	"math/rand"
	"net"
	"bufio"
	"strings"
	"sync"
	"time"
	"errors"
	"github.com/valyala/fasthttp"
)

var proxies []string

type PostData struct {
	Key       string `json:"key"`
	InstallID string `json:"install_id"`
	FcmToken  string `json:"fcm_token"`
	Referrer  string `json:"referrer"`
	TOS       string `json:"tos"`
	Type      string `json:"type"`
	Locale    string `json:"locale"`
}

func loadProxiesFromFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		proxy := strings.TrimSpace(scanner.Text())
		if proxy != "" {
			proxies = append(proxies, proxy)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if len(proxies) == 0 {
		return errors.New("no proxies found in the file")
	}

	return nil
}

func clearScreen() {
    var cmd *exec.Cmd
    if runtime.GOOS == "windows" {
        cmd = exec.Command("cmd", "/c", "cls")
    } else {
        cmd = exec.Command("clear")
    }
    cmd.Stdout = os.Stdout
    cmd.Run()
}

func getRandProxy() string {
	rand.Seed(time.Now().UnixNano())
	index := rand.Intn(len(proxies))
	return proxies[index]
}

func genRandString(length int) string {
	charSet := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var result strings.Builder
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < length; i++ {
		randomIndex := rand.Intn(len(charSet))
		result.WriteString(string(charSet[randomIndex]))
	}
	return result.String()
}

func FasthttpHTTPDialer(proxyAddr string) fasthttp.DialFunc {
	return func(addr string) (net.Conn, error) {
		conn, err := fasthttp.Dial(proxyAddr)
		if err != nil {
			return nil, err
		}

		req := "CONNECT " + addr + " HTTP/1.1\r\n\r\n"  
		if _, err := conn.Write([]byte(req)); err != nil {
			return nil, err
		}

		res := fasthttp.AcquireResponse()
		defer fasthttp.ReleaseResponse(res)
		res.SkipBody = true

		if err := res.Read(bufio.NewReader(conn)); err != nil {
			conn.Close()
			return nil, err
		}

		if res.Header.StatusCode() != 200 {
			conn.Close()
			return nil, fmt.Errorf("could not connect to proxy")
		}
		return conn, nil
	}
}

func sendRequest(warpdeviceid string, proxy string) {
	url2requests := fmt.Sprintf("https://api.cloudflareclient.com/v0a%d/reg", rand.Intn(900)+100)
	getInstallID := genRandString(22)
	getRandKey := genRandString(43) + "="
	RandFcmToken := fmt.Sprintf("%s:APA91b%s", getInstallID, genRandString(134))
	TOSTimed := time.Now().UTC().Format("2006-01-02T15:04:05.000Z07:00")
	postData := PostData{
		Key:       getRandKey,
		InstallID: getInstallID,
		FcmToken:  RandFcmToken,
		Referrer:  warpdeviceid,
		TOS:       TOSTimed,
		Type:      "Android",
		Locale:    "en_US",
	}

	jsonPayload, err := json.Marshal(postData)
	if err != nil {
		log.Println("Error creating JSON payload:", err)
		return
	}

	req := fasthttp.AcquireRequest()
	req.SetRequestURI(url2requests)
	req.Header.SetMethod("POST")
	req.Header.SetContentType("application/json")
	req.Header.Set("Host", "api.cloudflareclient.com")
	req.Header.Set("Connection", "Keep-Alive")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("User-Agent", "okhttp/3.12.1")
	req.SetBody(jsonPayload)

	client := &fasthttp.Client{
		Dial: FasthttpHTTPDialer(proxy),
		MaxConnsPerHost:     100,
		MaxIdleConnDuration: 5 * time.Second,
	}

	resp := fasthttp.AcquireResponse()

	err = client.Do(req, resp)

	if resp.StatusCode() != 200 || err != nil {
		log.Println("| FAILURE | Proxy -", proxy) //, "| Error:", err)
		fasthttp.ReleaseResponse(resp)
		fasthttp.ReleaseRequest(req)
		return
	}

	log.Println("| SUCCESS | Sent 1GB to", warpdeviceid, "| Proxy -", proxy)
	fasthttp.ReleaseResponse(resp)
	fasthttp.ReleaseRequest(req)

	// var body []byte
	// contentEncoding := string(resp.Header.Peek("Content-Encoding"))
	
	// if strings.Contains(contentEncoding, "gzip") {
	// 	reader, err := gzip.NewReader(bytes.NewReader(resp.Body()))
	// 	if err != nil {
	// 		log.Println("Error decompressing response:", err)
	// 		fasthttp.ReleaseResponse(resp)
	// 		fasthttp.ReleaseRequest(req)
	// 		return
	// 	}
	// 	defer reader.Close()
	
	// 	body, err = ioutil.ReadAll(reader)
	// 	if err != nil {
	// 		log.Println("Error reading response body:", err)
	// 		fasthttp.ReleaseResponse(resp)
	// 		fasthttp.ReleaseRequest(req)
	// 		return
	// 	}
	// } else {
	// 	body = resp.Body()
	// }

	// log.Println("Response body:", string(body))

	return
}

func main() {
    clearScreen()
	fmt.Println("  _    _                 ______")
	fmt.Println(" | |  | |                |  _  \\")
	fmt.Println(" | |  | | __ _ _ __ _ __ | | | |___  _ __")
	fmt.Println(" | |/\\| |/ _` | '__| '_ \\| | | / _ \\| '__|")
	fmt.Println(" \\  /\\  / (_| | |  | |_) | |/ / (_) | |")
	fmt.Println("  \\/  \\/ \\__,_|_|  | .__/|___/ \\___/|_|")
	fmt.Println("                   | |")
	fmt.Println("                   |_|")
	fmt.Println("")
	fmt.Println("                               Dorrrrr !")
	fmt.Println("")
	fmt.Println("Tools used to inject unlimited warp quota for free")
	fmt.Println("")
	fmt.Println("---------------------------------------------------")
	fmt.Println(" Program  : WarpDor | Version: 1.0.0")
	fmt.Println(" Author   : Wildy Sheverando")
	fmt.Println(" Email    : hai@wildy.id")
	fmt.Println(" License  : MIT License")
	fmt.Println("---------------------------------------------------")
	fmt.Println("")

	var proxyPath string
	fmt.Print("Proxy Path     : ")
	fmt.Scanln(&proxyPath)

	if proxyPath == "" {
		fmt.Println("Input proxy files path")
		return
	}

	var warpid string
	fmt.Print("Warp Device ID : ")
	fmt.Scanln(&warpid)

	if warpid == "" {
		fmt.Println("Input your warp device id")
		return
	}

	err := loadProxiesFromFile(proxyPath)
	if err != nil {
		fmt.Println("Cannot load proxy files path")
		return
	}

	var thread string
	fmt.Print("Thread         : ")
	fmt.Scanln(&thread)

	if thread == "" {
		fmt.Println("Input thread [ default : 5 ]")
		return
	}

	threads, err := strconv.Atoi(thread)
	if err != nil {
		fmt.Println("thread must integers")
		return
	}

	fmt.Println("\n\nStarting warp quota injection..")
	var wg sync.WaitGroup
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				proxy := getRandProxy()
				sendRequest(warpid, proxy)
				time.Sleep(1 * time.Second)
			}
		}()
	}
	wg.Wait()

}
