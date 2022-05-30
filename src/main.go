package main

import (
        //"fmt"
        "time"
        "strings"
        "bytes"
        "fmt"
        "net"
        "net/url"
        "math/rand"
        h "net/http"
        billy "github.com/go-git/go-billy/v5"
        memfs "github.com/go-git/go-billy/v5/memfs"
        git "github.com/go-git/go-git/v5"
        http "github.com/go-git/go-git/v5/plumbing/transport/http"
        memory "github.com/go-git/go-git/v5/storage/memory"
        object "github.com/go-git/go-git/v5/plumbing/object"
)

func getAllFilesRecursiveByPath(fs *billy.Filesystem, path string) ([]string, error) {
	var err error = nil
	ret := []string{}
	files, err := (*fs).ReadDir(path+"/")
	if err != nil { return ret, err }
	for _, file := range files {
		if file.IsDir() {
			rec, err := getAllFilesRecursiveByPath(fs, path+"/"+file.Name())
			if err != nil { return ret, err }
			ret = append(ret, rec...)
		}else{
			ret = append(ret, strings.TrimPrefix(path+"/"+file.Name(), "/"))
		}
	}
	return ret, err
}

func getAllFilesRecursive(fs *billy.Filesystem) ([]string, error) {
	return getAllFilesRecursiveByPath(fs, "")
}

func getPeersList() ([]string, error) {
	var err error = nil
	ret := []string{}
	var storer *memory.Storage
	var fs billy.Filesystem
       storer = memory.NewStorage()
       fs = memfs.New()
       auth := &http.BasicAuth{
               Username: "",
               Password: "",
       }
       repository := "https://github.com/yggdrasil-network/public-peers"
       _, err = git.Clone(storer, fs, &git.CloneOptions{
               URL:  repository,
               Auth: auth,
       })
       if err != nil { return ret, err }
	   files, err := getAllFilesRecursive(&fs)
       if err != nil { return ret, err }
       for _, filename := range files {
       	if filename == "README.md" { continue }
        file, err := fs.Open(filename)
        if err != nil { return ret, err }
        var b bytes.Buffer
        b.ReadFrom(file)
        text := strings.Split(b.String(), "\n")
        for _, row := range text {
        	if !strings.Contains(row, "`") { continue  }
        	if strings.Contains(row, "Peers:") { continue  }
        	if strings.Contains(row, "Note ") { continue  }
        	fr := strings.Split(row, "`")
        	if len(fr) < 2 { continue }
        	row = fr[1]
        	ret = append(ret, row)
        }
        file.Close()
       }
	return ret, err
}

func normaliseUris(peers []string) []string {
	ret := []string{}
	for _, peer := range peers {
		u, err := url.Parse(peer)
		if err != nil { continue }
		if u.Scheme == "socks" {
			host, _, _ := net.SplitHostPort(u.Host)
			if !(host == "localhost" || host == "127.0.0.1") { continue }
			peer = "tcp://"+strings.TrimPrefix(peer, "socks://"+u.Host+"/")
			ret = append(ret, peer)
		}else if u.Scheme == "tcp" || u.Scheme == "tls" {
			ret = append(ret, peer)
		}
	}
	return ret
}

func resolveNames(peers []string) []string {
	ret := []string{}
	for _, peer := range peers {
		ret = append(ret, peer)
		u, err := url.Parse(peer)
		if err != nil { continue }
		host, port, _ := net.SplitHostPort(u.Host)
		if host == ""{ continue }
		addr := net.ParseIP(host)
		if addr != nil { continue }
		ips ,err := net.LookupIP(host)
		if err != nil { continue }
		for _, ip := range ips {
			if ip.To4() == nil {
				// v6
				u.Host = "["+ip.String()+"]:"+port
				ret = append(ret, u.String())
			}else{
				//v4
				u.Host = ip.String()+":"+port
				ret = append(ret, u.String())
			}
		}
		if u.Scheme == "tls" {
			// add sni
			q, _ := url.ParseQuery(u.RawQuery)
			q.Add("sni", host)
			u.RawQuery = q.Encode()
		}
		for _, ip := range ips {
			if ip.To4() == nil {
				// v6
				u.Host = "["+ip.String()+"]:"+port
				ret = append(ret, u.String())
			}else{
				//v4
				u.Host = ip.String()+":"+port
				ret = append(ret, u.String())
			}
		}
	}
	return ret
}

func collectRows(rows []string) string {
	ret := ""
	for i, row := range rows {
		ret += row
		if i < len(rows)-1 {
			ret += "\n"
		}
	}
	return ret
}

type Deduplicator struct{
	latest *string
}

func (d *Deduplicator) get(s string) *string {
	if d.latest != nil { if s == *d.latest { return nil } }
	d.latest = &s
	return &s
}

func getTimestampRow() string {
	return "# Last chainged at "+time.Now().UTC().Format("02 January 2006 15:04")+" UTC\n"
}

func publish(repo string, file string, user string, pass string, text string) error {
	var err error = nil
	var storer *memory.Storage
	var fs billy.Filesystem
    storer = memory.NewStorage()
    fs = memfs.New()
    auth := &http.BasicAuth{
            Username: user,
            Password: pass,
    }
    r, err := git.Clone(storer, fs, &git.CloneOptions{
            URL:  repo,
            Auth: auth,
    })
    if err != nil { return err }
    w, err := r.Worktree()
    if err != nil { return err }
    file = strings.TrimPrefix(file, "/")
    newFile, err := fs.Create(file)
    if err != nil { return err }
    newFile.Write([]byte(text))
    newFile.Close()
    _, err = w.Add(file)
    if err != nil { return err }
    _, err = w.Commit("Update "+file, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Yplw Bot",
			Email: "",
			When:  time.Now(),
		},
    })
    if err != nil { return err }
	err = r.Push(&git.PushOptions{
		RemoteName: "origin",
		Auth: auth,
	})
	return err
}

func rowsMixer(s string) string {
	rows := strings.Split(s, "\n")
	for i := range rows {
	    j := rand.Intn(i + 1)
	    rows[i], rows[j] = rows[j], rows[i]
	}
	return collectRows(rows)
}

func storage(update chan string, req chan chan string){
	text := ""
	for {
		select {
			case t := <- update:
				text = t
			case resp := <- req:
				resp <- text
		}
	}
}

type Config struct {
	GitUser string
	GitPass string
	PubRepo string
	PubPath string
	Header string
	Http string
	UpdateDelay time.Duration
}

func run(conf Config, update chan string) {
	d := Deduplicator{}
	for {
		peers, err := getPeersList()
		if err != nil { panic(err) }
		peers = normaliseUris(peers)
		peers = resolveNames(peers)
		text := collectRows(peers)
		dd := d.get(text)
		if dd != nil {
			text = strings.TrimSuffix(conf.Header, "\n")+"\n"+getTimestampRow()+rowsMixer(text)
			err = publish(conf.PubRepo, conf.PubPath, conf.GitUser, conf.GitPass, text)
			if err != nil { panic(err) }
			update <- text
		}
		time.Sleep(conf.UpdateDelay)
	}
}

func listener(conf Config, req chan chan string) {
	httpHandler := func(w h.ResponseWriter, r *h.Request) {
		resp := make(chan string)
		req <- resp
		txt := <- resp
		fmt.Fprintf(w, "%s", txt)
	}
	h.HandleFunc(conf.PubPath, httpHandler)
	panic(h.ListenAndServe(conf.Http, nil))
}

func main() {
	update := make(chan string)
	req := make(chan chan string)
	conf := Config{
		"KEY",
		"",
		"https://github.com/DomesticMoth/MPL",
		"/yggdrasil.txt",
		"# Some header",
		"127.0.0.1:7788",
		time.Duration(time.Second * 10000),
	}
	go storage(update, req)
	go listener(conf, req)
	run(conf, update)
	// Add logging
	// Add json configuretion
}
