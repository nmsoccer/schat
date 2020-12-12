package table_desc

type ChatConfig struct {
	Name string `json:"name"`
	Value string `json:"value"`
}

type ChatConfigTable struct {
    Count int `json:"count"`
	Res []ChatConfig `json:"res"`
}

type ChatConfigJson struct {
	ConfigTable ChatConfigTable `json:"chat_config_table"` //each sheet
}