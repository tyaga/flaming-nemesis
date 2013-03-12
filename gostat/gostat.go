package main

import (
	"crypto/md5"
	"flag"
	"fmt"
	"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/native"
	"io"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
)

var (
	db       mysql.Conn
	projects ProjectColl
)

func main() {
	// options
	listen_port := flag.String("port", ":9090", "Port for listen to")
	db_dsn_flag := flag.String("db-dsn", "tcp://root:root@localhost:3306/statstat", "DB access DSN")

	flag.Parse()

	db_dsn, _ := url.Parse(*db_dsn_flag)
	pwd, _ := db_dsn.User.Password()

	// database
	db = mysql.New("tcp", "", db_dsn.Host, db_dsn.User.Username(), pwd, db_dsn.Path[1:])
	db.Connect()

	// retrieve all the projects from a'good old DB
	projects = *NewProjectColl()

	// background worker for analize and write stat
	chanPrepare := make(chan *Hash)
	chanWrite := make(chan *Hash)

	// preparator!
	go func() {
		for {
			q := <-chanPrepare
			log.Printf("[gostat] Analize: %s", *q)

			project, ok := getQueryProject(q)
			if !ok {
				continue
			}
			ok = project.checkSignature(q)
			if !ok {
				continue
			}
			ok = project.parseType(q)
			if !ok {
				continue
			}
			log.Printf("[gostat] %s", *q)

			// send to write channel
			chanWrite <- q
		}
	}()

	// writestignator!
	go func() {
		stmt, _ := db.Prepare("INSERT INTO stats (project_id, type_id, value, meta_type_id, meta_value) VALUES (?, ?, ?, ?, ?)")
		for {
			q := <-chanWrite
			log.Printf("[gostat] Write: %s", *q)
			// write down
			stmt.Run((*q)["project_id"], (*q)["type_id"], (*q)["value"], "0", "0")
		}
	}()

	// http
	http.HandleFunc("/stat", func(w http.ResponseWriter, r *http.Request) {
		// parse
		query := ParseQuery(&r.URL.RawQuery)

		// log
		log.Printf("[gostat] Got stat: %s", query)

		// send to prepare channel
		chanPrepare <- &query

		// response
		w.Header().Set("Content-type", "application/json")
		fmt.Fprint(w, `{"res": true}`)
	})
	http.ListenAndServe(*listen_port, nil)
}

// Storing and retrieving projects and types
type Type struct {
	id   int
	name string
}
type TypeColl map[int]Type
type Project struct {
	id     int
	name   string
	secret string
	types  TypeColl
}

func (self *Project) parseType(q *Hash) bool {
	// looking for key=>val in project.types
	for _, v := range self.types {
		if value, ok := (*q)[v.name]; ok {
			(*q)["type_id"] = strconv.Itoa(v.id)
			(*q)["value"] = value
			delete(*q, v.name)
			return true
		}
	}
	return false
}

func (self *Project) checkSignature(q *Hash) bool {
	defer printOnPanic()

	sig, ok := (*q)["sig"]
	if !ok {
		panic("No sig passed")
	}
	delete(*q, "sig")

	arr := []string{}
	for k, _ := range *q {
		arr = append(arr, k)
	}
	sort.Strings(arr)

	str := ""
	for _, v := range arr {
		str = str + v + "=" + (*q)[v]
	}
	str = str + self.secret

	expected := signMd5(str)
	if expected != sig {
		panic("Wrong signature, expected " + expected)
	}
	return true
}
func signMd5(str string) string {
	h := md5.New()
	io.WriteString(h, str)
	return fmt.Sprintf("%x", h.Sum(nil))
}

type ProjectColl map[int]Project

func NewProjectColl() *ProjectColl {
	projects := make(ProjectColl)

	rows, res, _ := db.Query("SELECT * FROM projects")
	for _, row := range rows {
		id, name, secret := row.Int(res.Map("id")), row.Str(res.Map("name")), row.Str(res.Map("secret"))
		projects[id] = Project{id, name, secret, make(TypeColl)}
	}
	rows, res, _ = db.Query("SELECT * FROM types")
	for _, row := range rows {
		id, name, project_id := row.Int(res.Map("id")), row.Str(res.Map("name")), row.Int(res.Map("project_id"))
		projects[project_id].types[id] = Type{id, name}
	}
	return &projects
}

type Hash map[string]string

func ParseQuery(rawQuery *string) Hash {
	queryMap, _ := url.ParseQuery(*rawQuery)
	vals := make(Hash)
	for k, v := range queryMap {
		vals[k] = v[0] // never ever assume having multiple stats
	}
	return vals
}

func getQueryProject(q *Hash) (Project, bool) {
	defer printOnPanic()

	project_idx, ok := (*q)["project_id"]
	if !ok {
		panic("No project_id passed")
	}
	project_id, err := strconv.Atoi(project_idx)
	if err != nil {
		panic("Wrong project_id")
	}
	project, ok := projects[project_id]
	if !ok {
		panic("Wrong project")
	}
	return project, true
}

func printOnPanic() {
	if r := recover(); r != nil {
		fmt.Println(r)
	}
}
