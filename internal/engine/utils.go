package engine

import (
	"ndeploy/v2/internal/db"
	"net"
)

func splitHostPort(host string) (string, string) {
	h, p, err := net.SplitHostPort(host)
	if err != nil {
		return host, ""
	}

	return h, p
}

func userAtHost(n *db.Node) string {
	host, _ := splitHostPort(n.Host)
	return n.User + "@" + host
}
