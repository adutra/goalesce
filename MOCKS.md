# Usage of mock objects

Mock objects are generated with [Mockery](https://github.com/vektra/mockery). 

To install mockery on macOS:

    brew install mockery

Currently, only a mock for the `Coalescer` interface is generated. To re-generate the mock, run the following command:

    mockery --name=Coalescer --structname=mockCoalescer --inpackage --filename=coalescer_test.go
