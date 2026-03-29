# Swagger API Documentation

This project uses **Swaggo** to automatically generate OpenAPI (Swagger) documentation from Go code annotations.

## ✨ Features

- ✅ **Automatic API Documentation**: Generated from Go code comments (annotations)
- ✅ **Interactive Swagger UI**: Browse and test endpoints at `/swagger/index.html`
- ✅ **Type-Safe Schemas**: All request/response DTOs are documented with examples
- ✅ **No Business Logic Changes**: Transport layer only - domain code remains untouched

## 📍 Access Swagger UI

Once the API is running, open your browser to:

```
http://localhost:8080/swagger/index.html
```

## 🔄 Regenerate Documentation

After making changes to handlers or DTOs, regenerate the documentation:

```bash
# Regenerate Swagger docs from annotations
swag init -g cmd/api/main.go

# Or use the Makefile
make docs
```

This will update:
- `docs/docs.go` - Embedded documentation (used at runtime)
- `docs/swagger.json` - OpenAPI 3.0 specification (JSON format)
- `docs/swagger.yaml` - OpenAPI 3.0 specification (YAML format)

## 📝 Adding Documentation to Endpoints

### Step 1: Add Handler Method Documentation

```go
// CreateCargo godoc
// @Summary Create a new cargo
// @Description Create a new cargo shipment with initial status
// @Tags cargo
// @Accept json
// @Produce json
// @Param request body CreateCargoRequest true "Cargo creation payload"
// @Success 201 {object} response.SuccessResponse{data=CargoResponse} "Cargo created successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request or validation failed"
// @Router /cargoes [post]
func (h *CargoHandler) CreateCargo(c *gin.Context) {
    // implementation...
}
```

### Step 2: Add DTO Field Documentation

```go
type CreateCargoRequest struct {
    Name        string    `json:"name" binding:"required" example:"Premium Electronics" description:"Cargo name"`
    Description string    `json:"description" example:"Laptop shipment" description:"Optional cargo description"`
    Weight      float64   `json:"weight" binding:"required,gt=0" example:"150.5" description:"Cargo weight in kg"`
    VesselID    uuid.UUID `json:"vessel_id" binding:"required" example:"123e4567-e89b-12d3-a456-426614174000" description:"UUID of the vessel carrying this cargo"`
}
```

## 📚 Documentation Annotations Reference

### Common Swagger Tags

| Tag | Purpose | Example |
|-----|---------|---------|
| `@Summary` | Short operation summary | `Create a new cargo` |
| `@Description` | Full operation description | `Create cargo with initial status` |
| `@Tags` | Group related endpoints | `cargo`, `vessel`, `tracking` |
| `@Accept` | Accepted content types | `json`, `xml` |
| `@Produce` | Response content types | `json` |
| `@Param` | Request parameters | `@Param id path string true "Cargo ID"` |
| `@Success` | Success response | `@Success 200 {object} CargoResponse` |
| `@Failure` | Error response | `@Failure 404 {object} ErrorResponse` |
| `@Router` | HTTP method and path | `@Router /cargoes [post]` |

### Parameter Types

- `path` - URL path parameter
- `query` - URL query parameter
- `body` - Request body
- `header` - HTTP header
- `formData` - Form data

### Response Types

- `{object}` - Single object (uses struct)
- `{array}` - Array of objects
- Nested example: `{object} response.SuccessResponse{data=CargoResponse}`

## 🔧 API Endpoints

### Cargo Management

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/cargoes` | Create a new cargo |
| `GET` | `/api/v1/cargoes` | List all cargoes |
| `GET` | `/api/v1/cargoes/{id}` | Get cargo by ID |
| `PATCH` | `/api/v1/cargoes/{id}/status` | Update cargo status |

### Vessel Management

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/vessels` | Create a new vessel |
| `GET` | `/api/v1/vessels` | List all vessels |
| `GET` | `/api/v1/vessels/{id}` | Get vessel by ID |
| `PATCH` | `/api/v1/vessels/{id}/location` | Update vessel location |

### Tracking

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/cargoes/{id}/tracking` | Add tracking entry |
| `GET` | `/api/v1/cargoes/{id}/tracking` | Get tracking history |

### Health Checks

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/health` | Liveness check |
| `GET` | `/ready` | Readiness check |

## 📦 Generated Files

After running `swag init`, the following files are generated:

```
docs/
├── docs.go          # Embedded documentation (imported in main.go)
├── swagger.json     # OpenAPI 3.0 specification (JSON)
└── swagger.yaml     # OpenAPI 3.0 specification (YAML)
```

- **docs.go**: Contains the embedded Swagger spec. Must be imported in main.go with `_ "path/to/docs"`
- **swagger.json/yaml**: Human-readable API specifications that can be downloaded or shared

## 🚀 Build & Run

```bash
# Build the project
make build

# Start the API server
./api

# Access Swagger UI
# Open browser: http://localhost:8080/swagger/index.html
```

## 🔍 Validation

The generated Swagger documentation validates against OpenAPI 3.0 specification. You can:

1. **Validate online**: Use [Swagger Editor](https://editor.swagger.io/) to validate `docs/swagger.json`
2. **View in local tools**: Import the JSON/YAML into tools like Insomnia, Postman, or ReDoc

## 📄 Swagger Specification

The API documentation is available in multiple formats:

- **Swagger UI** (Interactive): `http://localhost:8080/swagger/index.html`
- **JSON Format**: `http://localhost:8080/swagger/swagger.json`
- **YAML Format**: Available at `docs/swagger.yaml`

## ⚠️ Important Notes

1. **Always regenerate docs** when modifying handler annotations or DTO fields
2. **Keep annotations close to code** - makes maintenance easier
3. **Check the Swagger UI** regularly during development to catch documentation issues early
4. **Commit the docs/** folder** to git - it contains the final generated spec
5. **Do NOT manually edit** `docs/docs.go`, `swagger.json`, or `swagger.yaml` - they're generated

## 🐛 Troubleshooting

### Issue: Swagger UI shows no endpoints

**Solution**: Regenerate docs:
```bash
swag init -g cmd/api/main.go
```

### Issue: DTO fields not showing in schema

**Solution**: Ensure DTO struct fields are **exported** (capitalized) and have proper tags:
```go
type CargoResponse struct {
    ID    uuid.UUID `json:"id" example:"123e4567..." description:"Cargo ID"`  // ✅ Good
}
```

### Issue: Swagger UI returns 404

**Solution**: Ensure the route is registered in main.go:
```go
engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
```

## 🔗 References

- [Swaggo Documentation](https://github.com/swaggo/swag)
- [Gin-Swagger Integration](https://github.com/swaggo/gin-swagger)
- [OpenAPI 3.0 Specification](https://spec.openapis.org/oas/v3.0.0)
- [Swagger Annotation Guide](https://swaggo.github.io/docs/declarative_comments_format/)

---

**Last Updated**: 2026-03-29  
**Swagger Version**: 1.16.6  
**Gin-Swagger Version**: 1.6.1
