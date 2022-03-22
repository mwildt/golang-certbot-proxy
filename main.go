package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	serveHttp()
}

const (
	WELL_KNOWN_SEGMMENT = ".well-known/"
)

func serveHttp() {
	webroot, ok := os.LookupEnv("WEBROOT")

	if !ok {
		webroot = "/opt/mwcertbot/"
	}

	domainName, ok := os.LookupEnv("DOMAIN_NAME")

	if !ok {
		domainName = "localhost"
	}

	serviceUrl, ok := os.LookupEnv("SERVICE_URL")

	if !ok {
		log.Print("SERVICE_URL nicht gegeben")
		return
	}

	wellKnown := filepath.Join(webroot, WELL_KNOWN_SEGMMENT)
	fmt.Printf("webroot: %s\n", webroot)
	fmt.Printf("wellKnown: %s\n", wellKnown)

	go func() {
		fmt.Printf("Start tls Termination process\n")

		certroot := filepath.Join(webroot, "certs")
		domainCertDir := filepath.Join(certroot, domainName)

		if _, err := os.Stat(domainCertDir); os.IsNotExist(err) {
			runcertbot(domainName, certroot, wellKnown)
		}

		keyFile := filepath.Join(domainCertDir, "privkey.pem")
		certFile := filepath.Join(domainCertDir, "cert.pem")

		fmt.Printf("keyFile: %s\n", keyFile)
		fmt.Printf("certFile: %s\n", certFile)

		fmt.Printf("Listening for incomming https-requests on port 443\n")

		url, err := url.Parse(serviceUrl)
		if err != nil {
			log.Print(err)
			log.Printf("fehler invalid service url %s\n", serviceUrl)
			return
		}

		proxy := httputil.NewSingleHostReverseProxy(url)

		httpsHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			proxy.ServeHTTP(w, r)
		})

		err = http.ListenAndServeTLS(":443", certFile, keyFile, httpsHandler)
		if err != nil {
			log.Print("fehler beim starten des tls server")
			log.Print(err)
		}

	}()

	redirectBase := fmt.Sprintf("https://%s:443", domainName)
	dir := http.Dir(webroot)
	fs := http.FileServer(dir)

	httpHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%s %s -> ", r.Method, r.URL.Path)

		if strings.HasPrefix(r.URL.Path, "/"+WELL_KNOWN_SEGMMENT) {
			fmt.Printf("file-system-handler: ")
			if r.Method != "GET" {
				fmt.Printf("method not allowed\n")
				w.WriteHeader(http.StatusMethodNotAllowed)
			} else {
				fmt.Printf("handle\n")
				fs.ServeHTTP(w, r)
			}
		} else {
			http.Redirect(w, r, redirectBase+r.RequestURI, http.StatusMovedPermanently)
		}
	})

	fmt.Printf("Listening for incomming http-requests on port 80\n")

	if err := http.ListenAndServe(":80", httpHandler); err != nil {
		log.Fatal(err)
	}
}

func runcertbot(domainName string, certroot string, wellKnown string) {

	probeFile := filepath.Join(wellKnown, "probe.txt")
	fmt.Printf("writeing probefile %s\n", probeFile)

	if err := os.WriteFile(probeFile, []byte("probe"), 0644); err != nil {
		log.Panic(err)
	} else {

		probeUrl := fmt.Sprintf("http://%s:80/%sprobe.txt", domainName, WELL_KNOWN_SEGMMENT)

		client := http.Client{
			Timeout: 5 * time.Second,
		}
		fmt.Printf("run http probe [url %s]\n", probeUrl)
		resp, err := client.Get(probeUrl)
		for err != nil {
			log.Printf("Probe Result Error: %s\n", err.Error())
			time.Sleep(5 * time.Second)
			fmt.Printf("run http probe [url %s]\n", probeUrl)
			resp, err = client.Get(probeUrl)
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Print("Fehler beim lesen des Response body")
			log.Print(err)
			return
		}
		log.Printf("response body: %s\n", body)

		output, err := run("certbot", "certonly", "--webroot", "-w", "/opt/mwcertbot/", "--agree-tos", "-d", domainName, "-v", "--email", "certbot@maltewildt.de")
		if err != nil {
			log.Print("Fehler beim erneuern der certificate mit cerbot\n")
			log.Print(err)
			log.Print(output)
			return
		}
		log.Print(output)

		output, err = run("mkdir", "-p", certroot)
		if err != nil {
			log.Print("Fehler beim erstellen des certroot-dir\n")
			log.Print(err)
			return
		}
		log.Print(output)

		output, err = run("cp", "-L", "-r", filepath.Join("/etc/letsencrypt/live/", domainName), certroot+"/")
		if err != nil {
			log.Print("Fehler beim Kopieren der Zertifikate\n")
			log.Print(err)
			return
		}
		log.Print(output)
	}
}

func run(name string, argv ...string) (string, error) {
	fmt.Printf("runn command %s with args %v", name, argv)
	cmd := exec.Command(name, argv...)
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()

	if err != nil {
		return "", nil
	}

	return out.String(), nil
}
