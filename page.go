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
// "Template:foobar" -> "/Template/f/o/o/Template:Foobar"
func localFilename(nss []Namespace, title string) (string, error) {
	if title == "" {
		return "", errors.New("empty title")
	}

	path := ""
	ns, pageName := splitNamespace(nss, title)

	// start with the template, if it's not empty
	if ns.Name != "" {
		path += "/" + ns.Name
	}

	// add /f/o/o
	path += addComp(pageName, PathComponents)

	// add /[namespace:]<casefolded page name>
	path += "/"
	if ns.Name != "" {
		path += ns.Name + ":"
	}
	filename, err := caseFold(ns, pageName)
	if err != nil {
		return "", err
	}
	path += strings.Replace(filename, "/", "_", -1) // META: needed?
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

// splitNamespace looks through nss to find the namespace of page t, and
// returns that (or the empty namespace), and the page without namespace prefix.
func splitNamespace(nss []Namespace, t string) (Namespace, string) {
	var (
		tLc = strings.ToLower(t)
	)
	for _, ns := range nss {
		if strings.HasPrefix(tLc, strings.ToLower(ns.Name)+":") {
			parts := strings.SplitN(t, ":", 2)
			return ns, parts[1]
		}
	}
	// no namespace
	for _, ns := range nss {
		if ns.Name == "" {
			return ns, t
		}
	}
	// Oops. No empty namespace. Weird.
	return Namespace{}, t
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
