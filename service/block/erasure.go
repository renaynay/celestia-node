package block

import (
	"fmt"

	"github.com/celestiaorg/celestia-core/types"
)

// ErasureCodeBlock // TODO document
func (s *Service) ErasureCodeBlock(raw *types.Block) (*ErasureCodedBlock, error) {
	// TODO
	fmt.Println(raw.Header)
	return new(ErasureCodedBlock), nil
}
