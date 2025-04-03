package commands

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"errors"
)


type RMDISK struct{
	path string

}

func ParseRmdisk(tokens []string) (string, error){
	cmd := &RMDISK{}
	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-path="[^"]+"|-path=[^\s]+`)
	matches := re.FindAllString(args, -1)
	for _, match := range matches{
		kv := strings.SplitN(match, "=", 2)
		if len(kv) != 2{
			return "", fmt.Errorf("formato de parámetro inválido: %s", match)

		}

		key, value := strings.ToLower(kv[0]), kv[1]
		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\""){
			value = strings.Trim(value, "\"")

		}

		switch key{
			case "-path":
				cmd.path = value

			default:
				return "", fmt.Errorf("parámetro desconocido: %s", key)

		}

	}

	if cmd.path == ""{
		return "", errors.New("faltan parámetros requeridos: -path")

	}

	err := executeRmdisk(cmd)
	if err != nil{
		return "", err

	}

	return fmt.Sprintf("------------------------"+
		"RMDISK: Disco eliminado exitosamente\n-> Path: %s",
		cmd.path), nil

}

func executeRmdisk(rmdisk *RMDISK) error{
	if _, err := os.Stat(rmdisk.path); os.IsNotExist(err){
		return fmt.Errorf("el archivo no existe: %s", rmdisk.path)

	}

	err := os.Remove(rmdisk.path)
	if err != nil{
		return fmt.Errorf("no se pudo eliminar el disco: %v", err)

	}

	//fmt.Printf("\n Disco eliminado correctamente: %s\n", rmdisk.path)
	return nil

}
