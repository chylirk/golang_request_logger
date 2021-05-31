package main

import (
	"fmt"
	"io/ioutil"
	"html"
	"log"
	"net/http"
	"database/sql"
	_ "github.com/lib/pq"
	"time"
)

const (
	DB_USER 	= "lee"
	DB_PASSWORD 	= "password"
	DB_NAME 	= "test"
)

func sayHi(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fmt.Fprintf(w, "Latest Request Method, %q\n", html.EscapeString(r.Method))
	fmt.Fprintf(w, "Latest Request Host, %q\n", html.EscapeString(r.Host))
	fmt.Fprintf(w, "Latest Request Path, %q\n", html.EscapeString(r.URL.Path))
	fmt.Fprintf(w, "Latest Request URI, %q\n", html.EscapeString(r.RequestURI))
	fmt.Fprintf(w, "Latest Request Protocol, %q\n", html.EscapeString(r.Proto))
	fmt.Fprintf(w, "Latest Request Address, %q\n", html.EscapeString(r.RemoteAddr))
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(w, "Latest Request Body: %q\n", buf[:])
	for k, v := range r.Header {
		fmt.Fprintf(w, "Latest Request Header, %v: %v\n", k, v)
	}
}

func testSql(w http.ResponseWriter, _ *http.Request) {
        dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
            DB_USER, DB_PASSWORD, DB_NAME)
        db, err := sql.Open("postgres", dbinfo)
        checkErr(err)
        defer db.Close()

        fmt.Println("# Inserting values")

        var lastInsertId int
        err = db.QueryRow("INSERT INTO userinfo(username,departname,created) VALUES($1,$2,$3) returning uid;", "astaxie", "研发部门", "2012-12-09").Scan(&lastInsertId)
        checkErr(err)
        fmt.Println("last inserted id =", lastInsertId)

        fmt.Println("# Updating")
        stmt, err := db.Prepare("update userinfo set username=$1 where uid=$2")
        checkErr(err)

        res, err := stmt.Exec("astaxieupdate", lastInsertId)
        checkErr(err)

        affect, err := res.RowsAffected()
        checkErr(err)

        fmt.Println(affect, "rows changed")

        fmt.Println("# Querying")
        rows, err := db.Query("SELECT * FROM userinfo")
        checkErr(err)

        for rows.Next() {
            var uid int
            var username string
            var department string
            var created time.Time
            err = rows.Scan(&uid, &username, &department, &created)
            checkErr(err)
            fmt.Println("uid | username | department | created ")
            fmt.Fprintf(w, "%3v | %8v | %6v | %6v\n", uid, username, department, created)
        }

        fmt.Println("# Deleting")
        stmt, err = db.Prepare("delete from userinfo where uid=$1")
        checkErr(err)

        res, err = stmt.Exec(lastInsertId)
        checkErr(err)

        affect, err = res.RowsAffected()
        checkErr(err)

        fmt.Println(affect, "rows changed")
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
        dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
            DB_USER, DB_PASSWORD, "requests")
        db, err := sql.Open("postgres", dbinfo)
        checkErr(err)
        defer db.Close()

        fmt.Println("# processing request")
	r.ParseForm()
	method := html.EscapeString(r.Method)
	path := html.EscapeString(r.URL.Path)
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
	log.Fatal(err)
	}
	body := fmt.Sprintf("%q", buf[:])
	fmt.Println(body)

        var lastInsertId int
        err = db.QueryRow("INSERT INTO requests(method,path,body) VALUES($1,$2,$3) returning request_id;", method, path, body).Scan(&lastInsertId)
        checkErr(err)

	rows, err := db.Query("SELECT * FROM requests")

	fmt.Fprintf(w, "id  |  method  |    path   |    body\n")
        for rows.Next() {
            var id int
            var method string
            var path string
	    var body string
            err = rows.Scan(&id, &method, &path, &body)
            checkErr(err)
            fmt.Fprintf(w, "%3v | %8v | %6v | %8v\n", id, method, path, body)
        }
}

func getSeinfeld(w http.ResponseWriter, _ *http.Request) {
        dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
            DB_USER, DB_PASSWORD, "api")
        db, err := sql.Open("postgres", dbinfo)
        checkErr(err)
        defer db.Close()

        fmt.Println("# Querying")
        rows, err := db.Query("SELECT * FROM users")
        checkErr(err)

        for rows.Next() {
            var id int
            var name string
            var email string
            err = rows.Scan(&id, &name, &email)
            checkErr(err)
            fmt.Fprintf(w, "id  |   name   | email\n")
            fmt.Fprintf(w, "%3v | %8v | %v\n", id, name, email)
        }
}

func checkErr(err error) {
    if err != nil {
        panic(err)
    }
}

func main() {

	http.HandleFunc("/", sayHi)
	http.HandleFunc("/testSql", testSql)
	http.HandleFunc("/seinfeld", getSeinfeld)
	http.HandleFunc("/requests", handleRequest)
	log.Fatal(http.ListenAndServe(":3001", nil))
}
