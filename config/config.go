package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"go-winx-api/internal/utils"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
)

var ValueOf = &config{}

type config struct {
	ApiId          int      `envconfig:"API_ID" required:"true"`
	ApiHash        string   `envconfig:"API_HASH" required:"true"`
	BotToken       string   `envconfig:"BOT_TOKEN" required:"true"`
	ChannelId      int64    `envconfig:"CHANNEL_ID" required:"true"`
	Port           int      `envconfig:"PORT" default:"8080"`
	Host           string   `envconfig:"HOST" default:""`
	HashLength     int      `envconfig:"HASH_LENGTH" default:"6"`
	UserSession    string   `envconfig:"USER_SESSION"`
	UsePublicIP    bool     `envconfig:"USE_PUBLIC_IP" default:"false"`
	StringSessions []string `envconfig:"STRING_SESSIONS"`
}

func (c *config) loadFromEnvFile(log *zap.Logger) {
	envPath := filepath.Clean(".env")
	log.Sugar().Infof("trying to load ENV vars from %s", envPath)
	err := godotenv.Load(envPath)

	c.StringSessions = strings.Split(os.Getenv("STRING_SESSIONS"), ",")

	if err != nil {
		if os.IsNotExist(err) {
			log.Sugar().Errorf("ENV file not found: %s", envPath)
			log.Sugar().Info("Please create fsb.env file")
			log.Sugar().Info("Please ignore this message if you are hosting it in a service like Heroku or other alternatives.")
		} else {
			log.Fatal("Unknown error while parsing env file.", zap.Error(err))
		}
	}
}

func (c *config) setupEnvVars(log *zap.Logger) {
	c.loadFromEnvFile(log)

	err := envconfig.Process("", c)
	if err != nil {
		log.Fatal("error while parsing env variables", zap.Error(err))
	}
	var ipBlocked bool
	ip, err := utils.GetIP(c.UsePublicIP)
	if err != nil {
		log.Error("error while getting IP", zap.Error(err))
		ipBlocked = true
	}
	if c.Host == "" {
		c.Host = "http://" + ip + ":" + strconv.Itoa(c.Port)
		if c.UsePublicIP {
			if ipBlocked {
				log.Sugar().Warn("can't get public IP, using local IP")
			} else {
				log.Sugar().Warn("you are using a public IP, please be aware of the security risks while exposing your IP to the internet.")
				log.Sugar().Warn("use 'HOST' variable to set a domain name")
			}
		}
		log.Sugar().Info("HOST not set, automatically set to " + c.Host)
	}
}

func Load(log *zap.Logger) {
	log = log.Named("config")
	defer log.Info("loaded config")
	ValueOf.setupEnvVars(log)
	ValueOf.ChannelId = int64(stripInt(log, int(ValueOf.ChannelId)))
	if ValueOf.HashLength == 0 {
		log.Sugar().Info("HASH_LENGTH can't be 0, defaulting to 6")
		ValueOf.HashLength = 6
	}
	if ValueOf.HashLength > 32 {
		log.Sugar().Info("HASH_LENGTH can't be more than 32, changing to 32")
		ValueOf.HashLength = 32
	}
	if ValueOf.HashLength < 5 {
		log.Sugar().Info("HASH_LENGTH can't be less than 5, defaulting to 6")
		ValueOf.HashLength = 6
	}
}

func stripInt(log *zap.Logger, a int) int {
	strA := strconv.Itoa(abs(a))
	lastDigits := strings.Replace(strA, "100", "", 1)
	result, err := strconv.Atoi(lastDigits)
	if err != nil {
		log.Sugar().Fatalln(err)
		return 0
	}
	return result
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
