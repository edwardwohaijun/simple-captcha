package main

import (
	"github.com/edwardwohaijun/simple-captcha/pkg/captcha"
	"net/http"
	"log"
	"flag"
	"bytes"
	"strconv"
	"image/png"
	"fmt"
	"encoding/json"
)

func main() {
	var fontDir = flag.String("fontDir", "./fonts", "The font dir of true type fonts, default is ./fonts")
	var port = flag.String("port", "8080", "The port number server is listening, default is 8080")
	flag.Parse()

	var charSet = "ABCDEFHKLMNPQRTUVWXYabcdefhkmnpqrtuvwxy"; // omit: all numbers, Gg(9) Ii, Jj, Ll(i, 1), Oo(0), Ss(5), Zz(2), 0(o)
	err := captcha.Initialise(*fontDir, charSet)
	if err != nil {
		log.Fatal(err)
	}

	type jsonResponse struct {
		Text string `json:"text"`
		Base64 string `json:"base64"`
	}

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request){
		if req.URL.Path != "/" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		isBase64, _ := strconv.ParseBool(req.URL.Query().Get("base64"))
		if isBase64 {
			captchaText, _, captchaBase64, _ := captcha.New(true)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(jsonResponse{captchaText, captchaBase64})
			return
		}

		captchaText, captchaImg, _, _ := captcha.New(false)
		buffer := new(bytes.Buffer)
		if err := png.Encode(buffer, captchaImg); err != nil {
			log.Println("unable to encode image.")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Add("X-captcha-text", captchaText)
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))

		if _, err := w.Write(buffer.Bytes()); err != nil {
			log.Println("unable to write image.")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		fmt.Println("captchaText: ", captchaText)
	})

	log.Fatal(http.ListenAndServe(":" + *port, nil))
}
