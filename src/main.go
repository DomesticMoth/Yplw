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

func main() {
	peers, err := getPeersList()
    if err != nil { panic(err) }
    peers = normaliseUris(peers)
    for _, peer := range peers {
    	fmt.Println(peer)
    }
}


/*
socks://localhost:4447/7mx6ztmimo5nrnmydjjtkr6maupknr3zlyr33umly22pqnivyxcq.b32.i2p:46944
socks://localhost:4447/gqt6l2wox5jndysfllrgdr6mp473t24mdi7f3iz6lugpzv3z67wq.b32.i2p:63412
socks://localhost:4447/i6lbsjw7kh4gqmbylcsjtfh3juj3dbbk24bwzrpgvtalhs7xagoa.b32.i2p:2721
socks://localhost:4447/2ddoxwucodjfy6u34v3toxm2a463nzd3f523hhfzc3y5de7twj6a.b32.i2p:2281
socks://localhost:4447/ro7bwwx7ch6echfqwgivsi37tgojgcrjosgq5jrgx5ebadu4xsaq.b32.i2p:39565
socks://localhost:4447/5cejostqhllvdnbjgmtsua46z6wt5eiecirljvthlz45yvn5hswa.b32.i2p:61944
socks://localhost:4447/xmgzqfidm3zn4y3vfljqiuyfxorifntjca3rfnwv3dbbfcrx4uca.b32.i2p:30112
socks://localhost:4447/3qqi3lxscvx2ebatj36y6wmdzaah7eblf5fl7scayp6wgyhh3vpa.b32.i2p
socks://localhost:9050/yggnekkmyitzepgl5ltdl277y5wdg36n4pc45sualo3yesm3usnuwyad.onion:1863
socks://localhost:9050/x7dqdmjb7y5ykj4kgirwzj62wrrd3t5dv57oy7oyidnf7cpthd4k7ryd.onion:5222
socks://localhost:9050/nxuwjikhsirri2rbrdlphstsn3jr2qzjrsylwkt65rh2miycr5n24tid.onion:706
socks://localhost:9050/fllrj72kxnenalmmi3uz22ljqnmuex4h2jlhwnapxlzrnn7lknadxuqd.onion:706
socks://localhost:9050/douchedeiqqvyyylqorwpej4q3oz46n2shpngp7d27tlcnufnpwag7ad.onion:5222
socks://localhost:9050/p2pkbqdgvabddixbbr2y7vrra4qxq3sejfep2qknfu4owh7e3i622dqd.onion:133
*/
