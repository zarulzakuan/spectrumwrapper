# spectrum
--
    import "mygithub.gsk.com/zia13430/spectrum_wrapper"


## Usage

#### type AlarmCreationResults

    type AlarmCreationResults struct {
    	Mn     string `json:"modelname"`
    	Mh     string `json:"modelhandle"`
    	Status string `json:"alarmcreation"`
    	ID     string `json:"alarmid"`
    }


AlarmCreationResults is for alarm creation response

#### type AttributeModResults

    type AttributeModResults struct {
    	Mn     string `json:"modelname"`
    	Mh     string `json:"modelhandle"`
    	Status string `json:"attributemod"`
    }


AttributeModResults is for set attributes response

#### type AttributesList

    type AttributesList struct {
    	Mn         string            `json:"modelname"`
    	Mh         string            `json:"modelhandle"`
    	Attributes map[string]string `json:"attributes"`
    }


AttributesList is for get attributes response

#### type Connection

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


Connection holds connection Spectrum API details

#### func  NewConnection

    func NewConnection(baseURL string, un string, pw string) (*Connection, error)

NewConnection creates a new connection set to Spectrum Restful API (singleton)

#### func (*Connection) NewModels

    func (c *Connection) NewModels(names []string, searchType string) (*Models, error)

NewModels creates object of models that we can use against methods

#### type Event

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


Event holds event details for alarm creation

#### type Models

    type Models struct {
    	ModelNameHandlePair map[string]string
    	ModelTypes          []string
    }


Models holds models to be actioned details

#### func (*Models) CreateAlarm

    func (m *Models) CreateAlarm(c *Connection, event Event) ([]AlarmCreationResults, error)

CreateAlarm submits seperate XML for creating alarm for each model

#### func (*Models) GetModelAttributes

    func (m *Models) GetModelAttributes(c *Connection, attributes []string) ([]AttributesList, error)

GetModelAttributes submits single XML to Spectrum API for requesting multiple
model attributes for multiple models

#### func (*Models) SetModelAttributes

    func (m *Models) SetModelAttributes(c *Connection, attributes map[string]string) ([]AttributeModResults, error)

SetModelAttributes submits seperate XML for modifying multiple model attributes
for each model
