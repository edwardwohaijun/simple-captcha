# Introduction
A simple captcha generation lib and server app written in Golang. I didn't do any fancy transformations or distortions to the generated text.
Just randomly choose different font and size for each character, then a rotation within ±20°.

![simple-captcha sample1](https://github.com/edwardwohaijun/simple-captcha/sample1.png)

![simple-captcha sample2](https://github.com/edwardwohaijun/simple-captcha/sample2.png)

# Prerequisites
Go (>= 1.6)

# Installation
* import as a lib into your project:
```
import  "github.com/edwardwohaijun/simple-captcha/pkg/captcha"
// ... inside your main function, add the following:
err := captcha.Initialise(fontDir, charSet)
if err != nil {
  log.Fatal(err)
}
// generate a captcha image in base64 format
captchaText, _,         captchaBase64,  err := captcha.New(true)
// generate a captcha image in *image.RGBA format
captchaText, captchaImg, _,             err := captcha.New(false)
```
For base64 format, you must compose the final string for web use, e.g.,

`imgTag := "<img src='data:image/png;base64,' + captchaBase64 + '\' />"`

For *image.RGBA format, you must generate the final png yourself,
and figure out a way to pass the captchaText to the consuming server.
The following example use custom HTTP header to pass captcha text in a HandleFunc:
```
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
```

* run as standalone app
```bash
go get github.com/edwardwohaijun/simple-captcha
cd $GOPATH/src/github.com/edwardwohaijun/simple-captcha/
go build cmd/captchad.go
./captchad
```

# Usage
Whether import as a lib or run as a standalone app, you must first call

```
err := captcha.Initialise(fontDir, charSet)
```
and check the returned error.

`charSet` is the set of characters from which 4~6 characters are picked randomly. The default is:
```
"ABCDEFHKLMNPQRTUVWXYabcdefhkmnpqrtuvwxy"
```


As a standalone app, `fontDir` is passed as a string flag, e.g.,
```
./captchad -fontDir=./fonts
```
This folder must contain all the true type fonts. By default, the folder is assumed to be in the current working directory.
If you are in `/A/B/C`, the captchad binary is in `/D`, when you run `/D/captchad`, the fondDir is assumed to be `/A/B/C/fonts`.

Another flag is server port when running as standalone app.
```
./captchad -port=8080
```
The default port is `8080`

# License

This project is licensed under the [MIT License](/LICENSE).