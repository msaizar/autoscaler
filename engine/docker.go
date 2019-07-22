// Copyright 2018 Drone.IO Inc
// Use of this software is governed by the Business Source License
// that can be found in the LICENSE file.

package engine

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"time"

	docker "docker.io/go-docker"
	"docker.io/go-docker/api"
	"github.com/drone/autoscaler"
)

// clientFunc defines a builder funciton used to build and return
// the docker client from a Server. This is primarily used for
// mock unit testing.
type clientFunc func(*autoscaler.Server) (docker.APIClient, error)

// newDockerClient returns a new Docker client configured for the
// Server host and certificate chain.
func newDockerClient(server *autoscaler.Server) (docker.APIClient, error) {
	tlsCert, err := tls.X509KeyPair(server.TLSCert, server.TLSKey)
	if err != nil {
		return nil, err
	}
	tlsConfig := &tls.Config{
		ServerName:   server.Name,
		Certificates: []tls.Certificate{tlsCert},
	}
	tlsConfig.RootCAs = x509.NewCertPool()
	tlsConfig.RootCAs.AppendCertsFromPEM(server.CACert)

	client := &http.Client{
		Transport: &http.Transport{
			DialContext:(&net.Dialer{
					Timeout:   10 * time.Second,
					KeepAlive: 10 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 4 * time.Second,
			ResponseHeaderTimeout: 3 * time.Second,

			// Prevent endless redirects
			Timeout: 10 * time.Minute,

			TLSClientConfig: tlsConfig,
		},
	}
	host := fmt.Sprintf("https://%s:2376", server.Address)
	return docker.NewClient(host, api.DefaultVersion, client, nil)
}
