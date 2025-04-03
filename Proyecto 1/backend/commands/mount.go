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


type MOUNT struct{
	path string
	name string

}

func ParseMount(tokens []string) (string, error){
	cmd := &MOUNT{}
	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-path="[^"]+"|-path=[^\s]+|-name="[^"]+"|-name=[^\s]+`)
	matches := re.FindAllString(args, -1)
	for _, match := range matches {
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
				if value == ""{
					return "", errors.New("el path no puede estar vacío")

				}

				cmd.path = value

			case "-name":
				if value == ""{
					return "", errors.New("el nombre no puede estar vacío")

				}

				cmd.name = value

			default:
				return "", fmt.Errorf("parámetro desconocido: %s", key)

		}

	}

	if cmd.path == ""{
		return "", errors.New("faltan parámetros requeridos: -path")

	}

	if cmd.name == ""{
		return "", errors.New("faltan parámetros requeridos: -name")

	}

	idPartition, err := commandMount(cmd)
	if err != nil{
		return "", err

	}

	return fmt.Sprintf("------------------------"+
		"MOUNT: Partición montada exitosamente\n"+
		"-> Path: %s\n"+
		"-> Nombre: %s\n"+
		"-> ID: %s",
		cmd.path, cmd.name, idPartition), nil

}

func commandMount(mount *MOUNT) (string, error){
	var mbr structures.MBR
	err := mbr.DeserializeMBR(mount.path)
	if err != nil{
		fmt.Printf("error deserializando el MBR: %v\n", err)
		return "", err

	}

	partition, indexPartition := mbr.GetPartitionByName(mount.name)
	if partition == nil{
		return "", errors.New("la partición no existe")

	}

	for id, path := range stores.MountedPartitions{
		if path == mount.path && strings.HasSuffix(id, string(partition.Part_id[:])){
			return "", errors.New("la partición ya está montada")

		}

	}

	idPartition, partitionCorrelative, err := generatePartitionID(mount)
	if err != nil{
		fmt.Println("Error generando el id de partición:", err)
		return "", err

	}

	stores.MountedPartitions[idPartition] = mount.path
	partition.MountPartition(partitionCorrelative, idPartition)
	mbr.Mbr_partitions[indexPartition] = *partition
	err = mbr.SerializeMBR(mount.path)
	if err != nil{
		fmt.Println("Error serializando el MBR:", err)
		return "", err

	}

	return idPartition, nil

}

func generatePartitionID(mount *MOUNT) (string, int, error){
	letter, partitionCorrelative, err := utils.GetLetterAndPartitionCorrelative(mount.path)
	if err != nil{
		fmt.Println("Error obteniendo la letra:", err)
		return "", 0, err

	}

	idPartition := fmt.Sprintf("%s%d%s", stores.Carnet, partitionCorrelative, letter)
	return idPartition, partitionCorrelative, nil

}
