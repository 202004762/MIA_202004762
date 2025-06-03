package structures

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)


type EBR struct{
	Part_mount [1]byte
	Part_fit   [1]byte
	Part_start int32
	Part_s     int32
	Part_next  int32
	Part_name  [16]byte

}

/*
Part Mount:
	N: Disponible
	0: Creado
	1: Montado
*/

func (ebr *EBR) Serialize(path string, offset int64) error{
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil{
		return err

	}

	defer file.Close()
	_, err = file.Seek(offset, 0)
	if err != nil{
		return err

	}

	return binary.Write(file, binary.LittleEndian, ebr)

}

func (ebr *EBR) Deserialize(path string, offset int64) error{
	file, err := os.Open(path)
	if err != nil{
		return err

	}

	defer file.Close()
	_, err = file.Seek(offset, 0)
	if err != nil{
		return err

	}

	buffer := make([]byte, binary.Size(ebr))
	_, err = file.Read(buffer)
	if err != nil{
		return err

	}

	reader := bytes.NewReader(buffer)
	return binary.Read(reader, binary.LittleEndian, ebr)

}

func (ebr *EBR) PrintEBR(){
	fmt.Println("EBR:")
	fmt.Printf("  Mount: %c\n", ebr.Part_mount)
	fmt.Printf("  Fit: %c\n", ebr.Part_fit)
	fmt.Printf("  Start: %d\n", ebr.Part_start)
	fmt.Printf("  Size: %d\n", ebr.Part_s)
	fmt.Printf("  Next: %d\n", ebr.Part_next)
	fmt.Printf("  Name: %s\n", strings.Trim(string(ebr.Part_name[:]), "\x00"))

}

func PrintPartitions(diskPath string, start int32, size int32){
	fmt.Println("\n--- EBRs dentro de la partición extendida ---")
	offset := start
	limit := start + size
	var ebr EBR
	for offset < limit{
		err := ebr.Deserialize(diskPath, int64(offset))
		if err != nil{
			fmt.Println("Error al leer EBR:", err)
			break

		}

		if ebr.Part_s == 0{
			fmt.Println("EBR vacío encontrado. Fin del recorrido.")
			break

		}

		fmt.Printf("\n[EBR en offset %d]\n", offset)
		ebr.PrintEBR()
		if ebr.Part_next <= 0{
			fmt.Println("Último EBR encontrado (part_next = -1 o 0).")
			break

		}

		offset = ebr.Part_next

	}

}
