package auth

import (
	"fmt"
	"html/template"
	"testing"
)

func TestRenderEmail(t *testing.T) {
	rawURL := "https://www.google.com?foo=bar&baz=qux"
	for _, tpl := range []string{loginTpl, registerTpl} {
		out, err := renderEmail(tpl, data{URL: template.URL(rawURL)})
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(out)
	}
}
