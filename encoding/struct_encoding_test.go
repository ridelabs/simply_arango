package encoding

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestObject struct {
	Id             string `json:"id"`
	OrganizationId string `json:"organization_id"`
	CreatedDate    string `json:"created_date"`
	Name           string `json:"name"`
	Type           string `json:"type"`
	Description    string `json:"description"`
	Sent           int    `json:"sent"`
	Bounces        int    `json:"bounces"`
	Complaints     int    `json:"complaints"`
	Opened         int    `json:"opened"`
}

func TestMapToObject(t *testing.T) {
	input := map[string]interface{}{
		"id":              "imanid",
		"organization_id": "orgid",
		"name":            "myname",
		"type":            "mytype",
		"description":     "mydesc",
		"sent":            10,
		"bounces":         20,
		"opened":          30,
		"complaints":      40,
		"created_date":    "2024-01-01",
	}
	testObj := TestObject{}
	err := MapToObject(input, &testObj)
	assert.Nil(t, err)
	assert.Equal(t, "imanid", testObj.Id)
	assert.Equal(t, "orgid", testObj.OrganizationId)
	assert.Equal(t, "myname", testObj.Name)
	assert.Equal(t, "mytype", testObj.Type)
	assert.Equal(t, "mydesc", testObj.Description)
	assert.Equal(t, 10, testObj.Sent)
	assert.Equal(t, 20, testObj.Bounces)
	assert.Equal(t, 30, testObj.Opened)
	assert.Equal(t, 40, testObj.Complaints)
	assert.Equal(t, "2024-01-01", testObj.CreatedDate)
}

func TestObjectToMap(t *testing.T) {
	testObj := TestObject{
		Id:             "1imanid",
		OrganizationId: "1orgid",
		Name:           "1myname",
		Type:           "1mytype",
		Description:    "1mydesc",
		Sent:           110,
		Bounces:        220,
		Opened:         330,
		Complaints:     440,
		CreatedDate:    "2023-01-01",
	}
	m, err := ObjectToMap(&testObj)
	assert.Nil(t, err)
	assert.Equal(t, "1imanid", m["id"])
	assert.Equal(t, "1orgid", m["organization_id"])
	assert.Equal(t, "1myname", m["name"])
	assert.Equal(t, "1mytype", m["type"])
	assert.Equal(t, "1mydesc", m["description"])
	assert.Equal(t, 110, m["sent"])
	assert.Equal(t, 220, m["bounces"])
	assert.Equal(t, 330, m["opened"])
	assert.Equal(t, 440, m["complaints"])
	assert.Equal(t, "2023-01-01", m["created_date"])
}
