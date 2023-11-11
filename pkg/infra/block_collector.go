package infra

import (
	"context"
	"sync"
	"time"

	"github.com/hyperledger-twgc/tape/pkg/infra/bitmap"
	log "github.com/sirupsen/logrus"

	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/pkg/errors"
)

// BlockCollector keeps track of committed blocks on multiple peers.
// This is used when a block is considered confirmed only when committed
// on a certain number of peers within network.
type Registrytype map[uint64]*bitmap.BitMap
type BlockCollector struct {
	sync.Mutex
	thresholdP, totalP int
	registry           map[string]Registrytype
	logger             *log.Logger
}

// AddressedBlock describe the source of block
type AddressedBlock struct {
	*peer.FilteredBlock
	Address int // source peer's number
}

// NewBlockCollector creates a BlockCollector
func NewBlockCollector(threshold int, total int) (*BlockCollector, error) {
	registry := make(map[string]Registrytype)
	if threshold <= 0 || total <= 0 {
		return nil, errors.New("threshold and total must be greater than zero")
	}
	if threshold > total {
		return nil, errors.Errorf("threshold [%d] must be less than or equal to total [%d]", threshold, total)
	}
	logger := log.New()
	logger.SetLevel(log.InfoLevel)
	return &BlockCollector{
		thresholdP: threshold,
		totalP:     total,
		registry:   registry,
		logger:     logger,
	}, nil
}

func (bc *BlockCollector) Start(
	ctx context.Context,
	blockCh <-chan *AddressedBlock,
	now time.Time,
	printResult bool, // controls whether to print block commit message. Tests set this to false to avoid polluting stdout.
) {
	for {
		select {
		case block := <-blockCh:
			bc.commit(block, now, printResult)
		case <-ctx.Done():
			return
		}
	}
}

// TODO This function contains too many functions and needs further optimization
// commit commits a block to collector.
// If the number of peers on which this block has been committed has satisfied thresholdP,
// adds the number to the totalTx.
func (bc *BlockCollector) commit(block *AddressedBlock, now time.Time, printResult bool) {

	registry, ok := bc.registry[block.ChannelId]
	if !ok {
		registry = make(Registrytype)
		bc.registry[block.ChannelId] = registry
	}

	bitMap, ok := registry[block.Number]
	if !ok {
		b, err := bitmap.NewBitMap(32)
		if err != nil {
			panic("Can not make new bitmap for BlockCollector" + err.Error())
		}
		registry[block.Number] = &b
		bitMap = &b
	}
	// When the block from Address has been received before, return directly.
	if bitMap.Has(block.Address) {
		return
	}
	bitMap.Set(block.Address)
	cnt := bitMap.Count()

	if cnt == bc.thresholdP {
		bc.logger.Infof("ChannelId:%s  Block %6d Tx %6d ",
			block.ChannelId, block.Number, len(block.FilteredTransactions))
	}

	if cnt == bc.totalP {
		delete(bc.registry[block.ChannelId], block.Number)
	}
}
