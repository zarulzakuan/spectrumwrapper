package spectrum

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"text/template"

	"github.com/tidwall/gjson"
)

// Event holds event details for alarm creation
type Event struct {
	Severity  string `json:"severity"`
	Title     string `json:"title"`
	Desc      string `json:"desc"`
	CauseCode string `json:"causecode"`
	Ticket    string `json:"ticket"`
	Submitter string `json:"submitter"`
	DateTime  string `json:"datetime"`
	PID       string `json:"pid"`
}

// Models holds models to be actioned details
type Models struct {
	ModelNameHandlePair map[string]string
	ModelTypes          []string
}

// Connection holds connection Spectrum API details
type Connection struct {
	OneClickBaseURL  string
	OneClickPort     string
	RestfulAlarmsURL string
	RestfulEventsURL string
	RestfulModelsURL string
	RestfulModelURL  string
	RestfulLandscape string
	Landscapes       []string
	Username         string
	Password         string
}

var c *Connection

// AlarmCreationResults is for alarm creation response
type AlarmCreationResults struct {
	Mn     string `json:"modelname"`
	Mh     string `json:"modelhandle"`
	Status string `json:"alarmcreation"`
	ID     string `json:"alarmid"`
}

// AttributeModResults is for set attributes response
type AttributeModResults struct {
	Mn     string `json:"modelname"`
	Mh     string `json:"modelhandle"`
	Status string `json:"attributemod"`
}

// AttributesList is for get attributes response
type AttributesList struct {
	Mn         string            `json:"modelname"`
	Mh         string            `json:"modelhandle"`
	Attributes map[string]string `json:"attributes"`
}

// NewModels creates object of models that we can use against methods
func (c *Connection) NewModels(names []string, searchType string) (*Models, error) {

	var models Models
	models.ModelTypes = []string{
		"0x6330002",
		"0x6330008",
		"0x6330007",
		"0x4cb0002",
		"0x10290",
		"0x43b0003",
		"0x3d0002",
		"0x1160089",
		"0x430003",
		"0x630007",
		"0x3dc0000",
	}
	result, err := models.findModelHandle(c, names, searchType)
	if err != nil {
		return nil, err
	}
	models.ModelNameHandlePair = result

	return &models, nil
}

// NewConnection creates a new connection set to Spectrum Restful API (singleton)
func NewConnection(baseURL string, un string, pw string) (*Connection, error) {
	if c == nil {
		c = &Connection{
			OneClickBaseURL:  baseURL,
			OneClickPort:     ":8443",
			RestfulModelsURL: "/spectrum/restful/models", // for acquiring model handle (multiple)
			RestfulModelURL:  "/spectrum/restful/model",  // for model creation/deletion, read attr, update attr (single)
			RestfulEventsURL: "/spectrum/restful/events",
			RestfulAlarmsURL: "/spectrum/restful/alarms",
			RestfulLandscape: "/spectrum/restful/landscapes",
			Username:         un,
			Password:         pw,
		}

		//test connection
		myurl := fmt.Sprintf(`%s%s%s`, c.OneClickBaseURL, c.OneClickPort, c.RestfulLandscape)
		jsonByte, err := c.callAPI(myurl, "GET", "")
		if err != nil {
			return nil, err
		}

		bodyString := string(jsonByte)

		modelCount := gjson.Get(bodyString, "landscape-response.@total-landscapes").Int()
		var modelLandscapesStruct []gjson.Result
		if modelCount <= 0 {
			return nil, errors.New("Couldn't find any landscape")
		}

		if modelCount == 1 {
			modelLandscapesStruct = gjson.Get(bodyString, "landscape-response.landscape.id").Array()
		} else {
			modelLandscapesStruct = gjson.Get(bodyString, "landscape-response.landscape.#.id").Array()
		}

		var landscapeHandles []string
		for _, elem := range modelLandscapesStruct {
			landscapeHandles = append(landscapeHandles, elem.Str)
		}

		c.Landscapes = landscapeHandles

	}
	return c, nil
}

//////////////METHODS///////////////////

// findModelHandle submits a single XML for querying model handle and model name for multiple models
func (m *Models) findModelHandle(c *Connection, names []string, searchType string) (map[string]string, error) {
	modelNameHandlePair := make(map[string]string)

	if len(names) == 0 {
		return nil, errors.New("No model handles found for these models")
	}

	var tmpl = template.Must(template.New("findModelHandleTemplate").Parse(findModelHandleTemplate))
	contentPart := struct {
		Modelnames []string
		Modeltypes []string
		Searchtype string
	}{
		names,
		m.ModelTypes,
		searchType,
	}

	myurl := fmt.Sprintf(`%s%s%s`, c.OneClickBaseURL, c.OneClickPort, c.RestfulModelsURL)

	var xmlBodyBytes bytes.Buffer
	if err := tmpl.Execute(&xmlBodyBytes, contentPart); err != nil {
		return nil, err
	}

	xmlBody := xmlBodyBytes.String()

	jsonByte, err := c.callAPI(myurl, "POST", xmlBody)
	if err != nil {
		return nil, err
	}
	modelCount := gjson.GetBytes(jsonByte, "model-response-list.@total-models").Int()

	if modelCount <= 0 {
		return nil, errors.New("Couldn't find any model")
	}
	if modelCount == 1 {
		mh := gjson.GetBytes(jsonByte, "model-response-list.model-responses.model.@mh").String()
		mn := gjson.GetBytes(jsonByte, "model-response-list.model-responses.model.attribute.$").String()
		modelNameHandlePair[mn] = mh
	} else {
		result := gjson.GetBytes(jsonByte, "model-response-list.model-responses.model")
		result.ForEach(func(key, value gjson.Result) bool {
			mh := gjson.Get(value.String(), "@mh").String()
			mn := gjson.Get(value.String(), "attribute.$").String()
			modelNameHandlePair[mn] = mh
			return true // keep iterating
		})
	}

	return modelNameHandlePair, nil

}

// CreateAlarm submits seperate XML for creating alarm for each model
func (m *Models) CreateAlarm(c *Connection, event Event) ([]AlarmCreationResults, error) {

	createStatuses := []AlarmCreationResults{}
	if event.CauseCode == "" {
		return createStatuses, errors.New("CauseCode is required")
	}
	myurl := fmt.Sprintf(`%s%s%s`, c.OneClickBaseURL, c.OneClickPort, c.RestfulEventsURL)
	var tmpl = template.Must(template.New("createAlarmTemplate").Parse(createAlarmTemplate))

	for modelName, modelHandle := range m.ModelNameHandlePair {
		contentPart := struct {
			ModelHandle string
			CauseCode   string
			Title       string
			ModelName   string
			Severity    string
			Datetime    string
			Submitter   string
			Desc        string
			Ticket      string
			Pid         string
		}{
			modelHandle,
			event.CauseCode,
			event.Title,
			modelName,
			event.Severity,
			event.DateTime,
			event.Submitter,
			event.Desc,
			event.Ticket,
			event.PID,
		}

		var xmlBodyBytes bytes.Buffer
		if err := tmpl.Execute(&xmlBodyBytes, contentPart); err != nil {
			return nil, err
		}

		xmlBody := xmlBodyBytes.String()

		jsonByte, err := c.callAPI(myurl, "POST", xmlBody)
		if err != nil {
			return createStatuses, err
		}
		var createStatusStruct []gjson.Result
		var createIDStruct []gjson.Result
		if gjson.GetBytes(jsonByte, "event-response-list.event-response.@error").String() == "" {
			createStatusStruct = gjson.GetBytes(jsonByte, "event-response-list.event-response.#.@error").Array()
			createIDStruct = gjson.GetBytes(jsonByte, "event-response-list.event-response.#.@id").Array()
		} else {
			createStatusStruct = gjson.GetBytes(jsonByte, "event-response-list.event-response.@error").Array()
			createIDStruct = gjson.GetBytes(jsonByte, "event-response-list.event-response.@id").Array()
		}

		for i, elem := range createStatusStruct {
			eachElem := AlarmCreationResults{
				Mn:     modelName,
				Mh:     modelHandle,
				Status: elem.Str,
				ID:     createIDStruct[i].Str,
			}
			createStatuses = append(createStatuses, eachElem)
		}
	}
	return createStatuses, nil
}

// SetModelAttributes submits seperate XML for modifying multiple model attributes for each model
func (m *Models) SetModelAttributes(c *Connection, attributes map[string]string) ([]AttributeModResults, error) {
	attributeModResult := []AttributeModResults{}
	myurl := fmt.Sprintf(`%s%s%s`, c.OneClickBaseURL, c.OneClickPort, c.RestfulModelsURL)

	if len(attributes) == 0 {
		return attributeModResult, errors.New("Attributes are required")
	}

	for modelName, modelHandle := range m.ModelNameHandlePair {

		var tmpl = template.Must(template.New("setModelAttributesTemplate").Parse(setModelAttributesTemplate))

		contentPart := struct {
			ModelHandle string
			Attrs       map[string]string
		}{
			modelHandle,
			attributes,
		}

		var xmlBodyBytes bytes.Buffer
		if err := tmpl.Execute(&xmlBodyBytes, contentPart); err != nil {
			return nil, err
		}

		xmlBody := xmlBodyBytes.String()

		jsonByte, err := c.callAPI(myurl, "PUT", xmlBody)
		if err != nil {
			return attributeModResult, err
		}
		var modifyAttributeStruct []gjson.Result

		if gjson.GetBytes(jsonByte, "model-update-response-list.model-responses.model.@error").String() == "" {
			modifyAttributeStruct = gjson.GetBytes(jsonByte, "model-update-response-list.model-responses.model.#.@error").Array()

		} else {
			modifyAttributeStruct = gjson.GetBytes(jsonByte, "model-update-response-list.model-responses.model.@error").Array()
		}

		for _, elem := range modifyAttributeStruct {
			eachElem := AttributeModResults{
				Mn:     modelName,
				Mh:     modelHandle,
				Status: elem.Str,
			}
			attributeModResult = append(attributeModResult, eachElem)
		}
	}
	return attributeModResult, nil
}

// GetModelAttributes submits single XML to Spectrum API for requesting multiple model attributes for multiple models
func (m *Models) GetModelAttributes(c *Connection, attributes []string) ([]AttributesList, error) {
	attributeGetResult := []AttributesList{}
	var ats = make(map[string]string)
	myurl := fmt.Sprintf(`%s%s%s`, c.OneClickBaseURL, c.OneClickPort, c.RestfulModelsURL)

	if len(attributes) == 0 {
		return attributeGetResult, errors.New("Attributes are required")
	}
	var tmpl = template.Must(template.New("getModelAttributesTemplate").Parse(getModelAttributesTemplate))

	contentPart := struct {
		NameHandlePair map[string]string
		Attrs          []string
	}{
		m.ModelNameHandlePair,
		attributes,
	}

	var xmlBodyBytes bytes.Buffer
	if err := tmpl.Execute(&xmlBodyBytes, contentPart); err != nil {
		return nil, err
	}

	xmlBody := xmlBodyBytes.String()

	jsonByte, err := c.callAPI(myurl, "POST", xmlBody)
	if err != nil {
		return attributeGetResult, err
	}
	//fmt.Printf("SPECTRUM ==> %s", string(jsonByte))
	modelCount := gjson.GetBytes(jsonByte, "model-response-list.@total-models").Int()
	if modelCount > 0 {
		if modelCount == 1 {
			mh := gjson.GetBytes(jsonByte, "model-response-list.model-responses.model.@mh").String()
			mn := gjson.GetBytes(jsonByte, "model-response-list.model-responses.model.attribute.0.$").String()

			gjson.GetBytes(jsonByte, "model-response-list.model-responses.model.attribute").ForEach(func(key, value gjson.Result) bool {
				id := gjson.Get(value.String(), "@id").String()
				val := gjson.Get(value.String(), "$").String()
				ats[id] = val
				return true // keep iterating
			})
			attr := AttributesList{
				Mn:         mn,
				Mh:         mh,
				Attributes: ats,
			}
			attributeGetResult = append(attributeGetResult, attr)
		} else {
			gjson.GetBytes(jsonByte, "model-response-list.model-responses.model").ForEach(func(key, value gjson.Result) bool {
				mh := gjson.Get(value.String(), "@mh").String()
				mn := gjson.Get(value.String(), "attribute.0.$").String()
				ats = make(map[string]string) // reinitiate map

				gjson.Get(value.String(), "attribute").ForEach(func(key, value gjson.Result) bool {
					id := gjson.Get(value.String(), "@id").String()
					val := gjson.Get(value.String(), "$").String()
					ats[id] = val
					return true // keep iterating
				})
				attr := AttributesList{
					Mn:         mn,
					Mh:         mh,
					Attributes: ats,
				}
				attributeGetResult = append(attributeGetResult, attr)
				return true // keep iterating
			})
		}
	} else {
		//return attributeGetResult
	}

	return attributeGetResult, nil
}

// basicAuth gets user provided username and password and decode them to string
func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

// callAPI sends request to Spectrum Restful API with user provided credential
func (c *Connection) callAPI(url, method, xmlBody string) ([]byte, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest(method, url, strings.NewReader(xmlBody))
	req.Header.Add("Authorization", "Basic "+basicAuth(c.Username, c.Password))
	req.Header.Add("Content-Type", "application/xml;charset=UTF-8")
	req.Header.Add("Accept", "application/json")
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusOK {
		// OK
	} else {
		statusCode := strconv.Itoa(resp.StatusCode)
		return nil, errors.New("Couldn't connect: " + statusCode + ": " + string(xmlBody))
	}

	return bodyBytes, nil
}
