package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/influx6/faux/context"
	"github.com/influx6/fractals/fhttp"
)

func main() {

	var (
		addrs        string
		hasIndexFile bool
		basePath     string
		assetPath    string
		assetURL     string
		extraFiles   string
		files        []string
	)

	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	defaultAssets := filepath.Join(pwd, "assets")

	flag.StringVar(&extraFiles, "files", "", "files: Provides a argument to contain comma seperate paths to serve from directory\n\tExample: servex -files app.js:./app.js, db.svg:./assets/db.svg")
	flag.StringVar(&addrs, "addrs", ":4050", "addrs: The address and port to use for the http server.")
	flag.StringVar(&basePath, "base", pwd, "base: This values sets the path to be loaded as the base path.\n\t")
	flag.StringVar(&assetPath, "assets", defaultAssets, "assets: sets the absolute path to use for assets.\n\t")
	flag.StringVar(&assetURL, "assetURL", "/assets/*", "assetURL: Sets the path to be used for the server in serving assegs\n\tExample: servex -assetURL static/assets")
	flag.BoolVar(&hasIndexFile, "withIndex", true, "withIndex: Indicates whether we should serve index.html as root path.")
	flag.Parse()

	if strings.TrimSpace(extraFiles) != "" {
		files = strings.Split(extraFiles, ",")
		for index, fl := range files {
			files[index] = strings.TrimSpace(fl)
		}
	}

	assetURL = strings.TrimSpace(assetURL)
	if !strings.HasPrefix(assetURL, "/") {
		assetURL = "/" + assetURL
	}

	if !strings.HasSuffix(assetURL, "*") {
		if strings.HasSuffix(assetURL, "/") {
			assetURL += "*"
		} else {
			assetURL += "/*"
		}
	}

	basePath = filepath.Clean(basePath)
	assetPath = filepath.Clean(assetPath)

	if strings.HasPrefix(basePath, ".") || !strings.HasPrefix(basePath, "/") {
		basePath = filepath.Join(pwd, basePath)
	}

	if strings.HasPrefix(assetPath, ".") || !strings.HasPrefix(assetPath, "/") {
		assetPath = filepath.Join(pwd, assetPath)
	}

	apphttp := fhttp.Drive(fhttp.MW(fhttp.CORS(), fhttp.RequestLogger(os.Stdout)))(fhttp.MW(fhttp.ResponseLogger(os.Stdout)))

	approuter := fhttp.Route(apphttp)

	approuter(fhttp.Endpoint{
		Path:    assetURL,
		Method:  "GET",
		Action:  func(ctx context.Context, rw *fhttp.Request) error { return nil },
		LocalMW: fhttp.DirFileServer(assetPath, strings.TrimSuffix(assetURL, "*")),
	})

	approuter(fhttp.Endpoint{
		Path:    "/files/*",
		Method:  "GET",
		Action:  func(ctx context.Context, rw *fhttp.Request) error { return nil },
		LocalMW: fhttp.DirFileServer(basePath, "/files/"),
	})

	if hasIndexFile {
		approuter(fhttp.Endpoint{
			Path:    "/",
			Method:  "GET",
			Action:  func(ctx context.Context, rw *fhttp.Request) error { return nil },
			LocalMW: fhttp.IndexServer(basePath, "index.html", ""),
		})
	}

	for _, fl := range files {
		flset := strings.Split(fl, ":")

		if len(flset) < 2 {
			fmt.Printf("Unable to split extra file path: Path: %q, Splits: %+q\n", fl, flset)
			continue
		}

		if !strings.HasPrefix(flset[0], "/") {
			flset[0] = "/" + flset[0]
		}

		fmt.Printf("Adding file %q with endpoint URL: %q\n", flset[1], flset[0])

		approuter(fhttp.Endpoint{
			Path:    flset[0],
			Method:  "GET",
			Action:  func(ctx context.Context, rw *fhttp.Request) error { return nil },
			LocalMW: fhttp.IndexServer(basePath, flset[1], ""),
		})
	}

	fmt.Printf("Assets URL: %q\n", assetURL)
	fmt.Printf("Assets Path: %q\n", assetPath)
	fmt.Printf("Base Path: %q\n", basePath)

	apphttp.Serve(addrs)
}
