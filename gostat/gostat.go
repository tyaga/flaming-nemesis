package main

import ( 
	"fmt"
	"flag"
	"log"
	"net/http"
	"net/url"
    "github.com/ziutek/mymysql/mysql"
    _ "github.com/ziutek/mymysql/native"
)
var (
	db mysql.Conn
	projects ProjectColl
)
func main() {
	// options
	listen_port := flag.String("port", ":9090", "Port for listen to")
	db_dsn_flag := flag.String("db-dsn", "tcp://root:root@localhost:3306/statstat", "DB access DSN")

	flag.Parse()
	
	db_dsn, _ := url.Parse(*db_dsn_flag)
	pwd, _ := db_dsn.User.Password();

	// database
	db = mysql.New("tcp", "", db_dsn.Host, db_dsn.User.Username(), pwd, db_dsn.Path[1:])
	db.Connect()
    
    // retrieve all the projects from a'good old DB
    projects = *NewProjectColl()

	// background worker for analize and write stat
	chanCheck := make(chan *ParsedQuery)
	go func () {
		for {
			q := <-chanCheck
			log.Printf("[gostat] Analize: %s", *q)
		}
	}()

	// http
    http.HandleFunc("/stat", func(w http.ResponseWriter, r *http.Request) {
		// parse
        query :=  ParseQuery(r.URL.RawQuery)
        
        // log
        log.Printf("[gostat] Got stat: %s", query)
        
        // check and write in background
        go func () {
			chanCheck<-&query
		}()

		// response
		w.Header().Set("Content-type", "application/json")
		fmt.Fprint(w, `{"res": true}`)
    })
    http.ListenAndServe(*listen_port, nil)
}

// Storing and retrieving projects and types
type Type struct {
	id int
	name string
}
type TypeColl map[int]Type
type Project struct {
	id int
	name string
	secret string
	types TypeColl
}
type ProjectColl map[int]Project
func NewProjectColl() *ProjectColl {
	projects := make(ProjectColl)

    rows, res, _ := db.Query("SELECT * FROM projects")
    for _, row := range rows {
		id, name, secret := row.Int(res.Map("id")), row.Str(res.Map("name")), row.Str(res.Map("secret"));
		projects[id] = Project{id, name, secret, make(TypeColl)}
	}
	rows, res, _ = db.Query("SELECT * FROM types")
	for _, row := range rows {
		id, name, project_id := row.Int(res.Map("id")), row.Str(res.Map("name")), row.Int(res.Map("project_id"));
		projects[project_id].types[id] = Type{id, name}
	}
	return &projects
}

// Parsing GET query
type ParsedQuery map[string]string
func ParseQuery(rawQuery string) ParsedQuery {
	queryMap, _ := url.ParseQuery(rawQuery)
	vals := make(ParsedQuery)
	for k, v := range queryMap {
		vals[k] = v[0]
	}
	return vals
}

// 
