package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/rwirdemann/databasedragon/cmd"
	"github.com/rwirdemann/databasedragon/httpx/api"
	"html/template"
	"io"
	"log"
	"net/http"
	"time"
)

//go:embed templates
var templates embed.FS

var client = &http.Client{Timeout: 10 * time.Second}
var msgSuccess, msgError string

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/", indexHandler)
	router.HandleFunc("/new", newHandler)
	router.HandleFunc("/delete", deleteHandler)

	// start recording
	router.HandleFunc("/record", startRecording)

	// stop recording
	router.HandleFunc("/stoprecording", stopRecording)

	router.HandleFunc("/create", createHandler)
	router.HandleFunc("/run", startHandler)
	router.HandleFunc("/stop", stopHandler)
	router.HandleFunc("/show", showHandler)
	log.Println("Listening on :8081...")
	_ = http.ListenAndServe(":8081", router)
}

type ViewData struct {
	Title   string
	Message string
	Error   string
}

// render renders tmpl embedded in layout.html using the provided data.
func render(tmpl string, w http.ResponseWriter, data any) {
	if err := renderE(tmpl, w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// renderE works the same as render except returning the error instead of
// handling it.
func renderE(tmpl string, w http.ResponseWriter, data any) error {
	t, err := template.ParseFS(templates, "templates/layout.html", "templates/messages.html", fmt.Sprintf("templates/%s", tmpl))
	if err != nil {
		return err
	}
	return t.Execute(w, data)
}

// renderS renders tmpl embedded in layout.html and inserts title.
func renderS(tmpl string, w http.ResponseWriter, title string) {
	if err := renderE(tmpl, w, ViewData{
		Title:   title,
		Message: "",
		Error:   "",
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// redirectE redirects to url after setting the global msgError to err.
func redirectE(w http.ResponseWriter, r *http.Request, url string, err error) {
	msgError = err.Error()
	http.Redirect(w, r, url, http.StatusSeeOther)
}

func indexHandler(w http.ResponseWriter, _ *http.Request) {
	allTests := struct {
		Tests []api.Test `json:"tests"`
	}{}
	if r, err := client.Get("http://localhost:3000/tests"); err != nil {
		msgError = err.Error()
	} else {
		json.NewDecoder(r.Body).Decode(&allTests)
	}

	m, e := clearMessages()
	render("index.html", w, struct {
		ViewData
		Tests []api.Test
	}{ViewData: ViewData{
		Title:   "DataFrog Home",
		Message: m,
		Error:   e,
	}, Tests: allTests.Tests})
}

func showHandler(w http.ResponseWriter, request *http.Request) {
	if err := request.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	m, e := clearMessages()
	render("show.html", w, struct {
		ViewData
		Testname string
	}{ViewData: ViewData{
		Title:   "Show",
		Message: m,
		Error:   e,
	}, Testname: request.FormValue("testname")})
}

func newHandler(w http.ResponseWriter, _ *http.Request) {
	renderS("new.html", w, "New")
}

func startRecording(w http.ResponseWriter, request *http.Request) {
	if err := request.ParseForm(); err != nil {
		redirectE(w, request, "/", err)
		return
	}
	testname := request.FormValue("testname")
	r, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:3000/tests/%s/recordings", testname), nil)
	if err != nil {
		redirectE(w, request, "/", err)
		return
	}
	_, err = client.Do(r)
	if err != nil {
		redirectE(w, request, "/", err)
		return
	}
	msgSuccess = "Recording has been started. Run UI interactions and click 'Stop recording...' when finished"
	m, e := clearMessages()
	render("record.html", w, struct {
		ViewData
		Testname string
	}{ViewData: ViewData{
		Title:   "Record",
		Message: m,
		Error:   e,
	}, Testname: testname})
}

func deleteHandler(w http.ResponseWriter, request *http.Request) {
	testname := request.URL.Query().Get("testname")
	url := fmt.Sprintf("http://localhost:3000/tests/%s", testname)
	r, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		msgError = err.Error()
		http.Redirect(w, request, "/", http.StatusSeeOther)
		return
	}
	_, err = client.Do(r)
	if err != nil {
		msgError = err.Error()
		http.Redirect(w, request, "/", http.StatusSeeOther)
		return
	}
	msgSuccess = fmt.Sprintf("Test '%s' successfully deleted", testname)
	http.Redirect(w, request, fmt.Sprintf("/"), http.StatusSeeOther)
}

func clearMessages() (string, string) {
	e := msgError
	m := msgSuccess
	msgError = ""
	msgSuccess = ""
	return m, e
}

func startHandler(w http.ResponseWriter, request *http.Request) {
	testname := request.URL.Query().Get("testname")
	url := fmt.Sprintf("http://localhost:3000/tests/%s/runs", testname)
	r, err := http.NewRequest(http.MethodPut, url, nil)
	if err != nil {
		msgError = err.Error()
		http.Redirect(w, request, "/", http.StatusSeeOther)
		return
	}
	response, err := client.Do(r)
	if err != nil {
		msgError = err.Error()
		http.Redirect(w, request, "/", http.StatusSeeOther)
		return
	}
	statusOK := response.StatusCode >= 200 && response.StatusCode < 300
	if !statusOK {
		body, _ := io.ReadAll(response.Body)
		msgError = fmt.Sprintf("HTTP Status: %d => %s", response.StatusCode, body)
	} else {
		msgSuccess = fmt.Sprintf("Test '%s' has been started. Run test script and click 'Stop...' when you are done!", testname)
	}
	http.Redirect(w, request, fmt.Sprintf("/show?testname=%s", testname), http.StatusSeeOther)
}

func stopHandler(w http.ResponseWriter, request *http.Request) {
	testname := request.URL.Query().Get("testname")
	url := fmt.Sprintf("http://localhost:3000/tests/%s/runs", testname)
	r, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		redirectE(w, request, "/", err)
		return
	}
	response, err := client.Do(r)
	if err != nil {
		redirectE(w, request, "/", err)
		return
	}

	statusOK := response.StatusCode >= 200 && response.StatusCode < 300
	if !statusOK {
		body, _ := io.ReadAll(response.Body)
		msgError = fmt.Sprintf("HTTP Status: %d => %s", response.StatusCode, body)
		http.Redirect(w, request, "/", http.StatusSeeOther)
		return
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		redirectE(w, request, "/", err)
		return
	}
	var report cmd.Report
	err = json.Unmarshal(body, &report)
	if err != nil {
		redirectE(w, request, "/", err)
		return
	}

	m, e := clearMessages()
	render("result.html", w, struct {
		ViewData
		Report cmd.Report
	}{ViewData: ViewData{
		Title:   "Result",
		Message: m,
		Error:   e,
	}, Report: report})
}

func createHandler(w http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	testname := request.FormValue("testname")
	http.Redirect(w, request, fmt.Sprintf("/record&testname=%s", testname), http.StatusSeeOther)
}

func stopRecording(w http.ResponseWriter, request *http.Request) {
	testname := request.URL.Query().Get("testname")
	url := fmt.Sprintf("http://localhost:3000/tests/%s/recordings", testname)
	r, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		msgError = err.Error()
		http.Redirect(w, request, "/", http.StatusSeeOther)
		return
	}
	_, err = client.Do(r)
	if err != nil {
		msgError = err.Error()
		http.Redirect(w, request, "/", http.StatusSeeOther)
		return
	}

	http.Redirect(w, request, "/", http.StatusSeeOther)
}
