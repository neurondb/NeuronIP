'use client'

import { motion } from 'framer-motion'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import { useWarehouseSchemas } from '@/lib/api/queries'
import Loading from '@/components/ui/Loading'
import { Skeleton } from '@/components/ui/Loading'

export default function SchemaExplorer() {
  const { data: schemas, isLoading } = useWarehouseSchemas()

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <Skeleton variant="text" className="h-6 w-32" />
          <Skeleton variant="text" className="h-4 w-48 mt-2" />
        </CardHeader>
        <CardContent>
          <div className="space-y-2">
            {[1, 2, 3].map((i) => (
              <Skeleton key={i} variant="rectangular" className="h-16" />
            ))}
          </div>
        </CardContent>
      </Card>
    )
  }

  if (!schemas || (Array.isArray(schemas) && schemas.length === 0)) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Database Schemas</CardTitle>
          <CardDescription>Available database schemas</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="text-center text-muted-foreground py-8">
            No schemas available
          </div>
        </CardContent>
      </Card>
    )
  }

  const schemaList = Array.isArray(schemas) ? schemas : [schemas]

  return (
    <Card>
      <CardHeader>
        <CardTitle>Database Schemas</CardTitle>
        <CardDescription>Browse available database schemas</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-2">
          {schemaList.map((schema: any, index) => (
            <motion.div
              key={schema.id || index}
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: index * 0.1 }}
              className="p-3 rounded-lg border border-border hover:bg-accent cursor-pointer transition-colors"
            >
              <div className="font-medium">{schema.name || schema.schema_name || 'Unnamed Schema'}</div>
              {schema.description && (
                <div className="text-sm text-muted-foreground mt-1">{schema.description}</div>
              )}
            </motion.div>
          ))}
        </div>
      </CardContent>
    </Card>
  )
}
