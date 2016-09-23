package main

import (
	"testing"
)

func TestLocalFilename(t *testing.T) {
	nss := []Namespace{
		{
			Name: "",
			Case: "first-letter",
		},
		{
			Name: "Template",
			Case: "first-letter",
		},
	}

	for title, name := range map[string]string{
		"Accordion":                    "/a/c/c/Accordion",
		"101 Dalmatians (1961 movie)":  "/1/0/1/101 Dalmatians (1961 movie)",
		"A cappella":                   "/a/_/c/A cappella",
		"a cappella":                   "/a/_/c/A cappella",
		"Acarajé":                      "/a/c/a/Acarajé",
		"Açaí Palm":                    "/a/_/a/Açaí Palm",
		"-1":                           "/_/1/_/-1",
		"10":                           "/1/0/_/10",
		"A4":                           "/a/4/_/A4",
		"Aaa":                          "/a/a/a/Aaa",
		"A∴A∴":                         "/a/_/a/A∴A∴",
		"Not a Template:Abbreviations": "/n/o/t/Not a Template:Abbreviations",
		"Template:Abbreviations":       "/Template/a/b/b/Template:Abbreviations",
		"Template:AA":                  "/Template/a/a/_/Template:AA",
	} {
		ln, err := localFilename(nss, title)
		if err != nil {
			t.Fatal(err)
		}
		if have, want := ln, name; have != want {
			t.Errorf("have %q, want %q", have, want)
		}
	}
}
