package models

// ImportRequest representa uma requisição de importação (para uso futuro)
type ImportRequest struct {
	Source    string   `json:"source" example:"receita"`
	Files     []string `json:"files,omitempty" example:"empresas,estabelecimentos"`
	Overwrite bool     `json:"overwrite" example:"false"`
}
