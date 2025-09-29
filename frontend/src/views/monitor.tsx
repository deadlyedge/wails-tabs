type MonitorProps = {
  scanSummary?: {
    filesDiscovered: number
    filesPersisted: number
    filesSkipped: number
    duplicateGroups: number
    durationMs: number
  } | null
  scanProgress?: {
    path?: string
    filesProcessed?: number
    filesPersisted?: number
  } | null
  tidySummary?: {
    moved: number
    failed: number
    skipped: number
    durationMs: number
    total: number
    dryRun: boolean
  } | null
  tidyProgress?: {
    mediaId?: number
    source?: string
    target?: string
    status?: string
    completed?: number
    total?: number
    error?: string
  } | null
  error?: string | null
}

export const Monitor = ({
  scanSummary,
  scanProgress,
  tidySummary,
  tidyProgress,
  error,
}: MonitorProps) => {
  const formatDuration = (ms?: number) => {
    if (!ms) return "-"
    if (ms < 1000) return `${ms}ms`
    return `${(ms / 1000).toFixed(1)}s`
  }

  return (
    <div className='bg-blue-900/70 w-full h-full p-4 rounded flex flex-col gap-4 text-slate-100 text-sm'>
      {error && <div className='text-red-300 text-xs'>{error}</div>}

      <section>
        <h2 className='font-semibold text-base'>Scan</h2>
        {scanSummary ? (
          <ul className='mt-2 space-y-1 text-xs text-slate-200'>
            <li>Processed: {scanSummary.filesDiscovered}</li>
            <li>Persisted: {scanSummary.filesPersisted}</li>
            <li>Skipped: {scanSummary.filesSkipped}</li>
            <li>Duplicates: {scanSummary.duplicateGroups}</li>
            <li>Duration: {formatDuration(scanSummary.durationMs)}</li>
          </ul>
        ) : (
          <p className='text-xs text-slate-400 mt-1'>No scan yet.</p>
        )}
        {scanProgress?.path && (
          <p className='text-[11px] text-slate-300 mt-2 truncate'>
            {scanProgress.filesProcessed ?? 0} files seen · {scanProgress.filesPersisted ?? 0} stored
            <br />
            {scanProgress.path}
          </p>
        )}
      </section>

      <section>
        <h2 className='font-semibold text-base'>Tidy</h2>
        {tidySummary ? (
          <ul className='mt-2 space-y-1 text-xs text-slate-200'>
            <li>
              Moves: {tidySummary.moved} / {tidySummary.total}
            </li>
            <li>Skipped: {tidySummary.skipped}</li>
            <li>Failed: {tidySummary.failed}</li>
            <li>Mode: {tidySummary.dryRun ? "Dry run" : "Live"}</li>
            <li>Duration: {formatDuration(tidySummary.durationMs)}</li>
          </ul>
        ) : (
          <p className='text-xs text-slate-400 mt-1'>No tidy run yet.</p>
        )}
        {tidyProgress?.status && (
          <p className='text-[11px] text-slate-300 mt-2'>
            {tidyProgress.status} {tidyProgress.completed ?? 0}/{tidyProgress.total ?? 0}
          </p>
        )}
        {tidyProgress?.error && (
          <p className='text-[11px] text-amber-200 mt-1'>{tidyProgress.error}</p>
        )}
      </section>
    </div>
  )
}
