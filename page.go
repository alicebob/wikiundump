package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// storePage stores page p on disk
func storePage(nss []Namespace, p Page, makeSymlinks bool) error {
	filename, err := localFilename(nss, p.Title)
	if err != nil {
		return err
	}

	ff := filepath.Clean(*targetDir + "/" + filename)
	if err := os.MkdirAll(filepath.Dir(ff), 0700); err != nil {
		return err
	}
	if r := p.Redirect.Title; r != "" {
		if !makeSymlinks {
			return nil
		}
		to, err := localFilename(nss, r)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(filepath.Dir(filename), to)
		if err != nil {
			return err
		}
		if *verbose {
			fmt.Printf("symlink redirect %s -> %s\n", filename, rel)
		}
		return os.Symlink(rel, ff)
	}

	return ioutil.WriteFile(ff, []byte(p.Text), 0600)
}

// localFilename makes the full path for a page title
// "foobar" -> "/f/o/o/Foobar"
// "Template:foobar" -> "/Template/f/o/o/Foobar"
func localFilename(nss []Namespace, title string) (string, error) {
	if title == "" {
		return "", errors.New("empty title")
	}

	path := ""

	ns := findNamespace(nss, title)
	if ns.Name != "" {
		path += "/" + ns.Name
	}
	path += addComp(strings.TrimPrefix(title, ns.Name+":"), PathComponents)
	filename, err := caseFold(ns, title)
	if err != nil {
		return "", err
	}
	path += "/" + strings.Replace(filename, "/", "_", -1) // META: needed?
	return path, nil
}

func caseFold(ns Namespace, t string) (string, error) {
	switch c := ns.Case; c {
	case "first-letter":
		return ucFirst(t), nil
	default:
		return "", fmt.Errorf("unhandled namespace case: %q", c)
	}
}

// addComp adds the '/f' in '/f/foo'
func addComp(filename string, lvl int) string {
	if lvl == 0 {
		return ""
	}
	if len(filename) == 0 {
		return "/_" + addComp("", lvl-1)
	}

	ps := []rune(filename)
	sign := '_'
	// META: maybe allow unicode.IsLetter for less ASCIIish scripts?
	switch r := ps[0]; {
	case '0' <= r && r <= '9':
		sign = r
	case 'a' <= r && r <= 'z':
		sign = r
	case 'A' <= r && r <= 'Z':
		sign = unicode.ToLower(r)
	}
	return "/" + string(sign) + addComp(string(ps[1:]), lvl-1)
}

func ucFirst(t string) string {
	rs := []rune(t)
	rs[0] = unicode.ToUpper(rs[0])
	return string(rs)
}

// findNamespace looks through nss to find the namespace of page t
func findNamespace(nss []Namespace, t string) Namespace {
	var nspace Namespace
	for _, ns := range nss {
		if strings.HasPrefix(t, ns.Name+":") ||
			ns.Name == "" && nspace.Name == "" {
			nspace = ns
		}
	}
	return nspace
}

// saveNamespaces stores the namespaces as json. The namespaces are needed when
// finding pages later.
func saveNamespaces(nss []Namespace) error {
	fn := *targetDir + "/" + NamespaceFile
	f, err := os.Create(fn)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	enc.Encode(nss)
	return nil
}
