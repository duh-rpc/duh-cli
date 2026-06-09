package docs

import "strings"

// servePage is the bootstrap HTML returned by GET / in serve mode. It references
// the embedded ReDoc bundle at a local route and points ReDoc at the local
// /openapi.yaml endpoint. It subscribes to /events (SSE) and re-renders in place
// whenever a reload notification arrives. No external CDN references (BC-1).
const servePage = `<!DOCTYPE html>
<html>
  <head>
    <title>API Documentation</title>
    <meta charset="utf-8"/>
    <meta name="viewport" content="width=device-width, initial-scale=1"/>
    <style>body { margin: 0; padding: 0; }</style>
  </head>
  <body>
    <div id="redoc"></div>
    <script src="/redoc.standalone.js"></script>
    <script>
      function render() {
        Redoc.init('/openapi.yaml?t=' + Date.now(), {}, document.getElementById('redoc'));
      }
      render();
      var es = new EventSource('/events');
      es.onmessage = function () { render(); };
    </script>
  </body>
</html>
`

// exportTemplate is the single-file HTML for export. Both the ReDoc bundle and
// the raw spec are inlined; the spec is handed to ReDoc as a data URL so it is
// fetched and parsed in-browser with no network access (BC-1). The raw spec text
// remains present in the document (readable, grep-able) and is passed through
// unchanged — no YAML→JSON conversion.
const exportTemplate = `<!DOCTYPE html>
<html>
  <head>
    <title>API Documentation</title>
    <meta charset="utf-8"/>
    <meta name="viewport" content="width=device-width, initial-scale=1"/>
    <style>body { margin: 0; padding: 0; }</style>
  </head>
  <body>
    <div id="redoc"></div>
    <script type="application/yaml" id="openapi-spec">
%SPEC%
    </script>
    <script>
%REDOC%
    </script>
    <script>
      (function () {
        var text = document.getElementById('openapi-spec').textContent;
        var url = 'data:application/yaml;base64,' + btoa(unescape(encodeURIComponent(text)));
        Redoc.init(url, {}, document.getElementById('redoc'));
      })();
    </script>
  </body>
</html>
`

// buildExport renders the self-contained export HTML for the given spec bytes.
func buildExport(spec []byte) []byte {
	html := strings.Replace(exportTemplate, "%SPEC%", safeScriptText(string(spec)), 1)
	html = strings.Replace(html, "%REDOC%", string(redocJS), 1)
	return []byte(html)
}

// safeScriptText neutralizes the only sequence that would prematurely close the
// raw-text <script> block holding the spec, so the spec cannot break out of its
// container. All other spec text is preserved verbatim.
func safeScriptText(s string) string {
	return strings.ReplaceAll(s, "</script", "<\\/script")
}
