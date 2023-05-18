package adhelper

import (
	"crypto/x509"
	"fmt"
	"os"
)

func certPoolFromFile(filename string) (*x509.CertPool, error) {
	certPEM, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("could not read cert file: %w", err)
	}
	pool := x509.NewCertPool()
	ok := pool.AppendCertsFromPEM([]byte(certPEM))
	if !ok {
		return nil, fmt.Errorf("failed to parse certificate in %s", filename)
	}

	return pool, nil
}
