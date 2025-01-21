package settings

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

func InitConfig() *viper.Viper {
	viper_instance := viper.New()

	viper_instance.SetConfigType("yaml")
	viper_instance.SetConfigName("config")
	viper_instance.AddConfigPath(".")
	viper_instance.AddConfigPath("config")
	viper_instance.AddConfigPath("/etc/pipo-dispatcher")
	viper_instance.AutomaticEnv()
	viper_instance.SetEnvPrefix("pipo")                             // uppercased automatically
	viper_instance.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) // e.g. want to use . in Get() calls, but environmental variables to use _ delimiters (e.g. app.port -> APP_PORT)

	// Read the config file
	err := viper_instance.ReadInConfig()
	if err != nil {
		log.Fatalf("Error reading configuration file, %s", err)
	}
	return viper_instance
}
