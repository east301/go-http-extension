package httpext

import (
	"errors"
	"net/http"
)

func RunServer(server *http.Server, tlsCertificatePath *string, tlsPrivateKeyPath *string) error {
	if tlsCertificatePath != nil || tlsPrivateKeyPath != nil {
		if tlsCertificatePath == nil || tlsPrivateKeyPath == nil {
			return errors.New("both tlsCerfificatePath and tlsPrivateKeyPath are required")
		}
		return server.ListenAndServeTLS(*tlsCertificatePath, *tlsPrivateKeyPath)
	} else {
		return server.ListenAndServe()
	}
}
