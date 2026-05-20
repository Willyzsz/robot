package excel

import (
	"strings"

	"github.com/xuri/excelize/v2"
)

const (
	HeaderMarcaTemporal = "Marca temporal"
	HeaderEmailLeader   = "Correo electrónico"
	HeaderCategory      = "Categoría a participar"
	HeaderNameTeam      = "Nombre del equipo"
	HeaderMember        = "Nombre de los integrantes del equipo. Los equipos deberán estar conformados de mínimo dos y hasta cuatro integrantes."
	HeaderEmailMember   = "Correos de los Integrantes del equipo"
	HeaderSchool        = "Si eres alumno de la UTNC captura el nombre de tu carrera y tu grupo, si eres de otra institución compártenos donde estudias"
	HeaderGrade         = "¿Nivel de escolaridad que cursa?"
	HeaderNameLeader    = "Nombre del capitán del equipo"
	HeaderTeacher       = "Nombre del asesor"
)

type FormRow struct {
	EmailLeader  string
	Category     string
	NameTeam     string
	Members      []string
	EmailMembers []string
	School       string
	Grade        string
	NameLeader   string
	Teacher      string
}

func ParseRows(path string) ([]FormRow, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	sheet := f.GetSheetName(0)
	return getRows(f, sheet)
}

func getRows(f *excelize.File, sheet string) ([]FormRow, error) {
	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return []FormRow{}, nil
	}

	headers := headerIndexes(rows[0])
	formRows := make([]FormRow, 0, len(rows)-1)

	for _, row := range rows[1:] {
		formRows = append(formRows, FormRow{
			EmailLeader:  cell(row, headers, HeaderEmailLeader),
			Category:     cell(row, headers, HeaderCategory),
			NameTeam:     cell(row, headers, HeaderNameTeam),
			Members:      splitCell(cell(row, headers, HeaderMember)),
			EmailMembers: splitCell(cell(row, headers, HeaderEmailMember)),
			School:       cell(row, headers, HeaderSchool),
			Grade:        cell(row, headers, HeaderGrade),
			NameLeader:   cell(row, headers, HeaderNameLeader),
			Teacher:      cell(row, headers, HeaderTeacher),
		})
	}

	return formRows, nil
}

func headerIndexes(headers []string) map[string]int {
	indexes := make(map[string]int, len(headers))
	for index, header := range headers {
		indexes[normalizeHeader(header)] = index
	}

	return indexes
}

func cell(row []string, headers map[string]int, header string) string {
	index, ok := headers[normalizeHeader(header)]
	if !ok || index >= len(row) {
		return ""
	}

	return strings.TrimSpace(row[index])
}

func normalizeHeader(header string) string {
	return strings.Join(strings.Fields(header), " ")
}

func splitCell(value string) []string {
	parts := strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == '\n' || r == '\r'
	})

	values := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		values = append(values, part)
	}

	return values
  }