package web

import (
	"bytes"
	"fmt"
	"github.com/Masterminds/cookoo"
	"html/template"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"strings"
)

// Common web-oriented commands

// Flush sends content to output.
//
// If no writer is specified, this will attempt to write to whatever is in the
// Context with the key "http.ResponseWriter". If no suitable writer is found, it will
// not write to anything at all.
//
// Params:
// 	- writer: A Writer of some sort. This will try to write to the HTTP response if no writer
// 	is specified.
// 	- content: The content to write as a body. If this is a byte[], it is sent unchanged. Otherwise.
// 	we first try to convert to a string, then pass it into a writer.
// 	- contentType: The content type header (e.g. text/html). Default is text/plain
// 	- responseCode: Integer HTTP Response Code: Default is `http.StatusOK`.
// 	- headers: a map[string]string of HTTP headers. The keys will be run through
// 	 http.CannonicalHeaderKey()
//
// Note that this is optimized for writing from strings or arrays, not Readers. For larger
// objects, you may find it more efficient to use a different command.
//
// Context:
// - If this finds `web.ContentEncoding`, it will set a content-encoding header.
//
// Returns
//
// 	- boolean true
func Flush(cxt cookoo.Context, params *cookoo.Params) (interface{}, cookoo.Interrupt) {

	// Make sure we have a place to write this stuff.
	writer, ok := params.Has("writer")
	if writer == nil {
		writer, ok = cxt.Has("http.ResponseWriter")
		if !ok {
			return false, nil
		}
	}
	out := writer.(http.ResponseWriter)

	// Get the rest of the info.
	code := params.Get("responseCode", http.StatusOK).(int)
	header := out.Header()
	contentType := params.Get("contentType", "text/plain; charset=utf-8").(string)

	// Prepare the content.
	var content []byte
	rawContent, ok := params.Has("content")
	if !ok {
		// No content. Send nothing in the body.
		content = []byte("")
	} else if byteContent, ok := rawContent.([]byte); ok {
		// Got a byte[]; add it as is.
		content = byteContent
	} else {
		// Use the formatter to convert to a string, and then
		// cast it to bytes.
		content = []byte(fmt.Sprintf("%v", rawContent))
	}

	// Add headers:
	header.Set(http.CanonicalHeaderKey("content-type"), contentType)

	te := cxt.Get(ContentEncoding, "").(string)
	if len(te) > 0 {
		header.Set(http.CanonicalHeaderKey("transfer-encoding"), te)
	}

	headerO, ok := params.Has("headers")
	if ok {
		headers := headerO.(map[string]string)
		for k, v := range headers {
			header.Add(http.CanonicalHeaderKey(k), v)
		}
	}

	// Send the headers.
	out.WriteHeader(code)

	//io.WriteString(out, content)
	out.Write(content)

	return true, nil
}

// RenderHTML renders an HTML template.
//
// This uses the `html/template` system built into Go to render data into a writer.
//
// Params:
// 	- template (required): An html/templates.Template object.
// 	- templateName (required): The name of the template to render.
// 	- values: An interface{} with the values to be passed to the template. If
// 	  this is not specified, the contents of the Context are passed as a map[string]interface{}.
// 	  Note that datasources, in this model, are not accessible to the template.
// 	- writer: The writer that data should be sent to. By default, this will create a new
// 	  Buffer and put it into the context. (If no Writer was passed in, the returned writer
// 	  is actually a bytes.Buffer.) To flush the contents directly to the client, you can
// 	  use `.Using('writer').From('http.ResponseWriter')`.
//
// Returns
// 	- An io.Writer. The template's contents have already been written into the writer.
//
// Example:
//
//	reg.Route("GET /html", "Test HTML").
//		Does(cookoo.AddToContext, "_").
//			Using("Title").WithDefault("Hello World").
//			Using("Body").WithDefault("This is the body.").
//		Does(web.RenderHTML, "render").
//			Using("template").From('cxt:templateCache').
//			Using("templateName").WithDefault("index.html").
//		Does(web.Flush, "_").
//			Using("contentType").WithDefault("text/html").
//			Using("content").From("cxt:render")
//
// In the example above, we do three things:
// 	- Add Title and Body to the context. For the template rendered, it will see these as
// 	  {{.Title}} and {{.Body}}.
// 	- Render the template located in a local file called "index.html". It is recommended that
// 	  a template.Template object be created at startup. This way, all of the templates can
// 	  be cached immediately and shared throughout processing.
// 	- Flush the result out to the client. This gives you a chance to add any additional headers.
func RenderHTML(cxt cookoo.Context, params *cookoo.Params) (interface{}, cookoo.Interrupt) {
	ok, missing := params.Requires("template", "templateName")
	if !ok {
		return nil, &cookoo.FatalError{"Missing params: " + strings.Join(missing, ", ")}
	}

	var buf bytes.Buffer
	out := params.Get("writer", &buf).(io.Writer)
	tplName := params.Get("templateName", nil).(string)
	tpl := params.Get("template", nil).(*template.Template)
	vals := params.Get("values", cxt.AsMap())

	err := tpl.ExecuteTemplate(out, tplName, vals)
	if err != nil {
		log.Printf("Recoverable error parsing template: %s", err)
		// XXX: This outputs partially completed templates. Is this what we want?
		io.WriteString(out, "Template error. The error has been logged.")
		return out, &cookoo.RecoverableError{"Template failed to completely render."}
	}
	return out, nil
}

// ServerInfo gets the server info for this request.
//
// This assumes that `http.Request` and `http.ResponseWriter` are in the context, which
// they are by default.
//
// Returns:
// 	- boolean true
func ServerInfo(cxt cookoo.Context, params *cookoo.Params) (interface{}, cookoo.Interrupt) {
	req := cxt.Get("http.Request", nil).(*http.Request)
	out := cxt.Get("http.ResponseWriter", nil).(http.ResponseWriter)

	out.Header().Add("X-Foo", "Bar")
	out.Header().Add("Content-type", "text/plain; charset=utf-8")

	fmt.Fprintf(out, "Request:\n %+v\n", req)
	fmt.Fprintf(out, "\n\n\nResponse:\n%+v\n", out)
	return true, nil
}

const ContentEncoding = "web.ContentEncoding"

// GuessContentType guesses the MIME type of a given name.
//
// Name should be a path-like thing with an extension. E.g. foo.html,
// foo/bar/baz.css
//
// If this detects a file with extensions like gz, zip, or Z, it will also
// set the context `web.ContentEncoding` to the appropriate encoding.
//
// Params:
//	- name (string): The filename-like thing to use to guess the content type.
// Returns:
//	string content-type
func GuessContentType(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	n := p.Get("name", "").(string)

	ext := strings.ToLower(path.Ext(n))
	switch ext {
	case ".Z":
		c.Put(ContentEncoding, "compress")
	case ".gz", ".gzip", ".tgz":
		c.Put(ContentEncoding, "gzip")
		if ext == "tgz" {
			ext = "tar"
		}
	case ".bz2", ".bzip2", ".tbz2":
		c.Put(ContentEncoding, "bzip2")
		if ext == "tbz2" {
			ext = "tar"
		}
	}

	return mime.TypeByExtension(ext), nil
}

// ServeFiles is a cookoo command to serve files from a set of filesystem directories.
//
// If no writer is specified, this will attempt to write to whatever is in the
// Context with the key "http.ResponseWriter". If no suitable writer is found, it will
// not write to anything at all.
//
// Example:
//
//     registry.Route("GET /**", "Serve assets").
//         Does(web.ServeFiles, "fileServer").
//            Using("directory").WithDefault("static")
//
// Example 2:
//
//     registry.Route("GET /foo/**", "Serve assets").
//         Does(web.ServeFiles, "fileServer").
//             Using("directory").WithDefault("static").
//             Using("removePrefix").WithDefault("/foo")
//
// Params:
// 	- directory: A directory to serve files from.
// 	- removePrefix: A prefix to remove from the url before looking for it on the filesystem.
// 	- writer: A Writer of some sort. This will try to write to the HTTP response if no writer
// 	  is specified.
// 	- request: A request of some sort. This will try to use the HTTP request if no request
// 	  is specified.
func ServeFiles(cxt cookoo.Context, params *cookoo.Params) (interface{}, cookoo.Interrupt) {

	writer, ok := params.Has("writer")
	if writer == nil {
		writer, ok = cxt.Has("http.ResponseWriter")
		if !ok {
			return nil, &cookoo.Reroute{"@404"}
		}
	}
	out := writer.(http.ResponseWriter)

	req, ok := params.Has("request")
	if req == nil {
		req, ok = cxt.Has("http.Request")
		if !ok {
			return nil, &cookoo.Reroute{"@404"}
		}
	}

	in := req.(*http.Request)

	directory := params.Get("directory", nil)
	if directory == nil {
		return nil, &cookoo.Reroute{"@404"}
	}

	prefix := params.Get("removePrefix", "").(string)
	urlPath := strings.TrimPrefix(in.URL.Path, prefix)
	staticFile := path.Join(directory.(string), urlPath)

	info, err := os.Stat(staticFile)
	if err != nil {
		return nil, &cookoo.Reroute{"@404"}
	}

	if info.IsDir() == false {
		http.ServeFile(out, in, staticFile)
		return true, nil
	}
	return nil, &cookoo.Reroute{"@404"}
}
