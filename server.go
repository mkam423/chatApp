package main

import (
	//"fmt"
	"io/ioutil"
	"net/http"
	"html/template"
	"regexp"
	"fmt"
)

var validPath = regexp.MustCompile("^/(edit|save|view|refresh)/([a-zA-Z0-9]+)$")
var templates = template.Must(template.ParseFiles("edit.html", "view.html"))

type Page struct {
	Title string
	Body  []byte
}

//Write to file
func (p *Page) save() error {
	filename := p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

//Used to bring up file contents
func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

//Outline of view page
func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	//Load given url file
	p, err := loadPage(title)
  if err != nil {
      http.Redirect(w, r, "/edit/"+title, http.StatusFound)
      return
  }
  renderTemplate(w, "view", p)
}

//Outline of edit page
func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	//Load given url file
  p, err := loadPage(title)
  if err != nil {
      p = &Page{Title: title}
  }
   renderTemplate(w, "edit", p)
}

//Save current content (POST)
func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	//Grab body
	body, err1 := ioutil.ReadAll(r.Body)
	if err1 != nil {
  	fmt.Printf("FATAL IO reader issue %s ", err1.Error())
	}

	//Create page to save body
  p := &Page{Title: title, Body: []byte(body)}
  err2 := p.save()
  if err2 != nil {
      http.Error(w, err2.Error(), http.StatusInternalServerError)
      return
  }
}

//Refresh current content (GET)
func refreshHandler(w http.ResponseWriter, r *http.Request, title string) {
		//fmt.Printf("In refresh\n")
		p, err := loadPage(title)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(p.Body)
}

//Template handler to redirect to methods
func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    m := validPath.FindStringSubmatch(r.URL.Path)
    if m == nil {
      http.NotFound(w, r)
      return
    }
    fn(w, r, m[2])
  }
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
  err := templates.ExecuteTemplate(w, tmpl+".html", p)
  if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
  }
}


func main() {
	port:="8080"

  http.HandleFunc("/view/", makeHandler(viewHandler))
  http.HandleFunc("/edit/", makeHandler(editHandler))
  http.HandleFunc("/save/", makeHandler(saveHandler))
	http.HandleFunc("/refresh/", makeHandler(refreshHandler))

	fmt.Printf("Serving on port: %s\n", port)
  http.ListenAndServe(":" + port, nil)
}
