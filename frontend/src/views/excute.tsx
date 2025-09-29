import { Button } from "@/components/ui/button"

type ExcuteProps = {
  onScan: () => void
  onDryRun: () => void
  onReload: () => void
  scanning: boolean
  tidyRunning: boolean
  hasDuplicates: boolean
}

export function Excute({
  onScan,
  onDryRun,
  onReload,
  scanning,
  tidyRunning,
  hasDuplicates,
}: ExcuteProps) {
  return (
    <div className='flex items-center justify-end gap-2'>
      <Button
        variant='outline'
        onClick={onReload}
        disabled={scanning || tidyRunning}>
        Reload Settings
      </Button>
      <Button onClick={onScan} disabled={scanning || tidyRunning}>
        {scanning ? "Scanning…" : "Scan"}
      </Button>
      <Button
        variant='secondary'
        onClick={onDryRun}
        disabled={!hasDuplicates || scanning || tidyRunning}>
        {tidyRunning ? "Dry run…" : "Dry Run Tidy"}
      </Button>
    </div>
  )
}
