package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"text/template"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/gorilla/pat"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

type Post struct {
	Id                      uint16
	Title, Anons, Full_Text string
}

var posts = []Post{}
var showItems = Post{}

func mainPage(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("templates/mainPage.html", "templates/header.html", "templates/footer.html")

	t.ExecuteTemplate(w, "mainPage", nil)
}

func create(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("templates/create.html", "templates/header.html", "templates/footer.html")

	t.ExecuteTemplate(w, "create", nil)
}

func examples(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("templates/examples.html", "templates/header.html", "templates/footer.html")

	t.ExecuteTemplate(w, "examples", nil)
}

func googleSignIn(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("templates/googleSignIn.html", "templates/header.html", "templates/footer.html")

	t.ExecuteTemplate(w, "googleSignIn", nil)
}

func save_article(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	anons := r.FormValue("anons")
	full_text := r.FormValue("full_text")

	if title == "" || anons == "" || full_text == "" {
		fmt.Fprintf(w, "No")
	} else {
		db, err := sql.Open("mysql", "root@tcp(localhost:3306)/test-project")

		if err != nil {
			panic(err)
		}

		defer db.Close()

		inst, err := db.Query(fmt.Sprintf("INSERT INTO `articles` (`title`, `anons`, `full_text`) VALUES('%s', '%s', '%s')", title, anons, full_text))
		if err != nil {
			panic(err)
		}
		defer inst.Close()

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func post(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("templates/post.html", "templates/show.html", "templates/header.html", "templates/footer.html")

	db, err := sql.Open("mysql", "root@tcp(localhost:3306)/test-project")

	if err != nil {
		panic(err)
	}

	defer db.Close()

	res, err := db.Query("SELECT * FROM `articles`")
	if err != nil {
		panic(err)
	}

	posts = []Post{}
	for res.Next() {
		var post Post
		err = res.Scan(&post.Id, &post.Title, &post.Anons, &post.Full_Text)
		if err != nil {
			panic(err)
		}
		posts = append(posts, post)
	}
	t.ExecuteTemplate(w, "post", posts)
}

func showPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	t, _ := template.ParseFiles("templates/show.html", "templates/header.html", "templates/footer.html")

	db, err := sql.Open("mysql", "root@tcp(localhost:3306)/test-project")
	if err != nil {
		panic(err)
	}

	defer db.Close()

	res, err := db.Query(fmt.Sprintf("SELECT * FROM `articles` WHERE `id` = '%s'", vars["id"]))
	if err != nil {
		panic(err)
	}

	showItems = Post{}
	for res.Next() {
		var post Post
		err = res.Scan(&post.Id, &post.Title, &post.Anons, &post.Full_Text)
		if err != nil {
			panic(err)
		}
		showItems = post
	}
	t.ExecuteTemplate(w, "show", showItems)
}

func Google() {
	key := "Secret-session-key" // Replace with your SESSION_SECRET or similar
	maxAge := 86400 * 30        // 30 days
	isProd := false             // Set to true when serving over https

	store := sessions.NewCookieStore([]byte(key))
	store.MaxAge(maxAge)
	store.Options.Path = "/"
	store.Options.HttpOnly = true // HttpOnly should always be enabled
	store.Options.Secure = isProd

	gothic.Store = store

	goth.UseProviders(
		google.New("462871246413-a5a6oe2879hj133ovvqt50u09qlg6uag.apps.googleusercontent.com", "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", "http://localhost:3000/auth/google/callback", "email", "profile"),
	)

	p := pat.New()
	p.Get("/auth/{provider}/callback", func(res http.ResponseWriter, req *http.Request) {

		user, err := gothic.CompleteUserAuth(res, req)
		if err != nil {
			fmt.Fprintln(res, err)
			return
		}
		t, _ := template.ParseFiles("templates/complete.html")
		t.Execute(res, user)
	})

	p.Get("/auth/{provider}", func(res http.ResponseWriter, req *http.Request) {
		gothic.BeginAuthHandler(res, req)
	})

	p.Get("/", func(res http.ResponseWriter, req *http.Request) {
		t, _ := template.ParseFiles("templates/mainPage2.html")
		t.Execute(res, false)
	})
}

func HandleFunc() {
	router := mux.NewRouter()
	router.HandleFunc("/", mainPage).Methods("GET")
	router.HandleFunc("/create", create).Methods("GET")
	router.HandleFunc("/googleSignIn", googleSignIn).Methods("GET")
	router.HandleFunc("/examples", examples).Methods("GET")
	router.HandleFunc("/post", post).Methods("GET")
	router.HandleFunc("/show/{id:[0-9]+}", showPost).Methods("GET")
	router.HandleFunc("/save_article", save_article).Methods("POST")

	http.Handle("/", router)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	http.ListenAndServe(":8888", nil)
	log.Println("listening on localhost:8888")
}

func main() {
	Google()
	HandleFunc()
}
