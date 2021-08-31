package block

// Service represents the Block service that can be started / stopped on a `Full` node.
// Service contains 4 main functionalities:
//		1. Fetching "raw" blocks from Celestia-Core. // TODO note: this will eventually take place on the p2p level.
// 		2. Erasure coding the "raw" blocks and producing a DataAvailabilityHeader + verifying the Data root.
// 		3. Storing erasure coded blocks.
// 		4. Serving erasure coded blocks to other `Full` node peers. // TODO note: optional for devnet
type Service struct {
			
}

func NewBlockService() (*Service, error) {
	return nil, nil
}




