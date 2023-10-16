package da

import "github.com/Layr-Labs/eigenda/api/grpc/retriever"

type DAConfig struct {
	Rpc    string
	Client retriever.RetrieverClient
}
