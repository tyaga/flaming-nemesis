package main

import (
	"flag"
	"fmt"
	"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/native"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"runtime"
	"time"

	"./tools"
)
var (
	db       mysql.Conn
	projects ProjectColl
	verbose  *bool
)

func LOG(str ...interface{}) {
	if !*verbose {
		return;
	}
	log.Printf("[gostat] %s", str)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	
	// options
	verbose = flag.Bool("verbose", true, "")
	listen_port := flag.String("port", ":9090", "Port for listen to")
	db_dsn_flag := flag.String("db-dsn", "tcp://root:root@localhost:3306/statstat", "DB access DSN")

	flag.Parse()

	db_dsn, _ := url.Parse(*db_dsn_flag)
	pwd, _ := db_dsn.User.Password()

	// database
	db = mysql.New("tcp", "", db_dsn.Host, db_dsn.User.Username(), pwd, db_dsn.Path[1:])
	err := db.Connect()
	if err != nil {
        panic(err)
    }

	// retrieve all the projects from a'good old DB
	projects = *NewProjectColl()

	// background worker for prepare and then write stat
	chanPrepare := make(chan *tools.Hash)
	chanWrite := make(chan *tools.Hash)
	
	// preparator!
	go func() {
		for {
			q := <-chanPrepare
			LOG("Analize: ", *q)

			project, ok := getQueryProject(q)
			if !ok {
				continue
			}
			ok = project.checkSignature(q)
			if !ok { continue }
			
			ok = project.parseType(q)
			if !ok {
				continue
			}
			
			// send to write channel
			chanWrite <- q
		}
	}()

	// writestignator!
	go func() {
		stmt, err := db.Prepare("INSERT INTO stats (project_id, type_id, value, meta_type_id, meta_value) VALUES (?, ?, ?, ?, ?)")
		if err != nil { fmt.Println(err) }
		
		for {
			q := <-chanWrite
			LOG("Write: ", *q)
			// write down
			 _, err = stmt.Run((*q)["project_id"], (*q)["type_id"], (*q)["value"], "0", "0")
			 if err != nil { fmt.Println(err) }
		}
	}()
	
	/*func (writes <-chan []byte) { 
        var buf []byte 
        for { 
			data := <-writes
            buf = append(buf, data) 
        drain:
            for len(buf) < 256*1024 { 
				select {
					case d := <-writes: 
						buf = append(buf, d) 
                    default: 
						break drain 
                } 
            }
            f.Write(buf) 
            buf = buf[:0]
        } 
	}()*/

	// http server
	http.HandleFunc("/stat", func(w http.ResponseWriter, r *http.Request) {
		query := ParseQuery(&r.URL.RawQuery)
		LOG("Got stat: ", query)
		chanPrepare <- &query
		w.Header().Set("Content-type", "application/json")
		fmt.Fprint(w, `{"res": true}`)
	})
	go http.ListenAndServe(*listen_port, nil)
	
	for i:=0; i<1; i++{
		http.Get("http://localhost:9090/stat?project_id=1&wins=2&sig=b11e90a759c795dc1fb0cf59e624fea4")
	}
	time.Sleep(time.Microsecond*200);
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

func (self *Project) parseType(q *tools.Hash) bool {
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

func (self *Project) checkSignature(q *tools.Hash) bool {
	defer tools.PrintOnPanic()

	sig, ok := (*q)["sig"]
	if !ok {
		panic("No sig passed")
	}
	delete(*q, "sig")

	arr := q.SortedKeys();
	str := ""
	for _, v := range arr {
		str = str + v + "=" + (*q)[v]
	}
	str = str + self.secret

	expected := tools.SignMd5(str)
	if expected != sig {
		panic("Wrong signature, expected " + expected)
	}
	return true
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

func ParseQuery(rawQuery *string) tools.Hash {
	queryMap, _ := url.ParseQuery(*rawQuery)
	vals := make(tools.Hash)
	for k, v := range queryMap {
		vals[k] = v[0] // never ever assume having multiple stats
	}
	return vals
}

func getQueryProject(q *tools.Hash) (Project, bool) {
	defer tools.PrintOnPanic()

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
