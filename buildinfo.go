package statusz

import (
	"fmt"
	"html/template"
	"net/http"
	"runtime/debug"
)

const (
	buildinfo = "buildinfo"
)

func init() {
	RegisterPage("Build information", buildinfo, buildInfoHandler)
}

// buildInfoHandler displays the runtime build information
func buildInfoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, "<html><head></head><body>")
	fmt.Fprint(w, "<h1>Build Information</h1>")
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		fmt.Fprint(w, "No build information available<p>")
	} else {
		fmt.Fprintf(w, "Go toolchain version: %s<br>", template.HTMLEscapeString(bi.GoVersion))
		fmt.Fprintf(w, "Main package path: %s<br>", template.HTMLEscapeString(bi.Path))
		var e [][]string
		for _, s := range bi.Settings {
			e = append(e, []string{s.Key, s.Value})
		}
		biTable(w, "Build Settings", "Name", "Value", e)
		e = e[:0]
		for _, m := range bi.Deps {
			e = append(e, []string{m.Path, m.Version})
		}
		biTable(w, "Module Dependencies", "Path", "Version", e)
	}
	fmt.Fprint(w, "</body></html>")
}

func biTable(w http.ResponseWriter, h2, t1, t2 string, l [][]string) {
	fmt.Fprintf(w, "<h2>%s</h2>", h2)
	fmt.Fprintf(w, "<table border=\"1\"><thead><tr><th>%s</th><th>%s</th></tr></thead><tbody>", t1, t2)
	for _, s := range l {
		fmt.Fprintf(w, "<tr><td>%s</td><td style=\"text-align: end\">%s</td></tr>", template.HTMLEscapeString(s[0]), template.HTMLEscapeString(s[1]))
	}
	fmt.Fprint(w, "</tbody></table>")
}
