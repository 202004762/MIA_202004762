package structures

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)


type FileBlock struct{
	B_content [64]byte

}

func (fb *FileBlock) Serialize(path string, offset int64) error{
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil{
		return fmt.Errorf("error abriendo archivo para escritura: %w", err)

	}

	defer file.Close()
	if _, err = file.Seek(offset, 0); err != nil{
		return fmt.Errorf("error buscando offset %d: %w", offset, err)

	}

	if err = binary.Write(file, binary.LittleEndian, fb); err != nil{
		return fmt.Errorf("error escribiendo FileBlock: %w", err)

	}

	return nil

}

func (fb *FileBlock) Deserialize(path string, offset int64) error{
	file, err := os.Open(path)
	if err != nil{
		return fmt.Errorf("error abriendo archivo para lectura: %w", err)

	}

	defer file.Close()
	if _, err = file.Seek(offset, 0); err != nil{
		return fmt.Errorf("error buscando offset %d: %w", offset, err)

	}

	size := binary.Size(fb)
	if size <= 0{
		return fmt.Errorf("tamaño inválido de FileBlock: %d", size)

	}

	buffer := make([]byte, size)
	if _, err = file.Read(buffer); err != nil{
		return fmt.Errorf("error leyendo FileBlock: %w", err)
		
	}

	reader := bytes.NewReader(buffer)
	if err = binary.Read(reader, binary.LittleEndian, fb); err != nil{
		return fmt.Errorf("error deserializando FileBlock: %w", err)

	}

	return nil

}

func (fb *FileBlock) Print(){
	content := strings.Trim(string(fb.B_content[:]), "\x00")
	fmt.Println(content)

}
