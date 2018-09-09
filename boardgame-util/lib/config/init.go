package config

import (
	"path/filepath"
)

var defaultFirebase = &FirebaseConfig{
	ApiKey:            "AIzaSyDi0hhBgLPbpJgprVCDzDkk8zuFpb9XadM",
	AuthDomain:        "example-boardgame.firebaseapp.com",
	DatabaseURL:       "https://example-boardgame.firebaseio.com",
	ProjectID:         "example-boardgame",
	StorageBucket:     "example-boardgame.appspot.com",
	MessagingSenderID: "138149526364",
}

func defaultGamesNode() *GameNode {

	baseDir := "github.com/jkomoros/boardgame/example"

	games := []string{
		"blackjack",
		"checkers",
		"debuganimations",
		"memory",
		"pig",
		"tictactoe",
	}

	var fullList []string

	for _, game := range games {
		fullList = append(fullList, filepath.Join(baseDir, game))
	}

	return NewGameNode(fullList...)
}

//DefaultStarterConfig returns the default starting config, which is most appropriate
//starting point.
func DefaultStarterConfig() *Config {

	base := &RawConfigMode{
		ConfigModeCommon: ConfigModeCommon{
			DefaultStorageType: "mysql",
			Firebase:           defaultFirebase.copy(),
		},
		Games: defaultGamesNode(),
	}

	dev := &RawConfigMode{
		ConfigModeCommon: ConfigModeCommon{
			AllowedOrigins:       "http://localhost:8080",
			DisableAdminChecking: true,
			Storage: map[string]string{
				"mysql": "root:root@tcp(localhost:3306)/boardgame",
			},
		},
	}

	return starterConfig(base, dev, nil)
}

//SampleStaterConfig returns a full sample starting config, which is also
//provided in `boardgame/config.SAMPLE.json`.
func SampleStarterConfig() *Config {

	base := &RawConfigMode{
		ConfigModeCommon: ConfigModeCommon{
			DefaultStorageType: "bolt",
			GoogleAnalytics:    "UA-321674-10",
			Firebase:           defaultFirebase.copy(),
		},
		Games: defaultGamesNode(),
	}

	dev := &RawConfigMode{
		ConfigModeCommon: ConfigModeCommon{
			AllowedOrigins:       "http://localhost:8080",
			DisableAdminChecking: true,
			Storage: map[string]string{
				"mysql": "root:root@tcp(localhost:3306)/boardgame",
			},
		},
	}

	prod := &RawConfigMode{
		ConfigModeCommon: ConfigModeCommon{
			AllowedOrigins: "https://www.mygame.com",
			AdminUserIds: []string{
				"aH1TV07K47RC4mTNCai0ZPnQ9Kd2",
				"uYuZl1jXWXVJ9fEk7mDFifhTGmK2",
			},
			Storage: map[string]string{
				"mysql": "Your production server config goes here, See https://github.com/go-sql-driver/mysql for examples",
			},
		},
	}

	return starterConfig(base, dev, prod)

}

//MinimalStaterConfig returns a minimal config starter point, with minimal
//settings you might want to set.
func MinimialStaterConfig() *Config {
	base := &RawConfigMode{
		ConfigModeCommon: ConfigModeCommon{
			DefaultStorageType: "mysql",
		},
		Games: defaultGamesNode(),
	}

	dev := &RawConfigMode{
		ConfigModeCommon: ConfigModeCommon{
			AllowedOrigins:       "http://localhost:8080",
			DisableAdminChecking: true,
			Storage: map[string]string{
				"mysql": "root:root@tcp(localhost:3306)/boardgame",
			},
		},
	}

	return starterConfig(base, dev, nil)
}

func starterConfig(base, dev, prod *RawConfigMode) *Config {
	publicFile, _, err := DefaultFileNames("")

	if err != nil {
		return nil
	}

	raw := &RawConfig{
		Base: base,
		Dev:  dev,
		Prod: prod,
		path: publicFile,
	}

	return NewConfig(raw, nil)
}
