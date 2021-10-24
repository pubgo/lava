package config

import (
	"io"
	_ "unsafe"

	"github.com/spf13/viper"
)

//go:linkname unmarshalReader github.com/spf13/viper.(*Viper).unmarshalReader
func unmarshalReader(v *viper.Viper, in io.Reader, c map[string]interface{}) error
