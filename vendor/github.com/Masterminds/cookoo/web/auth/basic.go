package auth

import (
	"github.com/Masterminds/cookoo"

	"encoding/base64"
	"net/http"
	"strings"
	"fmt"
)

/**
 * Perform authentication.
 *
 * Params:
 * 	- realm (string): The name of the realm. (Default: "web")
 * 	- datasource (string): The name of the datasource that should be used to authenticate.
 * 	  This datasource must be an `auth.UserDatasource`. (Default: "auth.UserDatasource")
 *
 * Context:
 * 	- http.Request (*http.Request): The HTTP request. This is usually placed into the
 * 	context for you.
 * 	- http.ResponseWriter (http.ResponseWriter): The response. This is usually placed
 * 	 into the context for you.
 *
 * Datasource:
 * 	- An auth.UserDatasource. By default, this will look for a datasource named
 * 	  "auth.UserDatasource". This can be overridden by the `datasource` param.
 *
 * Returns:
 * 	- True if the user authenticated. If not, this will send a 401 and then stop
 * 	  the current chain.
 */
func Basic(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	realm := p.Get("realm", "web").(string)
	dsName := p.Get("datasource", "auth.UserDatasource").(string)

	req := c.Get("http.Request", nil).(*http.Request)
	res := c.Get("http.ResponseWriter", nil).(http.ResponseWriter)

	ds := c.Datasource(dsName).(UserDatasource)

	authz := strings.TrimSpace(req.Header.Get("Authorization"));
	if len(authz) == 0 || !strings.Contains(authz, "Basic ") {
		return sendUnauthorized(realm, res)
	}

	user, pass, err := parseBasicString(authz)
	if err != nil {
		c.Logf("info", "Basic authentication parsing failed: %s", err)
		return sendUnauthorized(realm, res)
	}

	ok, err := ds.AuthUser(user, pass)
	if !ok {
		if err != nil {
			c.Logf("info", "Basic authentication caused an error: %s", err)
		}
		return sendUnauthorized(realm, res)
	}

	return ok, err
}

func parseBasicString(header string) (user, pass string, err error) {
	parts := strings.Split(header, " ")
	user = ""
	pass = ""
	if len(parts) < 2 {
		err = &cookoo.RecoverableError{"No auth string found."}
		return
	}

	full, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return
	}

	parts = strings.SplitN(string(full), ":", 2)
	user = parts[0]
	if len(parts) > 0 {
		pass = parts[1]
	}
	return
}

func sendUnauthorized(realm string, res http.ResponseWriter) (interface{}, cookoo.Interrupt) {
	// Send a 403
	res.Header().Set("WWW-Authenticate", fmt.Sprintf("Basic realm=\"%s\"", realm))
	http.Error(res, "Authentication Required", http.StatusUnauthorized)

	// We've already notified the client. Issue a stop.
	return nil, &cookoo.Stop{}	
}

