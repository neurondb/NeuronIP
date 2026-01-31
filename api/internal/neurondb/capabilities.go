package neurondb

import (
	"context"
	"fmt"
	"regexp"

	"github.com/jackc/pgx/v5/pgxpool"
)

/* ExtensionVersion returns the installed NeuronDB extension version, or empty if not installed */
func (c *Client) ExtensionVersion(ctx context.Context) (string, error) {
	var version string
	err := c.pool.QueryRow(ctx, `SELECT extversion FROM pg_catalog.pg_extension WHERE extname = $1`, "neurondb").Scan(&version)
	if err != nil {
		return "", err
	}
	return version, nil
}

/* HasSchema returns true if the given schema exists */
func (c *Client) HasSchema(ctx context.Context, schemaName string) (bool, error) {
	if !isSafeIdentifier(schemaName) {
		return false, fmt.Errorf("invalid schema name: %s", schemaName)
	}
	var exists bool
	err := c.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM pg_catalog.pg_namespace WHERE nspname = $1)`, schemaName).Scan(&exists)
	return exists, err
}

/* HasFunction returns true if a function with the given schema and name exists */
func (c *Client) HasFunction(ctx context.Context, schemaName, funcName string) (bool, error) {
	if !isSafeIdentifier(schemaName) || !isSafeIdentifier(funcName) {
		return false, fmt.Errorf("invalid schema or function name")
	}
	var exists bool
	query := `SELECT EXISTS(
		SELECT 1 FROM pg_catalog.pg_proc p
		JOIN pg_catalog.pg_namespace n ON p.pronamespace = n.oid
		WHERE n.nspname = $1 AND p.proname = $2
	)`
	err := c.pool.QueryRow(ctx, query, schemaName, funcName).Scan(&exists)
	return exists, err
}

/* HasFunctionInPublic returns true if a function with the given name exists in the public schema (e.g. neurondb_embed) */
func (c *Client) HasFunctionInPublic(ctx context.Context, funcName string) (bool, error) {
	return c.HasFunction(ctx, "public", funcName)
}

/* isSafeIdentifier allows only identifiers that are safe for dynamic SQL (alphanumeric and underscore) */
var safeIdentifierRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

func isSafeIdentifier(s string) bool {
	return len(s) > 0 && len(s) <= 128 && safeIdentifierRegex.MatchString(s)
}

/* QuoteIdentifier double-quotes an identifier for use in dynamic SQL; returns error if not safe */
func QuoteIdentifier(name string) (string, error) {
	if !isSafeIdentifier(name) {
		return "", fmt.Errorf("invalid identifier: %s", name)
	}
	return `"` + name + `"`, nil
}

/* QuoteTableName returns "schema"."table" for use in dynamic SQL */
func QuoteTableName(schema, table string) (string, error) {
	sq, err := QuoteIdentifier(schema)
	if err != nil {
		return "", err
	}
	tq, err := QuoteIdentifier(table)
	if err != nil {
		return "", err
	}
	return sq + "." + tq, nil
}

/* Pool returns the underlying pool (for tests or when caller needs raw access) */
func (c *Client) Pool() *pgxpool.Pool {
	return c.pool
}

/* UseMultimodalEmbedding returns true if config allows and neurondb_embed_multimodal exists */
func (c *Client) UseMultimodalEmbedding(ctx context.Context) bool {
	if !c.multimodalEnabled() {
		return false
	}
	ok, _ := c.HasFunctionInPublic(ctx, "neurondb_embed_multimodal")
	return ok
}

/* UseImageEmbedding returns true if config allows and neurondb_embed_image exists */
func (c *Client) UseImageEmbedding(ctx context.Context) bool {
	if !c.imageEmbeddingsEnabled() {
		return false
	}
	ok, _ := c.HasFunctionInPublic(ctx, "neurondb_embed_image")
	return ok
}
