-- Create schema for tax regime datasets
CREATE SCHEMA IF NOT EXISTS tributario;

CREATE TABLE tributario.regimes (
    id BIGSERIAL PRIMARY KEY,
    ano INT NOT NULL CHECK (ano >= 1900 AND ano <= 2100),
    cnpj CHAR(14) NOT NULL,            
    cnpj_da_scp TEXT,                  
    forma_de_tributacao VARCHAR(100) NOT NULL,
    quantidade_de_escrituracoes INT NOT NULL,
    dataset VARCHAR(50) NOT NULL       
);

CREATE INDEX idx_regimes_cnpj ON tributario.regimes(cnpj);
CREATE INDEX idx_regimes_dataset ON tributario.regimes(dataset);
CREATE INDEX idx_regimes_year ON tributario.regimes(ano);
