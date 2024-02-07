package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type Curso struct {
	Nombre      [25]byte
	Codigo      int32
	Obligatorio int16
	Id          int32
}

var archivoBinario string = "./cursos.bin"

func main() {
	opcion := 0
	salir := false

	if err := CrearArchivo(archivoBinario); err != nil {
		return
	}

	for !salir {
		limpiarConsola()
		fmt.Println("1. Registro de Curso")
		fmt.Println("2. Ver Cursos")
		fmt.Println("3. Salir")
		fmt.Scanln(&opcion)
		switch opcion {
		case 1:
			RegistroCurso()
		case 2:
			VerCursos()
		case 3:
			salir = true
		}
	}
}

func validarObligatorio(input string) bool {
	return input == "S" || input == "s" || input == "N" || input == "n"
}

func validarCodigo(input string) (int, bool) {
	num, err := strconv.Atoi(input)
	return num, err == nil
}

func validarNombre(input string) bool {
	if len(input) > 25 {
		return false
	}
	match, _ := regexp.MatchString(`^[a-zA-Z0-9\s]+$`, input)
	return match
}

func RegistroCurso() {
	reader := bufio.NewReader(os.Stdin)

	lineaDoble(30)
	fmt.Println("Registro de Curso")
	lineaDoble(30)
	lineaEnBlanco()

	var nuevoCurso = new(Curso)
	var obligatorio string
	var nombre string
	var codigo int
	for {
		fmt.Println("¿Es obligatorio? (S/N)")
		obligatorio, _ = reader.ReadString('\n')
		obligatorio = strings.TrimSpace(obligatorio)
		if validarObligatorio(obligatorio) {
			break
		} else {
			fmt.Println("Entrada inválida. Por favor, ingresa 'S', 's', 'N', o 'n'.")
		}
	}

	for {
		fmt.Println("Ingresa el código")
		codigoInput, _ := reader.ReadString('\n')
		codigoInput = strings.TrimSpace(codigoInput)
		var valid bool
		codigo, valid = validarCodigo(codigoInput)
		if valid {
			break
		} else {
			fmt.Println("Entrada inválida. Por favor, ingresa un número entero.")
		}
	}

	for {
		fmt.Println("Ingresa el nombre del curso")
		nombre, _ = reader.ReadString('\n')
		nombre = strings.TrimSpace(nombre)
		if validarNombre(nombre) {
			break
		} else {
			fmt.Println("Nombre inválido. Debe ser alfanumérico y hasta 25 caracteres.")
		}
	}

	lineaDoble(30)

	if obligatorio == "S" || obligatorio == "s" {
		obligatorio = "true"
	} else {
		obligatorio = "false"
	}
	obligatorioBool, _ := strconv.ParseBool(obligatorio)
	if obligatorioBool {
		nuevoCurso.Obligatorio = 1
	} else {
		nuevoCurso.Obligatorio = 0
	}
	nuevoCurso.Id = ProximoId()
	nuevoCurso.Codigo = int32(codigo)
	copy(nuevoCurso.Nombre[:], nombre)

	archivo, err := AbrirArchivo(archivoBinario)
	if err != nil {
		fmt.Println("Error abriendo el archivo para registrar el nuevo curso:", err)
		return
	}
	defer archivo.Close()

	// Ir al final del archivo para agregar el nuevo curso.
	archivo.Seek(0, io.SeekEnd)

	if err := binary.Write(archivo, binary.LittleEndian, nuevoCurso); err != nil {
		fmt.Println("Error registrando el nuevo curso:", err)
		return
	}
}

func CrearArchivo(nombre string) error {
	dir := filepath.Dir(nombre)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		fmt.Println("Err CrearArchivo dir==", err)
		return err
	}

	// Crear archivo
	if _, err := os.Stat(nombre); os.IsNotExist(err) {
		file, err := os.Create(nombre)
		if err != nil {
			fmt.Println("Err CrearArchivo create==", err)
			return err
		}
		defer file.Close()
	}
	return nil
}

func AbrirArchivo(nombre string) (*os.File, error) {
	archivo, err := os.OpenFile(nombre, os.O_RDWR, 0644)
	if err != nil {
		fmt.Println("Err AbrirArchivo==", err)
		return nil, err
	}
	return archivo, nil
}

func RegistrarCursoArchivo(archivo *os.File, datos interface{}, posicion int64) error {
	archivo.Seek(posicion, 0)
	err := binary.Write(archivo, binary.LittleEndian, datos)
	if err != nil {
		fmt.Println("Err EscribirArchivo==", err)
		return err
	}
	return nil
}

func ProximoId() int32 {
	archivo, err := os.Open(archivoBinario)
	if err != nil {
		fmt.Println("Error abriendo el archivo para calcular el proximo ID:", err)
		return -1
	}
	defer archivo.Close()

	fileInfo, err := archivo.Stat()
	if err != nil {
		fmt.Println("Error obteniendo las estadisticas del archivo:", err)
		return -1
	}

	fileSize := fileInfo.Size()
	var curso Curso
	cursoSize := int64(binary.Size(curso))

	// Calcular la siguiente ID según el tamaño del archivo dividido por el tamaño de un registro de Curso
	nextID := int32(fileSize/cursoSize) + 1
	return nextID
}

func VerCursos() {
	lineaDoble(60)
	fmt.Println("Cursos Registrados")
	lineaDoble(60)
	fmt.Printf("%-5s %-10s %-30s %-12s\n", "ID", "Código", "Nombre", "Obligatorio")

	archivo, err := AbrirArchivo(archivoBinario)
	if err != nil {
		fmt.Println("Error abriendo el archivo:", err)
		return
	}
	defer archivo.Close()

	// Buscar al inicio del archivo.
	archivo.Seek(0, 0)

	var cursos []Curso // Slice para almacenar los Cursos leidos

	for {
		var cursoLeido Curso
		err := binary.Read(archivo, binary.LittleEndian, &cursoLeido)
		if err == io.EOF { // Fin de archivo, salir del ciclo
			break
		}
		if err != nil {
			fmt.Println("Error leyendo el curso:", err)
			return
		}
		cursos = append(cursos, cursoLeido)
	}

	if len(cursos) == 0 {
		fmt.Println("No hay cursos registrados.")
		Pausa()
		return
	}

	// Si hay cursos registrados, mostrarlos en un formato tabular
	for _, curso := range cursos {
		PrintCurso(curso)
	}
	lineaDoble(60)

	Pausa()
}

func LeerCursoArchivo(archivo *os.File, posicion int64) (*Curso, error) {
	archivo.Seek(posicion, 0)

	var curso Curso
	err := binary.Read(archivo, binary.LittleEndian, &curso)
	if err != nil {
		fmt.Println("Err LeerCursoArchivo==", err)
		return nil, err
	}

	return &curso, nil
}

func PrintCurso(curso Curso) {
	nombreStr := strings.TrimRight(string(curso.Nombre[:]), "\x00")
	nombrePadded := RellenarDerecha(nombreStr, 30, ' ')
	obligatorioStr := "No"
	if curso.Obligatorio == 1 {
		obligatorioStr = "Si"
	}
	fmt.Printf("%-5d %-10d %s %-12s\n", curso.Id, curso.Codigo, nombrePadded, obligatorioStr)
}

func Pausa() {
	mostrarMensaje("Presione ENTER para continuar...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func RellenarDerecha(str string, length int, padChar rune) string {
	padCountInt := length - len(str)
	if padCountInt > 0 {
		return str + strings.Repeat(string(padChar), padCountInt)
	}
	return str
}

func lineaEnBlanco() {
	fmt.Println("")
}

func mostrarMensaje(mensaje string) {
	lineaEnBlanco()
	fmt.Println(mensaje)
	lineaEnBlanco()
}

func lineaDoble(longitud int) {
	fmt.Println(strings.Repeat("=", longitud))
}

func limpiarConsola() {
	fmt.Print("\033[H\033[2J")
}
