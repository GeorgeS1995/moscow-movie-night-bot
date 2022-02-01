package tests

import (
	"errors"
	"fmt"
	"os"
	"testing"

	cfg "github.com/GeorgeS1995/moscow-movie-night-bot/internal/cfg"
)

type EnvSimpleTestData struct {
	env         string
	expectedEnv string
}

func (env EnvSimpleTestData) GetEnv() string {
	return env.env
}

type EnvTestData struct {
	EnvSimpleTestData
	err error
}

func (env EnvTestData) GetEnv() string {
	return env.env
}

type EnvTestInterface interface {
	GetEnv() string
}

func reloadEnv(envName string, data EnvTestInterface) error {
	err := os.Unsetenv(envName)
	if err != nil {
		return err
	}
	env := data.GetEnv()
	if env != "" {
		os.Setenv(envName, env)
	}
	return nil
}

var getBotDebugTestCases = []EnvTestData{
	{EnvSimpleTestData{"", "false"}, &cfg.ConfigError{Err: errors.New("strconv.ParseBool: parsing \"\": invalid syntax")}},
	{EnvSimpleTestData{"true", "true"}, nil},
	{EnvSimpleTestData{"0", "false"}, nil},
}

func TestGetBotDebug(t *testing.T) {
	for _, data := range getBotDebugTestCases {
		err := reloadEnv("BOT_DEBUG", data)
		if err != nil {
			t.Fatal(err)
			return
		}
		debug, err := cfg.GetBotDebug()
		debugStr := fmt.Sprint(debug)
		if debugStr != data.expectedEnv {
			t.Error("Not expected debug value:", debugStr, data.expectedEnv)
			continue
		}
		if err != nil && data.err != nil && err.Error() != data.err.Error() {
			t.Error("Not expected error response:", err.Error(), data.err.Error())
			continue
		}
		t.Logf("GetBotDebug testpased for case %+v", data)
	}
}

var getTelegramKeyTestCases = []EnvTestData{
	{EnvSimpleTestData{"", ""}, &cfg.ConfigError{Err: errors.New("TELEGRAM_KEY is required parameter")}},
	{EnvSimpleTestData{"1607551410:AAHIPMBPdeHnakAPioZRH4g9G5m4FtZoBlbE", "1607551410:AAHIPMBPdeHnakAPioZRH4g9G5m4FtZoBlbE"}, nil},
}

func TestTelegramKey(t *testing.T) {
	for _, data := range getTelegramKeyTestCases {
		err := reloadEnv("TELEGRAM_KEY", data)
		if err != nil {
			t.Fatal(err)
			return
		}
		tgKey, err := cfg.GetTelegramKey()
		if tgKey != data.expectedEnv {
			t.Error("Not expected telegram key:", tgKey, data.expectedEnv)
			continue
		}
		if err != nil && data.err != nil && err.Error() != data.err.Error() {
			t.Error("Not expected error response:", err.Error(), data.err.Error())
			continue
		}
		t.Logf("GetTelegramKey testpased for case %+v", data)
	}
}

var getTelegramLongpullingTimeoutTestCases = []EnvTestData{
	{EnvSimpleTestData{"", "60"}, nil},
	{EnvSimpleTestData{"10", "10"}, nil},
	{EnvSimpleTestData{"asdsa", "0"}, &cfg.ConfigError{Err: errors.New("strconv.ParseInt: parsing \"asdsa\": invalid syntax")}},
}

func TestGetTelegramLongpullingTimeout(t *testing.T) {
	for _, data := range getTelegramLongpullingTimeoutTestCases {
		err := reloadEnv("TELEGRAM_LONGPULLING_TIMEOUT", data)
		if err != nil {
			t.Fatal(err)
			return
		}
		timeout, err := cfg.GetTelegramLongpullingTimeout()
		timeoutStr := fmt.Sprint(timeout)
		if timeoutStr != data.expectedEnv {
			t.Error("Not expected timeout:", timeoutStr, data.expectedEnv)
			continue
		}
		if err != nil && data.err != nil && err.Error() != data.err.Error() {
			t.Error("Not expected error response:", err.Error(), data.err.Error())
			continue
		}
		t.Logf("GetTelegramLongpullingTimeout testpased for case %+v", data)
	}
}

var getGetDataBaseNameCases = []EnvSimpleTestData{
	{"", "movie.db"},
	{"new_name.db", "new_name.db"},
}

func TestGetDataBaseName(t *testing.T) {
	for _, data := range getGetDataBaseNameCases {
		err := reloadEnv("DB_NAME", data)
		if err != nil {
			t.Fatal(err)
			return
		}
		dbName := cfg.GetDataBaseName()
		if dbName != data.expectedEnv {
			t.Error("Not expected db name:", dbName, data.expectedEnv)
			continue
		}
		t.Logf("GetDataBaseName testpased for case %+v", data)
	}
}

var getFilmLimitTestCases = []EnvTestData{
	{EnvSimpleTestData{"", "0"}, nil},
	{EnvSimpleTestData{"0", "0"}, nil},
	{EnvSimpleTestData{"00", "0"}, nil},
	{EnvSimpleTestData{"10", "10"}, nil},
	{EnvSimpleTestData{"asdsa", "0"}, &cfg.ConfigError{Err: errors.New("strconv.ParseInt: parsing \"asdsa\": invalid syntax")}},
}

func TestGetAddingFilmLimit(t *testing.T) {
	for _, data := range getFilmLimitTestCases {
		err := reloadEnv("FILM_LIMIT", data)
		if err != nil {
			t.Fatal(err)
			return
		}
		filmLimit, err := cfg.GetAddingFilmLimit()
		filmLimitStr := fmt.Sprint(filmLimit)
		if filmLimitStr != data.expectedEnv {
			t.Error("Not expected timeout:", filmLimitStr, data.expectedEnv)
			continue
		}
		if err != nil && data.err != nil && err.Error() != data.err.Error() {
			t.Error("Not expected error response:", err.Error(), data.err.Error())
			continue
		}
		t.Logf("TestGetAddingFilmLimit testpased for case %+v", data)
	}
}
