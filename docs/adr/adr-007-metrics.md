# ADR #007: Celestia-Node Metrics Collection 

## Changelog
- 2022.05.04 - Initial draft

## Decision

### Dependencies
### Minimum required metrics
#### General performance
* Disk read / write count
* CPU
  * time spent by waiting on disk for all processes
  * time spent by the CPU working on all processes
  * time spent by the CPU working on *this* process
* Memory 
  * memstats from runtime pkg
#### Erasure coding (bridge node only)
* Average time it takes to extend a block received from celestia-core
*  
#### Sampling
* Average sample time (time it takes to sample a block during `DAS`)
*  

#### Nice to have
### Minimum required hot path tracing
#### Nice to have


