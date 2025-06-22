package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"text/template"

	captcha "github.com/s0nney/jerich0"
)

var (
	formTemplate = template.Must(template.New("example").Parse(formTemplateSrc))
	debugMode    = true // Set to false in production
	globalStore  = captcha.NewMemoryStore(captcha.CollectNum, captcha.Expiration)
)

func showFormHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	id := captcha.New()
	solution := globalStore.Get(id, false) // Get without clearing

	if debugMode {
		var solutionStr string
		for _, b := range solution {
			if b < 10 {
				solutionStr += string('0' + b)
			} else {
				solutionStr += string('A' + b - 10)
			}
		}
		log.Printf("Generated captcha: id=%s solution=%s", id, solutionStr)
	}

	d := struct {
		CaptchaId string
	}{
		id,
	}
	if err := formTemplate.Execute(w, &d); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func processFormHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	id := r.FormValue("captchaId")
	solution := r.FormValue("captchaSolution")

	// Get the actual solution for debugging
	actualSolution := globalStore.Get(id, false)
	var actualStr string
	if actualSolution != nil {
		for _, b := range actualSolution {
			if b < 10 {
				actualStr += string('0' + b)
			} else {
				actualStr += string('A' + b - 10)
			}
		}
	}

	log.Printf("Captcha verification:\n  ID: %s\n  Expected: %s\n  Received: %s",
		id, actualStr, solution)

	if !captcha.VerifyString(id, solution) {
		io.WriteString(w, "Wrong captcha solution! No robots allowed!\n")
		if debugMode {
			io.WriteString(w, fmt.Sprintf("Debug info: Expected %q, got %q\n", actualStr, solution))
		}
	} else {
		io.WriteString(w, "Great job, human! You solved the captcha.\n")
	}
	io.WriteString(w, "<br><a href='/'>Try another one</a>")
}

func main() {
	http.HandleFunc("/", showFormHandler)
	http.HandleFunc("/process", processFormHandler)
	http.Handle("/captcha/", captcha.Server(captcha.StdWidth, captcha.StdHeight))
	fmt.Println("Server is at localhost:8666")
	log.Println("Starting captcha server in debug mode...")
	if err := http.ListenAndServe("localhost:8666", nil); err != nil {
		log.Fatal(err)
	}
}

const formTemplateSrc = `<!doctype html>
<head><title>Captcha Example</title></head>
<body>
<script>
function setSrcQuery(e, q) {
	var src  = e.src;
	var p = src.indexOf('?');
	if (p >= 0) {
		src = src.substr(0, p);
	}
	e.src = src + "?" + q
}

function reload() {
	setSrcQuery(document.getElementById('image'), "reload=" + (new Date()).getTime());
	return false;
}
</script>
<form action="/process" method=post>
<p>Type the letters and numbers you see in the picture below:</p>
<p><img id=image src="/captcha/{{.CaptchaId}}.png" alt="Captcha image"></p>
<a href="#" onclick="reload()">Reload</a> 
<input type=hidden name=captchaId value="{{.CaptchaId}}"><br>
<input name=captchaSolution 
       title="Only letters and numbers are allowed"
       required autocomplete="off">
<input type=submit value=Submit>
</form>
`
