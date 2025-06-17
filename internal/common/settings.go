package common

import (
	"fmt"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

var (
	config_instance *viper.Viper
	config_once     sync.Once
)

// returns configuration singleton instance
func GetConfig() *viper.Viper {
	config_once.Do(func() {
		config_instance, _ = initConfig() // TODO handle error
	})
	return config_instance
}

func initConfig() (*viper.Viper, error) {
	viper_instance := viper.New()
	viper_instance.SetConfigName("config")
	viper_instance.SetConfigType("yaml")
	viper_instance.AddConfigPath("/etc/pipo-dispatcher")
	viper_instance.SetEnvPrefix("pipo")                             // uppercased automatically
	viper_instance.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) // e.g. want to use . in Get() calls, but environmental variables to use _ delimiters (e.g. app.port -> APP_PORT)
	viper_instance.AutomaticEnv()

	// Read the config file
	err := viper_instance.ReadInConfig()
	if err != nil { // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	return viper_instance, err
}
