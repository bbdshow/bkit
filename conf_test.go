package bkit

import (
	"testing"
)

type testConfig struct {
	Name    string
	Age     int `null:""`
	Address struct {
		Street       string
		StreetNumber string
	}
	Cards []card
	Key   string `null:""` // sensitive char
}
type card struct {
	Brand string
	Mode  string
}

func TestMarshalToFile(t *testing.T) {
	type args struct {
		config   interface{}
		filename string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "yaml",
			args: args{
				config: testConfig{
					Name: "yaml_test",
					Age:  24,
					Address: struct {
						Street       string
						StreetNumber string
					}{Street: "china", StreetNumber: "great"},
					Cards: []card{{Brand: "BYD", Mode: "SUV"}, {Brand: "ChangCheng", Mode: "Pika"}},
				},
				filename: "./config.yaml",
			},
			wantErr: false,
		},
		{
			name: "toml",
			args: args{
				config: testConfig{
					Name: "toml_test",
					Age:  24,
					Address: struct {
						Street       string
						StreetNumber string
					}{Street: "china", StreetNumber: "great"},
					Cards: []card{{Brand: "BYD", Mode: "SUV"}, {Brand: "ChangCheng", Mode: "Pika"}},
				},
				filename: "./config.toml",
			},
			wantErr: false,
		},
		{
			name: "json",
			args: args{
				config: testConfig{
					Name: "json_test",
					Age:  24,
					Address: struct {
						Street       string
						StreetNumber string
					}{Street: "china", StreetNumber: "great"},
					Cards: []card{{Brand: "BYD", Mode: "SUV"}, {Brand: "ChangCheng", Mode: "Pika"}},
					Key:   "i am key",
				},
				filename: "./config.json",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := MarshalToFile(tt.args.config, tt.args.filename); (err != nil) != tt.wantErr {
				t.Errorf("MarshalToFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReadConfig(t *testing.T) {
	type args struct {
		filename string
		config   interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "yaml",
			args: args{
				config:   &testConfig{},
				filename: "./config.yaml",
			},
			wantErr: false,
		},
		{
			name: "toml",
			args: args{
				config:   &testConfig{},
				filename: "./config.toml",
			},
			wantErr: false,
		},
		{
			name: "json",
			args: args{
				config:   &testConfig{},
				filename: "./config.json",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ReadConfig(tt.args.filename, tt.args.config); (err != nil) != tt.wantErr {
				t.Errorf("ReadConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			cfg, ok := tt.args.config.(*testConfig)
			if !ok {
				t.Errorf("type not testConfig")
			}
			if cfg.Address.Street != "china" {
				t.Errorf("config val unmarshal invalid")
			}
		})
	}
}
