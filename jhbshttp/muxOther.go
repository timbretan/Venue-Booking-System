package jhbshttp

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// InvalidCookie is for users having invalid cookie and landing at /invalid-cookie
// triggered by askForMemberCookie and askForAdminCookie
// at askForCookie.go
func InvalidCookie(w http.ResponseWriter, req *http.Request) {

	// find jhMemberCookie and jhAdminCookie
	// and delete them if found
	jhMemberCookie, err := req.Cookie("jhMemberCookie")
	if err == nil {
		jhMemberCookie.MaxAge = -1 // delete cookie
		http.SetCookie(w, jhMemberCookie)
	}
	jhAdminCookie, err := req.Cookie("jhAdminCookie")
	if err == nil {
		jhAdminCookie.MaxAge = -1 // delete cookie
		http.SetCookie(w, jhAdminCookie)
	}
	fmt.Println("done")

	// redirect user back to home page
	io.WriteString(w, `
		<html>
			<meta http-equiv="refresh" content="2;url=/" />
			<body style="text-align: center; display: block;">
			<div class="redirect-message">
				<h2>Welcome to June Holidays Booking System. Re-directing you to our home page...</h2>
			</div>
			<link href="css/style.css" type="text/css" rel="stylesheet">
			</body>
		</html>
	`)

}

// serveResource serves CSS, JS, PNG, JPG and miscellaneous file types
func ServeResource(w http.ResponseWriter, req *http.Request) {
	path := "." + req.URL.Path
	var contentType string
	if strings.HasSuffix(path, ".css") {
		contentType = "text/css"
	} else if strings.HasSuffix(path, ".png") {
		contentType = "image/png"
	} else {
		contentType = "text/plain"
	}

	f, err := os.Open(path)

	if err == nil {
		defer f.Close()
		w.Header().Add("Content-Type", contentType)

		br := bufio.NewReader(f)
		br.WriteTo(w)
	} else {
		w.WriteHeader(404)
	}

}
