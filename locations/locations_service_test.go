package locations

import (
	"os"
	"testing"

	"github.com/Financial-Times/neo-utils-go/neoutils"
	"github.com/jmcvetta/neoism"
	"github.com/stretchr/testify/assert"
)

const (
	locationUUID         = "12345"
	newLocationUUID      = "123456"
	tmeID                = "TME_ID"
	newTmeID             = "NEW_TME_ID"
	prefLabel            = "Test"
	specialCharPrefLabel = "Test 'special chars"
)

var defaultTypes = []string{"Thing", "Concept", "Location"}

func TestConnectivityCheck(t *testing.T) {
	assert := assert.New(t)
	locationsDriver := getLocationsCypherDriver(t)
	err := locationsDriver.Check()
	assert.NoError(err, "Unexpected error on connectivity check")
}

func TestPrefLabelIsCorrectlyWritten(t *testing.T) {
	assert := assert.New(t)
	locationsDriver := getLocationsCypherDriver(t)

	alternativeIdentifiers := alternativeIdentifiers{UUIDS: []string{locationUUID}}
	locationToWrite := Location{UUID: locationUUID, PrefLabel: prefLabel, AlternativeIdentifiers: alternativeIdentifiers}

	err := locationsDriver.Write(locationToWrite)
	assert.NoError(err, "ERROR happened during write time")

	storedLocation, found, err := locationsDriver.Read(locationUUID)
	assert.NoError(err, "ERROR happened during read time")
	assert.Equal(true, found)
	assert.NotEmpty(storedLocation)

	assert.Equal(prefLabel, storedLocation.(Location).PrefLabel, "PrefLabel should be "+prefLabel)
	cleanUp(assert, locationUUID, locationsDriver)
}

func TestPrefLabelSpecialCharactersAreHandledByCreate(t *testing.T) {
	assert := assert.New(t)
	locationsDriver := getLocationsCypherDriver(t)

	alternativeIdentifiers := alternativeIdentifiers{TME: []string{}, UUIDS: []string{locationUUID}}
	locationToWrite := Location{UUID: locationUUID, PrefLabel: specialCharPrefLabel, AlternativeIdentifiers: alternativeIdentifiers}

	assert.NoError(locationsDriver.Write(locationToWrite), "Failed to write location")

	//add default types that will be automatically added by the writer
	locationToWrite.Types = defaultTypes
	//check if locationToWrite is the same with the one inside the DB
	readLocationForUUIDAndCheckFieldsMatch(assert, locationsDriver, locationUUID, locationToWrite)
	cleanUp(assert, locationUUID, locationsDriver)
}

func TestCreateCompleteLocationWithPropsAndIdentifiers(t *testing.T) {
	assert := assert.New(t)
	locationsDriver := getLocationsCypherDriver(t)

	alternativeIdentifiers := alternativeIdentifiers{TME: []string{tmeID}, UUIDS: []string{locationUUID}}
	locationToWrite := Location{UUID: locationUUID, PrefLabel: prefLabel, AlternativeIdentifiers: alternativeIdentifiers}

	assert.NoError(locationsDriver.Write(locationToWrite), "Failed to write location")

	//add default types that will be automatically added by the writer
	locationToWrite.Types = defaultTypes
	//check if locationToWrite is the same with the one inside the DB
	readLocationForUUIDAndCheckFieldsMatch(assert, locationsDriver, locationUUID, locationToWrite)
	cleanUp(assert, locationUUID, locationsDriver)
}

func TestUpdateWillRemovePropertiesAndIdentifiersNoLongerPresent(t *testing.T) {
	assert := assert.New(t)
	locationsDriver := getLocationsCypherDriver(t)

	allAlternativeIdentifiers := alternativeIdentifiers{TME: []string{}, UUIDS: []string{locationUUID}}
	locationToWrite := Location{UUID: locationUUID, PrefLabel: prefLabel, AlternativeIdentifiers: allAlternativeIdentifiers}

	assert.NoError(locationsDriver.Write(locationToWrite), "Failed to write location")
	//add default types that will be automatically added by the writer
	locationToWrite.Types = defaultTypes
	readLocationForUUIDAndCheckFieldsMatch(assert, locationsDriver, locationUUID, locationToWrite)

	tmeAlternativeIdentifiers := alternativeIdentifiers{TME: []string{tmeID}, UUIDS: []string{locationUUID}}
	updatedLocation := Location{UUID: locationUUID, PrefLabel: specialCharPrefLabel, AlternativeIdentifiers: tmeAlternativeIdentifiers}

	assert.NoError(locationsDriver.Write(updatedLocation), "Failed to write updated location")
	//add default types that will be automatically added by the writer
	updatedLocation.Types = defaultTypes
	readLocationForUUIDAndCheckFieldsMatch(assert, locationsDriver, locationUUID, updatedLocation)

	cleanUp(assert, locationUUID, locationsDriver)
}

func TestDelete(t *testing.T) {
	assert := assert.New(t)
	locationsDriver := getLocationsCypherDriver(t)

	alternativeIdentifiers := alternativeIdentifiers{TME: []string{tmeID}, UUIDS: []string{locationUUID}}
	locationToDelete := Location{UUID: locationUUID, PrefLabel: prefLabel, AlternativeIdentifiers: alternativeIdentifiers}

	assert.NoError(locationsDriver.Write(locationToDelete), "Failed to write location")

	found, err := locationsDriver.Delete(locationUUID)
	assert.True(found, "Didn't manage to delete location for uuid %", locationUUID)
	assert.NoError(err, "Error deleting location for uuid %s", locationUUID)

	p, found, err := locationsDriver.Read(locationUUID)

	assert.Equal(Location{}, p, "Found location %s who should have been deleted", p)
	assert.False(found, "Found location for uuid %s who should have been deleted", locationUUID)
	assert.NoError(err, "Error trying to find location for uuid %s", locationUUID)
}

func TestCount(t *testing.T) {
	assert := assert.New(t)
	locationsDriver := getLocationsCypherDriver(t)

	alternativeIds := alternativeIdentifiers{TME: []string{tmeID}, UUIDS: []string{locationUUID}}
	locationOneToCount := Location{UUID: locationUUID, PrefLabel: prefLabel, AlternativeIdentifiers: alternativeIds}

	assert.NoError(locationsDriver.Write(locationOneToCount), "Failed to write location")

	nr, err := locationsDriver.Count()
	assert.Equal(1, nr, "Should be 1 locations in DB - count differs")
	assert.NoError(err, "An unexpected error occurred during count")

	newAlternativeIds := alternativeIdentifiers{TME: []string{newTmeID}, UUIDS: []string{newLocationUUID}}
	locationTwoToCount := Location{UUID: newLocationUUID, PrefLabel: specialCharPrefLabel, AlternativeIdentifiers: newAlternativeIds}

	assert.NoError(locationsDriver.Write(locationTwoToCount), "Failed to write location")

	nr, err = locationsDriver.Count()
	assert.Equal(2, nr, "Should be 2 locations in DB - count differs")
	assert.NoError(err, "An unexpected error occurred during count")

	cleanUp(assert, locationUUID, locationsDriver)
	cleanUp(assert, newLocationUUID, locationsDriver)
}

func readLocationForUUIDAndCheckFieldsMatch(assert *assert.Assertions, locationsDriver service, uuid string, expectedLocation Location) {

	storedLocation, found, err := locationsDriver.Read(uuid)

	assert.NoError(err, "Error finding location for uuid %s", uuid)
	assert.True(found, "Didn't find location for uuid %s", uuid)
	assert.Equal(expectedLocation, storedLocation, "locations should be the same")
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

func cleanUp(assert *assert.Assertions, uuid string, locationsDriver service) {
	found, err := locationsDriver.Delete(uuid)
	assert.True(found, "Didn't manage to delete location for uuid %", uuid)
	assert.NoError(err, "Error deleting location for uuid %s", uuid)
}
