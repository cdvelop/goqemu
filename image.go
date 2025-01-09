package goqemu

import (
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
)

// getImagePath devuelve la ruta local de la imagen descargada
func getImagePath(url string) string {
	// Usar el nombre del archivo de la URL como nombre local
	fileName := path.Base(url)
	return filepath.Join(os.Getenv("HOME"), "qemu", "img", fileName)
}

// downloadImage descarga una imagen desde una URL
func downloadImage(url, dest string) error {

	// Crear directorio si no existe
	err := os.MkdirAll(filepath.Dir(dest), 0755)
	if err != nil {
		return err
	}

	// Crear archivo destino
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	// Descargar imagen
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Copiar contenido
	_, err = io.Copy(out, resp.Body)
	return err
}
