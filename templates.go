package spectrum

const findModelHandleTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<rs:model-request throttlesize="10000"
    xmlns:rs="http://www.ca.com/spectrum/restful/schema/request"
    xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.ca.com/spectrum/restful/schema/request../../../xsd/Request.xsd ">
    <rs:target-models>
        <rs:models-search>
            <rs:search-criteria
                xmlns="http://www.ca.com/spectrum/restful/schema/filter">
                <filtered-models>
                    <and>
                        <or>
                            {{range .Modeltypes}}
                            <equals>
                                <attribute id="0x10001">
                                    <value>{{.}}</value>
                                </attribute>
                            </equals>
                            {{end}}
                        </or>
                        <or>
                            {{range .Modelnames}}
                            <{{$.Searchtype}}>
                                <attribute id="0x1006e">
                                    <value>{{.}}</value>
                                </attribute>
                            </{{$.Searchtype}}>
                            {{end}}
                        </or>
                    </and>
                </filtered-models>
            </rs:search-criteria>
        </rs:models-search>
    </rs:target-models>
<rs:requested-attribute id="0x1006e" />
</rs:model-request>`

const getModelAttributesTemplate = 
`<?xml version="1.0" encoding="UTF-8"?>
<rs:model-request throttlesize="10000"
    xmlns:rs="http://www.ca.com/spectrum/restful/schema/request"
    xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.ca.com/spectrum/restful/schema/request../../../xsd/Request.xsd ">
    <rs:landscape id="0x2000000" />
    <rs:target-models>
        <rs:models-search>
            <rs:search-criteria
                xmlns="http://www.ca.com/spectrum/restful/schema/filter">
                <filtered-models>
                    <and>
                        <or>
                            {{ range $name, $handle := .NameHandlePair }}
                            <equals>
                                <attribute id="0x129fa">
                                    <value>{{$handle}}</value>
                                </attribute>
                            </equals>
                            {{end}}
                        </or>
                    </and>
                </filtered-models>
            </rs:search-criteria>
        </rs:models-search>
    </rs:target-models>
	<rs:requested-attribute id="0x1006e" />
    {{range .Attrs}}
    <rs:requested-attribute id="{{.}}" />
    {{end}}
</rs:model-request>`

const setModelAttributesTemplate = 
`<?xml version="1.0" encoding="UTF-8"?>
<rs:update-models-request throttlesize="10000"
    xmlns:rs="http://www.ca.com/spectrum/restful/schema/request"
    xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
		xsi:schemaLocation="http://www.ca.com/spectrum/restful/schema/request ../../../xsd/Request.xsd ">
    <rs:target-models>
        <rs:model mh="{{.ModelHandle}}" />
    </rs:target-models>
    
    {{ range $key, $value := .Attrs }}
    <rs:attribute-value id="{{$key}}">{{$value}}</rs:attribute-value>
    {{end}}
		
</rs:update-models-request>`

const createAlarmTemplate = 
`<?xml version="1.0" encoding="UTF-8"?>
<rs:event-request
			throttlesize="10000"
    xmlns:rs="http://www.ca.com/spectrum/restful/schema/request"
    xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
			xsi:schemaLocation="http://www.ca.com/spectrum/restful/schema/request ../../../xsd/Request.xsd">
    <rs:event>
        <rs:target-models>
            <rs:model mh="{{.ModelHandle}}"/>
        </rs:target-models>
        <rs:event-type id="{{.CauseCode}}"/>
        <rs:varbind id="76620">{{.Title}}</rs:varbind>
        <rs:varbind id="1">{{.ModelName}}</rs:varbind>
        <rs:varbind id="103">{{.Severity}}</rs:varbind>
        <rs:varbind id="101">{{.Datetime}}</rs:varbind>
        <rs:varbind id="102">{{.Submitter}}</rs:varbind>
        <rs:varbind id="104">{{.Desc}}</rs:varbind>
        <rs:varbind id="105">{{.Ticket}}</rs:varbind>
        <rs:varbind id="106">{{.Pid}}</rs:varbind>
        <rs:varbind id="107">SPECMAC</rs:varbind>
    </rs:event>
</rs:event-request>`