package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"medb/server/stopwatch"
	"medb/server/user"
	"medb/storage"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

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
	http.HandleFunc("/", handlerTimer("rootView", rootViewHandler(staticDir)))
	http.HandleFunc("/edit/", handlerTimer("editView", editViewHandler(staticDir)))

	// Session manager setup
	manager := scs.NewCookieManager(sessionSecret)
	// Let the cookie stay in the browser across browser restarts
	manager.Persist(true)

	// User store setup
	store := user.NewStore(userFilePath)

	// API v1
	http.HandleFunc("/api/1/login", handlerTimer("login", loginHandler(manager, store)))
	http.HandleFunc("/api/1/list", handlerTimer("list", listHandler(manager)))
	http.HandleFunc("/api/1/search", handlerTimer("search", searchHandler(manager)))
	http.HandleFunc("/api/1/pull", handlerTimer("pull", pullHandler(manager)))
	http.HandleFunc("/api/1/push", handlerTimer("push", pushHandler(manager)))
	http.HandleFunc("/api/1/commit", handlerTimer("commit", commitHandler(manager)))
	http.HandleFunc("/api/1/edit", handlerTimer("edit", editHandler(manager)))
	http.HandleFunc("/api/1/load", handlerTimer("load", loadHandler(manager)))
	http.HandleFunc("/api/1/git/info", handlerTimer("git/info", gitInfoHandler(manager)))

	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		panic(err)
	}
}

const (
	rootPathCookieName = "rootPath"
	successJSON        = "{success: true}"
)

var logger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)

func handlerTimer(
	name string,
	handler func(w http.ResponseWriter, r *http.Request),
) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		handler(w, r)
		logger.Printf("Handled %s in %v", name, time.Since(start))
	}
}

func rootViewHandler(staticDir string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fullPath := staticDir + "/" + r.URL.Path[1:]
		http.ServeFile(w, r, fullPath)
	}
}

func editViewHandler(staticDir string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, staticDir+"/index.html")
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

func getDB(w http.ResponseWriter, r *http.Request, sessionManager *scs.Manager) storage.DB {
	defer stopwatch.Start("getDB").Stop(logger)
	session := sessionManager.Load(r)

	rootPath, err := session.GetString(rootPathCookieName)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return nil
	}
	if len(rootPath) == 0 {
		// User needs to login
		http.Redirect(w, r, "/login.html", 302)
		// TODO: Consider returning something in JSON here anyway that indicates where to redirect to
		// server-side rather than in the index file.
		return nil
	}

	return storage.NewDB(rootPath)
}

func listHandler(sessionManager *scs.Manager) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		db := getDB(w, r, sessionManager)
		if db == nil {
			// This doesn't write an error because we already did that
			return
		}

		filesAsJSON, err := db.AsJSON()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		raw, err := json.Marshal(filesAsJSON)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		fmt.Fprint(w, string(raw))
	}
}

func searchHandler(sessionManager *scs.Manager) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		db := getDB(w, r, sessionManager)
		if db == nil {
			// This doesn't write an error because we already did that
			return
		}

		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Failed to parse form.", 400)
			return
		}

		query := r.PostFormValue("query")
		if len(query) == 0 {
			http.Error(w, "Invalid search query", 400)
			return
		}

		results, err := db.Search(query, storage.SearchOptions{Limit: 5})
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		jsonResults := make([]storage.JSONFile, len(results))
		for i, result := range results {
			jsonResults[i] = storage.JSONFile{
				Name:  result.Name(),
				State: "file",
				Id:    result.ID(),
			}
		}

		raw, err := json.Marshal(jsonResults)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		fmt.Fprint(w, string(raw))
	}
}

func pullHandler(sessionManager *scs.Manager) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		db := getDB(w, r, sessionManager)
		if db == nil {
			// This doesn't write an error because we already did that
			return
		}
		err := db.Pull()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		fmt.Fprint(w, successJSON)
	}
}

func pushHandler(sessionManager *scs.Manager) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		db := getDB(w, r, sessionManager)
		if db == nil {
			// This doesn't write an error because we already did that
			return
		}
		err := db.Push()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		fmt.Fprint(w, successJSON)
	}
}

func commitHandler(sessionManager *scs.Manager) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		db := getDB(w, r, sessionManager)
		if db == nil {
			// This doesn't write an error because we already did that
			return
		}

		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Failed to parse form.", 400)
			return
		}

		filename := r.PostFormValue("filename")
		content := r.PostFormValue("content")

		if len(filename) == 0 {
			http.Error(w, "Invalid filename", 400)
			return
		}

		p := filename
		if strings.LastIndex(filename, "/") == -1 {
			p = path.Join("unfiled", p)
		}
		err = db.NewFile(p, content)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		err = db.CommitToGIT(fmt.Sprintf("MeDB Sync - saving note at %s", p))
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		fmt.Fprint(w, successJSON)
	}
}

func loadHandler(sessionManager *scs.Manager) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		db := getDB(w, r, sessionManager)
		if db == nil {
			// This doesn't write an error because we already did that
			return
		}

		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Failed to parse form.", 400)
			return
		}

		fileIDRaw := r.PostFormValue("fileID")
		fileID, err := uuid.Parse(fileIDRaw)
		if err != nil {
			http.Error(w, "unable to parse fileid", 400)
			return
		}
		f, err := db.LoadFile(fileID)
		if err != nil {
			http.Error(w, err.Error(), 404)
			return
		}

		fileAsJSON := struct {
			Id      string `json:"id"`
			Name    string `json:"name"`
			Content string `json:"content"`
		}{
			Id:      f.ID().String(),
			Name:    f.Name(),
			Content: f.Content(),
		}
		raw, err := json.Marshal(fileAsJSON)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		fmt.Fprint(w, string(raw))
	}
}

func editHandler(sessionManager *scs.Manager) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		db := getDB(w, r, sessionManager)
		if db == nil {
			// This doesn't write an error because we already did that
			return
		}

		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Failed to parse form.", 400)
			return
		}

		fileIDRaw := r.PostFormValue("fileID")
		content := r.PostFormValue("fileContent")

		fileID, err := uuid.Parse(fileIDRaw)
		if err != nil {
			http.Error(w, "unable to parse fileid", 400)
			return
		}
		f, err := db.LoadFile(fileID)
		if err != nil {
			http.Error(w, err.Error(), 404)
			return
		}
		f.Update(content)
		err = db.SaveFile(f)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		err = db.CommitToGIT(fmt.Sprintf("MeDB Sync - saving updated file %s", f.ID()))
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		fmt.Fprint(w, successJSON)
	}
}

func gitInfoHandler(sessionManager *scs.Manager) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		db := getDB(w, r, sessionManager)
		if db == nil {
			// This doesn't write an error because we already did that
			return
		}

		// Fetch so that our origin counts are accurate later
		err := db.Fetch()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		lastCommitTS, err := db.LastCommitTS()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		lastPullTS, err := db.LastPullTS()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		aheadBehind, err := db.AheadBehindOriginMaster()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		gitInfoStruct := struct {
			LastCommit    string `json:"lastCommit"`
			LastPull      string `json:"lastPull"`
			RemoteAheadBy string `json:"remoteAheadBy"`
			LocalAheadBy  string `json:"localAheadBy"`
		}{
			LastCommit:    fmt.Sprintf("Last Commit: %v ago.", time.Since(lastCommitTS)),
			LastPull:      fmt.Sprintf("Last Pull: %v ago.", time.Since(lastPullTS)),
			RemoteAheadBy: fmt.Sprintf("Remote ahead by: %d", aheadBehind.OriginAheadBy),
			LocalAheadBy:  fmt.Sprintf("Local ahead by: %d", aheadBehind.LocalAheadBy),
		}
		raw, err := json.Marshal(gitInfoStruct)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		fmt.Fprint(w, string(raw))
	}
}
