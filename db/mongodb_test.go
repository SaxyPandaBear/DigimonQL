package db

import "testing"

// Verify that the struct can be coerced into the interface
func TestConfirmsMongoDBToInterface(t *testing.T) {
	var _ DigimonRepository = &MongoDBRepository{}
	t.SkipNow()
}
