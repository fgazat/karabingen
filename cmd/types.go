package cmd

type Parameters struct {
	BasicToIfAloneTimeoutMilliseconds int `json:"basic.to_if_alone_timeout_milliseconds,omitempty"`
}

type Profile struct {
	Name                 string                `json:"name"`
	Selected             bool                  `json:"selected"`
	VirtualHIDKeyboard   *VirtualHIDKeyboard   `json:"virtual_hid_keyboard,omitempty"`
	SimpleModifications  []SimpleModification  `json:"simple_modifications,omitempty"`
	ComplexModifications *ComplexModifications `json:"complex_modifications,omitempty"`
	Devices              []interface{}         `json:"devices,omitempty"`
	Parameters           *Parameters           `json:"parameters,omitempty"`
}

type VirtualHIDKeyboard struct {
	KeyboardTypeV2 string `json:"keyboard_type_v2,omitempty"`
}

type SimpleModification struct {
	From KeyCode   `json:"from"`
	To   []KeyCode `json:"to"`
}

type ComplexModifications struct {
	Rules []Rule `json:"rules"`
}

type Rule struct {
	Description  string        `json:"description"`
	Manipulators []Manipulator `json:"manipulators"`
}

type Manipulator struct {
	Type         string      `json:"type"`
	Description  string      `json:"description,omitempty"`
	From         From        `json:"from"`
	To           []To        `json:"to,omitempty"`
	ToIfAlone    []To        `json:"to_if_alone,omitempty"`
	ToAfterKeyUp []To        `json:"to_after_key_up,omitempty"`
	Conditions   []Condition `json:"conditions,omitempty"`
	Parameters   *Parameters `json:"parameters,omitempty"`
}

type From struct {
	KeyCode        string     `json:"key_code,omitempty"`
	PointingButton string     `json:"pointing_button,omitempty"`
	Modifiers      *Modifiers `json:"modifiers,omitempty"`
}

type To struct {
	KeyCode          string            `json:"key_code,omitempty"`
	Modifiers        []string          `json:"modifiers,omitempty"`
	ShellCommand     string            `json:"shell_command,omitempty"`
	SetVariable      *SetVariable      `json:"set_variable,omitempty"`
	SoftwareFunction *SoftwareFunction `json:"software_function,omitempty"`
}

type KeyCode struct {
	KeyCode string `json:"key_code"`
}

type Modifiers struct {
	Mandatory []string `json:"mandatory,omitempty"`
	Optional  []string `json:"optional,omitempty"`
}

type SetVariable struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

type SoftwareFunction struct {
	OpenApplication *OpenApplication `json:"open_application,omitempty"`
}

type OpenApplication struct {
	FilePath string `json:"file_path"`
}

type Condition struct {
	Type              string   `json:"type"`
	Name              string   `json:"name,omitempty"`
	Value             int      `json:"value"`
	BundleIdentifiers []string `json:"bundle_identifiers,omitempty"`
}

type KarabinerConfig struct {
	Global   Global    `json:"global"`
	Profiles []Profile `json:"profiles"`
}

type Global struct {
	ShowProfileNameInMenuBar bool `json:"show_profile_name_in_menu_bar,omitempty"`
}
