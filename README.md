# Introduction
A simple captcha image generation package and server app written in Golang. There aren't any fancy transformations or distortions to the generated text.
Just randomly choose a true type font and size for each character, followed by a rotation within ±20°.
The following are 2 generated images.

![simple-captcha sample1](https://github.com/edwardwohaijun/simple-captcha/blob/master/sample1.png)

![simple-captcha sample2](https://github.com/edwardwohaijun/simple-captcha/blob/master/sample2.png)

# Prerequisites
* Go (>= 1.6)
* a font directory containing true type fonts
* supervisord if you want to run the server as daemon

# Installation
`go get github.com/edwardwohaijun/simple-captcha`
## import as a package into your project
```
import "github.com/edwardwohaijun/simple-captcha/pkg/captcha"
```

## run as standalone server
```bash
cd $GOPATH/src/github.com/edwardwohaijun/simple-captcha/
go build cmd/captchad.go
./captchad -fontDir=./fonts -port=8080
```
If you want to run the server as a daemon and auto-start it on system boot, install `supervisord` first,
then copy init/captchad.conf to /etc/supervisord.conf.d/captchad.conf. You need to make some changes in this file to suit your environment.

# Usage
## import as a package
Inside your main function, add the following:
```
err := captcha.Initialise(fontDir, charSet)
if err != nil {
  log.Fatal(err)
}
// generate a captcha image in base64 format
captchaText, _,         captchaBase64,  err := captcha.New(true)
// generate a captcha image in *image.RGBA format
captchaText, captchaImg, _,             err := captcha.New(false)
```
You must call `Initialise(fontDir, charSet)` first, and check the return error. `fontDir` is a string specifying the directory of true type fonts.
`charSet` is a string containing all characters from which 4~6 characters are randomly chosen from. The default is:
```
"ABCDEFHKLMNPQRTUVWXYabcdefhkmnpqrtuvwxy"
```

Then call `captchaText, captchaImage, captchaBase64, err := captcha.New(true)`,
the bool argument specifies whether you want a base64 string or native *image.RGBA format.

For base64 format, you must compose the final string for web use, e.g.,

`imgTag := "<img src='data:image/png;base64,' + captchaBase64 + '\' />"`

For *image.RGBA format, you must generate the final png yourself,
and figure out a way to pass the captchaText to the consuming server.
The following snippet use custom HTTP header to pass captcha text in a HandleFunc:
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

## run as a standalone app
`fontDir` is passed as a directory containing the true type fonts, e.g.,
```
./captchad -fontDir=./fonts -port=8080
```
By default, this folder is assumed to be in the current working directory.
If you are in `/A/B/C`, the captchad binary is in `/D`, when you run `/D/captchad`, the fondDir is assumed to be `/A/B/C/fonts`.

By default, `http://localhost:8080` return the png file,
and captcha text is returned as HTTP custom header: `X-captcha-text`. To return base64 string,
send the request as: `http://localhost:8080/?base64=true`, the return value is a JSON object:
```
{
  "text":"abcd",
  "base64":"iVBORw0KGgoAAAANSUhEUgAAAIwAAAAoCAYA..."
}
```

# License

This project is licensed under the [MIT License](/LICENSE).