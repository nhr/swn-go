# SWN Sector Generator (Go Edition)
A pure Go rewrite of the original [SWN Sector Generator](https://github.com/nhr/swn), built for use with the [Stars Without Number](https://sine-nomine-publishing.myshopify.com/collections/stars-without-number) role-playing game by [Sine Nomine Publishing](https://sine-nomine-publishing.myshopify.com/).

The rules for generating sectors are entirely based on the Stars Without Number role-playing game. The original generator was created out of love for the game and appreciation for the insane level of detail that the author put into his tables for rolling up a random galactic sector. This Go edition is a complete rewrite of the original Perl/CGI application as a single, self-contained binary.

## Getting Your Head Around This Code

Unlike the original Perl version with its CGI scripts and CPAN dependencies, this edition is written entirely in Go. All static assets, database files, fonts, and templates are embedded directly into the binary at compile time using Go's `embed` package. No external files or runtime dependencies are required.

### The Front End
The `static/index.html` file contains the entire user interface and most of the front-end logic, along with bundled jQuery library files. The JavaScript code handles the UI elements and makes AJAX calls back to the server:

**getRandSeed** - Called when the page loads to get a random seed from the server.

**getSector** - Hands the random seed back to the server and receives all of the details of a random galactic sector, then renders the data for display in the UI.

**Save Sector button** - Collects the current state of the sector as shown in the UI and passes it back to the server. The server generates two versions of a [TiddlyWiki](http://tiddlywiki.com/) file, compresses them into a .zip, and hands them back to the browser for download.

### The Back End
The Go back end serves a REST API using Go's standard `net/http` package:

- `GET /api/seed` - Returns a random seed token
- `GET /api/sector/{token}` - Returns full sector data as structured JSON
- `GET /api/sector/{token}/map` - Renders and returns the sector map as a PNG image
- `POST /api/sector/{token}/export` - Generates a TiddlyWiki ZIP export (accepts optional `{"stars": [...]}` body for custom star positions)

### Internal Packages
The application is organized into several internal packages:

- **generator** - Core sector generation logic: stars, worlds, NPCs, corporations, religions, political parties, and aliens
- **render** - Map image rendering using Go's image libraries
- **wiki** - TiddlyWiki file generation and export
- **dice** - Dice rolling utilities
- **names** - Name generation data and logic, including alien language files
- **conflux** - Conflux generation for sector maps
- **util** - Seed tokenization and general utilities
- **handlers** - HTTP request handlers

### Data
Sector generation tables are stored in an embedded SQLite database (`static/swn.sqlite`), and alien name generation uses language data files embedded from `static/Includes/`.

## Building

To build the `swn-go` binary, you need Go 1.25 or later installed.

Clone the repository and build:

    git clone <repo-url>
    cd swn-go
    go build -o swn-go .

This produces a single self-contained binary with all assets embedded.

### Building a Container

A `Containerfile` is included for building a Linux container image. You can build it with Podman or Docker:

    podman build -t swn-go .

Then run the container:

    podman run -p 8080:8080 swn-go

Replace `podman` with `docker` if you prefer Docker. The container uses a multi-stage build: the first stage compiles the Go binary, and the second stage copies it into a minimal Fedora image.

### Running

Start the server (defaults to port 8080):

    ./swn-go

Specify a custom port:

    ./swn-go 9090

Then open `http://localhost:8080` (or your custom port) in a browser.

## Legal Notes

* Check out the `LICENSE` file for information on how this app is licensed.

* Before making this tool publicly available, the original author contacted the game's creator and agreed to limit the tool's functionality to those rules described in the free version of Stars Without Number. Take care to avoid changing this tool in any way that may infringe the copyrights of Sine Nomine Publishing or any other RPG publishers.
