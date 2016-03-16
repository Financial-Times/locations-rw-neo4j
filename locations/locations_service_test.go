package locations

import (
	"os"
	"testing"

	"github.com/Financial-Times/base-ft-rw-app-go/baseftrwapp"
	"github.com/Financial-Times/neo-utils-go/neoutils"
	"github.com/jmcvetta/neoism"
	"github.com/stretchr/testify/assert"
)

var locationsDriver baseftrwapp.Service

func TestDelete(t *testing.T) {
	assert := assert.New(t)
	uuid := "12345"

	locationsDriver = getLocationsCypherDriver(t)

	locationToDelete := Location{UUID: uuid, CanonicalName: "Test", TmeIdentifier: "TME_ID"}

	assert.NoError(locationsDriver.Write(locationToDelete), "Failed to write location")

	found, err := locationsDriver.Delete(uuid)
	assert.True(found, "Didn't manage to delete location for uuid %", uuid)
	assert.NoError(err, "Error deleting location for uuid %s", uuid)

	p, found, err := locationsDriver.Read(uuid)

	assert.Equal(Location{}, p, "Found location %s who should have been deleted", p)
	assert.False(found, "Found location for uuid %s who should have been deleted", uuid)
	assert.NoError(err, "Error trying to find location for uuid %s", uuid)
}

func TestCreateAllValuesPresent(t *testing.T) {
	assert := assert.New(t)
	uuid := "12345"
	locationsDriver = getLocationsCypherDriver(t)

	locationToWrite := Location{UUID: uuid, CanonicalName: "Test", TmeIdentifier: "TME_ID"}

	assert.NoError(locationsDriver.Write(locationToWrite), "Failed to write location")

	readLocationForUUIDAndCheckFieldsMatch(t, uuid, locationToWrite)

	cleanUp(t, uuid)
}

func TestCreateHandlesSpecialCharacters(t *testing.T) {
	assert := assert.New(t)
	uuid := "12345"
	locationsDriver = getLocationsCypherDriver(t)

	locationToWrite := Location{UUID: uuid, CanonicalName: "Test 'special chars", TmeIdentifier: "TME_ID"}

	assert.NoError(locationsDriver.Write(locationToWrite), "Failed to write location")

	readLocationForUUIDAndCheckFieldsMatch(t, uuid, locationToWrite)

	cleanUp(t, uuid)
}

func TestCreateNotAllValuesPresent(t *testing.T) {
	assert := assert.New(t)
	uuid := "12345"
	locationsDriver = getLocationsCypherDriver(t)

	locationToWrite := Location{UUID: uuid, CanonicalName: "Test"}

	assert.NoError(locationsDriver.Write(locationToWrite), "Failed to write location")

	readLocationForUUIDAndCheckFieldsMatch(t, uuid, locationToWrite)

	cleanUp(t, uuid)
}

func TestUpdateWillRemovePropertiesNoLongerPresent(t *testing.T) {
	assert := assert.New(t)
	uuid := "12345"
	locationsDriver = getLocationsCypherDriver(t)

	locationToWrite := Location{UUID: uuid, CanonicalName: "Test", TmeIdentifier: "TME_ID"}

	assert.NoError(locationsDriver.Write(locationToWrite), "Failed to write location")
	readLocationForUUIDAndCheckFieldsMatch(t, uuid, locationToWrite)

	updatedLocation := Location{UUID: uuid, CanonicalName: "Test", TmeIdentifier: "TME_ID"}

	assert.NoError(locationsDriver.Write(updatedLocation), "Failed to write updated location")
	readLocationForUUIDAndCheckFieldsMatch(t, uuid, updatedLocation)

	cleanUp(t, uuid)
}

func TestConnectivityCheck(t *testing.T) {
	assert := assert.New(t)
	locationsDriver = getLocationsCypherDriver(t)
	err := locationsDriver.Check()
	assert.NoError(err, "Unexpected error on connectivity check")
}

func getLocationsCypherDriver(t *testing.T) service {
	assert := assert.New(t)
	url := os.Getenv("NEO4J_TEST_URL")
	if url == "" {
		url = "http://localhost:7474/db/data"
	}

	db, err := neoism.Connect(url)
	assert.NoError(err, "Failed to connect to Neo4j")
	return NewCypherLocationsService(neoutils.StringerDb{db}, db)
}

func readLocationForUUIDAndCheckFieldsMatch(t *testing.T, uuid string, expectedLocation Location) {
	assert := assert.New(t)
	storedLocation, found, err := locationsDriver.Read(uuid)

	assert.NoError(err, "Error finding location for uuid %s", uuid)
	assert.True(found, "Didn't find location for uuid %s", uuid)
	assert.Equal(expectedLocation, storedLocation, "locations should be the same")
}

func TestWritePrefLabelIsAlsoWrittenAndIsEqualToName(t *testing.T) {
	assert := assert.New(t)
	locationsDriver := getLocationsCypherDriver(t)
	uuid := "12345"
	locationToWrite := Location{UUID: uuid, CanonicalName: "Test", TmeIdentifier: "TME_ID"}

	assert.NoError(locationsDriver.Write(locationToWrite), "Failed to write location")

	result := []struct {
		PrefLabel string `json:"t.prefLabel"`
	}{}

	getPrefLabelQuery := &neoism.CypherQuery{
		Statement: `
				MATCH (t:Location {uuid:"12345"}) RETURN t.prefLabel
				`,
		Result: &result,
	}

	err := locationsDriver.cypherRunner.CypherBatch([]*neoism.CypherQuery{getPrefLabelQuery})
	assert.NoError(err)
	assert.Equal("Test", result[0].PrefLabel, "PrefLabel should be 'Test")
	cleanUp(t, uuid)
}

func cleanUp(t *testing.T, uuid string) {
	assert := assert.New(t)
	found, err := locationsDriver.Delete(uuid)
	assert.True(found, "Didn't manage to delete location for uuid %", uuid)
	assert.NoError(err, "Error deleting location for uuid %s", uuid)
}
