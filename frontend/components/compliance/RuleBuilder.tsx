'use client'

import { useState } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import Button from '@/components/ui/Button'

interface RuleBuilderProps {
  onSave?: (rules: any[]) => void
}

export default function RuleBuilder({ onSave }: RuleBuilderProps) {
  const [rules, setRules] = useState<any[]>([])
  const [showAddForm, setShowAddForm] = useState(false)
  const [newRule, setNewRule] = useState({
    pattern: '',
    pattern_type: 'substring',
    keywords: [] as string[],
  })

  const handleAddRule = () => {
    if (!newRule.pattern && newRule.keywords.length === 0) {
      return
    }

    setRules([...rules, { ...newRule }])
    setNewRule({ pattern: '', pattern_type: 'substring', keywords: [] })
    setShowAddForm(false)
  }

  const handleRemoveRule = (index: number) => {
    setRules(rules.filter((_, i) => i !== index))
  }

  const handleSave = () => {
    onSave?.(rules)
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle>Rule Builder</CardTitle>
            <CardDescription>Build custom compliance rules</CardDescription>
          </div>
          <Button onClick={() => setShowAddForm(!showAddForm)} size="sm">
            Add Rule
          </Button>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        {showAddForm && (
          <div className="p-4 rounded-lg border border-border space-y-3">
            <div>
              <label className="text-sm font-medium mb-2 block">Pattern Type</label>
              <select
                value={newRule.pattern_type}
                onChange={(e) => setNewRule({ ...newRule, pattern_type: e.target.value })}
                className="w-full rounded-lg border border-border bg-background px-4 py-2 focus:outline-none focus:ring-2 focus:ring-ring"
              >
                <option value="substring">Substring</option>
                <option value="regex">Regular Expression</option>
                <option value="keyword">Keyword</option>
                <option value="exact">Exact Match</option>
              </select>
            </div>
            <div>
              <label className="text-sm font-medium mb-2 block">Pattern</label>
              <input
                type="text"
                value={newRule.pattern}
                onChange={(e) => setNewRule({ ...newRule, pattern: e.target.value })}
                placeholder="Enter pattern..."
                className="w-full rounded-lg border border-border bg-background px-4 py-2 focus:outline-none focus:ring-2 focus:ring-ring"
              />
            </div>
            <div className="flex gap-2 justify-end">
              <Button type="button" variant="outline" onClick={() => setShowAddForm(false)}>
                Cancel
              </Button>
              <Button type="button" onClick={handleAddRule}>
                Add
              </Button>
            </div>
          </div>
        )}

        {rules.length === 0 ? (
          <div className="text-center py-8 text-muted-foreground">No rules defined</div>
        ) : (
          <div className="space-y-2">
            {rules.map((rule, index) => (
              <div key={index} className="p-3 rounded-lg border border-border flex items-center justify-between">
                <div>
                  <span className="text-sm font-medium">{rule.pattern_type}</span>
                  {rule.pattern && <span className="ml-2 text-sm text-muted-foreground">{rule.pattern}</span>}
                </div>
                <Button
                  onClick={() => handleRemoveRule(index)}
                  variant="ghost"
                  size="sm"
                >
                  Remove
                </Button>
              </div>
            ))}
          </div>
        )}

        {rules.length > 0 && onSave && (
          <div className="flex justify-end">
            <Button onClick={handleSave}>Save Rules</Button>
          </div>
        )}
      </CardContent>
    </Card>
  )
}
