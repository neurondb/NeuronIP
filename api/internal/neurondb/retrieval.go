package neurondb

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

/* HybridSearchFusionFromQueries runs semantic and lexical queries, then fuses results using neurondb.hybrid_search_fusion when available.
 * semQuery and lexQuery must each return two columns: id (UUID), score (float).
 * Returns list of map with id, combined_score, and optionally semantic_score, lexical_score.
 * If the fusion function is not available, returns (nil, nil) so caller can fall back to manual hybrid.
 */
func (c *Client) HybridSearchFusionFromQueries(ctx context.Context, semQuery string, semArgs []interface{}, lexQuery string, lexArgs []interface{}, alpha float64, limit int) ([]map[string]interface{}, error) {
	if !c.vectorOpsEnabled() {
		return nil, fmt.Errorf("NeuronDB vector ops disabled by config")
	}
	ok, err := c.HasFunction(ctx, "neurondb", "hybrid_search_fusion")
	if err != nil || !ok {
		return nil, nil
	}
	if limit <= 0 {
		limit = 10
	}

	conn, err := c.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()
	tx, err := conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `CREATE TEMP TABLE IF NOT EXISTS _ndb_hybrid_sem(id uuid, score float) ON COMMIT DROP`)
	if err != nil {
		return nil, fmt.Errorf("create temp sem table: %w", err)
	}
	_, err = tx.Exec(ctx, `CREATE TEMP TABLE IF NOT EXISTS _ndb_hybrid_lex(id uuid, score float) ON COMMIT DROP`)
	if err != nil {
		return nil, fmt.Errorf("create temp lex table: %w", err)
	}

	_, err = tx.Exec(ctx, `TRUNCATE _ndb_hybrid_sem`)
	if err != nil {
		return nil, err
	}
	_, err = tx.Exec(ctx, `TRUNCATE _ndb_hybrid_lex`)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(ctx, `INSERT INTO _ndb_hybrid_sem (id, score) (`+semQuery+`)`, semArgs...)
	if err != nil {
		return nil, fmt.Errorf("insert semantic results: %w", err)
	}
	_, err = tx.Exec(ctx, `INSERT INTO _ndb_hybrid_lex (id, score) (`+lexQuery+`)`, lexArgs...)
	if err != nil {
		return nil, fmt.Errorf("insert lexical results: %w", err)
	}

	q := `SELECT * FROM neurondb.hybrid_search_fusion('_ndb_hybrid_sem', '_ndb_hybrid_lex', 'id', 'score', 'score', $1) ORDER BY combined_score DESC LIMIT $2`
	rows, err := tx.Query(ctx, q, alpha, limit)
	if err != nil {
		return nil, fmt.Errorf("hybrid_search_fusion: %w", err)
	}
	defer rows.Close()
	out, err := rowsToMapSlice(ctx, rows)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return out, nil
}

/* RerankWithRRF runs reciprocal_rank_fusion on multiple ranking tables when available.
 * Each table must have idCol and rankCol. rankTableNames is a list of table names.
 * kParam is the RRF constant (e.g. 60). Returns id and fused score.
 */
func (c *Client) RerankWithRRF(ctx context.Context, rankTableNames []string, idCol, rankCol string, kParam int) ([]map[string]interface{}, error) {
	if len(rankTableNames) == 0 {
		return nil, nil
	}
	ok, err := c.HasFunction(ctx, "neurondb", "reciprocal_rank_fusion")
	if err != nil || !ok {
		return nil, nil
	}
	if !isSafeIdentifier(idCol) || !isSafeIdentifier(rankCol) {
		return nil, fmt.Errorf("invalid id or rank column name")
	}
	for _, t := range rankTableNames {
		if !isSafeIdentifier(t) {
			return nil, fmt.Errorf("invalid table name: %s", t)
		}
	}

	query := `SELECT * FROM neurondb.reciprocal_rank_fusion($1::text[], $2, $3, $4)`
	rows, err := c.pool.Query(ctx, query, rankTableNames, idCol, rankCol, kParam)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return rowsToMapSlice(ctx, rows)
}

/* RerankWithMMR runs MMR reranking on a table with a vector column when neurondb.mmr_rerank_with_scores is available.
 * Returns rows with id, content (or full row), score. topK is the number of results, lambda balances relevance (1.0) vs diversity (0.0).
 */
func (c *Client) RerankWithMMR(ctx context.Context, tableName, vectorCol, queryEmbedding string, topK int, lambda float64) ([]map[string]interface{}, error) {
	if !c.vectorOpsEnabled() {
		return nil, fmt.Errorf("NeuronDB vector ops disabled by config")
	}
	ok, err := c.HasFunction(ctx, "neurondb", "mmr_rerank_with_scores")
	if err != nil || !ok {
		return nil, nil
	}
	if topK <= 0 {
		topK = 10
	}
	if !isSafeIdentifier(tableName) || !isSafeIdentifier(vectorCol) {
		return nil, fmt.Errorf("invalid table or vector column name")
	}

	query := "SELECT * FROM neurondb.mmr_rerank_with_scores($1, $2, $3::vector, $4, $5)"
	rows, err := c.pool.Query(ctx, query, tableName, vectorCol, queryEmbedding, topK, lambda)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return rowsToMapSlice(ctx, rows)
}

/* UseNeuronDBHybridFusion returns true if neurondb.hybrid_search_fusion is available (for callers to choose path). */
func (c *Client) UseNeuronDBHybridFusion(ctx context.Context) bool {
	ok, _ := c.HasFunction(ctx, "neurondb", "hybrid_search_fusion")
	return ok
}

/* UseNeuronDBRRF returns true if neurondb.reciprocal_rank_fusion is available. */
func (c *Client) UseNeuronDBRRF(ctx context.Context) bool {
	ok, _ := c.HasFunction(ctx, "neurondb", "reciprocal_rank_fusion")
	return ok
}

/* UseNeuronDBMMR returns true if neurondb.mmr_rerank_with_scores is available. */
func (c *Client) UseNeuronDBMMR(ctx context.Context) bool {
	ok, _ := c.HasFunction(ctx, "neurondb", "mmr_rerank_with_scores")
	return ok
}
