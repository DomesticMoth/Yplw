package main

import (
        "fmt"
        "strings"
        "bytes"
        "net"
        "net/url"
        billy "github.com/go-git/go-billy/v5"
        memfs "github.com/go-git/go-billy/v5/memfs"
        git "github.com/go-git/go-git/v5"
        http "github.com/go-git/go-git/v5/plumbing/transport/http"
        memory "github.com/go-git/go-git/v5/storage/memory"
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

func main() {
	peers, err := getPeersList()
    if err != nil { panic(err) }
    peers = normaliseUris(peers)
    peers = resolveNames(peers)
    fmt.Println(collectRows(peers))
}
