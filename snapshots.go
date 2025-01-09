package goqemu

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"time"
)

// Snapshot representa un punto de restauración de la VM
type Snapshot struct {
	ID          string // ID corto único
	Description string // Formato: [desc1, desc2, desc3]
	CreatedAt   time.Time
}

// generateSnapshotID genera un ID único para snapshots
func generateSnapshotID() string {
	return time.Now().Format("20060102150405")
}

// CreateSnapshot crea un nuevo snapshot de la VM
func (vm *QemuVM) CreateSnapshot(description string) (*Snapshot, error) {
	if vm.config == nil {
		return nil, errors.New("VM no configurada")
	}

	snapshot := &Snapshot{
		ID:          generateSnapshotID(),
		Description: description,
		CreatedAt:   time.Now(),
	}

	// Crear directorio de snapshots si no existe
	snapDir := filepath.Join(os.Getenv("HOME"), "qemu", "snapshots")
	err := os.MkdirAll(snapDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("error creando directorio de snapshots: %v", err)
	}

	// Crear archivo de metadatos
	metaPath := filepath.Join(snapDir, snapshot.ID+".meta")
	metaFile, err := os.Create(metaPath)
	if err != nil {
		return nil, fmt.Errorf("error creando archivo de metadatos: %v", err)
	}
	defer metaFile.Close()

	// Escribir metadatos
	_, err = fmt.Fprintf(metaFile, "ID: %s\nDescription: %s\nCreatedAt: %s\n",
		snapshot.ID, snapshot.Description, snapshot.CreatedAt.Format(time.RFC3339))
	if err != nil {
		return nil, fmt.Errorf("error escribiendo metadatos: %v", err)
	}

	// Crear snapshot en memoria o disco
	if vm.config.SnapshotsInMemory {
		// TODO: Implementar lógica para snapshots en memoria
	} else {
		// Crear snapshot en disco
		snapPath := filepath.Join(snapDir, snapshot.ID+".qcow2")
		cmd := exec.Command("qemu-img", "snapshot", "-c", snapshot.ID, snapPath)
		err := cmd.Run()
		if err != nil {
			return nil, fmt.Errorf("error creando snapshot en disco: %v", err)
		}
	}

	return snapshot, nil
}

// RestoreSnapshot restaura la VM a un snapshot existente
func (vm *QemuVM) RestoreSnapshot(id string) error {
	if vm.config == nil {
		return errors.New("VM no configurada")
	}

	// Verificar existencia del snapshot
	snapDir := filepath.Join(os.Getenv("HOME"), "qemu", "snapshots")
	metaPath := filepath.Join(snapDir, id+".meta")
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		return fmt.Errorf("snapshot no encontrado: %s", id)
	}

	// Restaurar snapshot
	if vm.config.SnapshotsInMemory {
		// TODO: Implementar lógica para snapshots en memoria
	} else {
		// Restaurar snapshot en disco
		snapPath := filepath.Join(snapDir, id+".qcow2")
		cmd := exec.Command("qemu-img", "snapshot", "-a", id, snapPath)
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("error restaurando snapshot: %v", err)
		}
	}

	// Reiniciar conexión SSH
	if vm.sshClient != nil {
		vm.sshClient.Close()
	}
	err := vm.connectSSH()
	if err != nil {
		return fmt.Errorf("error reconectando SSH: %v", err)
	}

	return nil
}

// ListSnapshots lista todos los snapshots disponibles
func (vm *QemuVM) ListSnapshots() ([]Snapshot, error) {
	if vm.config == nil {
		return nil, errors.New("VM no configurada")
	}

	snapDir := filepath.Join(os.Getenv("HOME"), "qemu", "snapshots")
	files, err := os.ReadDir(snapDir)
	if err != nil {
		return nil, fmt.Errorf("error leyendo directorio de snapshots: %v", err)
	}

	var snapshots []Snapshot

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".meta" {
			metaPath := filepath.Join(snapDir, file.Name())
			content, err := os.ReadFile(metaPath)
			if err != nil {
				return nil, fmt.Errorf("error leyendo metadatos: %v", err)
			}

			var snapshot Snapshot
			_, err = fmt.Sscanf(string(content),
				"ID: %s\nDescription: %s\nCreatedAt: %s\n",
				&snapshot.ID, &snapshot.Description, &snapshot.CreatedAt)
			if err != nil {
				return nil, fmt.Errorf("error parseando metadatos: %v", err)
			}

			snapshots = append(snapshots, snapshot)
		}
	}

	// Ordenar snapshots por fecha (más reciente primero)
	sort.Slice(snapshots, func(i, j int) bool {
		return snapshots[i].CreatedAt.After(snapshots[j].CreatedAt)
	})

	return snapshots, nil
}

// DeleteSnapshot elimina un snapshot existente
func (vm *QemuVM) DeleteSnapshot(id string) error {
	if vm.config == nil {
		return errors.New("VM no configurada")
	}

	snapDir := filepath.Join(os.Getenv("HOME"), "qemu", "snapshots")

	// Verificar existencia del snapshot
	metaPath := filepath.Join(snapDir, id+".meta")
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		return fmt.Errorf("snapshot no encontrado: %s", id)
	}

	// Eliminar snapshot en memoria o disco
	if vm.config.SnapshotsInMemory {
		// TODO: Implementar lógica para snapshots en memoria
	} else {
		// Eliminar snapshot en disco
		snapPath := filepath.Join(snapDir, id+".qcow2")
		err := os.Remove(snapPath)
		if err != nil {
			return fmt.Errorf("error eliminando snapshot en disco: %v", err)
		}
	}

	// Eliminar metadatos
	err := os.Remove(metaPath)
	if err != nil {
		return fmt.Errorf("error eliminando metadatos: %v", err)
	}

	return nil
}

// ListCachedImages lista las imágenes almacenadas en caché
func (vm *QemuVM) ListCachedImages() ([]string, error) {
	if vm.config == nil {
		return nil, errors.New("VM no configurada")
	}

	cacheDir := filepath.Join(os.Getenv("HOME"), "qemu", "img")
	files, err := os.ReadDir(cacheDir)
	if err != nil {
		return nil, fmt.Errorf("error leyendo directorio de caché: %v", err)
	}

	var images []string
	for _, file := range files {
		if !file.IsDir() {
			images = append(images, file.Name())
		}
	}

	return images, nil
}

// DeleteCachedImage elimina una imagen de la caché
func (vm *QemuVM) DeleteCachedImage(name string) error {
	if vm.config == nil {
		return errors.New("VM no configurada")
	}

	cacheDir := filepath.Join(os.Getenv("HOME"), "qemu", "img")
	imgPath := filepath.Join(cacheDir, name)

	// Verificar existencia de la imagen
	if _, err := os.Stat(imgPath); os.IsNotExist(err) {
		return fmt.Errorf("imagen no encontrada: %s", name)
	}

	// Eliminar archivo de caché
	err := os.Remove(imgPath)
	if err != nil {
		return fmt.Errorf("error eliminando imagen: %v", err)
	}

	return nil
}
