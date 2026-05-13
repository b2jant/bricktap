# 🧱 Bricktap

[![Go Reference](https://pkg.go.dev/badge/github.com/b2jant/bricktap.svg)](https://pkg.go.dev/github.com/b2jant/bricktap)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**Bricktap** is a blazing-fast, interactive CLI tool written in Go that generates target-specific SQL code from YAML semantic models. 

Tired of writing boilerplate SQL for your data transformations? Bricktap allows you to define your business logic once in a generic YAML format and seamlessly compile it into data-warehouse-specific SQL for modern data orchestration frameworks.

---

## ✨ Features

- **Interactive TUI**: Built with [Bubbletea](https://github.com/charmbracelet/bubbletea), Bricktap provides a beautiful, interactive Terminal UI to configure your code generation. No more remembering complex CLI flags!
- **Multi-Framework Support**: Generate code ready to be dropped into:
  - 🛠️ [dbt](https://www.getdbt.com/)
  - 🚀 [SQLMesh](https://sqlmesh.com/)
  - 🐻 [Bruin](https://bruin.data/)
- **Dialect Aware**: Emits optimized, valid SQL for your specific data warehouse:
  - ❄️ Snowflake
  - 🔍 BigQuery
  - 🐘 PostgreSQL
- **Global Type Casting Rules**: Automatically inject type casts, trims, or null handling defined globally for your target dialect.

## 🚀 Installation

Ensure you have [Go](https://go.dev/) (1.20+) installed.

### Option 1: Install via `go install`
```bash
go install github.com/b2jant/bricktap/cmd/bricktap@latest
```

### Option 2: Clone and Build manually
```bash
git clone https://github.com/b2jant/bricktap.git
cd bricktap
go build -o bricktap ./cmd/bricktap
```

## 💻 Usage

Bricktap expects a specific directory structure to operate out-of-the-box. 

1. Create a `semantic_models/` directory in your current path and add your `.yaml` definitions there.
2. Run `bricktap`.

```bash
bricktap
```

You will be greeted by the interactive configuration menu:
1. Select your **Target Framework** (dbt, SQLMesh, Bruin).
2. Select your **Target Data Warehouse** (Snowflake, BigQuery, Postgres).
3. Watch the magic happen as your models are parsed and written to the `./models` directory!

### Example Directory Structure
```text
.
├── semantic_models/
│   └── sales/
│       └── orders.yaml      # Your semantic definition
└── models/                  # Generated files will appear here
```

## 🏗️ Architecture

Bricktap operates using an extensible internal architecture:

1. **Scanner**: Recursively finds `.yaml` files in the input directory.
2. **Parser**: Reads YAML files and converts them into an Internal Representation (`core.IR`).
3. **Dialects**: Applies SQL-dialect-specific syntax formatting to the IR (e.g., Snowflake-specific macros vs. Postgres syntax).
4. **Adapters (Generators)**: Takes the finalized IR and writes out the framework-specific files (e.g., creating both an `orders.sql` and `orders_schema.yml` for dbt).

## 🤝 Contributing

Contributions are welcome! If you'd like to add a new Dialect, a new Generator Framework, or improve the Parser:

1. Fork the repository.
2. Create your feature branch (`git checkout -b feature/AmazingFeature`).
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`).
4. **Run the tests locally** (`go test ./...`) and ensure they pass. The maintainer will check out your branch and verify the tests manually before merging.
5. Push to the branch (`git push origin feature/AmazingFeature`).
6. Open a Pull Request.

## 📝 License

Distributed under the MIT License. See `LICENSE` for more information.
