package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	PathComponents = 3 // "Abcdef -> "/a/b/c/Abcdef"
	NamespaceFile  = "namespaces.json"
)

var (
	targetDir      = flag.String("dir", "./wiki/", "target directory")
	symlinkRedirs  = flag.Bool("symlink", true, "make symlinks for redirects, or ignore them")
	keepNamespaces = flag.String("keep", "", "comma separated list of namespaces to keep. e.g.: ',Template'. Empty means everything")
	verbose        = flag.Bool("verbose", false, "print every page name")
)

func main() {
	flag.Parse()

	if err := os.MkdirAll(*targetDir, 0700); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if len(flag.Args()) == 0 {
		if err := parseFile(os.Stdin); err != nil {
			fmt.Printf("stdin: %v\n", err)
			os.Exit(2)
		}
		return
	}
	for _, file := range flag.Args() {
		f, err := os.Open(file)
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
		if err := parseFile(f); err != nil {
			fmt.Printf("%s: %v\n", file, err)
			os.Exit(2)
		}
		f.Close()
	}
}

type Namespace struct {
	Key  string `xml:"key,attr" json:"-"`
	Case string `xml:"case,attr" json:"case"`
	Name string `xml:",chardata" json:"name"`
}

type SiteInfo struct {
	SiteName   string      `xml:"sitename"`
	Namespaces []Namespace `xml:"namespaces>namespace"`
}

type Page struct {
	Title       string `xml:"title"`
	NamespaceID string `xml:"ns"`
	Text        string `xml:"revision>text"`
	Redirect    struct {
		Title string `xml:"title,attr"`
	} `xml:"redirect"`
}

func parseFile(fh io.Reader) error {
	dec := xml.NewDecoder(fh)
	var namespaces []Namespace

	for {
		elem, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		switch e := elem.(type) {
		case xml.StartElement:
			switch n := e.Name.Local; n {
			case "mediawiki":
				// root element
			case "siteinfo":
				// there is a single <siteinfo> at the top which tells us what
				// the namespaces are
				var siteinfo SiteInfo
				if dec.DecodeElement(&siteinfo, &e); err != nil {
					return nil
				}
				namespaces = siteinfo.Namespaces
				if err := saveNamespaces(namespaces); err != nil {
					return err
				}
			case "page":
				var p Page
				if dec.DecodeElement(&p, &e); err != nil {
					return nil
				}
				ns, _ := splitNamespace(namespaces, p.Title)
				if !keepNamespace(ns) {
					continue
				}
				if *verbose {
					fmt.Printf("page: %+v\n", p.Title)
				}
				if err := storePage(namespaces, p, *symlinkRedirs); err != nil {
					return err
				}
			default:
				return fmt.Errorf("unhandled toplevel element: %q\n", n)
			}
		case xml.EndElement:
		case xml.CharData:
		default:
			fmt.Printf("oops. unhandled XML construct %T\n", elem)
			os.Exit(42)
		}
	}
	return nil
}

// keepNamespace tells whether we want to store pages in this namespace on disk
func keepNamespace(ns Namespace) bool {
	if *keepNamespaces == "" {
		return true
	}
	for _, n := range strings.Split(*keepNamespaces, ",") {
		if n == ns.Name {
			return true
		}
	}
	return false
}
