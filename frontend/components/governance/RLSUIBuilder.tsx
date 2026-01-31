'use client'

import { useState } from 'react'

import Button from '@/components/ui/Button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import Input from '@/components/ui/Input'

export default function RLSUIBuilder() {
  const [schemaName, setSchemaName] = useState('')
  const [tableName, setTableName] = useState('')
  const [policyName, setPolicyName] = useState('')
  const [policyType, setPolicyType] = useState('select')
  const [condition, setCondition] = useState('')

  const createPolicy = async () => {
    try {
      await fetch('/api/v1/governance/rls/policies', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          schema_name: schemaName,
          table_name: tableName,
          policy_name: policyName,
          policy_type: policyType,
          condition,
        }),
      })
      alert('RLS policy created successfully')
    } catch (error) {
      console.error('Failed to create policy:', error)
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Create RLS Policy</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div>
          <label className="block text-sm font-medium mb-2">Schema Name</label>
          <Input value={schemaName} onChange={(e) => setSchemaName(e.target.value)} />
        </div>
        <div>
          <label className="block text-sm font-medium mb-2">Table Name</label>
          <Input value={tableName} onChange={(e) => setTableName(e.target.value)} />
        </div>
        <div>
          <label className="block text-sm font-medium mb-2">Policy Name</label>
          <Input value={policyName} onChange={(e) => setPolicyName(e.target.value)} />
        </div>
        <div>
          <label className="block text-sm font-medium mb-2">Policy Type</label>
          <select
            value={policyType}
            onChange={(e) => setPolicyType(e.target.value)}
            className="w-full p-2 border rounded"
          >
            <option value="select">SELECT</option>
            <option value="insert">INSERT</option>
            <option value="update">UPDATE</option>
            <option value="delete">DELETE</option>
            <option value="all">ALL</option>
          </select>
        </div>
        <div>
          <label className="block text-sm font-medium mb-2">Condition (SQL)</label>
          <Input
            value={condition}
            onChange={(e) => setCondition(e.target.value)}
            placeholder="e.g., user_id = current_user_id()"
            className="font-mono"
          />
        </div>
        <Button onClick={createPolicy}>Create Policy</Button>
      </CardContent>
    </Card>
  )
}
