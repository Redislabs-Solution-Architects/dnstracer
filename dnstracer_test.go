package main

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/Redislabs-Solution-Architects/dnstracer/collection"
	"github.com/Redislabs-Solution-Architects/dnstracer/rules"
)

var validCollection = collection.Collection{
	LocalA:          []string{"1.1.1.1", "2.2.2.2"},
	GoogleA:         []string{"1.1.1.1", "2.2.2.2"},
	CFlareA:         []string{"1.1.1.1", "2.2.2.2"},
	LocalNS:         []string{"ns1.example.com", "ns2.example.com", "ns3.example.com"},
	GoogleNS:        []string{"ns1.example.com", "ns2.example.com", "ns3.example.com"},
	CFlareNS:        []string{"ns1.example.com", "ns2.example.com", "ns3.example.com"},
	LocalGlue:       []string{"3.3.3.3", "1.1.1.1", "2.2.2.2"},
	GoogleGlue:      []string{"3.3.3.3", "1.1.1.1", "2.2.2.2"},
	CFlareGlue:      []string{"3.3.3.3", "1.1.1.1", "2.2.2.2"},
	PublicMatchA:    true,
	LocalMatchA:     true,
	PublicMatchNS:   true,
	LocalMatchNS:    true,
	PublicMatchGlue: true,
	LocalMatchGlue:  true,
	EndpointStatus:  []bool{true, true, true},
}

/* TestCollection : just make sure the data structure initializes to false */
func TestCollection(t *testing.T) {
	a := collection.Collection{}
	if a.PublicMatchA {
		t.Errorf("Struct should initialize to false")
	}
}

/* TestRulesAgainstEmptyCollection : everything should return false if the Collection is empty */
func TestRulesAgainstEmptyCollection(t *testing.T) {
	a := collection.Collection{}
	r := rules.Check(a, false)
	e := rules.Results{ResultA: false, ResultNS: false, ResultGlue: false, ResultAccess: false}
	if reflect.DeepEqual(e, r) {
		t.Log("Empty collection works")
	} else {
		t.Error("Empty collection should equal all false rules\n")
		fmt.Printf("%+v\n", e)
		fmt.Printf("%+v\n", r)
	}
}

/* TestRulesAgainstEmptyCollection : everything should return false if the Collection is empty */
func TestRulesAgainstFullCollection(t *testing.T) {
	r := rules.Check(validCollection, false)
	e := rules.Results{ResultA: true, ResultNS: true, ResultGlue: true, ResultAccess: true}
	if reflect.DeepEqual(e, r) {
		t.Log("Full valid collection works")
	} else {
		fmt.Printf("%+v\n", r)
		t.Error("Full valid collection should return all True\n")
	}
}
