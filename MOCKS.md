# Usage of mock objects

Mock objects are generated with [Mockery](https://github.com/vektra/mockery). 

To install Mockery on macOS:

    brew install mockery

For more installation options, see [installation](https://github.com/vektra/mockery#installation). Note that using `go
install` is deprecated.

Currently, only a mock for the `Coalescer` interface is generated. To re-generate the mock, run the following command:

    mockery --name=Coalescer --structname=mockCoalescer --inpackage --filename=coalescer_test.go

The Makefile has a `mocks` target that will regenerate all mock objects. It currently uses `go install` to install 
a local Mockery version, for portability reasons. This may need to be revisited in the future.
