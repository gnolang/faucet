package http

import (
	"context"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/amino"
	rpcClient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	coreTypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// Client is the TM2 HTTP client
type Client struct {
	client rpcClient.Client
}

// NewClient creates a new TM2 HTTP client
func NewClient(remote string) (*Client, error) {
	client, err := rpcClient.NewHTTPClient(remote)
	if err != nil {
		return nil, fmt.Errorf("unable to create HTTP client, %w", err)
	}

	return &Client{
		client: client,
	}, nil
}

func (c *Client) GetAccount(address crypto.Address) (std.Account, error) {
	path := fmt.Sprintf("auth/accounts/%s", address.String())

	queryResponse, err := c.client.ABCIQuery(context.Background(), path, []byte{})
	if err != nil {
		return nil, fmt.Errorf("unable to execute ABCI query, %w", err)
	}

	var queryData struct{ BaseAccount std.BaseAccount }

	if err := amino.UnmarshalJSON(queryResponse.Response.Data, &queryData); err != nil {
		return nil, err
	}

	return &queryData.BaseAccount, nil
}

func (c *Client) SendTransactionSync(tx *std.Tx) (*coreTypes.ResultBroadcastTx, error) {
	aminoTx, err := amino.Marshal(tx)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal transaction, %w", err)
	}

	return c.client.BroadcastTxSync(context.Background(), aminoTx)
}

func (c *Client) SendTransactionCommit(tx *std.Tx) (*coreTypes.ResultBroadcastTxCommit, error) {
	aminoTx, err := amino.Marshal(tx)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal transaction, %w", err)
	}

	return c.client.BroadcastTxCommit(context.Background(), aminoTx)
}

func (c *Client) Status() (*coreTypes.ResultStatus, error) {
	return c.client.Status(context.Background(), nil)
}
