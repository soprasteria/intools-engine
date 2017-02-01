package utils

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/gin-gonic/gin"
	"github.com/soprasteria/dockerapi"
	"gopkg.in/redis.v3"
)

var (
	redisOptions *redis.Options
)

func StringTransform(s string) string {
	v := make([]rune, 0, len(s))
	for i, r := range s {
		if r == utf8.RuneError {
			_, size := utf8.DecodeRuneInString(s[i:])
			if size == 1 {
				continue
			}
		}
		//check unicode chars
		if !unicode.IsControl(r) {
			v = append(v, r)
		}
	}
	s = string(v)

	return strings.TrimSpace(s)
}

func ReadLogs(reader io.Reader) (string, error) {
	scanner := bufio.NewScanner(reader)
	var text string
	for scanner.Scan() {
		b := []byte(scanner.Text())
		if len(b) > 7 && b[0] == 1 {
			finalText := string(b[8:])
			text += StringTransform(finalText)
		}
		text += StringTransform("\n")
	}
	err := scanner.Err()
	if err != nil {
		log.WithError(err).Error("There was an error with the scanner")
	}
	log.Debug(text)
	return text, err
}

func GetRedis(c *cli.Context) (*redis.Client, error) {
	redisOptions = &redis.Options{
		Addr:     c.GlobalString("redis"),
		Password: c.GlobalString("redis-password"),
		DB:       int64(c.GlobalInt("redis-db")),
	}

	client, err := GetRedisClient()

	log.WithFields(log.Fields{"http": c.GlobalString("redis"), "db": c.GlobalString("redis")}).Info("Connected to Redis Host")
	return client, err
}

func GetRedisClient() (*redis.Client, error) {
	client := redis.NewClient(redisOptions)

	_, err := client.Ping().Result()
	if err != nil {
		log.Error("Unable to connect to redis host")
		return nil, err
	}

	return client, nil
}

func GetDockerCient(c *cli.Context) (*dockerapi.Client, string, error) {
	host := c.GlobalString("host")
	if host == "" {
		log.Error("Incorrect usage, please set the docker host")
		return nil, "", errors.New("Unable to connect to docker host")
	}

	tlsConfig := &tls.Config{}

	certPath := c.GlobalString("cert")
	if certPath != "" {
		caFile := filepath.Join(certPath, "ca.pem")
		if _, err := os.Stat(caFile); os.IsNotExist(err) {
			log.WithField("file", caFile).Error("Cannot open file")
			log.Error("Incorrect usage, please set correct cert files")
			return nil, host, errors.New("Unable to connect to docker host")
		}

		certFile := filepath.Join(certPath, "cert.pem")
		if _, err := os.Stat(certFile); os.IsNotExist(err) {
			log.WithField("file", certFile).Error("Cannot open file")
			log.Error("Incorrect usage, please set correct cert files")
			return nil, host, errors.New("Unable to connect to docker host")
		}

		keyFile := filepath.Join(certPath, "key.pem")
		if _, err := os.Stat(keyFile); os.IsNotExist(err) {
			log.WithField("file", keyFile).Error("Cannot open file")
			log.Error("Incorrect usage, please set correct cert files")
			return nil, host, errors.New("Unable to connect to docker host")
		}

		cert, _ := tls.LoadX509KeyPair(certFile, keyFile)
		pemCerts, _ := ioutil.ReadFile(caFile)

		tlsConfig.RootCAs = x509.NewCertPool()
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		tlsConfig.Certificates = []tls.Certificate{cert}
		tlsConfig.RootCAs.AppendCertsFromPEM(pemCerts)
	}
	var dockerClient *dockerapi.Client
	var err error
	if certPath == "" {
		dockerClient, err = dockerapi.NewClient(host)
	} else {
		dockerClient, err = dockerapi.NewClient(host)
	}
	if err != nil {
		log.Error("Unable to connect to docker host")
		return nil, host, err
	}
	env, err := dockerClient.Docker.Version()
	if err != nil {
		log.WithError(err).Error("Unable to ping docker host")
		return nil, host, err
	}
	log.Info("Connected to Docker Host " + host)
	log.Debug("Docker Version: " + env.Get("Version"))
	log.Debug("Git Commit:" + env.Get("GitCommit"))
	log.Debug("Go Version:" + env.Get("GoVersion"))

	return dockerClient, host, err
}

func Contains(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}
	_, ok := set[item]
	return ok
}

func IndexOf(slice []string, item string) (ind int, ok bool) {
	ind = -1
	for i, str := range slice {
		if str == item {
			return i, true
		}
	}
	return ind, false
}

func HandleError(message string, err error, c *gin.Context) map[string]string {
	log.Warn("*******************************************************************************")
	log.Warn("RemoteAddr :\t" + c.Request.RemoteAddr)
	log.Warn("RequestURI :\t" + c.Request.RequestURI)
	log.Warn("Message :\t" + message)
	log.Warn("Error :\t" + err.Error())
	log.Warn("*******************************************************************************")

	return map[string]string{
		"message": message,
		"details": err.Error(),
	}
}
