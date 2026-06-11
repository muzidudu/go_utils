package htmlminify

import (
	"bytes"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
)

func TestKeepDocumentTags(t *testing.T) {
	in := `<!DOCTYPE html><html><head><title>t</title></head><body><p>x</p></body></html>`

	var without bytes.Buffer
	m := minify.New()
	m.Add("text/html", &html.Minifier{})
	if err := m.Minify("text/html", &without, bytes.NewReader([]byte(in))); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(without.String(), "</body>") || strings.Contains(without.String(), "</html>") {
		t.Fatalf("default should drop document tags, got %q", without.String())
	}

	var with bytes.Buffer
	m2 := minify.New()
	m2.Add("text/html", &html.Minifier{KeepDocumentTags: true})
	if err := m2.Minify("text/html", &with, bytes.NewReader([]byte(in))); err != nil {
		t.Fatal(err)
	}
	got := with.String()
	if !strings.Contains(got, "</body>") || !strings.Contains(got, "</html>") {
		t.Fatalf("KeepDocumentTags should keep closing tags, got %q", got)
	}
}

func TestHTMLMinifyMiddleware_KeepDocumentTags(t *testing.T) {
	app := fiber.New()
	app.Use(HTMLMinify(HTMLMinifyConfig{KeepDocumentTags: true}))
	app.Get("/", func(c fiber.Ctx) error {
		return c.Type("html").SendString(`<!DOCTYPE html><html><body><p>hi</p></body></html>`)
	})

	req := httptest.NewRequest(fiber.MethodGet, "/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	body := string(raw)
	if !strings.Contains(body, "</body>") || !strings.Contains(body, "</html>") {
		t.Fatalf("middleware KeepDocumentTags, got %q", body)
	}
}
