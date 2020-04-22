package metrics

import "github.com/openfaas/faas-provider/auth"

type MetadataQuery struct {
	Credentials *auth.BasicAuthCredentials
}

func NewMetadataQuery(credentials *auth.BasicAuthCredentials) *MetadataQuery {
	return &MetadataQuery{Credentials: credentials}
}
