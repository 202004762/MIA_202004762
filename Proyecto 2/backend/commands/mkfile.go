package commands

import (
	stores "backend/stores"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)


type MKFILE struct{
	path string
	p    bool
	size int
	cont string

}

func ParseMkfile(tokens []string) (string, error){
	cmd := &MKFILE{}
	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-path="?([^"\s]+)"?|(-p)|-size=([0-9]+)|-cont="?([^"\s]+)"?`)
	matches := re.FindAllStringSubmatch(args, -1)
	if len(matches) == 0{
		return "", errors.New("parámetros inválidos")

	}

	for _, match := range matches{
		if match[1] != ""{
			cmd.path = match[1]

		}else if match[2] == "-p"{
			cmd.p = true

		}else if match[3] != ""{
			fmt.Sscanf(match[3], "%d", &cmd.size)

		}else if match[4] != ""{
			cmd.cont = match[4]

		}

	}

	if cmd.path == ""{
		return "", errors.New("el parámetro -path es obligatorio")

	}

	err := commandMkfile(cmd)
	if err != nil{
		return "", err

	}

	return fmt.Sprintf("------------------------"+
		"MKFILE: exito creando el archivo"+
		"Archivo %s creado correctamente.", cmd.path), nil

}

func commandMkfile(mkfile *MKFILE) error{
	if !stores.Auth.IsAuthenticated(){
		return errors.New("no se ha iniciado sesión")

	}

	partitionID := stores.Auth.GetPartitionID()
	sb, mountedPartition, partitionPath, err := stores.GetMountedPartitionSuperblock(partitionID)
	if err != nil{
		return fmt.Errorf("error al obtener superbloque: %w", err)

	}

	parentDirs, fileName := getParentDirectories(mkfile.path)
	if mkfile.p{
		err = sb.CreateFolder(partitionPath, parentDirs, "")
		if err != nil{
			return fmt.Errorf("error al crear directorios padres: %w", err)

		}

	}

	var content string
	if mkfile.cont != ""{
		data, err := os.ReadFile(mkfile.cont)
		if err != nil{
			return fmt.Errorf("no se pudo leer el archivo del parámetro -cont: %w", err)

		}

		content = string(data)

	}else if mkfile.size > 0{
		content = strings.Repeat("A", mkfile.size)

	}

	err = sb.CreateFile(partitionPath, parentDirs, fileName, content)
	if err != nil{
		return fmt.Errorf("error al crear el archivo: %w", err)

	}

	return sb.Serialize(partitionPath, int64(mountedPartition.Part_start))

}

func getParentDirectories(path string) ([]string, string){
	path = filepath.Clean(path)
	parts := strings.Split(path, string(filepath.Separator))
	var parents []string
	for i := 1; i < len(parts)-1; i++{
		parents = append(parents, parts[i])

	}

	return parents, parts[len(parts)-1]

}
