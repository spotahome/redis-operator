/*
Package mocks will have all the mocks of the library, we'll try to use mocking using blackbox
testing and integration tests whenever is possible.
*/
package mocks // import "github.com/spotahome/kooper/mocks"

// Operator tooling mocks.
//go:generate mockery -output ./operator/resource -outpkg resource -dir ../operator/resource -name CRD
//go:generate mockery -output ./operator/controller -outpkg controller -dir ../operator/controller -name Controller
//go:generate mockery -output ./operator/handler -outpkg handler -dir ../operator/handler -name Handler

// Wrappers mocks
//go:generate mockery -output ./wrapper/time -outpkg time -dir ../wrapper/time -name Time
