package healthbox

type CurrentData struct {
	DeviceType     string `json:"device_type"`
	Description    string `json:"description"`
	Serial         string `json:"serial"`
	WarrantyNumber string `json:"warranty_number"`
	Global         struct {
		Parameter struct {
			DeviceName struct {
				Unit  string `json:"unit"`
				Value string `json:"value"`
			} `json:"device name"`
			LegislationCountry struct {
				Unit  string `json:"unit"`
				Value string `json:"value"`
			} `json:"legislation country"`
			Warranty struct {
				Unit  string `json:"unit"`
				Value string `json:"value"`
			} `json:"warranty"`
		} `json:"parameter"`
	} `json:"global"`
	Room   []RoomInfo   `json:"room"`
	Sensor []SensorInfo `json:"sensor"`
}

type RoomInfo struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Parameter struct {
		DoorsOpen struct {
			Unit  string `json:"unit"`
			Value bool   `json:"value"`
		} `json:"doors_open,omitempty"`
		DoorsPresent struct {
			Unit  string `json:"unit"`
			Value bool   `json:"value"`
		} `json:"doors_present,omitempty"`
		Icon struct {
			Unit  string `json:"unit"`
			Value string `json:"value"`
		} `json:"icon"`
		MeasuredPower struct {
			Unit  string  `json:"unit"`
			Value float64 `json:"value"`
		} `json:"measured_power,omitempty"`
		MeasuredVoltage struct {
			Unit  string  `json:"unit"`
			Value float64 `json:"value"`
		} `json:"measured_voltage,omitempty"`
		Measurement struct {
			Unit  string  `json:"unit"`
			Value float64 `json:"value"`
		} `json:"measurement"`
		Offset struct {
			Unit  string  `json:"unit"`
			Value float64 `json:"value"`
		} `json:"offset"`
		Subzone struct {
			Unit  string `json:"unit"`
			Value string `json:"value"`
		} `json:"subzone"`
		Valve struct {
			Unit  string `json:"unit"`
			Value string `json:"value"`
		} `json:"valve"`
		ValveWarranty struct {
			Unit  string `json:"unit"`
			Value string `json:"value"`
		} `json:"valve_warranty"`
		Nominal struct {
			Unit  string  `json:"unit"`
			Value float64 `json:"value"`
		} `json:"nominal"`
		RoomOrderIndex struct {
			Unit  string `json:"unit"`
			Value string `json:"value"`
		} `json:"room_order_index,omitempty"`
	} `json:"parameter"`
	Actuator []Actuator `json:"actuator"`
}

type Actuator struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	BasicId   int    `json:"basic id"`
	Parameter struct {
		FlowRate struct {
			Unit  string  `json:"unit"`
			Value float64 `json:"value"`
		} `json:"flow_rate"`
	} `json:"parameter"`
}

type SensorInfo struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	BasicId   int    `json:"basic id"`
	Parameter struct {
		Index struct {
			Unit  string  `json:"unit"`
			Value float64 `json:"value"`
		} `json:"index"`
		MainPollutant struct {
			Unit  string `json:"unit"`
			Value string `json:"value"`
		} `json:"main_pollutant"`
		Room struct {
			Unit  string `json:"unit"`
			Value string `json:"value"`
		} `json:"room"`
	} `json:"parameter"`
}
