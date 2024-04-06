package config

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
)

func LoadConfigFromEnv(cfg interface{}) error {
    v := reflect.ValueOf(cfg).Elem()
    for i := 0; i < v.NumField(); i++ {
        field := v.Field(i)
        tag := v.Type().Field(i).Tag.Get("env")

        // If the field is another struct, recurse
        if field.Kind() == reflect.Struct {
            err:= LoadConfigFromEnv(field.Addr().Interface())
			if err != nil {
				return err
			}
            continue
        }

        // Skip fields without an env tag
        if tag == "" {
            continue
        }

        envValue := os.Getenv(tag)
        if envValue == "" {
			return fmt.Errorf("missing env variable: %s", tag)
        }

        // Assign the env value to the field, converting the type as necessary
        switch field.Kind() {
        case reflect.String:
            field.SetString(envValue)
        case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
            intValue, err := strconv.ParseInt(envValue, 10, 64)
            if err != nil {
                return fmt.Errorf("can't parse env variable %s=%s to int", tag, envValue)
            }
            field.SetInt(intValue)
        case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
            uintValue, err := strconv.ParseUint(envValue, 10, 64)
            if err != nil {
				return fmt.Errorf("can't parse env variable %s=%s to uint", tag, envValue)
            }
            field.SetUint(uintValue)
        // Other cases
        }
    }
    return nil
}
