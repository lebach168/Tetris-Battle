package main

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/joho/godotenv"

	"log"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"tetris-be/internal/validator"
)

const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func LoadConfig() *Config {
	var envPath string

	flag.StringVar(&envPath, "env", "../../.env.local", "Path to .env file")
	flag.Parse()

	if err := godotenv.Load(envPath); err != nil {
		log.Fatalf("Error loading env file: %s", envPath)
		return nil
	}

	var cfg Config
	cfg.port = getIntEnv("PORT", 8080)

	return &cfg
}

func getIntEnv(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	num, err := strconv.Atoi(val)
	if err != nil {
		log.Printf("Invalid int for %s: %s", key, val)
		return defaultVal
	}
	return num
}

func healthcheck(w http.ResponseWriter, r *http.Request) {
	data := envelope{
		"status": "available",
		"system_info": map[string]string{
			"version": "1.0",
		},
	}
	err := encodeComplexType(w, http.StatusOK, data, nil)
	if err != nil {
		serverErrorResponse(w, r, err)
	}
}

func readIntParam(r *http.Request, name string) (int64, error) {
	v, err := strconv.ParseInt(r.PathValue(name), 10, 64)
	if err != nil {
		return 0, errors.New("invalid param name")
	}
	return v, nil
}

// read from query string
func readInt(qs url.Values, key string, defaultVal int, v *validator.Validator) int {
	s := qs.Get(key)
	if s == "" {
		return defaultVal
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddError(key, "must be an integer value")
		return defaultVal
	}
	return i
}
func readString(qs url.Values, key string, defaultVal string) string {
	s := qs.Get(key) //if len qs[key]>1 -> return only qs[0]
	if s == "" {
		return defaultVal
	}
	return s
}
func readCSV(qs url.Values, key string, defaultVal []string) []string {
	csv := qs.Get(key)
	if csv == "" {
		return defaultVal
	}
	return strings.Split(csv, ",")
}
func readListString(qs url.Values, key string, defaultVal []string) []string {
	ls := qs[key]
	if len(ls) == 0 {
		return defaultVal
	}
	var res []string
	copy(res, ls)
	return res
}

func encodeComplexType(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	jsonData, err := json.MarshalIndent(data, "", "  ") //for improved readability only
	if err != nil {
		return err
	}
	jsonData = append(jsonData, '\n')
	for key, value := range headers {
		w.Header()[key] = value
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(jsonData)
	return nil
}

func encode[T any](w http.ResponseWriter, status int, v T, headers http.Header) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	for key, value := range headers {
		w.Header()[key] = value
	}
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}

func decode[T any](r *http.Request) (T, error) {
	//maxBytes := 3_145_728
	//r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("decode json: %w", err)
	}

	return v, nil
}

func generateID(n int) (string, error) {
	b := make([]byte, n)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		b[i] = letters[num.Int64()]
	}
	return string(b), nil
}

type envelope map[string]interface{}
