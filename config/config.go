package config

import (
	"log"

	"github.com/spf13/viper"
)

var config *viper.Viper

func Init(env string) {
	var err error
	config = viper.New()
	config.SetConfigType("yaml")
	log.Printf("Setting config for \"%s\"", env)
	config.SetConfigName(env)
	config.AddConfigPath("../config/")
	config.AddConfigPath("../../config/")
	config.AddConfigPath("config/")
	config.AddConfigPath(".")
	err = config.ReadInConfig()
	if err != nil {
		log.Fatal("error on parsing configuration file")
	}
}

func GetConfig() *viper.Viper {
	return config
}
