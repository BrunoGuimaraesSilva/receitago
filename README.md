# ReceitaGo

ReceitaGo is a Go project that downloads and processes open datasets from:

* 📊 **Receita Federal do Brasil** (CNPJ open data)
* 💰 **Tesouro Nacional** (SIAFI dataset)

The project follows **DDD** and **Clean Architecture**.

---

## 🚀 How to Run

```bash
go run ./cmd/api
```

Server will start at:

```
http://localhost:8080
```

---

## 🔗 Endpoints

* `GET /download/receita` → Downloads the latest Receita CNPJ dataset (big `.zip` files).
* `GET /download/tesouro` → Downloads the Tesouro Nacional dataset (`.csv` files).

---

## 📂 Output

* Files are stored in `./data/receita/` and `./data/tesouro/`.
* Metadata (last version downloaded) is stored in `./data/meta/`.

---

## 📊 Example Response

```json
[
  {
    "id": "receita-2025-09-0",
    "filename": "Empresas0.zip",
    "success": true,
    "download_time": "45.2s",
    "unzip_time": "12.4s",
    "attempts": 1
  }
]
```

---

## 👨‍💻 Author

Bruno Guimarães Silva