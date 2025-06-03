package commands

import (
	"backend/stores"
	"backend/structures"
	"fmt"
)


func ParseMounted() error{
	if len(stores.MountedPartitions) == 0{
		fmt.Println("\nNo hay particiones montadas actualmente.")
		return nil

	}

	fmt.Println("\nParticiones montadas:")
	for id, path := range stores.MountedPartitions{
		var mbr structures.MBR
		err := mbr.DeserializeMBR(path)
		if err != nil{
			fmt.Printf("- %s (error al leer disco: %v)\n", id, err)
			continue

		}

		partition, _ := mbr.GetPartitionByID(id)
		if partition == nil{
			continue

		}

		fmt.Printf("- %s (nombre: %s, disco: %s)\n", id, string(partition.Part_name[:]), path)

	}

	return nil

}
