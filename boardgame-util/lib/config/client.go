package config

//ClientConfig is the struct representation of a client config that the client
//web application expects. It can be generated from a config object via
//Client(). Its json representation is what the client web app expects.
type ClientConfig struct {
	Firebase        *FirebaseConfig `json:"firebase"`
	GoogleAnalytics string          `json:"google_analytics"`
	Host            string          `json:"host"`
	DevHost         string          `json:"dev_host"`
	//This property will be set if DisableAuthChecking is true in config
	OfflineDevMode bool `json:"offline_dev_mode,omitempty"`
}

//Client returns a ClientConfig derived from the given config. The returned
//ClientConfig is reasonable to marshal to json and encode for the client web
//app. Will use the firebase and google analytics block from dev mode unless
//prodMode is true. In practice those blocks rarely differ in dev or prod mode
//so that parameter shouldn't matter.
func (c *Config) Client(prodMode bool) *ClientConfig {

	var host string
	var devHost string

	if c.Dev != nil {
		devHost = c.Dev.ApiHost
	}

	if c.Prod != nil {
		host = c.Prod.ApiHost
	}

	mode := c.Dev

	if prodMode || mode == nil {
		mode = c.Prod
	}

	//Neither prod nor dev were defined, apparently
	if mode == nil {
		return nil
	}

	return &ClientConfig{
		Firebase:        mode.Firebase,
		GoogleAnalytics: mode.GoogleAnalytics,
		Host:            host,
		DevHost:         devHost,
		OfflineDevMode:  mode.OfflineDevMode,
	}
}
