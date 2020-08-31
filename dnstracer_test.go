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
	DNS2A:           []string{"1.1.1.1", "2.2.2.2"},
	DNS1A:           []string{"1.1.1.1", "2.2.2.2"},
	LocalNS:         []string{"ns1.example.com", "ns2.example.com", "ns3.example.com"},
	DNS2NS:          []string{"ns1.example.com", "ns2.example.com", "ns3.example.com"},
	DNS1NS:          []string{"ns1.example.com", "ns2.example.com", "ns3.example.com"},
	LocalGlue:       []string{"3.3.3.3", "1.1.1.1", "2.2.2.2"},
	DNS2Glue:        []string{"3.3.3.3", "1.1.1.1", "2.2.2.2"},
	DNS1Glue:        []string{"3.3.3.3", "1.1.1.1", "2.2.2.2"},
	PublicMatchA:    true,
	LocalMatchA:     true,
	PublicMatchNS:   true,
	LocalMatchNS:    true,
	PublicMatchGlue: true,
	LocalMatchGlue:  true,
	EndpointStatus:  []bool{true, true, true},
}

// TestCollection : just make sure the data structure initializes to false
func TestCollection(t *testing.T) {
	a := collection.Collection{}
	if a.PublicMatchA {
		t.Errorf("Struct should initialize to false")
	}
}

// TestRulesAgainstEmptyCollection : everything should return false if the Collection is empty
func TestRulesAgainstEmptyCollection(t *testing.T) {
	a := &collection.Collection{}
	r := rules.Check(a, false, false)
	e := rules.Results{ResultA: false, ResultNS: false, ResultGlue: false, ResultAccess: false}
	if reflect.DeepEqual(e, r) {
		t.Log("Empty collection works")
	} else {
		t.Error("Empty collection should equal all false rules\n")
		fmt.Printf("%+v\n", e)
		fmt.Printf("%+v\n", r)
	}
}

// TestRulesAgainstEmptyCollection : everything should return false if the Collection is empty
func TestRulesAgainstFullCollection(t *testing.T) {
	coll := validCollection
	r := rules.Check(&coll, false, false)
	e := rules.Results{ResultA: true, ResultNS: true, ResultGlue: true, ResultAccess: true}
	if reflect.DeepEqual(e, r) {
		t.Log("Full valid collection works")
	} else {
		fmt.Printf("%+v\n", r)
		t.Error("A full correct collection should return all true\n")
	}
}

// ---------------------------------- A RECORD TESTS ------------------------------------------

// TestRulesAgainstMissingARecord : everything should return false if the Collection is empty
func TestRulesAgainstMissingARecord(t *testing.T) {
	coll := validCollection
	coll.LocalA = nil
	r := rules.Check(&coll, false, false)
	e := rules.Results{ResultA: false, ResultNS: true, ResultGlue: true, ResultAccess: true}
	if reflect.DeepEqual(e, r) {
		t.Log("Missing A Record works")
	} else {
		fmt.Printf("Was      : %+v\n", r)
		fmt.Printf("Should Be: %+v\n", e)
		fmt.Printf("Collection: %+v\n", coll)
		t.Error("Only ResultA should return false\n")
	}
}

// TestRulesAgainstExtraARecord : if we add an extra record we should fail
func TestRulesAgainstExtraARecord(t *testing.T) {
	coll := validCollection
	coll.LocalA = []string{"1.1.1.1", "2.2.2.2", "4.4.4.4"}
	r := rules.Check(&coll, false, false)
	e := rules.Results{ResultA: false, ResultNS: true, ResultGlue: true, ResultAccess: true}
	if reflect.DeepEqual(e, r) {
		t.Log("Extra A Record works")
	} else {
		fmt.Printf("Was      : %+v\n", r)
		fmt.Printf("Should Be: %+v\n", e)
		fmt.Printf("Collection: %+v\n", coll)
		t.Error("Only ResultA should return false\n")
	}
}

// TestRulesAgainstMismatchARecord : if we have a different record we should fail
func TestRulesAgainstMismatchARecord(t *testing.T) {
	coll := validCollection
	coll.LocalA = []string{"1.1.1.1", "4.4.4.4"}
	r := rules.Check(&coll, false, false)
	e := rules.Results{ResultA: false, ResultNS: true, ResultGlue: true, ResultAccess: true}
	if reflect.DeepEqual(e, r) {
		t.Log("Extra A Record works")
	} else {
		fmt.Printf("Was      : %+v\n", r)
		fmt.Printf("Should Be: %+v\n", e)
		fmt.Printf("Collection: %+v\n", coll)
		t.Error("Only ResultA should return false\n")
	}
}

// ---------------------------------- NS RECORD TESTS ------------------------------------------

// TestRulesAgainstMissingNSRecord : everything should return false if the Collection is empty
func TestRulesAgainstMissingNSRecord(t *testing.T) {
	coll := validCollection
	coll.LocalNS = nil
	r := rules.Check(&coll, false, false)
	e := rules.Results{ResultA: true, ResultNS: false, ResultGlue: false, ResultAccess: false}
	if reflect.DeepEqual(e, r) {
		t.Log("Missing NS Record works")
	} else {
		fmt.Printf("Was      : %+v\n", r)
		fmt.Printf("Should Be: %+v\n", e)
		fmt.Printf("Collection: %+v\n", coll)
		t.Error("Full valid collection should return all True\n")
	}
}

// TestRulesAgainstExtraNSRecord : if we add an extra record we should fail
func TestRulesAgainstExtraNSRecord(t *testing.T) {
	coll := validCollection
	coll.LocalNS = append(coll.LocalNS, "ns4.example.com")
	r := rules.Check(&coll, false, false)
	e := rules.Results{ResultA: true, ResultNS: false, ResultGlue: false, ResultAccess: false}
	if reflect.DeepEqual(e, r) {
		t.Log("Extra NS Record works")
	} else {
		fmt.Printf("Was      : %+v\n", r)
		fmt.Printf("Should Be: %+v\n", e)
		fmt.Printf("Collection: %+v\n", coll)
		t.Error("Full valid collection should return all True\n")
	}
}

// TestRulesAgainstMismatchNSRecord : if we have a different record we should fail
func TestRulesAgainstMismatchNSRecord(t *testing.T) {
	coll := validCollection
	coll.LocalNS = []string{"ns1.example.com", "ns2.example.com"}
	r := rules.Check(&coll, false, false)
	e := rules.Results{ResultA: true, ResultNS: false, ResultGlue: false, ResultAccess: false}
	if reflect.DeepEqual(e, r) {
		t.Log("Missing NS Record works")
	} else {
		fmt.Printf("Was      : %+v\n", r)
		fmt.Printf("Should Be: %+v\n", e)
		fmt.Printf("Collection: %+v\n", coll)
		t.Error("Full valid collection should return all True\n")
	}
}
