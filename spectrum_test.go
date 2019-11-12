package spectrum

import "testing"

const un = "xxxxx"
const pw = "xxxxx"
const OCBaseURL = "https://xxxxxxxx"

var testDevices = []string{"xxxxxx"}
var testAttributes = []string{"0x12c02"}
var testNewAttributes = map[string]string{"0x12c02": "test"}

func TestFindModelHandle(t *testing.T) {
	connection, err := NewConnection(OCBaseURL, un, pw)
	if err != nil {
		t.Error(err)
	}
	_, err = connection.NewModels(testDevices, "has-substring-ignore-case")
	if err != nil {
		t.Error(err)
	}
}

func TestGetModelAttributes(t *testing.T) {
	connection, err := NewConnection(OCBaseURL, un, pw)
	if err != nil {
		t.Error(err)
	}
	models, err := connection.NewModels(testDevices, "has-substring-ignore-case")
	if err != nil {
		t.Error(err)
	}
	_, err = models.GetModelAttributesByModel(connection, testAttributes)
	if err != nil {
		t.Error(err)
	}
}

func TestSetModelAttributes(t *testing.T) {
	connection, err := NewConnection(OCBaseURL, un, pw)
	if err != nil {
		t.Error(err)
	}
	models, err := connection.NewModels(testDevices, "has-substring-ignore-case")
	if err != nil {
		t.Error(err)
	}
	_, err = models.SetModelAttributes(connection, testNewAttributes)
	if err != nil {
		t.Error(err)
	}
}
