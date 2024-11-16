package chain

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/waynewu411/blocktasks/pkg/config"
	"github.com/waynewu411/blocktasks/pkg/request"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestBase_GetBlocksWithFullTxnsAndLogs(t *testing.T) {
	request := request.NewMockRequest(gomock.NewController(t))
	base := NewBaseChain(zap.NewNop(), config.ChainConfig{}, request)

	testData, err := os.ReadFile("test_data/getblocks_withfulltxns_withlogs.json")
	require.NoError(t, err)

	request.EXPECT().MakeRequest(
		http.MethodPost,
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
	).Return(testData, nil)

	blocks, err := base.GetBlocks(context.Background(), 21646720, 21646723, true, true, []string{}, []string{})
	require.NoError(t, err)
	require.Len(t, blocks, 4)
	for _, block := range blocks {
		require.Greater(t, block.Timestamp, int64(1000000000000)) // make sure block timestamp is milliseconds
		require.NotEmpty(t, block.Txns)
		require.Empty(t, block.TxnHashes)
		require.NotEmpty(t, block.Logs)
	}
}

func TestBase_GetBlocksWithoutFullTxnsAndWithLogs(t *testing.T) {
	request := request.NewMockRequest(gomock.NewController(t))
	base := NewBaseChain(zap.NewNop(), config.ChainConfig{}, request)

	testData, err := os.ReadFile("test_data/getblocks_withoutfulltxns_withlogs.json")
	require.NoError(t, err)

	request.EXPECT().MakeRequest(
		http.MethodPost,
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
	).Return(testData, nil)

	blocks, err := base.GetBlocks(context.Background(), 21646720, 21646723, false, true, []string{}, []string{})
	require.NoError(t, err)
	require.Len(t, blocks, 4)
	for _, block := range blocks {
		require.Greater(t, block.Timestamp, int64(1000000000000)) // make sure block timestamp is milliseconds
		require.Empty(t, block.Txns)
		require.NotEmpty(t, block.TxnHashes)
		require.NotEmpty(t, block.Logs)
	}
}

func TestBase_GetBlocksWithFullTxnsAndWithoutLogs(t *testing.T) {
	request := request.NewMockRequest(gomock.NewController(t))
	base := NewBaseChain(zap.NewNop(), config.ChainConfig{}, request)

	testData, err := os.ReadFile("test_data/getblocks_withfulltxns_withoutlogs.json")
	require.NoError(t, err)

	request.EXPECT().MakeRequest(
		http.MethodPost,
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
	).Return(testData, nil)

	blocks, err := base.GetBlocks(context.Background(), 21646720, 21646723, true, false, []string{}, []string{})
	require.NoError(t, err)
	require.Len(t, blocks, 4)
	for _, block := range blocks {
		require.Greater(t, block.Timestamp, int64(1000000000000)) // make sure block timestamp is milliseconds
		require.NotEmpty(t, block.Txns)
		require.Empty(t, block.TxnHashes)
		require.Empty(t, block.Logs)
	}
}

func TestBase_GetBlocksWithoutFullTxnsAndWithoutLogs(t *testing.T) {
	request := request.NewMockRequest(gomock.NewController(t))
	base := NewBaseChain(zap.NewNop(), config.ChainConfig{}, request)

	testData, err := os.ReadFile("test_data/getblocks_withoutfulltxns_withoutlogs.json")
	require.NoError(t, err)

	request.EXPECT().MakeRequest(
		http.MethodPost,
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
	).Return(testData, nil)

	blocks, err := base.GetBlocks(context.Background(), 21646720, 21646723, false, false, []string{}, []string{})
	require.NoError(t, err)
	require.Len(t, blocks, 4)
	for _, block := range blocks {
		require.Greater(t, block.Timestamp, int64(1000000000000)) // make sure block timestamp is milliseconds
		require.Empty(t, block.Txns)
		require.NotEmpty(t, block.TxnHashes)
		require.Empty(t, block.Logs)
	}
}
