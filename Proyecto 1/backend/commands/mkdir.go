package commands

import (
	stores "backend/stores"
	structures "backend/structures"
	utils "backend/utils"
	"errors"
	"fmt"
	"regexp"
	"strings"
)


type MKDIR struct{
	path string
	p    bool

}


func ParseMkdir(tokens []string) (string, error){
	cmd := &MKDIR{}
	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-path=[^\s]+|-p`)
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
		key := strings.ToLower(kv[0])
		switch key{
			case "-path":
				if len(kv) != 2{
					return "", fmt.Errorf("formato de parámetro inválido: %s", match)

				}

				value := kv[1]
				if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\""){
					value = strings.Trim(value, "\"")

				}

				cmd.path = value

			case "-p":
				cmd.p = true

			default:
				return "", fmt.Errorf("parámetro desconocido: %s", key)

		}

	}

	if cmd.path == ""{
		return "", errors.New("faltan parámetros requeridos: -path")

	}

	err := commandMkdir(cmd)
	if err != nil{
		return "", err

	}

	return fmt.Sprintf("------------------------"+
		"MKDIR: exito creando el direcotrio"+
		"Directorio %s creado correctamente.", cmd.path), nil

}

func commandMkdir(mkdir *MKDIR) error{
	var partitionID string
	if stores.Auth.IsAuthenticated(){
		partitionID = stores.Auth.GetPartitionID()

	}else{
		return errors.New("no se ha iniciado sesión en ninguna partición")

	}

	partitionSuperblock, mountedPartition, partitionPath, err := stores.GetMountedPartitionSuperblock(partitionID)
	if err != nil{
		return fmt.Errorf("error al obtener la partición montada: %w", err)

	}

	err = createDirectory(mkdir.path, partitionSuperblock, partitionPath, mountedPartition)
	if err != nil{
		err = fmt.Errorf("error al crear el directorio: %w", err)

	}

	return err

}

func createDirectory(dirPath string, sb *structures.SuperBlock, partitionPath string, mountedPartition *structures.PARTITION) error{
	fmt.Println("\nCreando directorio:", dirPath)
	parentDirs, destDir := utils.GetParentDirectories(dirPath)
	fmt.Println("\nDirectorios padres:", parentDirs)
	fmt.Println("Directorio destino:", destDir)
	err := sb.CreateFolder(partitionPath, parentDirs, destDir)
	if err != nil{
		return fmt.Errorf("error al crear el directorio: %w", err)

	}

	//sb.PrintInodes(partitionPath)
	//sb.PrintBlocks(partitionPath)
	err = sb.Serialize(partitionPath, int64(mountedPartition.Part_start))
	if err != nil{
		return fmt.Errorf("error al serializar el superbloque: %w", err)

	}

	return nil

}
