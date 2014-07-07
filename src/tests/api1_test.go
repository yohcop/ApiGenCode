package main

import (
	"api1"
	"testing"
)

func TestStruct(t *testing.T) {
	r := &api1.Location{12, 88}
	t.Log(r)
}

// ========================================================
// Create a simple implementation of APISample, and test a call
// to a function that only accepts implementations of this interface.
type impl struct{}

func (i *impl) LocationsList() (*api1.Locations, error) {
	return nil, nil
}
func (i *impl) NewLocation(input *api1.Location) (*api1.Location, error) {
	return nil, nil
}

func testInterfaceFakeSetup(api api1.APISample, t *testing.T) {
	_, err := api.LocationsList()
	if err != nil {
		t.Fail()
	}
}

func TestInterface(t *testing.T) {
	i := &impl{}
	testInterfaceFakeSetup(i, t)
}
