'use client'

import { useState } from 'react'
import { motion } from 'framer-motion'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import MetricCard from '@/components/dashboard/MetricCard'
import LineChart from '@/components/charts/LineChart'
import ChartContainer from '@/components/charts/ChartContainer'
import { useQualityDashboard, useQualityTrends } from '@/lib/api/queries'
import { CheckCircleIcon, XCircleIcon, ExclamationTriangleIcon } from '@heroicons/react/24/outline'

export default function QualityDashboard() {
  const [trendLevel, setTrendLevel] = useState<'overall' | 'connector' | 'dataset' | 'column'>('overall')
  const { data: dashboard, isLoading } = useQualityDashboard()
  const { data: trends } = useQualityTrends(trendLevel, 30)

  if (isLoading) {
    return (
      <Card>
        <CardContent className="py-8 text-center">
          <p className="text-muted-foreground">Loading quality dashboard...</p>
        </CardContent>
      </Card>
    )
  }

  const chartData = trends?.map((point: any) => ({
    date: new Date(point.date).toLocaleDateString(),
    score: point.score,
  })) || []

  return (
    <div className="space-y-4">
      {/* Overview Metrics */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <MetricCard
          title="Overall Quality Score"
          value={`${dashboard?.overall_score?.toFixed(1) || 0}%`}
          description="Average quality across all datasets"
          trend={
            dashboard?.overall_score
              ? {
                  value: 5,
                  isPositive: dashboard.overall_score >= 80,
                }
              : undefined
          }
        />
        <MetricCard
          title="Total Rules"
          value={dashboard?.total_rules || 0}
          description={`${dashboard?.active_rules || 0} active`}
        />
        <MetricCard
          title="Passing Rules"
          value={dashboard?.passing_rules || 0}
          description={`${dashboard?.failing_rules || 0} failing`}
          icon={<CheckCircleIcon className="h-5 w-5 text-green-600" />}
        />
        <MetricCard
          title="Rule Violations"
          value={dashboard?.rule_violations?.length || 0}
          description="Requires attention"
          icon={<ExclamationTriangleIcon className="h-5 w-5 text-yellow-600" />}
        />
      </div>

      {/* Quality Trends */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle>Quality Score Trends</CardTitle>
            <div className="flex gap-2">
              {(['overall', 'connector', 'dataset', 'column'] as const).map((level) => (
                <button
                  key={level}
                  onClick={() => setTrendLevel(level)}
                  className={`px-3 py-1 text-sm rounded-lg border ${
                    trendLevel === level
                      ? 'bg-primary text-primary-foreground border-primary'
                      : 'bg-background border-border hover:bg-muted'
                  }`}
                >
                  {level.charAt(0).toUpperCase() + level.slice(1)}
                </button>
              ))}
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <div className="h-[300px]">
            <LineChart
              data={chartData}
              dataKeys={['score']}
              xAxisKey="date"
              colors={['#0ea5e9']}
            />
          </div>
        </CardContent>
      </Card>

      {/* Dataset Quality Scores */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <Card>
          <CardHeader>
            <CardTitle>Lowest Quality Datasets</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              {dashboard?.dataset_scores?.slice(0, 10).map((dataset: any, idx: number) => (
                <div
                  key={idx}
                  className="flex items-center justify-between p-2 border border-border rounded-lg"
                >
                  <div>
                    <div className="font-medium text-sm">
                      {dataset.schema_name}.{dataset.table_name}
                    </div>
                    <div className="text-xs text-muted-foreground">
                      {dataset.rule_count} rules
                    </div>
                  </div>
                  <div className="text-right">
                    <div
                      className={`font-bold ${
                        dataset.score >= 80
                          ? 'text-green-600'
                          : dataset.score >= 60
                          ? 'text-yellow-600'
                          : 'text-red-600'
                      }`}
                    >
                      {dataset.score.toFixed(1)}%
                    </div>
                    <div className="text-xs text-muted-foreground">
                      {new Date(dataset.last_checked).toLocaleDateString()}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Top Rule Violations</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              {dashboard?.rule_violations?.map((violation: any, idx: number) => (
                <div
                  key={idx}
                  className="flex items-center justify-between p-2 border border-border rounded-lg"
                >
                  <div>
                    <div className="font-medium text-sm">{violation.rule_name}</div>
                    <div className="text-xs text-muted-foreground">
                      Severity: {violation.severity}
                    </div>
                  </div>
                  <div className="text-right">
                    <div className="font-bold text-red-600">{violation.violations}</div>
                    <div className="text-xs text-muted-foreground">violations</div>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
