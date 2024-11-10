package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"time"
)

func root_handler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://www.regexplanet.com/advanced/golang/index.html", http.StatusFound)
}

func write_with_callback(w http.ResponseWriter, callback string, v interface{}) {

	w.Header().Set("Content-Type", "text/plain; charset=utf8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET")
	w.Header().Set("Access-Control-Max-Age", "604800") // 1 week

	var b []byte
	var err error
	b, err = json.Marshal(v)
	if err != nil {
		b = []byte("{\"success\":false,\"html\":\"<p>json.Marshal failed</p>\"}")
	}

	if callback > "" {
		w.Write([]byte(callback))
		w.Write([]byte("("))
		w.Write(b)
		w.Write([]byte(");"))
	} else {
		w.Write(b)
	}
}

type Status struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
	Commit    string `json:"commit"`
	Lastmod   string `json:"lastmod"`
	Tech      string `json:"tech"`
	Version   string `json:"version"`
	Environ   []string
	Getwd     string
	Hostname  string
	Seconds   int64
	TempDir   string
}

func status_handler(w http.ResponseWriter, r *http.Request) {
	var err error
	status := Status{}

	status.Success = true
	status.Message = "OK"
	status.Timestamp = time.Now().UTC().Format(time.RFC3339)
	status.Commit = os.Getenv("COMMIT")
	status.Lastmod = os.Getenv("LASTMOD")
	status.Tech = runtime.Version()

	status.Getwd, err = os.Getwd()
	if err != nil {
		status.Getwd = "ERROR: " + err.Error()
	}

	status.Hostname, err = os.Hostname()
	if err != nil {
		status.Hostname = "ERROR: " + err.Error()
	}

	status.TempDir = os.TempDir()
	status.Version = runtime.Version()
	status.Seconds = time.Now().Unix()

	write_with_callback(w, r.FormValue("callback"), status)
}

type TestResult struct {
	Success bool   `json:"success"`
	Html    string `json:"html,omitempty"`
	Message string `json:"message,omitempty"`
}

func write_ints(buffer *bytes.Buffer, data [][]int) {

	if data == nil {
		buffer.WriteString("<i>nil</i>")
		return
	}

	for loop := 0; loop < len(data); loop++ {
		if loop > 0 {
			buffer.WriteString("<br/>")
		}
		buffer.WriteString("[")
		buffer.WriteString(html.EscapeString(fmt.Sprintf("%d", loop)))
		buffer.WriteString("]: ")

		for inner := 0; inner < len(data[loop]); inner++ {
			if inner > 0 {
				buffer.WriteString(", ")
			}
			buffer.WriteString(html.EscapeString(fmt.Sprintf("%d", data[loop][inner])))
		}
	}
}

func write_strings(buffer *bytes.Buffer, data []string) {

	if data == nil {
		buffer.WriteString("<i>nil</i>")
		return
	}

	buffer.WriteString("[")
	for loop := 0; loop < len(data); loop++ {
		if loop > 0 {
			buffer.WriteString(", ")
		}
		buffer.WriteString("<code>")
		buffer.WriteString(html.EscapeString(data[loop]))
		buffer.WriteString("</code>")
	}
	buffer.WriteString("]")
}

func test_handler(w http.ResponseWriter, r *http.Request) {

	var strRegex = r.FormValue("regex")
	var replacement = r.FormValue("replacement")
	var callback = r.FormValue("callback")
	var options = r.Form["option"]

	if strRegex == "" {
		write_with_callback(w, callback, TestResult{false, "", "No regex to test"})
		return
	}

	var buffer bytes.Buffer

	//buffer.WriteString("<div class=\"alert alert-warning\">Actually, it is a lot less than beta: the real code isn't even written yet</div>\n")

	buffer.WriteString("<table class=\"table table-bordered table-striped bordered-table zebra-striped\" style=\"width:auto;\">\n")
	buffer.WriteString("\t<tbody>\n")

	buffer.WriteString("\t\t<tr>\n")
	buffer.WriteString("\t\t\t<td>Regular Expression</td>\n")
	buffer.WriteString("\t\t\t<td><code>")
	buffer.WriteString(html.EscapeString(strRegex))
	buffer.WriteString("</code></td>\n")
	buffer.WriteString("\t\t</tr>\n")

	if replacement > "" {
		buffer.WriteString("\t\t<tr>\n")
		buffer.WriteString("\t\t\t<td>Replacement</td>\n")
		buffer.WriteString("\t\t\t<td><code>")
		buffer.WriteString(html.EscapeString(replacement))
		buffer.WriteString("</code></td>\n")
		buffer.WriteString("\t\t</tr>\n")
	}

	buffer.WriteString("\t\t<tr>\n")
	buffer.WriteString("\t\t\t<td>Escaped (<code>regexp.QuoteMeta(s)</code>)</td>\n")
	buffer.WriteString("\t\t\t<td><code>")
	buffer.WriteString(html.EscapeString(regexp.QuoteMeta(strRegex)))
	buffer.WriteString("</code></td>\n")
	buffer.WriteString("\t\t</tr>\n")

	ifPosix := false

	if len(options) > 0 {
		for loop := 0; loop < len(options); loop++ {
			if options[loop] == "posix" {
				ifPosix = true
			}
		}
	}

	buffer.WriteString("\t\t<tr>\n")
	buffer.WriteString("\t\t\t<td>Option</td>\n")
	buffer.WriteString("\t\t\t<td><code>")
	if ifPosix {
		buffer.WriteString(html.EscapeString("CompilePOSIX()"))
	} else {
		buffer.WriteString(html.EscapeString("Compile()"))
	}
	buffer.WriteString("</code></td>\n")
	buffer.WriteString("\t\t</tr>\n")

	re, err := regexp.Compile(strRegex)
	if err != nil {
		buffer.WriteString("\t\t<tr>\n")
		buffer.WriteString("\t\t\t<td>Error</td>\n")
		buffer.WriteString("\t\t\t<td><code>")
		buffer.WriteString(html.EscapeString(err.Error()))
		buffer.WriteString("</code></td>\n")
		buffer.WriteString("\t\t</tr>\n")
		buffer.WriteString("\t</tbody>\n")
		buffer.WriteString("</table>\n")
		write_with_callback(w, callback, TestResult{false, buffer.String(), "Error when compiling regex"})
		return
	}
	buffer.WriteString("\t\t<tr>\n")
	buffer.WriteString("\t\t\t<td># of groups (<code>re.NumSubexp()</code>)</td>\n")
	buffer.WriteString("\t\t\t<td>")
	buffer.WriteString(html.EscapeString(fmt.Sprintf("%d", re.NumSubexp())))
	buffer.WriteString("</td>\n")
	buffer.WriteString("\t\t</tr>\n")

	buffer.WriteString("\t</tbody>\n")
	buffer.WriteString("</table>\n")

	if r.FormValue("input") == "" {
		buffer.WriteString("<div class=\"alert alert-warning\">No inputs to test</div>")
		write_with_callback(w, callback, TestResult{true, buffer.String(), ""})
		return
	}

	buffer.WriteString("<table class=\"table table-bordered table-striped bordered-table zebra-striped\" style=\"width:auto;\">\n")
	buffer.WriteString("\t<thead>\n")
	buffer.WriteString("\t\t<tr>\n")
	buffer.WriteString("\t\t\t<th>Test</th>\n")
	buffer.WriteString("\t\t\t<th>Target String</th>\n")
	buffer.WriteString("\t\t\t<th>MatchString()</th>\n")
	if replacement > "" {
		buffer.WriteString("\t\t\t<th>ReplaceAllString()</th>\n")
	}
	buffer.WriteString("\t\t\t<th>FindAllString()</th>\n")
	buffer.WriteString("\t\t\t<th>FindAllStringIndex()</th>\n")
	buffer.WriteString("\t\t\t<th>FindAllStringSubmatch()</th>\n")
	buffer.WriteString("\t\t</tr>\n")
	buffer.WriteString("\t</thead>\n")

	buffer.WriteString("\t<tbody>\n")

	var inputs = r.Form["input"]

	for loop := 0; loop < len(inputs); loop++ {
		var input = inputs[loop]

		if len(input) == 0 {
			continue
		}

		buffer.WriteString("\t\t<tr>\n")

		buffer.WriteString("\t\t\t<td style=\"text-align:center\">")
		buffer.WriteString(html.EscapeString(fmt.Sprintf("%d", loop+1)))
		buffer.WriteString("</td>\n")

		buffer.WriteString("\t\t\t<td>")
		buffer.WriteString(html.EscapeString(input))
		buffer.WriteString("</td>\n")

		buffer.WriteString("\t\t\t<td>")
		buffer.WriteString(strconv.FormatBool(re.MatchString(input)))
		buffer.WriteString("</td>\n")

		if replacement > "" {
			buffer.WriteString("\t\t\t<td>")
			buffer.WriteString(html.EscapeString(re.ReplaceAllString(input, replacement)))
			buffer.WriteString("</td>\n")
		}

		buffer.WriteString("\t\t\t<td>")
		write_strings(&buffer, re.FindAllString(input, -1))
		buffer.WriteString("</td>\n")

		buffer.WriteString("\t\t\t<td>")
		write_ints(&buffer, re.FindAllStringIndex(input, -1))
		buffer.WriteString("</td>\n")

		buffer.WriteString("\t\t\t<td>")
		var data = re.FindAllStringSubmatch(input, -1)
		if data == nil {
			buffer.WriteString("<i>nil</i>")
		} else {
			for dataLoop := 0; dataLoop < len(data); dataLoop++ {
				buffer.WriteString("[")
				buffer.WriteString(html.EscapeString(fmt.Sprintf("%d", dataLoop)))
				buffer.WriteString("]: ")
				write_strings(&buffer, data[dataLoop])
				buffer.WriteString("<br/>")
			}
		}
		buffer.WriteString("</td>\n")
		buffer.WriteString("\t</tr>\n")
	}

	buffer.WriteString("\t</tbody>\n")
	buffer.WriteString("<table>\n")

	write_with_callback(w, callback, TestResult{true, buffer.String(), ""})
}

func main() {

	http.HandleFunc("/", root_handler)
	http.HandleFunc("/status.json", status_handler)
	http.HandleFunc("/test.json", test_handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "4000"
	}

	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
