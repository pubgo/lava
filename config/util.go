package config

import (
	"github.com/spf13/viper"

	"io"
	_ "unsafe"
)

//go:linkname unmarshalReader github.com/spf13/viper.(*Viper).unmarshalReader
func unmarshalReader(v *viper.Viper, in io.Reader, c map[string]interface{}) error
