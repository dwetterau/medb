package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"medb/storage"
	"net/http"
	"path"

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
	http.HandleFunc("/edit/", editPageHandler(staticDir))

	// Session manager setup
	manager := scs.NewCookieManager(sessionSecret)
	// Let the cookie stay in the browser across browser restarts
	manager.Persist(true)

	// User store setup
	store := user.NewStore(userFilePath)

	// API v1
	http.HandleFunc("/api/1/login", loginHandler(manager, store))
	http.HandleFunc("/api/1/list", listHandler(manager))
	http.HandleFunc("/api/1/sync/pull", syncPullHandler(manager))
	http.HandleFunc("/api/1/save", saveHandler(manager))
	http.HandleFunc("/api/1/edit", editHandler(manager))
	http.HandleFunc("/api/1/load", loadHandler(manager))

	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		panic(err)
	}
}

const (
	rootPathCookieName = "rootPath"
	successJSON        = "{success: true}"
)

func rootHandler(staticDir string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fullPath := staticDir + "/" + r.URL.Path[1:]
		http.ServeFile(w, r, fullPath)
	}
}

func editPageHandler(staticDir string) func(w http.ResponseWriter, r *http.Request) {
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

func syncPullHandler(sessionManager *scs.Manager) func(w http.ResponseWriter, r *http.Request) {
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

func saveHandler(sessionManager *scs.Manager) func(w http.ResponseWriter, r *http.Request) {
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
		}

		// Pull first to minimize the chance of a conflict
		err = db.Pull()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		p := path.Join("unfiled", filename)
		err = db.NewFile(p, content)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		_, err = db.CommitToGIT(fmt.Sprintf("MeDB Sync - saving unfiled note %s", filename))
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		err = db.Push()
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

		// Pull first to minimize the chance of a conflict
		err = db.Pull()
		if err != nil {
			http.Error(w, err.Error(), 500)
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

		_, err = db.CommitToGIT(fmt.Sprintf("MeDB Sync - saving updated file %s", f.ID()))
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		err = db.Push()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		fmt.Fprint(w, successJSON)
	}
}
