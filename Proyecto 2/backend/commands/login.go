package commands

import (
	stores "backend/stores"
	"errors"
	"fmt"
	"regexp"
	"strings"
)


type LOGIN struct {
	User string
	Pass string
	Id   string

}

func ParseLogin(tokens []string) (string, error){
	cmd := &LOGIN{}
	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-user=[^\s]+|-pass=[^\s]+|-id=[^\s]+`)
	matches := re.FindAllString(args, -1)
	if len(matches) != len(tokens){
		for _, token := range tokens{
			if !re.MatchString(token){
				return "", fmt.Errorf("parámetro inválido: %s", token)

			}

		}

	}

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
			case "-user":
				if value == ""{
					return "", errors.New("el usuario no puede estar vacío")

				}

				cmd.User = value

			case "-pass":
				if value == ""{
					return "", errors.New("la contraseña no puede estar vacía")

				}

				cmd.Pass = value

			case "-id":
				if value == ""{
					return "", errors.New("el id no puede estar vacío")

				}

				cmd.Id = value

			default:
				return "", fmt.Errorf("parámetro desconocido: %s", key)

		}

	}

	if cmd.Id == ""{
		return "", errors.New("faltan parámetros requeridos: -id")

	}

	if cmd.User == ""{
		return "", errors.New("faltan parámetros requeridos: -user")

	}

	if cmd.Pass == ""{
		return "", errors.New("faltan parámetros requeridos: -pass")

	}

	err := commandLogin(cmd)
	if err != nil{
		return "", err

	}

	return fmt.Sprintf("------------------------" +
	"LOGIN: sesión iniciada con éxito\n" + 
	"Usuario: %s, Contraseña: %s, ID: %s", cmd.User, cmd.Pass, cmd.Id), nil

}

func CommandLoginPublic(login *LOGIN) error{
	return commandLogin(login)

}

func commandLogin(login *LOGIN) error{
	if stores.Auth.IsAuthenticated(){
		return fmt.Errorf("ya hay una sesión iniciada. Debe cerrarla antes de iniciar una nueva")

	}

	partitionSuperblock, _, partitionPath, err := stores.GetMountedPartitionSuperblock(login.Id)
	if err != nil{
		return fmt.Errorf("error al obtener la partición montada: %w", err)

	}

	usersBlock, err := partitionSuperblock.GetUsersBlock(partitionPath)
	if err != nil{
		return fmt.Errorf("error al obtener el bloque de usuarios: %w", err)

	}

	//fmt.Println(usersBlock)

	content := strings.Trim(usersBlock, "\x00")
	lines := strings.Split(content, "\n")

	//fmt.Println(content)

	var foundUser bool
	var userPassword string
	for _, line := range lines{
		fields := strings.Split(line, ",")
		for i := range fields{
			fields[i] = strings.TrimSpace(fields[i])

		}

		if len(fields) == 5 && fields[1] == "U"{
			if strings.EqualFold(fields[3], login.User){
				foundUser = true
				userPassword = fields[4]
				break

			}

		}

	}

	if !foundUser{
		return fmt.Errorf("el usuario %s no existe", login.User)

	}

	if !strings.EqualFold(userPassword, login.Pass){
		return fmt.Errorf("la contraseña no coincide")

	}

	stores.Auth.Login(login.User, login.Pass, login.Id)
	return nil

}