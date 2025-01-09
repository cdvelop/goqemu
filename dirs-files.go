package goqemu

import (
	"os"
	"path/filepath"
)

// createQemuDirs crea los directorios base para QEMU si no existen
func createQemuDirs() error {
	// Crear directorio de imágenes si no existe
	imgDir := filepath.Join(os.Getenv("HOME"), "qemu", "img")
	if _, err := os.Stat(imgDir); os.IsNotExist(err) {
		err := os.MkdirAll(imgDir, 0755)
		if err != nil {
			return err
		}
	}

	// Crear directorio de snapshots si no existe
	snapDir := filepath.Join(os.Getenv("HOME"), "qemu", "snapshots")
	if _, err := os.Stat(snapDir); os.IsNotExist(err) {
		err := os.MkdirAll(snapDir, 0755)
		if err != nil {
			return err
		}
	}

	return nil
}
