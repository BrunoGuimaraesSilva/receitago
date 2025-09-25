-- Create separate schema for reference tables
CREATE SCHEMA IF NOT EXISTS dictionaries;

-- Dictionary tables
CREATE TABLE dictionaries.cnaes (
    id BIGINT PRIMARY KEY,
    description TEXT NOT NULL
);

CREATE TABLE dictionaries.motivos (
    id BIGINT PRIMARY KEY,
    description TEXT NOT NULL
);

CREATE TABLE dictionaries.qualificacoes (
    id BIGINT PRIMARY KEY,
    description TEXT NOT NULL
);

CREATE TABLE dictionaries.municipios (
    id BIGINT PRIMARY KEY,
    description TEXT NOT NULL
);

CREATE TABLE dictionaries.paises (
    id BIGINT PRIMARY KEY,
    description TEXT NOT NULL
);

CREATE TABLE dictionaries.naturezas (
    id BIGINT PRIMARY KEY,
    description TEXT NOT NULL
);

-- Main business table (separate from dictionaries)
CREATE TABLE public.cnpjs (
    id SERIAL PRIMARY KEY,
    cnpj CHAR(14) UNIQUE NOT NULL
);
