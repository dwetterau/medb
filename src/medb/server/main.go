package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"medb/storage"
	"net/http"
	"sort"
	"strings"

	"medb/server/user"

	"github.com/alexedwards/scs"
	"github.com/google/uuid"
)

func main() {
	var staticDir string
	var userFilePath string
	var sessionSecret string
	port := 3000

	flag.StringVar(&staticDir, "static", staticDir, "path to the static dir for the ui")
	flag.StringVar(&userFilePath, "usersFilePath", userFilePath, "path to the file with user information")
	flag.StringVar(&sessionSecret, "sessionSecret", sessionSecret, "32 char random string to use for the sessions")
	flag.IntVar(&port, "port", port, "The port to listen on")
	flag.Parse()

	if staticDir == "" {
		panic("Must specify static dir!")
	}
	if userFilePath == "" {
		panic("Must specify path to the users file!")
	}
	if len(sessionSecret) != 32 {
		panic("Must specify 32 char session secret!")
	}

	staticServer := http.FileServer(http.Dir(staticDir))

	http.Handle("/static/", staticServer)
	http.HandleFunc("/", rootHandler(staticDir))

	// Session manager setup
	manager := scs.NewCookieManager(sessionSecret)
	// Let the cookie stay in the browser across browser restarts
	manager.Persist(true)

	// User store setup
	store := user.NewStore(userFilePath)

	// API v1
	http.HandleFunc("/api/1/login", loginHandler(manager, store))
	http.HandleFunc("/api/1/list", listHandler(manager))
	http.HandleFunc("/api/1/sync", syncHandler)

	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		panic(err)
	}
}

const rootPathCookieName = "rootPath"

func rootHandler(staticDir string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fullPath := staticDir + "/" + r.URL.Path[1:]
		http.ServeFile(w, r, fullPath)
	}
}

func loginHandler(sessionManager *scs.Manager, store user.Store) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Failed to parse form.", 400)
			return
		}

		username := r.PostFormValue("username")
		password := r.PostFormValue("password")
		u, err := store.Login(username, password)
		if err != nil {
			// User failed to login, send a 401
			http.Error(w, "Failed to login.", 401)
			return
		}

		// Put the user's path in the cookie, login succeeded
		session := sessionManager.Load(r)
		err = session.PutString(w, rootPathCookieName, u.Path())
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		// Login succeeded, redirect to root
		http.Redirect(w, r, "/", 303)
	}
}

func listHandler(sessionManager *scs.Manager) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		session := sessionManager.Load(r)

		rootPath, err := session.GetString(rootPathCookieName)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		if len(rootPath) == 0 {
			// User needs to login
			http.Redirect(w, r, "/login.html", 302)
			// TODO: Consider returning something in JSON here anyway that indicates where to redirect to
			// server-side rather than in the index file.
			return
		}

		db := storage.NewDB(rootPath)
		files, err := db.AllFiles()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		filesAsJSON := convertToJSON(rootPath, files)

		raw, err := json.Marshal(filesAsJSON)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		fmt.Fprint(w, string(raw))
	}
}

type jsonFile struct {
	Name     string      `json:"name""`
	State    string      `json:"state"`
	Contents []*jsonFile `json:"contents"`
	Id       uuid.UUID   `json:"id"`
}

func convertToJSON(rootPath string, files []storage.File) []*jsonFile {
	// Sort shortest paths first to keep things stable
	sort.Slice(files, func(i, j int) bool {
		p1 := files[i].Path()
		p2 := files[j].Path()
		if len(p1) != len(p2) {
			return len(p1) < len(p2)
		}
		return strings.Compare(p1, p2) < 0
	})
	tree := &jsonFile{}

	for _, file := range files {
		cur := tree
		path := file.Path()[len(rootPath):]
		pathStrings := strings.Split(path, "/")
		for componentIndex, component := range pathStrings {
			if component == "" {
				continue
			}
			foundIndex := -1
			for i, val := range cur.Contents {
				if val.Name == component {
					foundIndex = i
					break
				}
			}
			if foundIndex == -1 {
				cur.Contents = append(
					cur.Contents,
					&jsonFile{Name: component},
				)
				foundIndex = len(cur.Contents) - 1
			}
			if componentIndex == len(pathStrings)-1 {
				cur.Contents[foundIndex].Id = file.ID()
			}
			// Move down the tree
			cur = cur.Contents[foundIndex]
		}
	}
	toProcess := []*jsonFile{tree}
	for len(toProcess) > 0 {
		cur := toProcess[len(toProcess)-1]
		toProcess = toProcess[:len(toProcess)-1]

		if len(cur.Contents) > 0 {
			cur.State = "expanded"
		} else {
			cur.State = "collapsed"
		}
		for _, c := range cur.Contents {
			toProcess = append(toProcess, c)
		}
	}

	return tree.Contents
}

func syncHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Print the synced status
	http.Error(w, "not implemented", 500)
}
