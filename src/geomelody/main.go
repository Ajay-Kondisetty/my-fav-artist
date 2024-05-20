package main

import (
	"log"

	"geomelody/constants"
	"geomelody/routers"

	_ "github.com/beego/beego/v2/core/config/yaml"
	"github.com/beego/beego/v2/server/web"
	"github.com/joho/godotenv"
)

func main() {
	// Generated using http://patorjk.com/software/taag/#p=display&f=Graffiti
	log.Println(`
                                    .__             .___      
   ____   ____  ____   _____   ____ |  |   ____   __| _/__.__.
  / ___\_/ __ \/  _ \ /     \_/ __ \|  |  /  _ \ / __ <   |  |
 / /_/  >  ___(  <_> )  Y Y  \  ___/|  |_(  <_> ) /_/ |\___  |
 \___  / \___  >____/|__|_|  /\___  >____/\____/\____ |/ ____|
/_____/      \/            \/     \/                 \/\/     
	`)
	web.BConfig.Log.AccessLogs = true
	web.Run()
}

func init() {
	// Load app config.
	appConfigFile := "conf/local.app.yaml"
	if err := web.LoadAppConfig("yaml", appConfigFile); err != nil {
		log.Fatal("Error loading app config: ", err)
	} else {
		log.Printf("Loaded app config: %v", appConfigFile)
	}

	// Load env vars and init const from envs
	if err := godotenv.Load("local_env"); err != nil {
		log.Fatal("Error loading env variables: ", err)
	}
	constants.InitConstantsVars()

	// Init routes
	routers.InitRoutes()
}
