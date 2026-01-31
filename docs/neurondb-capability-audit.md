# NeuronDB Capability Audit

Inventory of NeuronDB usage in NeuronIP vs capabilities available in `/Users/ibrarahmed/pgelephant/pge/neurondb/NeuronDB/sql/`.

## 1. NeuronDB functions currently used in NeuronIP

| Function / capability | Where defined (NeuronIP) | Where called | Fallback |
|------------------------|--------------------------|--------------|----------|
| `neurondb_embed` | `api/internal/neurondb/client.go` | semantic, rag, support, warehouse, workflows, agent/memory, compliance/policy, knowledgegraph, catalog, models, pipeline, dataquality | None |
| `neurondb_embed_batch` | `api/internal/neurondb/client.go` | support, semantic (batch chunk embedding), catalog | Fallback to per-item `GenerateEmbedding` in client |
| `neurondb_embed_multimodal` | `api/internal/neurondb/client.go` | semantic (multimodal doc embedding) | Fallback to text-only embed in semantic |
| `neurondb_embed_image` | `api/internal/neurondb/client.go` | semantic (image embedding) | Fallback to text embed |
| `neurondb_classify` | `api/internal/neurondb/client.go` | models, compliance/anomaly, dataquality, knowledgegraph, profiling | None |
| `neurondb_regress` | `api/internal/neurondb/client.go` | models, warehouse (anomaly score) | None |
| `neurondb_train_model` | `api/internal/neurondb/client.go` | models | None |
| `neurondb_predict` / `neurondb_predict_batch` | `api/internal/neurondb/client.go` | models | None |
| `neurondb_evaluate_model` | `api/internal/neurondb/client.go` | models | None |
| `neurondb_process_document` | `api/internal/neurondb/client.go` | semantic (chunking) | `simpleChunkDocument` in client |
| `neurondb_generate_response` | `api/internal/neurondb/client.go` | semantic, rag | None |
| `neurondb_answer_with_citations` | `api/internal/neurondb/client.go` | semantic (RAG), knowledgegraph | Fallback to `GenerateResponse` in semantic |
| `neurondb_cluster_data` | `api/internal/neurondb/client.go` | analytics | None |
| `neurondb_detect_outliers` | `api/internal/neurondb/client.go` | compliance/anomaly, profiling, dataquality | Heuristic fallbacks in some callers |
| `neurondb_reduce_dimensionality` | `api/internal/neurondb/client.go` | (client only) | None |
| `neurondb_timeseries_analysis` | `api/internal/neurondb/client.go` | analytics | None |
| `neurondb.models` catalog | `api/internal/neurondb/client.go` (ListModels, GetModelInfo, DeleteModel) | models | Empty list on missing table |
| Raw vector search (`<=>`, `<->`, `<#>`) | `api/internal/neurondb/client.go` (VectorSearch, VectorSearchL2, VectorSearchCosine, VectorSearchInnerProduct) | semantic, rag, support, warehouse, ai | Uses pgvector operators; no NeuronDB wrapper |
| Manual hybrid (semantic + keyword CTEs) | `api/internal/neurondb/client.go` (HybridSearch) | semantic, support, warehouse | Pure SQL in client; no `neurondb.hybrid_search_fusion` |
| CreateVectorIndex (HNSW/IVF) | `api/internal/neurondb/client.go` | semantic, models | None |

**Services that hold a NeuronDB client:** semantic, rag, support, warehouse, workflows, compliance (service + policy + anomaly), analytics, models, knowledgegraph, agent/memory, catalog, ai, dataquality, profiling, integrations (slack, teams).

**Server wiring:** `api/cmd/server/main.go` builds a single `neurondbClient := neurondb.NewClient(pool)` from the default `neuronip` pool; `cfg.NeuronDB` is never used.

## 2. NeuronDB features in local repo not yet used in NeuronIP

| Feature | NeuronDB SQL file(s) | Suggested use in NeuronIP |
|---------|------------------------|---------------------------|
| Hybrid fusion | `20_ml_hybrid_search.sql` (`neurondb.hybrid_search_fusion`) | Replace manual hybrid CTE in semantic/support/warehouse with fusion when available |
| LTR rerank | `20_ml_hybrid_search.sql` (`neurondb.ltr_rerank_pointwise`, `ltr_score_features`) | Rerank top-N retrieval in RAG/semantic/support |
| MMR / RRF / ensemble rerank | `16_ml_reranking.sql` (`mmr_rerank`, `reciprocal_rank_fusion`, `rerank_ensemble_weighted`, `rerank_ensemble_borda`) | Optional rerank stage after vector/hybrid search |
| Flash / long-context rerank | `24_reranking_flash.sql` (`rerank_flash`, `rerank_long_context`) | High-quality reranking when available |
| Sparse vectors + dense/sparse hybrid | `22_sparse_vectors.sql` (`sparse_vector`, `splade_embed`, `hybrid_dense_sparse_search`, `sparse_index_create`, `rrf_fusion`) | Better keyword recall; new columns + indexes |
| Drift detection | `19_ml_drift.sql` (`detect_centroid_drift`, `compute_distribution_divergence`) | dataquality, profiling, analytics |
| Quantization (FP8/INT4) | `23_quantization_fp8.sql`, `11_quantization_detail.sql` | Optional index/storage optimization |
| Vector comparison operators | `23_vector_comparison_operators.sql` | May already be covered by pgvector; verify |
| Multimodal embeddings (extension) | `25_multimodal_embeddings.sql` | Align with existing `neurondb_embed_image` / `neurondb_embed_multimodal` usage |

## 3. Config and wiring gaps

- **NeuronDBConfig** in `api/internal/config/config.go`: `EnableVectorOps`, `EnableMLOps`, `EnableRAGOps`, `EnableMultimodal`, `EnableImageEmbeddings`, `EnableBatchOps`, `EnableVectorIndexing`, `AutoCreateIndexes`, `DefaultIndexType` are loaded but never read by any code.
- **Pool:** NeuronDB client always uses the same pool as the app DB. No separate `NEURONDB_*` pool is created in `multipool.go`; docs/compose sometimes imply a separate NeuronDB database.
- **Startup:** No check that the `neurondb` extension is installed or which functions exist.

## 4. Summary

- **Used:** Embeddings (single/batch/multimodal/image), classify, regress, train/predict/evaluate, process_document, generate_response, answer_with_citations, cluster_data, detect_outliers, reduce_dimensionality, timeseries_analysis, raw vector/hybrid SQL, CreateVectorIndex, neurondb.models catalog.
- **Not used:** Hybrid fusion function, LTR/MMR/RRF/ensemble/Flash reranking, sparse vectors and dense+sparse search, drift detection, config flags, capability/version checks, and separate NeuronDB pool when desired.
