package domain

import (
	"fmt"
	"strings"
	"time"
)

type Record struct {
	InscricaoEstadual string    `json:"inscricao_estadual"`
	CNPJ              string    `json:"cnpj"`
	InicioAtividade   time.Time `json:"inicio_atividade"`
	MunicipioCodigo   string    `json:"municipio_codigo"`
	AtividadeCodigo   string    `json:"atividade_codigo"`
	CNAE              string    `json:"cnae"`
}

func ParseRecord(line string) (*Record, error) {
	parts := strings.Split(line, ";")
	if len(parts) < 6 {
		return nil, fmt.Errorf("invalid line: %s", line)
	}
	t, err := time.Parse("20060102", parts[2]+"01")
	if err != nil {
		return nil, err
	}
	return &Record{
		InscricaoEstadual: parts[0],
		CNPJ:              parts[1],
		InicioAtividade:   t,
		MunicipioCodigo:   parts[3],
		AtividadeCodigo:   parts[4],
		CNAE:              parts[5],
	}, nil
}
