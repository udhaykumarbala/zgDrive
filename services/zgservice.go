package services

import (
	"context"
	"fmt"
	"os"

	"github.com/0glabs/0g-storage-client/common/blockchain"
	"github.com/0glabs/0g-storage-client/core"
	"github.com/0glabs/0g-storage-client/indexer"
	"github.com/0glabs/0g-storage-client/node"
	"github.com/0glabs/0g-storage-client/transfer"
	"github.com/ethereum/go-ethereum/common"
	"github.com/openweb3/web3go"
)

type ZgService struct {
	evmRpc     string
	privateKey string
	flowAddr   string
	indRpc     string
	w3client   *web3go.Client
	Indexer    *indexer.Client
}

func NewZgService() (*ZgService, error) {

	// get all from env
	evmRpc := os.Getenv("EVM_RPC")
	privateKey := os.Getenv("PRIVATE_KEY")
	flowAddr := os.Getenv("FLOW_ADDR")
	indRpc := os.Getenv("IND_RPC")

	fmt.Println("evmRpc:", evmRpc)
	fmt.Println("privateKey:", privateKey)
	fmt.Println("flowAddr:", flowAddr)
	fmt.Println("indRpc:", indRpc)

	w3client := blockchain.MustNewWeb3(evmRpc, privateKey)
	defer w3client.Close()

	standardIndexer, err := indexer.NewClient(indRpc)
	if err != nil {
		return nil, err
	}

	return &ZgService{
		evmRpc:     evmRpc,
		privateKey: privateKey,
		flowAddr:   flowAddr,
		indRpc:     indRpc,
		w3client:   w3client,
		Indexer:    standardIndexer,
	}, nil
}

func FileHash(filePath string) (string, error) {
	rootHash, err := core.MerkleRoot(filePath)
	if err != nil {
		return "", err
	}

	return rootHash.String(), nil
}

func (z *ZgService) getNodes(ctx context.Context) ([]*node.ZgsClient, error) {
	nodes, err := z.Indexer.SelectNodes(ctx, 0, 1, []string{})
	if err != nil {
		return nil, err
	}

	return nodes, nil
}

func (z *ZgService) UploadFile(ctx context.Context, file string) (string, error) {
	fmt.Println("Uploading file:", file)
	nodes, err := z.getNodes(ctx)
	if err != nil {
		return "", err
	}

	fmt.Println("Nodes:", nodes)

	uploader, err := transfer.NewUploader(ctx, z.w3client, nodes)
	if err != nil {
		return "", err
	}
	fmt.Println("Uploader:", uploader)
	tx, err := uploader.UploadFile(ctx, file)
	if err != nil {
		return "", err
	}
	fmt.Println("Transaction:", tx)
	return tx.String(), nil
}

func (z *ZgService) CheckFileStatus(ctx context.Context, rootHash string) (bool, error) {
	nodes, err := z.getNodes(ctx)
	if err != nil {
		return false, err
	}

	hash := common.HexToHash(rootHash)

	for _, v := range nodes {
		info, err := v.GetFileInfo(ctx, hash)
		if err != nil {
			fmt.Println("Error getting file info:", err)
			continue
		}

		if info == nil {
			fmt.Println("File not found on node", v.URL())
			continue
		}

		fmt.Println("File found on node", v.URL())
		return true, nil
	}

	return false, nil
}

func (z *ZgService) DownloadFile(ctx context.Context, file string, hash string) (bool, error) {
	nodes, err := z.getNodes(ctx)
	if err != nil {
		return false, err
	}

	downloader, err := transfer.NewDownloader(nodes)
	if err != nil {
		return false, err
	}

	err = downloader.Download(ctx, hash, file, false)
	if err != nil {
		return false, err
	}

	return true, nil
}
