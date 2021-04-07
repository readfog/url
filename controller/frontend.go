package controller

import (
	"log"
	"net/http"

	"github.com/readfog/url/assets"
	"github.com/readfog/url/response"
	"github.com/readfog/url/service/url"
)

// Index is the controller for root aka index page
// It responds to `GET /` and does not require auth token.
func Index(res http.ResponseWriter, req *http.Request) {
	// http.ServeFile(res, req, "assets/home.html")
	FileFromFS(res, req, "home.html")
}

// Banner is the controller for favicon.ico
// It responds to `GET /banner.png` and does not require auth token.
func Banner(res http.ResponseWriter, req *http.Request) {
	// http.ServeFile(res, req, "assets/urlsh-light.png")
	FileFromFS(res, req, "banner.png")
}

// Favicon is the controller for favicon.ico
// It responds to `GET /favicon.ico` and does not require auth token.
func Favicon(res http.ResponseWriter, req *http.Request) {
	// http.ServeFile(res, req, "assets/u.png")
	FileFromFS(res, req, "logo.png")
}

// Robots is the controller for robots.txt
// It responds to `GET /robots.txt` and does not require auth token.
func Robots(res http.ResponseWriter, req *http.Request) {
	// http.ServeFile(res, req, "assets/robots.txt")
	FileFromFS(res, req, "robots.txt")
}

// Status is the controller for health/status check
// It responds to `GET /status` and does not require auth token.
func Status(res http.ResponseWriter, _ *http.Request) {
	response.JSON(res, http.StatusOK, response.Body{"message": "正常"})
}

// NotFound is the controller for handling unregistered request path
// It is auto triggered if router does not find controller for request path.
func NotFound(res http.ResponseWriter, _ *http.Request) {
	response.JSON(res, http.StatusNotFound, response.Body{"message": "請求資源未找到"})
}

// ServeShortURL is the controller for serving short urls
// It responds to `GET /{shortCode}` and does not require auth token.
func ServeShortURL(res http.ResponseWriter, req *http.Request) {
	shortCode := req.URL.Path[1:]
	urlModel, status, cached := url.LookupOriginURL(shortCode)

	if cached {
		res.Header().Add("X-Cached", "true")
	}

	if status != http.StatusMovedPermanently {
		response.JSON(res, status, response.Body{"message": "請求資源未找到"})
		return
	}

	go url.IncrementHits(urlModel)
	log.Printf("ServeShortURL urlModel=%+v, cached=%v", urlModel, cached)
	http.Redirect(res, req, urlModel.OriginURL, status)
}

// FileFromFS writes the specified file from http.FileSytem into the body stream in an efficient way.
func FileFromFS(res http.ResponseWriter, req *http.Request, filepath string) {
	defer func(old string) {
		req.URL.Path = old
	}(req.URL.Path)

	req.URL.Path = filepath
	fs := assets.Assets.HTTPFileSystem()
	http.FileServer(fs).ServeHTTP(res, req)
}
