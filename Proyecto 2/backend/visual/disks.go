package visual

import (
	"backend/structures"
	"encoding/json"
	"io/fs"
	"path/filepath"
	"strings"
	"time"
)


type DiskInfo struct{
	Path       string          `json:"path"`
	Size       int32           `json:"size"`
	CreatedAt  string          `json:"created_at"`
	Fit        string          `json:"fit"`
	Partitions []PartitionInfo `json:"partitions"` 

}

type PartitionInfo struct{
	Name        string `json:"name"`
	Start       int32  `json:"start"`
	Size        int32  `json:"size"`
	Type        string `json:"type"`
	Status      string `json:"status"`
	ID          string `json:"id"`
	Correlative int32  `json:"correlative"`
	Fit         string `json:"fit"` 

}

// GetAllDisksInfo busca discos .mia en la carpeta base y devuelve su información en JSON
func GetAllDisksInfo(baseDir string) (string, error){
	var disks []DiskInfo

	err := filepath.WalkDir(baseDir, func(path string, d fs.DirEntry, err error) error{
		if err != nil{
			return err

		}

		if !d.IsDir() && strings.HasSuffix(d.Name(), ".mia"){
			mbr := &structures.MBR{}
			err := mbr.DeserializeMBR(path)
			if err != nil{
				return nil

			}

			disk := DiskInfo{
				Path:      path,
				Size:      mbr.Mbr_size,
				CreatedAt: time.Unix(int64(mbr.Mbr_creation_date), 0).Format("2006-01-02 15:04:05"),
				Fit:       strings.Trim(string(mbr.Mbr_disk_fit[:]), "\x00 "),

			}

			for _, part := range mbr.Mbr_partitions{
				if part.Part_status[0] != 'N' && part.Part_size > 0{
					partition := PartitionInfo{
						Name:        strings.Trim(string(part.Part_name[:]), "\x00 "),
						Start:       part.Part_start,
						Size:        part.Part_size,
						Type:        string(part.Part_type[0]),
						Status:      string(part.Part_status[0]),
						ID:          strings.Trim(string(part.Part_id[:]), "\x00 "),
						Correlative: part.Part_correlative,
						Fit:         string(part.Part_fit[0]),

					}

					disk.Partitions = append(disk.Partitions, partition)

				}

			}

			disks = append(disks, disk)

		}

		return nil

	})

	if err != nil{
		return "", err

	}

	result, err := json.MarshalIndent(disks, "", "  ")
	if err != nil{
		return "", err

	}

	return string(result), nil

}
