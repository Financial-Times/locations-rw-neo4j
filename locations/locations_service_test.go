package locations

import (
	"os"
	"testing"

	"github.com/Financial-Times/neo-utils-go/neoutils"
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
	locationsDriver := getLocationsCypherDriver(t)
	err := locationsDriver.Check()
	assert.NoError(t, err, "Unexpected error on connectivity check")
}

func TestPrefLabelIsCorrectlyWritten(t *testing.T) {
	locationsDriver := getLocationsCypherDriver(t)

	alternativeIdentifiers := alternativeIdentifiers{UUIDS: []string{locationUUID}}
	locationToWrite := Location{UUID: locationUUID, PrefLabel: prefLabel, AlternativeIdentifiers: alternativeIdentifiers}

	err := locationsDriver.Write(locationToWrite)
	assert.NoError(t, err, "ERROR happened during write time")

	storedLocation, found, err := locationsDriver.Read(locationUUID)
	assert.NoError(t, err, "ERROR happened during read time")
	assert.Equal(t, true, found)
	assert.NotEmpty(t, storedLocation)

	assert.Equal(t, prefLabel, storedLocation.(Location).PrefLabel, "PrefLabel should be "+prefLabel)
	cleanUp(t, locationUUID, locationsDriver)
}

func TestPrefLabelSpecialCharactersAreHandledByCreate(t *testing.T) {
	locationsDriver := getLocationsCypherDriver(t)

	alternativeIdentifiers := alternativeIdentifiers{TME: []string{}, UUIDS: []string{locationUUID}}
	locationToWrite := Location{UUID: locationUUID, PrefLabel: specialCharPrefLabel, AlternativeIdentifiers: alternativeIdentifiers}

	assert.NoError(t, locationsDriver.Write(locationToWrite), "Failed to write location")

	//add default types that will be automatically added by the writer
	locationToWrite.Types = defaultTypes
	//check if locationToWrite is the same with the one inside the DB
	readLocationForUUIDAndCheckFieldsMatch(t, locationsDriver, locationUUID, locationToWrite)
	cleanUp(t, locationUUID, locationsDriver)
}

func TestCreateCompleteLocationWithPropsAndIdentifiers(t *testing.T) {
	locationsDriver := getLocationsCypherDriver(t)

	alternativeIdentifiers := alternativeIdentifiers{TME: []string{tmeID}, UUIDS: []string{locationUUID}}
	locationToWrite := Location{UUID: locationUUID, PrefLabel: prefLabel, AlternativeIdentifiers: alternativeIdentifiers}

	assert.NoError(t, locationsDriver.Write(locationToWrite), "Failed to write location")

	//add default types that will be automatically added by the writer
	locationToWrite.Types = defaultTypes
	//check if locationToWrite is the same with the one inside the DB
	readLocationForUUIDAndCheckFieldsMatch(t, locationsDriver, locationUUID, locationToWrite)
	cleanUp(t, locationUUID, locationsDriver)
}

func TestUpdateWillRemovePropertiesAndIdentifiersNoLongerPresent(t *testing.T) {
	locationsDriver := getLocationsCypherDriver(t)

	allAlternativeIdentifiers := alternativeIdentifiers{TME: []string{}, UUIDS: []string{locationUUID}}
	locationToWrite := Location{UUID: locationUUID, PrefLabel: prefLabel, AlternativeIdentifiers: allAlternativeIdentifiers}

	assert.NoError(t, locationsDriver.Write(locationToWrite), "Failed to write location")
	//add default types that will be automatically added by the writer
	locationToWrite.Types = defaultTypes
	readLocationForUUIDAndCheckFieldsMatch(t, locationsDriver, locationUUID, locationToWrite)

	tmeAlternativeIdentifiers := alternativeIdentifiers{TME: []string{tmeID}, UUIDS: []string{locationUUID}}
	updatedLocation := Location{UUID: locationUUID, PrefLabel: specialCharPrefLabel, AlternativeIdentifiers: tmeAlternativeIdentifiers}

	assert.NoError(t, locationsDriver.Write(updatedLocation), "Failed to write updated location")
	//add default types that will be automatically added by the writer
	updatedLocation.Types = defaultTypes
	readLocationForUUIDAndCheckFieldsMatch(t, locationsDriver, locationUUID, updatedLocation)

	cleanUp(t, locationUUID, locationsDriver)
}

func TestDelete(t *testing.T) {
	locationsDriver := getLocationsCypherDriver(t)

	alternativeIdentifiers := alternativeIdentifiers{TME: []string{tmeID}, UUIDS: []string{locationUUID}}
	locationToDelete := Location{UUID: locationUUID, PrefLabel: prefLabel, AlternativeIdentifiers: alternativeIdentifiers}

	assert.NoError(t, locationsDriver.Write(locationToDelete), "Failed to write location")

	found, err := locationsDriver.Delete(locationUUID)
	assert.True(t, found, "Didn't manage to delete location for uuid %", locationUUID)
	assert.NoError(t, err, "Error deleting location for uuid %s", locationUUID)

	p, found, err := locationsDriver.Read(locationUUID)

	assert.Equal(t, Location{}, p, "Found location %s who should have been deleted", p)
	assert.False(t, found, "Found location for uuid %s who should have been deleted", locationUUID)
	assert.NoError(t, err, "Error trying to find location for uuid %s", locationUUID)
}

func TestCount(t *testing.T) {

	locationsDriver := getLocationsCypherDriver(t)

	alternativeIds := alternativeIdentifiers{TME: []string{tmeID}, UUIDS: []string{locationUUID}}
	locationOneToCount := Location{UUID: locationUUID, PrefLabel: prefLabel, AlternativeIdentifiers: alternativeIds}

	assert.NoError(t, locationsDriver.Write(locationOneToCount), "Failed to write location")

	nr, err := locationsDriver.Count()
	assert.Equal(t, 1, nr, "Should be 1 locations in DB - count differs")
	assert.NoError(t, err, "An unexpected error occurred during count")

	newAlternativeIds := alternativeIdentifiers{TME: []string{newTmeID}, UUIDS: []string{newLocationUUID}}
	locationTwoToCount := Location{UUID: newLocationUUID, PrefLabel: specialCharPrefLabel, AlternativeIdentifiers: newAlternativeIds}

	assert.NoError(t, locationsDriver.Write(locationTwoToCount), "Failed to write location")

	nr, err = locationsDriver.Count()
	assert.Equal(t, 2, nr, "Should be 2 locations in DB - count differs")
	assert.NoError(t, err, "An unexpected error occurred during count")

	cleanUp(t, locationUUID, locationsDriver)
	cleanUp(t, newLocationUUID, locationsDriver)
}

func readLocationForUUIDAndCheckFieldsMatch(t *testing.T, locationsDriver service, uuid string, expectedLocation Location) {

	storedLocation, found, err := locationsDriver.Read(uuid)

	assert.NoError(t, err, "Error finding location for uuid %s", uuid)
	assert.True(t, found, "Didn't find location for uuid %s", uuid)
	assert.Equal(t, expectedLocation, storedLocation, "locations should be the same")
}

func getLocationsCypherDriver(t *testing.T) service {
	url := os.Getenv("NEO4J_TEST_URL")
	if url == "" {
		url = "http://localhost:7474/db/data"
	}

	conf := neoutils.DefaultConnectionConfig()
	conf.Transactional = false
	db, err := neoutils.Connect(url, conf)
	assert.NoError(t, err, "Failed to connect to Neo4j")
	service := NewCypherLocationsService(db)
	service.Initialise()
	return service
}

func cleanUp(t *testing.T, uuid string, locationsDriver service) {
	found, err := locationsDriver.Delete(uuid)
	assert.True(t, found, "Didn't manage to delete location for uuid %", uuid)
	assert.NoError(t, err, "Error deleting location for uuid %s", uuid)
}
