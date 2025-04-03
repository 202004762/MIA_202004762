package commands

import (
	stores "backend/stores"
	"errors"
	"fmt"
)


type LOGOUT struct{}

func ParseLogout(tokens []string) (string, error){
	if len(tokens) > 1{
		return "", fmt.Errorf("no hay ninguna sesión iniciada")

	}

	err := commandLogout()
	if err != nil{
		return "", err

	}

	return fmt.Sprintf("------------------------" +
	"LOGOUT: sesión cerrada con éxito\n"), nil

}

func commandLogout() error{
	if !stores.Auth.IsAuthenticated(){
		return errors.New("no hay ninguna sesión activa")

	}

	stores.Auth.Logout()
	return nil

}
