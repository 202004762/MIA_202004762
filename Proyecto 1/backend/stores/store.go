package stores

import (
	structures "backend/structures"
	"errors"
)


const Carnet string = "62" // 202004762
var (MountedPartitions map[string]string = make(map[string]string))
func GetMountedPartition(id string) (*structures.PARTITION, string, error){
	path := MountedPartitions[id]
	if path == ""{
		return nil, "", errors.New("la partición no está montada")

	}

	var mbr structures.MBR
	err := mbr.DeserializeMBR(path)
	if err != nil{
		return nil, "", err

	}

	partition, err := mbr.GetPartitionByID(id)
	if partition == nil{
		return nil, "", err

	}

	return partition, path, nil

}

func GetMountedPartitionRep(id string) (*structures.MBR, *structures.SuperBlock, string, error){
	path := MountedPartitions[id]
	if path == ""{
		return nil, nil, "", errors.New("la partición no está montada")

	}

	var mbr structures.MBR
	err := mbr.DeserializeMBR(path)
	if err != nil{
		return nil, nil, "", err

	}

	partition, err := mbr.GetPartitionByID(id)
	if partition == nil{
		return nil, nil, "", err

	}

	var sb structures.SuperBlock
	err = sb.Deserialize(path, int64(partition.Part_start))
	if err != nil{
		return nil, nil, "", err

	}

	return &mbr, &sb, path, nil

}

func GetMountedPartitionSuperblock(id string) (*structures.SuperBlock, *structures.PARTITION, string, error){
	path := MountedPartitions[id]
	if path == ""{
		return nil, nil, "", errors.New("la partición no está montada")

	}

	var mbr structures.MBR
	err := mbr.DeserializeMBR(path)
	if err != nil{
		return nil, nil, "", err

	}

	partition, err := mbr.GetPartitionByID(id)
	if partition == nil{
		return nil, nil, "", err

	}

	var sb structures.SuperBlock
	err = sb.Deserialize(path, int64(partition.Part_start))
	if err != nil{
		return nil, nil, "", err

	}

	return &sb, partition, path, nil

}