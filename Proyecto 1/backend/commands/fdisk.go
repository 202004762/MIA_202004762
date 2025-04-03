package commands

import (
	"backend/structures"
	"backend/utils"
	"encoding/binary"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)


type FDISK struct{
	size int
	unit string
	fit  string
	path string
	typ  string
	name string

}

func ParseFdisk(tokens []string) (string, error){
	cmd := &FDISK{}
	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-size=\d+|-unit=[bBkKmM]|-fit=[bBfF]{2}|-path="[^"]+"|-path=[^\s]+|-type=[pPeElL]|-name="[^"]+"|-name=[^\s]+`)
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
			case "-size":
				size, err := strconv.Atoi(value)
				if err != nil || size <= 0{
					return "", errors.New("el tamaño debe ser un número entero positivo")
					
				}

				cmd.size = size

			case "-unit":
				if value != "B" && value != "K" && value != "M"{
					return "", errors.New("la unidad debe ser B, K o M")

				}

				cmd.unit = strings.ToUpper(value)

			case "-fit":
				value = strings.ToUpper(value)
				if value != "BF" && value != "FF" && value != "WF"{
					return "", errors.New("el ajuste debe ser BF, FF o WF")

				}

				cmd.fit = value

			case "-path":
				if value == ""{
					return "", errors.New("el path no puede estar vacío")

				}

				cmd.path = value

			case "-type":
				value = strings.ToUpper(value)
				if value != "P" && value != "E" && value != "L"{
					return "", errors.New("el tipo debe ser P, E o L")

				}

				cmd.typ = value

			case "-name":
				if value == ""{
					return "", errors.New("el nombre no puede estar vacío")

				}

				cmd.name = value

			default:
				return "", fmt.Errorf("parámetro desconocido: %s", key)

		}

	}

	if cmd.size == 0{
		return "", errors.New("faltan parámetros requeridos: -size")

	}

	if cmd.path == ""{
		return "", errors.New("faltan parámetros requeridos: -path")

	}

	if cmd.name == ""{
		return "", errors.New("faltan parámetros requeridos: -name")

	}

	if cmd.unit == ""{
		cmd.unit = "K"

	}

	if cmd.fit == ""{
		cmd.fit = "WF"

	}

	if cmd.typ == ""{
		cmd.typ = "P"

	}

	err := commandFdisk(cmd)
	if err != nil{
		return "", err

	}

	return fmt.Sprintf("------------------------"+
		"FDISK: Partición creada exitosamente\n"+
		"-> Path: %s\n"+
		"-> Nombre: %s\n"+
		"-> Tamaño: %d%s\n"+
		"-> Tipo: %s\n"+
		"-> Fit: %s",
		cmd.path, cmd.name, cmd.size, cmd.unit, cmd.typ, cmd.fit), nil

}

func commandFdisk(fdisk *FDISK) error{
	sizeBytes, err := utils.ConvertToBytes(fdisk.size, fdisk.unit)
	if err != nil{
		fmt.Println("Error convirtiendo tamaño:", err)
		return err

	}

	switch fdisk.typ{
		case "P":
			err = createPrimaryPartition(fdisk, sizeBytes)

		case "E":
			err = createExtendedPartition(fdisk, sizeBytes)

		case "L":
			err = createLogicalPartition(fdisk, sizeBytes)

		default:
			err = errors.New("tipo de partición inválido")

	}

	return err

}

func createPrimaryPartition(fdisk *FDISK, sizeBytes int) error{
	var mbr structures.MBR
	err := mbr.DeserializeMBR(fdisk.path)
	if err != nil{
		fmt.Println("Error deserializando el MBR:", err)
		return err

	}

	used := int32(0)
	for _, p := range mbr.Mbr_partitions{
		if p.Part_status[0] != 'N'{
			used += p.Part_size

		}

	}

	if used+int32(sizeBytes) > mbr.Mbr_size{
		return fmt.Errorf("no hay suficiente espacio en el disco para crear la partición")

	}

	//fmt.Println("\nMBR original:")
	//mbr.PrintMBR()

	availablePartition, startPartition, indexPartition := mbr.GetFirstAvailablePartition()
	if availablePartition == nil{
		return errors.New("no hay particiones disponibles")

	}

	if int32(startPartition) + int32(sizeBytes) > mbr.Mbr_size{
		return errors.New("el espacio disponible no es suficiente para la nueva partición")

	}

	//fmt.Println("\nPartición disponible:")
	//availablePartition.PrintPartition()

	availablePartition.CreatePartition(startPartition, sizeBytes, fdisk.typ, fdisk.fit, fdisk.name)

	//fmt.Println("\nPartición creada")
	//availablePartition.PrintPartition()

	mbr.Mbr_partitions[indexPartition] = *availablePartition

	//fmt.Println("\nParticiones actuales:")
	//mbr.PrintPartitions()

	err = mbr.SerializeMBR(fdisk.path)
	if err != nil{
		fmt.Println("Error serializando MBR:", err)

	}

	return err

}

func createExtendedPartition(fdisk *FDISK, sizeBytes int) error{
	var mbr structures.MBR
	err := mbr.DeserializeMBR(fdisk.path)
	if err != nil{
		return fmt.Errorf("error deserializando el MBR: %v", err)

	}

	//fmt.Println("\nMBR original:")
	//mbr.PrintMBR()

	extendedExists := false
	used := int32(0)
	for _, p := range mbr.Mbr_partitions{
		if p.Part_type[0] == 'E'{
			extendedExists = true
			break

		}

		if p.Part_status[0] != 'N'{
			used += p.Part_size + int32(binary.Size(structures.PARTITION{}))

		}

	}

	if extendedExists{
		return errors.New("ya existe una particion extendida en el disco")

	}

	if used+int32(sizeBytes) > mbr.Mbr_size{
		return errors.New("no hay suficiente espacio en el disco para crear la partición extendida")

	}

	availablePartition, startPartition, index := mbr.GetFirstAvailablePartition()
	if availablePartition == nil{
		return errors.New("no hay particiones disponibles en el MBR")

	}

	if int32(startPartition)+int32(sizeBytes) > mbr.Mbr_size{
		return errors.New("el espacio disponible no es suficiente para la nueva partición")
	}

	//fmt.Println("\nPartición disponible:")
	//availablePartition.PrintPartition()

	availablePartition.CreatePartition(startPartition, sizeBytes, "E", fdisk.fit, fdisk.name)
	mbr.Mbr_partitions[index] = *availablePartition

	//fmt.Println("\nPartición creada")
	//availablePartition.PrintPartition()

	ebr := structures.EBR{
		Part_mount: [1]byte{'N'},
		Part_fit:   [1]byte{fdisk.fit[0]},
		Part_start: int32(startPartition),
		Part_s:     0,
		Part_next:  -1,

	}

	copy(ebr.Part_name[:], "")
	err = ebr.Serialize(fdisk.path, int64(startPartition))
	if err != nil{
		return fmt.Errorf("error escribiendo el primer EBR: %v", err)

	}

	//fmt.Println("\nEBR inicial dentro de la partición extendida:")
	//ebr.PrintEBR()

	//fmt.Println("\nParticiones actuales:")
	//mbr.PrintPartitions()

	//fmt.Println("\nDebug: EBRs actuales dentro de la partición extendida:")
	//structures.PrintPartitions(fdisk.path, int32(startPartition), int32(sizeBytes))

	return mbr.SerializeMBR(fdisk.path)

}

func createLogicalPartition(fdisk *FDISK, sizeBytes int) error{
	var mbr structures.MBR
	err := mbr.DeserializeMBR(fdisk.path)
	if err != nil{
		return fmt.Errorf("error deserializando el MBR: %v", err)

	}

	var extended structures.PARTITION
	extendedFound := false
	for _, p := range mbr.Mbr_partitions{
		if p.Part_type[0] == 'E'{
			extended = p
			extendedFound = true
			break

		}

	}

	if !extendedFound{
		return errors.New("no se encontro una particion extendida para crear la logica")

	}

	offset := extended.Part_start
	limit := extended.Part_start + extended.Part_size
	var lastEBR structures.EBR
	found := false
	for{
		err = lastEBR.Deserialize(fdisk.path, int64(offset))
		if err != nil{
			return fmt.Errorf("error leyendo EBR: %v", err)

		}

		if lastEBR.Part_mount[0] == 'N'{
			found = true
			break

		}

		if lastEBR.Part_next == -1{
			break

		}

		offset = lastEBR.Part_next

	}

	if !found{
		offset = lastEBR.Part_start + lastEBR.Part_s + int32(binary.Size(structures.EBR{}))
		if offset+int32(sizeBytes) > limit{
			return errors.New("no hay suficiente espacio en la extendida para crear la particion logica")

		}

	}

	newEBR := structures.EBR{
		Part_mount: [1]byte{'0'},
		Part_fit:   [1]byte{fdisk.fit[0]},
		Part_start: offset,
		Part_s:     int32(sizeBytes),
		Part_next:  -1,

	}

	copy(newEBR.Part_name[:], fdisk.name)
	err = newEBR.Serialize(fdisk.path, int64(offset))
	if err != nil{
		return fmt.Errorf("error escribiendo EBR: %v", err)

	}

	//fmt.Println("\nEBR de partición lógica creada")
	//newEBR.PrintEBR()

	if lastEBR.Part_s != 0{
		lastEBR.Part_next = offset
		err = lastEBR.Serialize(fdisk.path, int64(lastEBR.Part_start))
		if err != nil{
			return fmt.Errorf("error actualizando EBR anterior: %v", err)

		}

		//fmt.Println("\nEBR anterior actualizado:")
		//lastEBR.PrintEBR()

	}

	//fmt.Println("\nDebug: EBRs actuales dentro de la partición extendida:")
	//structures.PrintPartitions(fdisk.path, extended.Part_start, extended.Part_size)

	return nil

}
