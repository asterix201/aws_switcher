package config

import (
	"encoding/json"
	"testing"
)

func BenchmarkConfigMarshal(b *testing.B) {
	cfg := &Config{
		StartURL:  "https://test.awsapps.com/start/",
		SSORegion: "us-west-2",
		Accounts: []Account{
			{ID: "111111111111", Name: "Account 1", RoleName: "Role1"},
			{ID: "222222222222", Name: "Account 2", RoleName: "Role2"},
			{ID: "333333333333", Name: "Account 3", RoleName: "Role3"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(cfg)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkConfigUnmarshal(b *testing.B) {
	cfg := &Config{
		StartURL:  "https://test.awsapps.com/start/",
		SSORegion: "us-west-2",
		Accounts: []Account{
			{ID: "111111111111", Name: "Account 1", RoleName: "Role1"},
			{ID: "222222222222", Name: "Account 2", RoleName: "Role2"},
			{ID: "333333333333", Name: "Account 3", RoleName: "Role3"},
		},
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var loadedConfig Config
		err := json.Unmarshal(data, &loadedConfig)
		if err != nil {
			b.Fatal(err)
		}
	}
}
