'use client'

import { useState } from 'react'

import DatasetDetail from '@/components/catalog/DatasetDetail'
import DatasetList from '@/components/catalog/DatasetList'
import PageTemplate from '@/components/layout/PageTemplate'
import Button from '@/components/ui/Button'

export default function CatalogPage() {
  const [selectedDatasetId, setSelectedDatasetId] = useState<string | null>(null)

  return (
    <PageTemplate
      title="Data Catalog"
      description="Browse datasets, fields, owners, descriptions. Semantic and relational discovery."
      archetype="search"
      actions={
        selectedDatasetId ? (
          <Button variant="outline" onClick={() => setSelectedDatasetId(null)}>
            Back to List
          </Button>
        ) : undefined
      }
    >
      {selectedDatasetId ? (
        <DatasetDetail datasetId={selectedDatasetId} />
      ) : (
        <DatasetList onSelectDataset={setSelectedDatasetId} />
      )}
    </PageTemplate>
  )
}
