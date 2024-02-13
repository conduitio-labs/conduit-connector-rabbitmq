package rabbitmq

import "encoding/json"

// cfgToMap converts a config struct to a map. This is useful for more type
// safety on tests.
func cfgToMap(cfg any) map[string]string {
	bs, err := json.Marshal(cfg)
	if err != nil {
		panic(err)
	}

	m := map[string]string{}
	err = json.Unmarshal(bs, &m)
	if err != nil {
		panic(err)
	}

	return m
}
