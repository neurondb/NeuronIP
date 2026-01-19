'use client'

import { motion } from 'framer-motion'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/Table'
import { CheckCircleIcon, XCircleIcon } from '@heroicons/react/24/outline'
import { slideUp, transition } from '@/lib/animations/variants'

interface ComplianceCheck {
  id: string
  entityType: string
  entityId: string
  status: 'compliant' | 'non-compliant'
  checkedAt: Date
  matches: number
}

interface ComplianceTableProps {
  checks: ComplianceCheck[]
}

export default function ComplianceTable({ checks }: ComplianceTableProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Compliance Checks</CardTitle>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Entity</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Matches</TableHead>
              <TableHead>Checked</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {checks.map((check, index) => (
              <motion.tr
                key={check.id}
                variants={slideUp}
                initial="hidden"
                animate="visible"
                transition={{ ...transition, delay: index * 0.05 }}
              >
                <TableCell>
                  {check.entityType} - {check.entityId}
                </TableCell>
                <TableCell>
                  <div className="flex items-center gap-2">
                    {check.status === 'compliant' ? (
                      <CheckCircleIcon className="h-5 w-5 text-green-600" />
                    ) : (
                      <XCircleIcon className="h-5 w-5 text-red-600" />
                    )}
                    <span className="capitalize">{check.status}</span>
                  </div>
                </TableCell>
                <TableCell>{check.matches}</TableCell>
                <TableCell>{check.checkedAt.toLocaleString()}</TableCell>
              </motion.tr>
            ))}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  )
}
